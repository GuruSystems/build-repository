package models

import (
    "golang.org/x/net/context"
    datastore "cloud.google.com/go/datastore"
    datastoreAE "google.golang.org/appengine/datastore"
)

type TX struct {
    Id string
    Height int
    Hash string
}

func NewTransaction(srcUser, destUser *UserAccount, srcAddress, destAddress *Address, currency *Currency, quantity float64) *Transaction {

    return &Transaction{
        Sender: srcUser.Username,
        Receiver: destUser.Username,
        SendUser: srcUser,
        SendAddr: srcAddress.Addr,
        SendAddress: srcAddress,
        ReceiveUser: destUser,
        ReceiveAddr: destAddress.Addr,
        ReceiveAddress: destAddress,
        Currency: currency.Alias(),
        Inputs: []Input{},
        Outputs: []Output{},
        OutputsTotal: quantity,
    }
}

type Transaction struct {
	TxId string `json:"txid"`
    Sender string `json:"sender"`
    Receiver string `json:"receiver"`
    SendUser *UserAccount `json:"-" datastore:"-"`
    SendAddr string `json:"sendAddress"`
	SendAddress *Address `json:"-" datastore:"-"`
    ReceiveUser *UserAccount `json:"-" datastore:"-"`
    ReceiveAddr string `json:"receiveAddress"`
    ReceiveAddress *Address `json:"-" datastore:"-"`
    Currency string `json:"currency"`
    Inputs []Input `json:"inputs"`
	Outputs []Output `json:"outputs"`
    OutputsTotal float64 `json:"outputsTotal"`
    Size int `json:"outputs"`
}

func (tx *Transaction) DatastoreKey(ctx ...context.Context) interface{} {
	if len(ctx) > 0 {
		return datastoreAE.NewIncompleteKey(ctx[0], CONST_DS_ENTITY_TRANSACTION, nil)
	}
	return datastore.IncompleteKey(CONST_DS_ENTITY_TRANSACTION, nil)
}
