package models

import (
	"fmt"
	"bytes"
	"golang.org/x/crypto/sha3"
)

func hash128(b []byte) []byte {
	digest := make([]byte, 16)
	sha3.ShakeSum128(digest, b)
	return digest
}

func hash256(b []byte) []byte {
	digest := make([]byte, 32)
	sha3.ShakeSum256(digest, b)
	return digest
}

func DeterministicUID(i interface{}) string {

	var inputs [][]byte

	switch v := i.(type) {

		case *Currency:

			inputs = append(
				inputs,
				[]byte(v.Project),
			)
			inputs = append(
				inputs,
				[]byte(v.Name),
			)

		case *Network:

			inputs = append(
				inputs,
				[]byte(v.Chain),
			)
			inputs = append(
				inputs,
				[]byte(v.Customer),
			)

		default:

			panic("DeterministicUID: unrecognised type")

	}

	return fmt.Sprintf(
		"%X",
		hash128(
			bytes.Join(inputs, nil),
		),
	)
}
