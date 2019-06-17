package hash

import "github.com/spikeekips/mitum/common"

type HashAlgorithm interface {
	Type() common.DataType
	GenerateHash([]byte) []byte
	IsValid([]byte) error
}
