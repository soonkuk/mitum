// +build test

package isaac

import (
	"crypto/rand"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
)

func init() {
	common.SetTestLogger(Log())
}

func NewRandomProposalHash() hash.Hash {
	b := make([]byte, 4)
	_, _ = rand.Read(b)

	h, _ := NewProposalHash(b)
	return h
}

func NewRandomBlockHash() hash.Hash {
	b := make([]byte, 4)
	_, _ = rand.Read(b)

	h, _ := NewBlockHash(b)
	return h
}
