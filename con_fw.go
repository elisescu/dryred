package dryred

import (
	"log"
	"net"
	"net/rpc"
	"time"
	"errors"
	"io"
)

// FWServer describes the server listing and serving FW connections
type FWServer interface {
	// Serve and Forward connections
	ServeAndForward() error
	// Stop the server
	Stop() error
}

// FWClient describes the client  connecting to a remote forwarded address via the FW server
type FWClient interface {
	// Connect to the remote address, accessible via the server
	Connect(address string) (net.Conn, error)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
type rpcFWServer struct {
	connection io.ReadWriteCloser
	rpcServer  *rpc.Server
}

// FWServerRPCNew creates new FWServer instance
func FWServerRPCNew(connection io.ReadWriteCloser) (server FWServer) {
	server = &rpcFWServer{
		rpcServer: rpc.NewServer(),
		connection: connection,
	}
	return
}

// Start the server and serve
func (server *rpcFWServer) ServeAndForward() error {
	server.rpcServer.Register(NETRPCNetConnNew())
	server.rpcServer.ServeConn(server.connection)
	return nil
}

// Stop the server
func (server *rpcFWServer) Stop() error {
	return server.connection.Close()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Client that connects to the server and can create remote fw connections on the server side
type rpcFWClient struct {
	connection io.ReadWriteCloser
	rpcClient  *rpc.Client
}

type rpcNetConn struct {
	ConnID     int64
	client     *rpcFWClient
	remoteAddr net.Addr
	localAddr  net.Addr
}

// FWClientRPCNew creates new FWClient instance
func FWClientRPCNew(connection io.ReadWriteCloser) (client FWClient) {
	return &rpcFWClient{
		connection: connection,
		rpcClient: rpc.NewClient(connection),
	}
}

// Connect to an address on the FWServer side and return a connection to that address
func (client *rpcFWClient) Connect(address string) (connection net.Conn, err error) {
	var ret NETRPCConnectRet
	args := NETRPCConnectArgs{Address: address}
	err = client.rpcClient.Call("NETRPCNetConn.Connect", args, &ret)
	if err != nil {
		log.Print("Returned from the connect call with err ", err)
		return nil, err
	}

	rpcConn := &rpcNetConn{
		ConnID: ret.ConnID,
		client: client,
	};
	return  rpcConn, nil
}

// Write data to the forwarding connection
func (connection *rpcNetConn) Write(data []byte) (numWritten int, err error) {
	args := NETRPCWriteArgs{
		ConnID: connection.ConnID,
		Data:   data,
	}
	var ret NETRPCWriteRet

	if len(data) == 0 {
		return
	}

	err = connection.client.rpcClient.Call("NETRPCNetConn.Write", args, &ret)
	//TODO: check properly for errors
	numWritten = ret.NumWritten
	return
}

// Read data from the forwarding connection
func (connection *rpcNetConn) Read(data []byte) (numRead int, err error) {
	args := NETRPCReadArgs{
		MaxRead: len(data),
		ConnID:  connection.ConnID,
	}
	var ret NETRPCReadRet

	if len(data) == 0 {
		return
	}

	//TODO: check properly for errors
	ret.Data = data
	err = connection.client.rpcClient.Call("NETRPCNetConn.Read", args, &ret)
	numRead = ret.NumBytesRead
	return
}

// Close the forwarding connection
func (connection *rpcNetConn) Close() (err error) {
	args := NETRPCCloseArgs{
		ConnID: connection.ConnID,
	}
	var ret NETRPCCloseRet
	err = connection.client.rpcClient.Call("NETRPCNetConn.Close", args, &ret)
	return
}


// Return the remote address of the connection
func (connection *rpcNetConn) RemoteAddr() net.Addr {
	// TODO: implement this
	return nil
}

// Set deadline. F**k you golint
func (connection *rpcNetConn) SetDeadline(t time.Time) error {
	// TODO: implement this
	return nil
}

// Set read deadline.
func (connection *rpcNetConn) SetReadDeadline(t time.Time) error {
	// TODO: implement this
	return errors.New("not implemented")
}

// Set write deadline.
func (connection *rpcNetConn) SetWriteDeadline(t time.Time) error {
	// TODO: implement this
	return errors.New("not implemented")
}

// Get local address
func (connection *rpcNetConn) LocalAddr() net.Addr {
	// TODO: implement this
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// NETRPCNetConn used by the rpc package
type NETRPCNetConn struct {
	connections map[int64]net.Conn
	readBuffer  []byte
	currConnID  int64
}

// NETRPCNetConnNew creates new connection
func NETRPCNetConnNew() (rpcNetConn *NETRPCNetConn) {
	rpcNetConn = &NETRPCNetConn{
		connections:  make(map[int64]net.Conn),
		currConnID: 0,
		readBuffer:  make([]byte, 8*1024),
	}
	return
}

// NETRPCReadArgs used by the rpc package
type NETRPCReadArgs struct {
	ConnID  int64
	MaxRead int
}

// NETRPCReadRet used by rpc package
type NETRPCReadRet struct {
	Data         []byte
	NumBytesRead int
	Error        error
}

// NETRPCWriteArgs used by the rpc package
type NETRPCWriteArgs struct {
	ConnID int64
	Data   []byte
}

// NETRPCWriteRet used by the rpc package
type NETRPCWriteRet struct {
	NumWritten int
	Error      error
}

// NETRPCConnectArgs used by the rpc package
type NETRPCConnectArgs struct {
	Address string
}

// NETRPCConnectRet used by the rpc package
type NETRPCConnectRet struct {
	ConnID int64
}

// NETRPCCloseArgs used by the rpc package
type NETRPCCloseArgs struct {
	ConnID int64
}

// NETRPCCloseRet used by the rpc package
type NETRPCCloseRet bool

// Connect is used to connect to the remote side via RPC
func (server *NETRPCNetConn) Connect(args NETRPCConnectArgs, ret *NETRPCConnectRet) (err error) {
	currID := server.currConnID
	server.connections[currID], err = net.Dial("tcp", args.Address)
	ret.ConnID = currID
	server.currConnID++
	return
}

func (server *NETRPCNetConn) Read(args NETRPCReadArgs, ret *NETRPCReadRet) (err error) {
	toRead := args.MaxRead
	if toRead > len(server.readBuffer) {
		toRead = len(server.readBuffer)
	}
	ret.Data = server.readBuffer[:toRead]
	ret.NumBytesRead, err = server.connections[args.ConnID].Read(ret.Data)
	/* Re-slice to send back only the right data */
	ret.Data = ret.Data[:ret.NumBytesRead]
	return
}

func (server *NETRPCNetConn) Write(args NETRPCWriteArgs, ret *NETRPCWriteRet) (err error) {
	for written, remain := 0, len(args.Data); remain > 0 && err == nil; {
		written, err = server.connections[args.ConnID].Write(args.Data)
		remain -= written
	}
	ret.NumWritten = len(args.Data)
	return
}

// Close Used to close  the remote connection
func (server *NETRPCNetConn) Close(args NETRPCCloseArgs, ret *NETRPCCloseRet) (err error) {
	err = server.connections[args.ConnID].Close()
	return
}