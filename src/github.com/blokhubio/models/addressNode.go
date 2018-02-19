package models

import (
    "fmt"
    //
    "golang.org/x/net/context"
    datastore "cloud.google.com/go/datastore"
    datastoreAE "google.golang.org/appengine/datastore"
)

func NewAddressNode(address *Address, node *Node) *AddressNode {
    return &AddressNode{
        Address: address.Addr,
        Node: node.UID,
    }
}

type AddressNode struct {
    Address string
    Node string
}

func (addrNode *AddressNode) UID() string {
    return fmt.Sprintf(
        "%x",
        hash128(
            []byte(
                addrNode.Address + " " + addrNode.Node,
            ),
        ),
    )
}

func (addrNode *AddressNode) DatastoreKey(ctx ...context.Context) interface{} {
    entityType := CONST_DS_ENTITY_ADDRESSNODE
    key := addrNode.UID()
	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
