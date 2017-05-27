package dryred

import (
	"net"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"fmt"
	"log"
	"encoding/json"
)

type DRClient interface {
	Connect(serverAddress string, backClientName string, forwardAddress string) (net.Conn, error)
	ListenViaServer(string) (net.Conn, error)
}

type DRClientSSHConfig struct {
	SSHKeyFileName string
	SSHKeyPassPhrase string
	User string
}

type sshDRClient struct {
	sshConfig ssh.ClientConfig
	config DRClientSSHConfig
}

func DRClientSSHNew(config DRClientSSHConfig) (DRClient, error) {
	privateBytes, err := ioutil.ReadFile(config.SSHKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot read SSH private key: %s", err.Error())
	}

	sshPrivateKey, err := ssh.ParsePrivateKeyWithPassphrase(privateBytes,
		[]byte(config.SSHKeyPassPhrase))

	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %s", err.Error())
	}

	sshConfig := ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(sshPrivateKey)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// TODO: ask the user here, or check known hosts
			log.Printf("Automatically accepting server key %s",
				ssh.FingerprintSHA256(key))
			return nil
		},
	}
	return &sshDRClient{
		sshConfig: sshConfig,
		config: config,
	}, nil
}

func buildSSHConnectionToServer(client *sshDRClient, serverAddress string) (ssh.Conn, net.Conn, error) {
	conn, err := net.Dial("tcp", serverAddress)
	should_close := false

	if err != nil {
		return nil, nil, fmt.Errorf("Cannot connect to %s: %s", serverAddress, err.Error())
	}

	connWrapper := ConnWrapperNoCloserNew(conn, func(closingConnection net.Conn) error {
		if should_close {
			return conn.Close()
		}
		return nil
	})

	sshConn, sshChans, sshReqs, err := ssh.NewClientConn(connWrapper, serverAddress,
		&client.sshConfig)

	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("Cannot perform SSH connection %s", err.Error())
	}

	// Discard the requests and channels. Server doesn't open channels or sends request to us
	go discardSSHChans(sshChans)
	go discardSSHRequests(sshReqs)
	return sshConn, conn, nil
}

func (client *sshDRClient)ListenViaServer(serverAddress string) (net.Conn, error) {
	sshConn, tcpConn, err := buildSSHConnectionToServer(client, serverAddress)

	if err != nil {
		return nil, fmt.Errorf("Cannot create SSH secure connection to %s : %s",
			serverAddress, err.Error())
	}

	listenRequest, err := json.Marshal(FWListenRequest_V1{
		Name: client.sshConfig.User,
	})

	if err != nil {
		panic(err)
	}

	_, data, err := sshConn.SendRequest(FWListenRequestName, true, listenRequest)

	if err != nil {
		return nil, fmt.Errorf("Cannot send listen request: ", err.Error())
	}

	var listenReply FWListenReply_V1
	err = json.Unmarshal(data, &listenReply)

	if err != nil {
		return nil, fmt.Errorf("Cannot parse listen reply: ", err.Error())
	}

	if !listenReply.Status {
		return nil, fmt.Errorf("Cannot listen connection. Reply: %s", string(data))
	}

	sshConn.Close()
	return tcpConn, nil
}

func (client *sshDRClient)Connect(serverAddress, backClient, backAddress string) (net.Conn, error) {
	sshConn, tcpConn, err := buildSSHConnectionToServer(client, serverAddress)

	if err != nil {
		return nil, fmt.Errorf("Cannot create SSH secure connection to %s : %s",
			serverAddress, err.Error())
	}

	fwRequest, err := json.Marshal(FWConnectRequest_V1{
		BackClientName: backClient,
		BackConnectionAddress: backAddress,
	})

	if err != nil {
		panic(err)
	}

	_, data, err := sshConn.SendRequest(FWConnectRequestName, true, fwRequest)

	if err != nil {
		return nil, fmt.Errorf("Cannot send request for %s, to %s: ", backClient,
			backAddress, err.Error())
	}

	var fwReply FWConnectReply_V1
	err = json.Unmarshal(data, &fwReply)

	if err != nil {
		return nil, fmt.Errorf("Cannot parse fw reply for %s, to %s: ", backClient,
			backAddress, err.Error())
	}

	if !fwReply.Status {
		return nil, fmt.Errorf("Cannot open a fw connection. Reply: %s",
			string(data))
	}

	sshConn.Close()
	return tcpConn, nil
}