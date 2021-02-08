package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"qtunnel/src/tunnel"
	"syscall"
)

func waitSignal() {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan)
	for sig := range sigChan {
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			log.Printf("terminated by signal %v\n", sig)
			return
		} else {
			log.Printf("received signal: %v, ignore\n", sig)
		}
	}
}

func main() {
	var faddr, baddr, local, cryptoMethod, secret, logTo string
	var clientMode bool
	flag.StringVar(&logTo, "logto", "stdout", "stdout or syslog")
	flag.StringVar(&faddr, "listen", ":8080", "host:port qtunnel listen on")
	flag.StringVar(&baddr, "backend", ":1111", "host:port of the backend")
	flag.StringVar(&cryptoMethod, "crypto", "rc4", "encryption method")
	flag.StringVar(&secret, "secret", "secret", "password used to encrypt the data")
	flag.BoolVar(&clientMode, "clientmode", true, "if running at client mode")
	flag.Parse()

	log.SetOutput(os.Stdout)
	if logTo == "syslog" {
		w, err := syslog.New(syslog.LOG_INFO, "qtunnel")
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(w)
	}

	t := tunnel.NewReverseTunnel(faddr, baddr, local, clientMode, cryptoMethod, secret, 4096)
	log.Println("qtunnel started.")
	if clientMode {
		go t.StartClient()
	} else {
		go t.StartServer()
	}
	waitSignal()
}
