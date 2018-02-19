package models

import (
    "fmt"
	"golang.org/x/net/context"
    //
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
)

func NewPermission(wallet *Wallet, address *Address, action string, state bool) *Permission {
    return &Permission{
        Wallet: wallet.UID,
        Address: address.Addr,
        Action: action,
        State: state,
    }
}

type Permission struct {
    Wallet string `json:"wallet"`
    Address string `json:"address"`
    Action string `json:"action"`
    State bool `json:"state"`
}

func (perm *Permission) Key() string {
    return fmt.Sprintf(
        "%x",
        hash128(
            []byte(
                fmt.Sprintf(
                    "%v %v %v",
                    perm.Wallet,
                    perm.Address,
                    perm.Action,
                ),
            ),
        ),
    )
}

func (perm *Permission) DatastoreKey(ctx ...context.Context) interface{} {
    entityType := CONST_DS_ENTITY_PERMISSION
    key := perm.Key()

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
