package dryred

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/elisescu/speakeasy"
	"golang.org/x/crypto/ssh"
	"net"
	"fmt"
	"time"
	"encoding/base64"
)

type CloseCallback func(net.Conn) error

type ConnWrapperNoCloser struct {
	innerConnection net.Conn
	closeCallback   CloseCallback
}

func ConnWrapperNoCloserNew(conn net.Conn, callback CloseCallback) *ConnWrapperNoCloser {
	return &ConnWrapperNoCloser{
		innerConnection: conn,
		closeCallback: callback,
	}
}

func (conn ConnWrapperNoCloser)Read(b []byte) (n int, err error) {
	return conn.innerConnection.Read(b)
}

func (conn ConnWrapperNoCloser)Write(b []byte) (n int, err error) {
	return conn.innerConnection.Write(b)
}

func (conn ConnWrapperNoCloser)Close() error {
	return conn.closeCallback(conn.innerConnection)
}

func (conn ConnWrapperNoCloser)LocalAddr() net.Addr {
	return conn.innerConnection.LocalAddr()
}

func (conn ConnWrapperNoCloser)RemoteAddr() net.Addr {
	return conn.innerConnection.RemoteAddr()
}

func (conn ConnWrapperNoCloser)SetDeadline(t time.Time) error {
	return nil // conn.innerConnection.SetDeadline(t)
}

func (conn ConnWrapperNoCloser)SetReadDeadline(t time.Time) error {
	return nil //conn.innerConnection.SetReadDeadline(t)
}

func (conn ConnWrapperNoCloser)SetWriteDeadline(t time.Time) error {
	return nil //conn.innerConnection.SetWriteDeadline(t)
}

func decryptSSHKey(key []byte, password string) ([]byte, error) {
	block, rest := pem.Decode(key)
	if len(rest) > 0 {
		return nil, fmt.Errorf("Can't decode pem key")
	}

	if x509.IsEncryptedPEMBlock(block) {
		if password == "" {
			pass, err := speakeasy.Ask("Enter ssh key passphrase:")
			if err != nil {
				return nil, fmt.Errorf("Can't read password from stdin: %s",
					err.Error())
			}
			password = pass
		}
		der, err := x509.DecryptPEMBlock(block, []byte(password))
		if err != nil {
			return nil, fmt.Errorf("Can't descrypt pem key: %s", err.Error())
		}
		return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}), nil
	}
	return key, nil
}

func encodeSSHPubKey(pubKey ssh.PublicKey) string {
	return pubKey.Type() + " " + base64.StdEncoding.EncodeToString(pubKey.Marshal())
}

func discardSSHRequests(in <-chan *ssh.Request) {
	for req := range in {
		if req.WantReply {
			req.Reply(false, nil)
		}
	}
}

func discardSSHChans(in <-chan ssh.NewChannel) {
	for range in {
	}
}
