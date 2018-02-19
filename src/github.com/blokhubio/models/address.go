package models

import (
    "golang.org/x/net/context"
    datastore "cloud.google.com/go/datastore"
    datastoreAE "google.golang.org/appengine/datastore"
)

type Address struct {
    user *UserAccount
    Wallet string `json:"wallet"`
    Salt string `json:"-"`
    Addr string `json:"addr"`
    Created int64 `json:"created"`
}

func (address *Address) DatastoreKey(ctx ...context.Context) interface{} {
    entityType := CONST_DS_ENTITY_ADDRESS
    key := address.Salt
	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}

func (address *Address) Balance(assetName string) *Balance {
    return &Balance{
        address: address,
        Address: address.Addr,
        Asset: assetName,
    }
}
