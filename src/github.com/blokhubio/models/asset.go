package models

type Asset struct {
	Project string `json:"project,omitempty"`
	Currency string `json:"currency"`
	Recipient string `json:"recipient,omitempty"`
	Quantity float64 `json:"quantity"`
	Time int64 `json:"time,omitempty"`
}
