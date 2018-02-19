package models

import (
    //
	"golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
)

func NewNetworkConfig(chainConfig *ChainConfig) (*NetworkConfig, error) {

	return &NetworkConfig{
        UID: chainConfig.UID,
        Chain: chainConfig.UID,
        ChainName: chainConfig.ChainName,
        DefaultRpcPort: chainConfig.DefaultRpcPort,
        PrivateKeyVersion: chainConfig.PrivateKeyVersion,
        AddressPubkeyhashVersion: chainConfig.AddressPubkeyhashVersion,
        AddressChecksumValue: chainConfig.AddressChecksumValue,
        Created: timeNow(),
    }, nil
}

type NetworkConfig struct {
    UID string `json:"uid"`
    Chain string `json:"chain"`
    //
    RPCUser string `json:"rpcuser"`
    RPCPassword string `json:"rpcpassword"`
    //
    ChainName string `json:"chain-name"`
    DefaultRpcPort int `json:"default-rpc-port"`
    PrivateKeyVersion string `json:"-"`
    AddressPubkeyhashVersion string `json:"-"`
    AddressChecksumValue string `json:"-"`
    //
    Created int64 `json:"created"`
}

func (netCfg *NetworkConfig) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_NETWORKCONFIG
    key := netCfg.UID

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}

func toBool(s string) bool {
    return s == "true"
}
