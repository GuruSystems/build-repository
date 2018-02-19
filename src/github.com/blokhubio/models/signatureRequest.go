package models

import (
    "github.com/blokhubio/multichain"
)

// Sent from App Engine to the cold node.
type SignatureRequest struct {
    TX string
    Array []*multichain.TxData
    Key string
    Args []string
    NetworkConfig
}
