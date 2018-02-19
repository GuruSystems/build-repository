package models

func NewNetworkMessageAck(msg, sig string) *NetworkMessageAck {
    return &NetworkMessageAck{
        Msg: msg,
        Sig: sig,
    }
}

type NetworkMessageAck struct {
    Msg string `json:"msg"`
    Sig string `json:"sig"`
}
