package dryred

import (
	"log"
	"time"
	"testing"
	"net"
)

func TestDRServer(t *testing.T) {
	const serverAddress string = "localhost:7000"
	const destAddress string = "localhost:7001"

	var backClient, frontClient DRClient
	var frontConn, backConn net.Conn

	clientsDB , err := ClientsDBFromToml("testdata/clients.toml")
	if err != nil {
		t.Fatal("Cant read the clients.toml file: ", err)
	}
	serverConfig := DRServerSSHConfig{
		SSHServerKeyPassphrase: "TestTest",
		SSHServerKeyFileName: "testdata/server_id_rsa",
		ClientsDB: clientsDB,
	}
	server , err := DRServerSSHNew(serverConfig)

	if err != nil {
		t.Fatal("Can't create server: ", err)
	}
	log.Printf("Starting server on :7000")
	go server.ListenAndServe(":7000")

	frontClientConfig := DRClientSSHConfig{
		SSHKeyFileName:"testdata/front_id_rsa",
		SSHKeyPassPhrase:"TestTest",
		User:"elisescu",
	}
	backClientConfig := DRClientSSHConfig{
		SSHKeyFileName:"testdata/back_id_rsa",
		SSHKeyPassPhrase:"TestTest",
		User:"pi",
	}


	if frontClient, err = DRClientSSHNew(frontClientConfig); err != nil {
		t.Fatal("Can't create front client: ", err)
	}

	if backClient, err = DRClientSSHNew(backClientConfig); err != nil {
		t.Fatal("Can't create back client: ", err)
	}


	if backConn, err = backClient.ListenViaServer(serverAddress); err != nil {
		t.Fatalf("Can't listen via the server: %s", err.Error())
	}

	if frontConn, err = frontClient.Connect(serverAddress, backClientConfig.User, destAddress); err != nil {
		t.Fatalf("Can't connect to the remote back address: %s", err.Error())
	}

	time.Sleep(200 * time.Second)

	backConn.Close()
	frontConn.Close()
	server.Stop()

	time.Sleep(2 * time.Second)
}
