package isaac

import (
	"github.com/spikeekips/mitum/hash"
)

var (
	ProposalHashHint string = "pp"
)

func NewProposalHash(hashes *hash.Hashes, b []byte) (hash.Hash, error) {
	return hashes.NewHash(ProposalHashHint, b)
}

// TODO create func to check proposal hash
