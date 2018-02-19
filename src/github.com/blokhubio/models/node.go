package models

import (
    "strings"
    "strconv"
    //
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
    //
    "github.com/golangdaddy/tarantula/web/validation"
    //
    "github.com/golangdaddy/go.uuid"
)

func NewNode(networkConfig *NetworkConfig, ipAddress string) (*Node, error) {

    salt, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

    ipv4 := validation.IPv4{}
    for x, s := range strings.Split(ipAddress, ".") {
        i, err := strconv.Atoi(
            strings.TrimSpace(s),
        )
        if err != nil {
            return nil, err
        }
        ipv4[x] = i
    }

    return &Node{
        UID: salt.String(),
        Network: networkConfig.UID,
        IPv4: ipv4.String(),
    }, nil
}

type Node struct {
    UID string `json:"uid"`
    Network string `json:"network"`
    IPv4 string `json:"ipv4"`
}

func (node *Node) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_NODE
    key := node.UID

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
