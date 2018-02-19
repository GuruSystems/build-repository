package models

import (
	"fmt"
	"encoding/hex"
	//
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
)

type Balance struct {
    address *Address
    Address string `json:"address"`
    Asset string `json:"asset"`
    Value float64 `json:"value"`
}

func (balance *Balance) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_BALANCE

    key := hex.EncodeToString(
        hash128(
            []byte(
                fmt.Sprintf("%s@%s", balance.Asset, balance.Address),
            ),
        ),
    )

    parent := balance.address.DatastoreKey(ctx...)

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, parent.(*datastoreAE.Key))
	}
	return datastore.NameKey(entityType, key, parent.(*datastore.Key))
}
