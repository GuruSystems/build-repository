package models

import (
    "golang.org/x/net/context"
    datastore "cloud.google.com/go/datastore"
    datastoreAE "google.golang.org/appengine/datastore"
    //
    "github.com/golangdaddy/go.uuid"
)

func NewWallet(user *UserAccount) (*Wallet, error) {

    uid, err := uuid.NewV4()
    if err != nil {
        return nil, err
    }

    salt, err := uuid.NewV4()
    if err != nil {
        return nil, err
    }

    return &Wallet{
        UID: uid.String(),
        User: user.UID,
        Salt: salt.String(),
        Created: timeNow(),
    }, nil
}

type Wallet struct {
    UID string `json:"uid"`
    User string `json:"owner"`
    Salt string `json:"-"`
    DefaultAddress string `json:"defaultAddress"`
    Created int64 `json:"created"`
}

func (wallet *Wallet) DatastoreKey(ctx ...context.Context) interface{} {
	entityType := CONST_DS_ENTITY_WALLET
	key := wallet.UID
	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
