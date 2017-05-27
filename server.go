package dryred

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"
)

// DRServer interface describing a forwarding server
type DRServer interface {
	ListenAndServe(address string) error
	Stop() error
}

type sshDRServer struct {
	listener  net.Listener
	clientsDB ClientsDB
	sshConfig ssh.ServerConfig
}

// DRServerSSHConfig describes the configuration for a DRServer based on SSH auth/encrypt
type DRServerSSHConfig struct {
	ClientsDB              ClientsDB
	SSHServerKeyFileName   string
	SSHServerKeyPassphrase string
}

func DRServerSSHNew(config DRServerSSHConfig) (DRServer, error) {
	server := &sshDRServer{
		clientsDB: config.ClientsDB,
	}

	server.sshConfig = ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			ok, front := authorizeKey(server, pubKey)
			if ok {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp":    ssh.FingerprintSHA256(pubKey),
						"pubkey":       string(pubKey.Marshal()),
						"front-client": strconv.FormatBool(front),
					},
				}, nil
			}
			return nil, fmt.Errorf("Unknown public key for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile(config.SSHServerKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot read SSH server key: %s", err.Error())
	}

	sshPrivateKey, err := ssh.ParsePrivateKeyWithPassphrase(privateBytes,
		[]byte(config.SSHServerKeyPassphrase))

	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %s", err.Error())
	}

	server.sshConfig.AddHostKey(sshPrivateKey)
	return server, nil
}

func (server *sshDRServer) ListenAndServe(address string) (err error) {
	server.listener, err = net.Listen("tcp", address)

	if err != nil {
		return fmt.Errorf("Cannot create ssh DRServer: %s" + err.Error())
	}

	for {
		conn, err := server.listener.Accept()

		if err != nil {
			// The server was shut down
			return err
		}

		go handleConn(server.sshConfig, conn)
	}
	return nil
}

func (server *sshDRServer) Stop() error {
	return server.listener.Close()
}

func authorizeKey(server *sshDRServer, pubKey ssh.PublicKey) (authorized bool, frontClient bool) {
	pubKeyString := encodeSSHPubKey(pubKey)
	client := server.clientsDB.FindClientByPubKey(pubKeyString)

	if client != nil {
		log.Printf("Authorized client with key: %s", ssh.FingerprintSHA256(pubKey))
		return true, client.FrontClient()
	}

	return false, false
}

func handleConn(sshConfig ssh.ServerConfig, conn net.Conn) {
	// Proceed with the SSH handshake and authenticate the remote
	connWrapper := ConnWrapperNoCloserNew(conn, func(closingConnection net.Conn) error {
		return nil
	})

	sshConn, sshChan, sshReq, err := ssh.NewServerConn(connWrapper, &sshConfig)

	if err != nil {
		log.Printf("SSH handshake error with client %s", conn.RemoteAddr().String())
		return
	}
	go discardSSHChans(sshChan)

	select {
	case newRequest := <-sshReq:
		switch newRequest.Type {
		case FWListenRequestName:
			replyBuilder := func(allowed bool, errorMsg string) []byte {
				reply, err := json.Marshal(FWListenReply_V1{
					Status:       allowed,
					ErrorMessage: errorMsg,
				})
				if err != nil {
					panic(err)
				}
				return reply
			}
			var listenRequest FWListenRequest_V1
			err = json.Unmarshal(newRequest.Payload, &listenRequest)
			log.Printf("Got Listen request to from %s", listenRequest.Name)

			allowed := sshConn.Permissions.Extensions["front-client"] != "true"

			if allowed {
				newRequest.Reply(true, replyBuilder(true, ""))
				// Close the SSH layer on top of the connection. Not needed anymore
				sshConn.Close()

				handleBackConnection(conn)
				return
			}

			newRequest.Reply(true, replyBuilder(false, "Not a back client"))
			sshConn.Close()
			conn.Close()
		case FWConnectRequestName:
			replyBuilder := func(allowed bool, errorMsg string) []byte {
				reply, err := json.Marshal(FWConnectReply_V1{
					Status:       allowed,
					ErrorMessage: errorMsg,
				})
				if err != nil {
					panic(err)
				}
				return reply
			}

			var fwRequest FWConnectRequest_V1
			err = json.Unmarshal(newRequest.Payload, &fwRequest)
			log.Printf("Got Connect request to %s, addr: %s, %s",
				fwRequest.BackClientName, fwRequest.BackConnectionAddress,
				string(newRequest.Payload))

			allowed := sshConn.Permissions.Extensions["front-client"] == "true"

			if allowed {
				newRequest.Reply(true, replyBuilder(true, ""))
				// Close the SSH layer on top of the connection. Not needed anymore
				sshConn.Close()

				handleFrontConnection(conn)
				return
			}

			newRequest.Reply(true, replyBuilder(false, "Not a front client"))
			sshConn.Close()
			conn.Close()
		default:
			newRequest.Reply(false, []byte("Unknown request: "+newRequest.Type))
			sshConn.Close()
			conn.Close()
		}
	case <-time.After(time.Second * 5):
		sshConn.Close()
		conn.Close()
		break
	}
}

func handleBackConnection(backConn net.Conn) {
	// The connection is authenticated at this point. Keep it open and keep track of it
}

func handleFrontConnection(frontConn net.Conn) {
	// Find the corresponding back connection and hook them up?

	// TODO: get here the back conn and pass it to the function bellow
	go forwardConnections(frontConn, frontConn)
}

func forwardConnections(conn1, conn2 net.Conn) {
	go io.Copy(conn1, conn2)
	//go io.Copy(conn2, conn1)
}
