package isaac

import (
	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/hash"
)

var (
	BlockHashHint string = "bk"
)

type Height struct {
	big.Big
}

func NewBlockHeight(height uint64) Height {
	return Height{Big: big.NewBig(height)}
}

func NewBlockHash(b []byte) (hash.Hash, error) {
	return hash.NewArgon2Hash(BlockHashHint, b)
}

// TODO create func to check block hash
