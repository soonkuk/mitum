package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

func NewBallot(n node.Address, height Height, round Round, stage Stage, proposal, currentBlock, nextBlock hash.Hash) (BaseBallot, error) {
	body := BaseBallotBody{
		Node:         n,
		Height:       height,
		Round:        round,
		Stage:        stage,
		Proposal:     proposal,
		CurrentBlock: currentBlock,
		NextBlock:    nextBlock,
	}
	return NewBaseBallot(body)
}

func IsValidINITBallot(BaseBallotBody) error {
	return nil
}

func IsValidSIGNBallot(BaseBallotBody) error {
	return nil
}

func IsValidACCEPTBallot(BaseBallotBody) error {
	return nil
}
