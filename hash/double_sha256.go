package hash

import (
	"crypto/sha256"

	"github.com/spikeekips/mitum/common"
)

var (
	emptyDoubleSHA256HashValue                 = [32]byte{}
	DoubleSHA256HashType       common.DataType = common.NewDataType(1, "double-sha256")
)

type DoubleSHA256Hash struct {
}

func NewDoubleSHA256Hash() DoubleSHA256Hash {
	return DoubleSHA256Hash{}
}

func (d DoubleSHA256Hash) Type() common.DataType {
	return DoubleSHA256HashType
}

func (d DoubleSHA256Hash) GenerateHash(b []byte) []byte {
	f := sha256.Sum256(b)
	s := sha256.Sum256(f[:])

	return s[:]
}

func (d DoubleSHA256Hash) IsValid(b []byte) error {
	if len(b) != 32 {
		return HashFailedError.Newf("double sha256 length should be 32; length=%d", len(b))
	}

	{
		empty := true
		for i, a := range b {
			if a != emptyDoubleSHA256HashValue[i] {
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
