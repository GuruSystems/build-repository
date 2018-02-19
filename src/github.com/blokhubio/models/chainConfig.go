package models

import (
    "errors"
    "strconv"
    "strings"
    //
    "golang.org/x/net/context"
	datastore "cloud.google.com/go/datastore"
	datastoreAE "google.golang.org/appengine/datastore"
    //
    "github.com/golangdaddy/go.uuid"
    //
    "github.com/blokhubio/multichain/params"
)

func NewChainConfig(params_dat string) (*ChainConfig, error) {

    uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	chain := &ChainConfig{
        UID: uid.String(),
        Created: timeNow(),
    }

    return chain, chain.ParseParams(params_dat)
}

type ChainConfig struct {
    UID string `json:"uid"`
    ChainName string `json:"chain-name"`
    DefaultRpcPort int `json:"default-rpc-port"`
    PrivateKeyVersion string `json:"private-key-version"`
    AddressPubkeyhashVersion string `json:"address-pubkeyhash-version"`
    AddressChecksumValue string `json:"address-checksum-value"`
    //
    RootStreamName string `json:"root-stream-name"`
    RootStreamOpen bool `json:"root-stream-open"`
    ChainIsTestnet bool `json:"chain-is-testnet"`
    TargetBlockTime int `json:"target-block-time"`
    MaximumBlockSize int `json:"maximum-block-size"`

    MaximumPerOutput int64 `json:"maximum-per-output"`
    AnyoneCanConnect bool `json:"anyone-can-connect"`
    AnyoneCanSend bool `json:"anyone-can-send"`
    AnyoneCanReceive bool `json:"anyone-can-receive"`
    AnyoneCanReceiveEmpty bool `json:"anyone-can-receive-empty"`
    AnyoneCanCreate bool `json:"anyone-can-create"`
    AnyoneCanIssue bool `json:"anyone-can-issue"`
    AnyoneCanMine bool `json:"anyone-can-mine"`
    AnyoneCanActivate bool `json:"anyone-can-activate"`
    AnyoneCanAdmin bool `json:"anyone-can-admin"`
    SupportMinerPrecheck bool `json:"support-miner-precheck"`
    AllowArbitraryOutputs bool `json:"allow-arbitrary-outputs"`
    AllowP2shOutputs bool `json:"allow-p2sh-outputs"`
    AllowMultisigOutputs bool `json:"allow-multisig-outputs"`
    //
    Created int64 `json:"created"`
}

func (chain *ChainConfig) DatastoreKey(ctx ...context.Context) interface{} {

    entityType := CONST_DS_ENTITY_CHAINCONFIG
    key := chain.UID

	if len(ctx) > 0 {
		return datastoreAE.NewKey(ctx[0], entityType, key, 0, nil)
	}
	return datastore.NameKey(entityType, key, nil)
}

func (chain *ChainConfig) ParseParams(params_dat string) error {

    params := params.Params{}

    for _, line := range strings.Split(params_dat, "\n") {

        line = strings.TrimSpace(line)

        parts := strings.Split(line, "#")

        if len(parts[0]) == 0 { continue }

        kv := strings.Split(strings.TrimSpace(parts[0]), "=")

        if len(kv) == 1 {
            continue
        }

        k := strings.TrimSpace(kv[0])
        v := strings.TrimSpace(kv[1])

        params[k] = v

    }

    for {
        var ok bool

        chain.ChainName, ok = params["chain-name"]
        if !ok {
            break
        }

        s, ok := params["default-rpc-port"]
        if !ok {
            break
        }
        i, err := strconv.Atoi(s)
        if err != nil {
            break
        }
        chain.DefaultRpcPort = i

        chain.PrivateKeyVersion, ok = params["private-key-version"]
        if !ok {
            break
        }

        chain.AddressPubkeyhashVersion, ok = params["address-pubkeyhash-version"]
        if !ok {
            break
        }

        chain.AddressChecksumValue, ok = params["address-checksum-value"]
        if !ok {
            break
        }

        chain.RootStreamName = params["root-stream-name"]
        if !ok {
            break
        }

        chain.RootStreamOpen = toBool(params["root-stream-open"])
        if !ok {
            break
        }

        chain.AnyoneCanConnect = toBool(params["anyone-can-connect"])
        if !ok {
            break
        }

        chain.AnyoneCanSend = toBool(params["anyone-can-send"])
        if !ok {
            break
        }

        chain.AnyoneCanReceive = toBool(params["anyone-can-receive"])
        if !ok {
            break
        }

        chain.AnyoneCanReceiveEmpty = toBool(params["anyone-can-receive-empty"])
        if !ok {
            break
        }

        chain.AnyoneCanCreate = toBool(params["anyone-can-create"])
        if !ok {
            break
        }

        chain.AnyoneCanIssue = toBool(params["anyone-can-issue"])
        if !ok {
            break
        }

        chain.AnyoneCanMine = toBool(params["anyone-can-mine"])
        if !ok {
            break
        }

        chain.AnyoneCanActivate = toBool(params["anyone-can-activate"])
        if !ok {
            break
        }

        chain.AnyoneCanAdmin = toBool(params["anyone-can-admin"])
        if !ok {
            break
        }

        return nil
    }
    return errors.New("FAILED TO PARSE NECCESSARY FIELDS FROM params.dat!")
}
