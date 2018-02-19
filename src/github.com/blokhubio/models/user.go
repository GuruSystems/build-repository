package models

import (
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
	//
	"github.com/golangdaddy/go.uuid"
)

func NewUserAccount(project *Namespace, username string, agent bool) (*UserAccount, error) {

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	salt, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &UserAccount{
		UID: uid.String(),
		Project: project.UID,
		Salt: salt.String(),
		Agent: agent,
		Username: username,
		Created: timeNow(),
	}, nil
}

type UserAccount struct {
	UID string `json:"uid"`
	Agent bool `json:"agent"`
	Project string `json:"namespaceUID"`
	Salt string `json:"-"`
	Username string `json:"username"`
	Created int64 `json:"created"`
}

func (user *UserAccount) DatastoreKey(ctx ...context.Context) interface{} {
	entityType := CONST_DS_ENTITY_ACCOUNT_USER
	key := user.UID
	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
