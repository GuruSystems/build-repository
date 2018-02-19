package models

import (
	"errors"
	//
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
    //
    "github.com/golangdaddy/go.uuid"
)

func NewNetwork(customer *Customer, networkConfig *NetworkConfig, roleName string) (*Network, error) {

	var role int

	switch roleName {
		case "user":
		case "editor": role = 1
		case "owner": role = 2
		default:
			return nil, errors.New("(NewNetwork) NO NETWORK ROLE DEFINED FOR: "+roleName)
	}

	salt, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

    return &Network{
        UID: salt.String(),
        Customer: customer.UID,
		Chain: networkConfig.UID,
		Role: role,
    }, nil
}

type Network struct {
    UID string `json:"uid"`
    Customer string `json:"-"`
	Chain string `json:"chain"`
    Role int `json:"role"`
    Public bool `json:"-"`
    params string
}

func (network *Network) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_NETWORK
    key := DeterministicUID(network)

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
