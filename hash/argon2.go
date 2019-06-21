package hash

import (
	"golang.org/x/crypto/argon2"
)

var (
	// TODO Argon2Sault should be changed at start time
	Argon2Sault []byte = []byte("argon2-default-sault; please set manually")
)

func NewArgon2Hash(hint string, b []byte) (Hash, error) {
	body := argon2.IDKey(b, Argon2Sault, 2, 64*1024, 2, 32)

	return NewHash(hint, body)
}
