package models

import (
    "golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
    //
	"github.com/golangdaddy/go.uuid"
)

type ApiToken struct {
    UID string `json:"uid"`
    Label string `json:"label"`
    Resource string `json:"resource"`
    Subject string `json:"subject"`
    Secret string `json:"secret"`
    Created int64 `json:"created"`
}

func NewApiToken(label, resource, subject string) (*ApiToken, error) {

    uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

    secret, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

    return &ApiToken{
        UID: uid.String(),
        Label: label,
        Resource: resource,
        Subject: subject,
        Secret: secret.String(),
        Created: timeNow(),
    }, nil
}

func (token *ApiToken) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_APITOKEN
    key := token.UID

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
