package node

import (
	"github.com/spikeekips/mitum/hash"
)

var (
	AddressHashHint string = "na"
)

type Address struct {
	hash.Hash
}

func NewAddress(hashes *hash.Hashes, b []byte) (Address, error) {
	h, err := hashes.NewHash(AddressHashHint, b)
	if err != nil {
		return Address{}, err
	}

	return Address{Hash: h}, nil
}
