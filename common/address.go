package common

import "github.com/ethereum/go-ethereum/rlp"

type Address interface {
	rlp.Encoder
	IsValid() error
	Equal(Address) bool
	String() string
}
