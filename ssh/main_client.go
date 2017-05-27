package main

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/elisescu/speakeasy"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
)

func decrypt(key []byte) []byte {
	block, rest := pem.Decode(key)
	if len(rest) > 0 {
		log.Fatalf("Extra data included in key")
	}

	if x509.IsEncryptedPEMBlock(block) {
		password, err := speakeasy.Ask("Enter ssh key passphrase:")
		if err != nil {
			panic(err)
		}
		der, err := x509.DecryptPEMBlock(block, []byte(password))
		if err != nil {
			log.Fatalf("Decrypt failed: %v", err)
		}
		return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	}
	return key
}

func main() {
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key_data, err := ioutil.ReadFile("/Users/elisescu/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(decrypt(key_data))
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "elisescu",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			log.Printf("Do you accept server pub key %s (%s)", ssh.FingerprintSHA256(key), ssh.FingerprintLegacyMD5(key))
			return nil
		},
	}

	// Connect to the remote server and perform the SSH
	// handshake.
	client, err := ssh.Dial("tcp", "localhost:2022", config)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}

	// OpenChannel(name string, data []byte) (Channel, <-chan *Request, error)
	channel, _, err := client.OpenChannel("test", []byte("test"))
	if err != nil {
		panic(err)
	}

	channel.Write([]byte("test"))
	buf := make([]byte, 1025)
	len, _ := channel.Read(buf)
	log.Printf("got from server-> %s", string(buf[:len]))
	defer client.Close()
}
