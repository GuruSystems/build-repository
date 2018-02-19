package models

import (
    "fmt"
    "encoding/hex"
    "encoding/json"
    //
    "golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
    //
    "github.com/golangdaddy/go.uuid"
)

func NewNetworkMessage(networkConfig *NetworkConfig, action string, args []interface{}) (*NetworkMessage, error) {

    salt, err := uuid.NewV4()
    if err != nil {
        return nil, err
    }

    b, err := json.Marshal(args)
    if err != nil {
        return nil, err
    }

    return &NetworkMessage{
        UID: salt.String(),
        Network: networkConfig.UID,
        Action: action,
        Args: hex.EncodeToString(b),
        Created: timeNow(),
    }, nil
}

type NetworkMessage struct {
    UID string `json:"uid"`
    Network string `json:"network"`
    Action string `json:"action"`
    Args string `json:"args"`
    Signature string `json:"-"`
    Created int64 `json:"created"`
}

func (netMsg *NetworkMessage) Digest() ([]byte, error) {

    return hash128(
        []byte(
            fmt.Sprintf(
                "%s %s %s %s %d",
                netMsg.UID,
                netMsg.Network,
                netMsg.Action,
                netMsg.Args,
                netMsg.Created,
            ),
        ),
    ), nil
}

func (netMsg *NetworkMessage) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_NETWORKMESSAGE
    key := netMsg.UID

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}
