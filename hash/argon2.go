package hash

import (
	"golang.org/x/crypto/argon2"
)

var (
	emptyArgon2HashValue = [32]byte{}
	// TODO Argon2Sault should be changed at start time
	Argon2Sault    []byte            = []byte("argon2-default-sault; please set manually")
	Argon2HashType HashAlgorithmType = NewHashAlgorithmType(2, "argon2")
)

type Argon2Hash struct {
}

func NewArgon2Hash() Argon2Hash {
	return Argon2Hash{}
}

func (a Argon2Hash) Type() HashAlgorithmType {
	return Argon2HashType
}

func (a Argon2Hash) GenerateHash(b []byte) ([]byte, error) {
	return argon2.IDKey(b, Argon2Sault, 2, 64*1024, 2, 32), nil
}

func (a Argon2Hash) IsValid(b []byte) error {
	if len(b) != 32 {
		return HashFailedError.Newf("argon2 length should be 32; length=%d", len(b))
	}

	{
		empty := true
		for i, a := range b {
			if a != emptyArgon2HashValue[i] {
				empty = false
				break
			}
		}

		if empty {
			return EmptyHashError.Newf("empty hash body")
		}
	}

	return nil
}
