package dryred

import (
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

const ( DEFAULT_TIMEOUT_MS = 10000 )

func init() {
	log.SetFlags(log.Lshortfile)
}

func slice_eq(s1, s2 []byte) bool {
	if len(s1) != len(s2) {
		log.Printf("Slices not equal. Size1 %d, Size2 %d", len(s1), len(s2))
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			log.Printf("Slices not equal. Offset %d, bytes: %d != %d", i, s1[i], s2[i])
			return false
		}
	}
	return true
}

func TestFWDataTransfer(t *testing.T) {
	var wg sync.WaitGroup
	rpc_address := "localhost:3000"
	dest_service_address := "localhost:3001"
	wg.Add(2)

	/* Remote RPC side, where the connection will be open */
	go func() {
		defer wg.Done()
		defer log.Printf("Exit the RPC server goroutine")
		rpc_listener, err := net.Listen("tcp", rpc_address)
		defer rpc_listener.Close()
		if err != nil {
			t.Fatalf("Cannot listen RPC address %s", rpc_address)

		}
		rpc_connection, err := rpc_listener.Accept()

		forwarder := FWServerRPCNew(rpc_connection)
		forwarder.ServeAndForward()
	}()

	/* The local service to which we forward the connection via the RPC forwarder */
	go func() {
		defer wg.Done()
		defer log.Printf("Exit the remote service echo-connection")
		listener, err := net.Listen("tcp", dest_service_address)
		defer listener.Close()
		if err != nil {
			t.Fatalf("Cannot bind for the destination fw service %s",
				dest_service_address)
		}
		dest_service_conn, err := listener.Accept()
		log.Printf("Accepted connection on the destination. Echoing back all data")

		io.Copy(dest_service_conn, dest_service_conn)
	}()

	time.Sleep(time.Millisecond * 500)

	rpc_connection, err := net.Dial("tcp", rpc_address)
	if err != nil {
		t.Fatalf("Cannot connect to the RPC address %s, error: %s", rpc_address, err)
	}
	rpc_client := FWClientRPCNew(rpc_connection)

	connection, err := rpc_client.Connect(dest_service_address)

	if err != nil {
		t.Fatalf("Couldn't connect to the destination service, via the RPC forwarder")

	}

	for i := 1; i < 5; i++ {
		size := i * 1024
		input := make([]byte, size)
		output := make([]byte, size)
		rand.Read(input)
		num_written, err := connection.Write(input)
		if err != nil {
			t.Fatalf("Error when writing to the remote connection: %s", err)
		}
		num_read, err := connection.Read(output)
		if err != nil {
			t.Fatalf("Error when reading from the remote connection: %s", err)
		}
		equal := slice_eq(input[:num_written], output[:num_read])
		log.Printf("Data read from the remote connection: expected %d bytes, got %d bytes. "+
			"Equal: %t", num_written, num_read, equal)

		if !equal {
			t.Fatalf("What read not equal to what written")
		}
	}
	connection.Close()
	rpc_connection.Close()
	/* Wait for the two goroutines to end*/
	wg.Wait()
}
