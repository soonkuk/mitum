package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type Ballot interface {
	seal.Seal
	Node() node.Address
	Height() Height
	Round() Round
	Stage() Stage
	//Proposer() hash.Hash // TODO should be included
	Proposal() hash.Hash
	NextBlock() hash.Hash
}
