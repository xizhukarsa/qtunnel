package tunnel

import (
	"io"
	"log"
	"net"
	"time"
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
	}
}

func (p *ReverseTunnel) Start() {
	if p.clientMode {
		p.startClient()
	} else {
		p.startServer()
	}
}

func (p *ReverseTunnel) pipe(dst, src *Conn, c chan int64) {
	defer func() {
		dst.CloseWrite()
		src.CloseRead()
	}()
	n, err := io.Copy(dst, src)
	if err != nil {
		log.Print(err)
	}
	c <- n
}

func (p *ReverseTunnel) startClient() {
	connPool := make(chan net.Conn, 100)
	go func() {
		for {
			conn1, err := net.DialTCP("tcp", nil, p.ternelAddr)
			if nil != err {
				log.Fatal(err)
			}
			connPool <- conn1

			conn2, err := net.DialTCP("tcp", nil, p.addr)
			if nil != err {
				log.Fatal(err)
			}

			cipher := NewCipher(p.cryptoMethod, p.secret)

			var bconn, fconn *Conn
			fconn = NewConn(conn1, cipher, p.pool)
			bconn = NewConn(conn2, nil, p.pool)

			readChan := make(chan int64)
			writeChan := make(chan int64)

			go p.pipe(bconn, fconn, writeChan)
			go p.pipe(fconn, bconn, readChan)

			time.Sleep(time.Second)
		}
	}()
}

func (p *ReverseTunnel) startServer() {
	connPool := make(chan net.Conn, 100)
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
			connPool <- cnn
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
			cnn2 := <-connPool

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
