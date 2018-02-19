package models

type Input struct {
	TxId string `json:"txid"`
	Vout int `json:"vout"`
	Value float64 `json:"value,omitempty"`
}

type Output struct {
	Asset string `json:"asset"`
	Amount float64 `json:"amount"`
}
