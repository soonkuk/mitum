// +build test

package account

import "crypto/rand"

func NewRandomAddress() Address {
	b := make([]byte, 4)
	_, _ = rand.Read(b)

	a, _ := NewAddress(b)
	return a
}
