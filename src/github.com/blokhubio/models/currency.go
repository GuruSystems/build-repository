package models

import (
	"fmt"
	//
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
	//
	"github.com/golangdaddy/go.uuid"
)

func NewCurrency(project *Namespace, currencyName string, units int) (*Currency, error) {

	salt, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &Currency{
		UID: salt.String(),
		Project: project.UID,
		Name: currencyName,
		Units: units,
		Created: timeNow(),
	}, nil
}

type Currency struct {
	UID string `json:"uid"`
	Project string `json:"project"`
	Name string `json:"name"`
	Units int `json:"units"`
	Created int64 `json:"created"`
}

func (currency *Currency) Alias() string {

	return fmt.Sprintf(
		"%x",
		hash128(
			[]byte(currency.UID),
		),
	)
}

func (currency *Currency) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_CURRENCY
	key := DeterministicUID(currency)

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
