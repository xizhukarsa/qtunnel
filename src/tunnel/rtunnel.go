package tunnel

import (
	"io"
	"log"
	"net"
)

type ReverseTunnel struct {
	faddr, baddr  *net.TCPAddr
	clientMode    bool
	cryptoMethod  string
	secret        []byte
	sessionsCount int32
	pool          *recycler
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
	return &ReverseTunnel{
		faddr:         a1,
		baddr:         a2,
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
}

func (t *ReverseTunnel) StartClient() {
}
