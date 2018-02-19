package models

import (
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
	//
	"github.com/golangdaddy/go.uuid"
)

func NewStream(project *Namespace, name string) (*Stream, error) {

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

    salt, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

    return &Stream{
		UID: uid.String(),
		Project: project.UID,
        Salt: salt.String(),
        Name: name,
        Created: timeNow(),
    }, nil
}

type Stream struct {
	UID string `json:"uid"`
	Project string `json:"project"`
    Salt string `json:"-"`
    Name string `json:"name"`
    Created int64 `json:"created"`
}

func (stream *Stream) DatastoreKey(ctx ...context.Context) interface{} {
	entityType := CONST_DS_ENTITY_STREAM
	key := stream.UID
	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
