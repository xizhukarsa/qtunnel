package tunnel

import (
	"io"
	"log"
	"net"
	"sync"
)

/**
How to reverse proxy:
	1. client start tunnel to server
	2. server store tunnel for user
	3. when get user data, pick one
*/

type ReverseTunnel struct {
	addr, ternelAddr *net.TCPAddr
	clientMode       bool
	cryptoMethod     string
	secret           []byte
	sessionsCount    int32
	pool             *recycler
	connPoll1        chan net.Conn
	connPoll2        chan net.Conn
	connPoll3        chan net.Conn
}

func NewReverseTunnel(addr, ternelAddress string, clientMode bool, cryptoMethod, secret string, size uint32) *ReverseTunnel {
	a1, err := net.ResolveTCPAddr("tcp", ternelAddress)
	if err != nil {
		log.Fatalln("resolve frontend error:", err)
	}
	a2, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalln("resolve backend error:", err)
	}
	return &ReverseTunnel{
		ternelAddr:    a1,
		addr:          a2,
		clientMode:    clientMode,
		cryptoMethod:  cryptoMethod,
		secret:        []byte(secret),
		sessionsCount: int32(size),
		pool:          NewRecycler(size),
		connPoll1:     make(chan net.Conn, size),
		connPoll2:     make(chan net.Conn, size),
		connPoll3:     make(chan net.Conn, 1024),
	}
}

func (p *ReverseTunnel) Start() {
	if p.clientMode {
		p.startClient()
	} else {
		p.startServer()
	}
}

func (p *ReverseTunnel) Stop() {
	closeAllTunn := func(ch chan net.Conn) {
		log.Println("close all tunn")
		for {
			select {
			case c := <-ch:
				c.Close()
			default:
				return
			}
		}
	}
	closeAllTunn(p.connPoll1)
	closeAllTunn(p.connPoll2)
	closeAllTunn(p.connPoll3)
}

func (p *ReverseTunnel) pipe(dst, src *Conn, c chan int64) {
	defer func() {
		dst.CloseWrite()
		src.CloseRead()
	}()
	n, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("tunnel data err : %v\n", err)
	}
	c <- n
}

func (p *ReverseTunnel) startClient() {
	go func() {
		for {
			conn1, err := net.DialTCP("tcp", nil, p.ternelAddr)
			if nil != err {
				log.Fatal(err)
			}
			p.connPoll1 <- conn1

			conn2, err := net.DialTCP("tcp", nil, p.addr)
			if nil != err {
				log.Fatal(err)
			}
			p.connPoll2 <- conn2

			cipher := NewCipher(p.cryptoMethod, p.secret)

			var bconn, fconn *Conn
			fconn = NewConn(conn1, cipher, p.pool)
			bconn = NewConn(conn2, nil, p.pool)

			readChan := make(chan int64)
			writeChan := make(chan int64)

			go func() {
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					p.pipe(bconn, fconn, writeChan)
					wg.Done()
				}()
				go func() {
					p.pipe(fconn, bconn, readChan)
					wg.Done()
				}()
				wg.Wait()
				<-p.connPoll1
				<-p.connPoll2
			}()
		}
	}()
}

func (p *ReverseTunnel) startServer() {
	// list on client
	go func() {
		ln, err := net.ListenTCP("tcp", p.ternelAddr)
		if nil != err {
			log.Fatal(err)
		}

		for {
			cnn, err := ln.Accept()
			if nil != err {
				log.Fatal(err)
			}
			p.connPoll3 <- cnn
			log.Println("client connected")
		}
	}()

	// list on user
	go func() {
		ln, err := net.ListenTCP("tcp", p.addr)
		if nil != err {
			log.Fatal(err)
		}
		for {
			cnn1, err := ln.Accept()
			if nil != err {
				log.Fatal(err)
			}
			cnn2 := <-p.connPoll3

			log.Println("user connected")

			cipher := NewCipher(p.cryptoMethod, p.secret)

			var bconn, fconn *Conn
			fconn = NewConn(cnn1, nil, p.pool)
			bconn = NewConn(cnn2, cipher, p.pool)

			readChan := make(chan int64)
			writeChan := make(chan int64)

			go p.pipe(bconn, fconn, writeChan)
			go p.pipe(fconn, bconn, readChan)
		}
	}()
}
