package models

import (
	"golang.org/x/net/context"
)

type Account interface {
	Type() string
	Name() string
	Addr() string
	DatastoreKey(...context.Context) interface{}
	Ns() *Namespace
}
