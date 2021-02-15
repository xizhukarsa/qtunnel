package main

import (
	"log"
	"os"
	"os/signal"
	"qtunnel/src/tunnel"
	"syscall"
	"time"
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
	go tunnel.NewReverseTunnel(":9091", ":9092", false, "rc4", "abc", 100).StartServer()
	time.Sleep(time.Second)
	tunnel.NewReverseTunnel(":8080", "127.0.0.1:9092", false, "rc4", "abc", 10).StartClient()
	waitSignal()
}
