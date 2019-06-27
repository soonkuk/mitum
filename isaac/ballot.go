package isaac

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

var (
	BallotType     common.DataType = common.NewDataType(1, "ballot")
	BallotHashHint string          = "ballot"
)

type Ballot interface {
	seal.Seal
	Node() node.Address
	Height() Height
	Round() Round
	Stage() Stage
	//Proposer() hash.Hash // TODO should be included
	Proposal() hash.Hash
	CurrentBlock() hash.Hash
	NextBlock() hash.Hash
}
