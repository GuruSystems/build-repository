package models

import (
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
	//
	"github.com/golangdaddy/go.uuid"
)

func NewProject(networkConfig *NetworkConfig, customer *Customer, title string) (*Namespace, error) {

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	salt, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &Namespace{
		UID: uid.String(),
		Network: networkConfig.UID,
		Customer: customer.UID,
		Salt: salt.String(),
		Title: title,
		Created: timeNow(),
	}, nil
}

type Namespace struct {
	UID string `json:"uid"`
	Network string `json:"network"`
	Customer string `json:"customer"`
	Salt string `json:"-"`
	Public bool `json:"public"`
	Title string `json:"title"`
	Description string `json:"description"`
	PaidTier bool `json:"paidTier"`
	//
	Usage int64 `json:",omitempty" datastore:"-"`
	Created int64 `json:"created"`
}

func (namespace *Namespace) Ns() *Namespace {
	return namespace
}

func (namespace *Namespace) DatastoreKey(ctx ...context.Context) interface{} {
	entityType := CONST_DS_ENTITY_NAMESPACE
	key := namespace.UID
	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
