// +build test

package node

import (
	"crypto/rand"

	"github.com/spikeekips/mitum/common"
)

func init() {
	common.SetTestLogger(Log())
}

func NewRandomAddress() Address {
	b := make([]byte, 4)
	rand.Read(b)

	h, _ := NewAddress(b)
	return h
}
