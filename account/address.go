package account

import (
	"github.com/spikeekips/mitum/hash"
)

var (
	AccountHashHint string = "ac"
)

type Address struct {
	hash.Hash
}

func NewAddress(hashes *hash.Hashes, b []byte) (Address, error) {
	h, err := hashes.NewHash(AccountHashHint, b)
	if err != nil {
		return Address{}, err
	}

	return Address{Hash: h}, nil
}
