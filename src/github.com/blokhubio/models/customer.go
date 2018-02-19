package models

import (
	"fmt"
	"encoding/hex"
	//
)

type Customer struct {
	UID string `json:"-"`
	Email string `json:"email"`
	Created int64 `json:"created"`
}

// Constructs a unique username for the descendant
func (customer *Customer) NamespaceUID(title string) string {
	return hex.EncodeToString(
		hash128(
			[]byte(fmt.Sprintf("%s.%s", title, customer.UID)),
		),
	)
}
