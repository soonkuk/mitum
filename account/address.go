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

func NewAddress(b []byte) (Address, error) {
	h, err := hash.NewArgon2Hash(AccountHashHint, b)
	if err != nil {
		return Address{}, err
	}

	return Address{Hash: h}, nil
}
