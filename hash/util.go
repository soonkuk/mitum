package hash

import (
	"encoding"

	"github.com/spikeekips/mitum/common"
)

type Hashable interface {
	Hash() Hash
}

func MakeInstanceHash(hashes *Hashes, hint string, i interface{}) (Hash, error) {
	if hashable, ok := i.(Hashable); ok {
		return hashable.Hash(), nil
	}

	var err error
	var b []byte
	if bm, ok := i.(encoding.BinaryMarshaler); ok {
		b, err = bm.MarshalBinary()
	} else {
		b, err = common.RLPEncode(i)
	}

	if err != nil {
		return Hash{}, HashFailedError.New(err)
	}

	return hashes.NewHash(hint, b)
}
