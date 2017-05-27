package main

import (
	"github.com/elisescu/dryred"
	"os"
	"os/signal"
	"flag"
	"log"
	"sync"
	"time"
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
	client := dryred.DRBackClientNew()
	// TODO: Fix this way of exiting the main loop
	should_stop := false
	var wg sync.WaitGroup
	wg.Add(1)

	install_signal(func () {
		log.Printf("Caught Ctrl+C.")
		should_stop = true
		client.Close()
		wg.Done()
	})

	server_hostname := flag.String("server", "localhost", "Server hostname to connect to")
	server_port := flag.Int("port", 1234, "Server port to connect to")
	flag.Parse()

	log.Printf("Connecting to %s:%d", *server_hostname, *server_port);
	go func() {

		for !should_stop {
			log.Printf("Launching client..")
			client.Launch(*server_hostname, *server_port)
			time.Sleep(time.Second * 1)
		}
	} ()
	wg.Wait()
}