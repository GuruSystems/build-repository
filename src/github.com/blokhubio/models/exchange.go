package models

import (
    "time"
    //
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
    //
    "github.com/golangdaddy/go.uuid"
)

func NewExchange(project *Namespace, user *UserAccount, giveCurrency *Currency, giveQuantity float64, recvCurrency *Currency, recvQuantity float64, tx string) (*Exchange, error) {

    uid, err := uuid.NewV4()
    if err != nil {
        return nil, err
    }

    return &Exchange{
        UID: uid.String(),
        Expiry: time.Now().Add(48 * time.Hour).UTC().Unix(),
        Project: project.UID,
        User: user.UID,
        GiveCurrency: giveCurrency.UID,
        GiveQuantity: giveQuantity,
        RecvCurrency: recvCurrency.UID,
        RecvQuantity: recvQuantity,
        Tx: tx,
        Created: timeNow(),
    }, nil
}

type Exchange struct {
    UID string `json:"uid"`
    Expiry int64 `json:"expiry"`
    Project string `json:"project"`
    User string `json:"-"`
    GiveCurrency string `json:"giveCurrency"`
    GiveQuantity float64 `json:"giveQuantity"`
    RecvCurrency string `json:"recvCurrency"`
    RecvQuantity float64 `json:"recvQuantity"`
    Tx string `json:"-"`
    Created int64 `json:"created"`
}

func (exchange *Exchange) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_EXCHANGE

    key := exchange.UID

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
