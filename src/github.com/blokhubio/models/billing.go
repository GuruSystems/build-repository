package models

import (
    "time"
    //
    "golang.org/x/net/context"
    datastore "cloud.google.com/go/datastore"
    datastoreAE "google.golang.org/appengine/datastore"
)

func NewBillingRecord(namespace *Namespace, transaction *Transaction) *BillingRecord {
    return &BillingRecord{
        namespace,
        namespace.UID,
        time.Now().UTC().Unix(),
        *transaction,
    }
}

type BillingRecord struct {
    namespace *Namespace
    Namespace string `json:"namespace"`
    Time int64 `json:"unixTimeUTC"`
    Transaction
}

func (br *BillingRecord) DatastoreKey(ctx ...context.Context) interface{} {

    namespaceKey := br.namespace.DatastoreKey()

    if len(ctx) > 0 {
        return datastoreAE.NewIncompleteKey(
            ctx[0],
            CONST_DS_ENTITY_BILLING_RECORD,
            namespaceKey.(*datastoreAE.Key),
        )
    }
    return datastore.IncompleteKey(
        CONST_DS_ENTITY_BILLING_RECORD,
        namespaceKey.(*datastore.Key),
    )
}
