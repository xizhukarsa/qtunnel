package tunnel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"time"
)

type MsgType int

const (
	MsgTypeNone MsgType = iota
	MsgTypeData
	MsgTypeHeartBeat
	MsgTypeQuit
)

const (
	split = "\n"
)

type Msg struct {
	T    MsgType
	Data string
}

type ReverseTunnel struct {
	addr, ternelAddr *net.TCPAddr
	clientMode       bool
	cryptoMethod     string
	secret           []byte
	sessionsCount    int32
	tconn            *net.TCPConn
	reader           chan Msg
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
		reader:        make(chan Msg, 1024),
	}
}

func (t *ReverseTunnel) StartClient() {
	conn2, err := net.DialTCP("tcp", nil, t.ternelAddr)
	if nil != err {
		log.Fatal(err)
	}
	t.tconn = conn2
	// t.startReceiveData()

	go func() {
		for {
			t.sendMsg(Msg{
				T:    MsgTypeHeartBeat,
				Data: "hello world !",
			})
			// rand.Seed(time.Now().UnixNano())
			// time.Sleep(time.Second * time.Duration(rand.Int63n(10)))
		}
	}()

}

func (t *ReverseTunnel) StartServer() {
	ln1, err := net.ListenTCP("tcp", t.ternelAddr)
	if nil != err {
		log.Fatal(err)
	}

	conn, err := ln1.AcceptTCP()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("tunnel client connected")
	t.tconn = conn
	t.startReceiveData()

	go func() {
		for msg := range t.reader {
			if msg.T == MsgTypeHeartBeat {
				log.Println("heart beat")
			}
		}
	}()

	// ln2, err := net.ListenTCP("tcp", t.addr)
	// if nil != err {
	// 	log.Fatal(err)
	// }

	// go func() {
	// 	clientConn, err := ln2.AcceptTCP()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	log.Printf("receive user req %v\n", clientConn)
	// }()
}

func (t *ReverseTunnel) startReceiveData() {
	go func() {
		splitBuf := []byte(split)
		l := len(splitBuf)
		data := []byte{}
		for {
			err := t.tconn.SetReadDeadline(time.Now().Add(5 * time.Second))
			if nil != err {
				log.Fatal(err)
			}
			recvBuf := make([]byte, 1024)
			n, err := t.tconn.Read(recvBuf[:]) // recv data
			if nil != err {
				log.Fatal(err)
			}
			log.Printf("receive data %v\n", n)
			if n <= 0 {
				continue
			}

			data = append(data, recvBuf[:n]...)

			i := bytes.Index(data, splitBuf)
			if i < 0 {
				continue
			}
			j := bytes.Index(data[i+l:], splitBuf)
			if j < 0 {
				continue
			}

			tmpBuf := data[i+l : i+j+l]
			dataBuf, err := base64.StdEncoding.DecodeString(string(tmpBuf))
			if nil != err {
				log.Fatal(err)
			}

			var msg Msg
			err = json.Unmarshal(dataBuf, &msg)
			if nil != err {
				log.Fatal(err)
			}
			t.reader <- msg

			if msg.T == MsgTypeQuit {
				break
			}

			data = data[j+2*l:]
		}
	}()
}

func (t *ReverseTunnel) sendMsg(msg Msg) {
	log.Println("sned msg")
	_, err := t.tconn.Write([]byte(split))
	if nil != err {
		log.Fatal(err)
	}
	buf, err := json.Marshal(msg)
	if nil != err {
		log.Fatal(err)
	}
	_, err = t.tconn.Write([]byte(base64.StdEncoding.EncodeToString(buf)))
	if nil != err {
		log.Fatal(err)
	}
	_, err = t.tconn.Write([]byte(split))
	if nil != err {
		log.Fatal(err)
	}
}

func (t *ReverseTunnel) readMsg() *Msg {
	msg := <-t.reader
	return &msg
}
