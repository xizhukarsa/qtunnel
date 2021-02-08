package tunnel

import (
	"io"
	"log"
	"net"
)

type ReverseTunnel struct {
	faddr, baddr, local *net.TCPAddr
	clientMode          bool
	cryptoMethod        string
	secret              []byte
	sessionsCount       int32
	pool                *recycler
}

func NewReverseTunnel(faddr, baddr, local string, clientMode bool, cryptoMethod, secret string, size uint32) *ReverseTunnel {
	a1, err := net.ResolveTCPAddr("tcp", faddr)
	if err != nil {
		log.Fatalln("resolve frontend error:", err)
	}
	a2, err := net.ResolveTCPAddr("tcp", baddr)
	if err != nil {
		log.Fatalln("resolve backend error:", err)
	}
	a3, err := net.ResolveTCPAddr("tcp", local)
	if err != nil && clientMode {
		log.Fatalln("resolve backend error:", err)
	}
	return &ReverseTunnel{
		faddr:         a1,
		baddr:         a2,
		local:         a3,
		clientMode:    clientMode,
		cryptoMethod:  cryptoMethod,
		secret:        []byte(secret),
		sessionsCount: 0,
		pool:          NewRecycler(size),
	}
}

func (t *ReverseTunnel) pipe(dst, src *Conn, c chan int64) {
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

func (t *ReverseTunnel) StartServer() {
	ch1 := make(chan *net.TCPConn, 1)
	ch2 := make(chan *net.TCPConn, 1)

	cnf := func(addr *net.TCPAddr, ch chan *net.TCPConn) {
		ln1, err := net.ListenTCP("tcp", t.faddr)
		if err != nil {
			log.Fatal(err)
		}
		defer ln1.Close()

		conn, err := ln1.AcceptTCP()
		if err != nil {
			log.Fatal("accept:", err)
		}
		ch <- conn
	}

	go cnf(t.baddr, ch1)
	go cnf(t.faddr, ch2)

	cn1 := <-ch1
	cn2 := <-ch2

	cipher := NewCipher(t.cryptoMethod, t.secret)
	fconn := NewConn(cn1, nil, t.pool)
	bconn := NewConn(cn2, cipher, t.pool)

	readChan := make(chan int64)
	writeChan := make(chan int64)
	go t.pipe(bconn, fconn, writeChan)
	go t.pipe(fconn, bconn, readChan)

	var readBytes, writeBytes int64
	readBytes = <-readChan
	writeBytes = <-writeChan
	log.Printf("r:%d w:%d", readBytes, writeBytes)
}

func (t *ReverseTunnel) StartClient() {
	log.Println("start client ")
	defer log.Println("client end")

	conn1, err := net.DialTCP("tcp", nil, t.baddr)
	if err != nil {
		log.Print(err)
		return
	}

	conn2, err := net.DialTCP("tcp", nil, t.local)
	if err != nil {
		log.Print(err)
		return
	}

	cipher := NewCipher(t.cryptoMethod, t.secret)
	readChan := make(chan int64)
	writeChan := make(chan int64)
	var bconn, fconn *Conn
	fconn = NewConn(conn2, nil, t.pool)
	bconn = NewConn(conn1, cipher, t.pool)
	go t.pipe(bconn, fconn, writeChan)
	go t.pipe(fconn, bconn, readChan)
}
