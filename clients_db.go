package dryred

type ClientsDB interface {
	// FindClientByName finds the client in the data base by its name and returns a ClientInfo
	// interface, if found, or nil otherwise.
	FindClientByName(string) (ClientInfo)

	// FindClientByPubKey finds the client in the data base by its public key and returns a
	// ClientInfo interface, if found, or nil otherwise.
	// The public key format has to be of the form: "keyTypeString keyData"
	FindClientByPubKey(string) (ClientInfo)

	// ForwardAllowedFromTo returns true if first client is allowed to forward data to the
	// second one.
	ForwardAllowedFromTo(ClientInfo, ClientInfo) bool
}

type ClientInfo interface {
	Name() string
	PubKey() string
	FrontClient() bool
}
