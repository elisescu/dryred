package dryred

import (
	"github.com/elisescu/toml"
	"fmt"
	"strings"
)

type clientToml struct {
	Client_name string
	Public_key  string
}

type dbToml struct {
	// Don't change these two variables - they are filled in vya toml DecodeFile
	To   []clientToml
	From []clientToml
}

type clientInfo struct {
	name string
	pubKey string
	front bool
}

type clientsDB struct {
	clients dbToml
}

func (client *clientInfo)Name() string {
	return client.name
}

func (client *clientInfo)PubKey() string {
	return client.pubKey
}

func (client *clientInfo)FrontClient() bool {
	return client.front
}
// TODO: avoid code duplication and use a lambda as search criteria
func (db *clientsDB) FindClientByPubKey(pubKey string) (ClientInfo) {
	for _, client := range db.clients.From {
		if strings.HasPrefix(client.Public_key, pubKey) {
			return &clientInfo{
				name: client.Client_name,
				pubKey: client.Public_key,
				front: true,
			}
		}
	}

	for _, client := range db.clients.To {
		if strings.HasPrefix(client.Public_key, pubKey) {
			return &clientInfo{
				name: client.Client_name,
				pubKey: client.Public_key,
				front: false,
			}
		}
	}

	return nil
}

func (db *clientsDB) FindClientByName(name string) (ClientInfo) {
	for _, client := range db.clients.From {
		if client.Client_name == name {
			return &clientInfo{
				name: client.Client_name,
				pubKey: client.Public_key,
				front: true,
			}
		}
	}

	for _, client := range db.clients.To {
		if client.Client_name == name {
			return &clientInfo{
				name: client.Client_name,
				pubKey: client.Public_key,
				front: false,
			}
		}
	}

	return nil
}

func (db clientsDB) ForwardAllowedFromTo(client1 ClientInfo, client2 ClientInfo) bool {
	return client1.FrontClient() && !client2.FrontClient()
}

func ClientsDBFromToml(file string) (ClientsDB, error) {
	var clients dbToml
	_, err := toml.DecodeFile(file, &clients)
	if err != nil {
		return nil, fmt.Errorf("Can't parse file: %s", err.Error())

	}
	return &clientsDB{
		clients: clients,
	}, nil
}
