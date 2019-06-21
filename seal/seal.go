package seal

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
)

type Seal struct {
	header Header
	body   Body
}

type Header struct {
	signedAt  common.Time
	signature []byte
	hash      hash.Hash
}

func (h Header) SignedAt() common.Time {
	return h.signedAt
}

func (h Header) Signature() []byte {
	return h.signature
}

func (h Header) Hash() hash.Hash {
	return h.hash
}

type Body struct {
	source common.Address
}

func (b Body) Source() common.Address {
	return b.source
}
