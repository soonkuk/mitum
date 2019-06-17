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

func NewBlockHash(hashes *hash.Hashes, b []byte) (hash.Hash, error) {
	return hashes.NewHash(BlockHashHint, b)
}

// TODO create func to check block hash
