package isaac

import (
	"github.com/spikeekips/mitum/hash"
)

var (
	ProposalHashHint string = "pp"
)

func NewProposalHash(b []byte) (hash.Hash, error) {
	return hash.NewArgon2Hash(ProposalHashHint, b)
}

// TODO create func to check proposal hash
