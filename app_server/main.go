package main

import (
	"github.com/elisescu/dryred"
	"os"
	"os/signal"
	"flag"
	"log"
	"sync"
	"golang.org/x/crypto/ssh"
)

func install_signal(fn func ()) {
	signal_channel := make (chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	signal.Notify(signal_channel, os.Kill)

	go func() {
		<- signal_channel
		fn()
	}()
}

func main() {
	var wg sync.WaitGroup
	server_port := flag.Int("port", 1234, "Port to listen to")
	server := dryred.DRServerNew(*server_port)

	wg.Add(1)
	install_signal(func() {
		log.Printf("Caught Ctrl+C.")
		server.Stop()
		//wg.Done()
	})

	log.Printf("Listening on port %d", *server_port);

	// TODO: find a way to stop the server. Right now, it stops because main exits.
	server.Run()
	//wg.Wait()
}
