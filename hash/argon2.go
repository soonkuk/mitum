package hash

import (
	"crypto/sha256"
)

var (
	// TODO Argon2Sault should be changed at start time
	Argon2Sault []byte = []byte("argon2-default-sault; please set manually")
)

/*
func NewArgon2Hash(hint string, b []byte) (Hash, error) {
	body := argon2.IDKey(b, Argon2Sault, 1, 64*1024, ^uint8(0), 32)

	return NewHash(hint, body)
}
*/

func NewArgon2Hash(hint string, b []byte) (Hash, error) {
	f := sha256.Sum256(b)
	s := sha256.Sum256(f[:])

	return NewHash(hint, s[:])
}

/*
func NewArgon2Hash(hint string, b []byte) (Hash, error) {
	f := sha512.Sum512(b)
	s := sha512.Sum512(f[:])

	return NewHash(hint, s[:])
}
*/
