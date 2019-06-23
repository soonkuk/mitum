package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

func NewINITBallot(n node.Address, height Height, round Round) (BaseBallot, error) {
	body := BaseBallotBody{
		Node:   n,
		Height: height,
		Round:  round,
		Stage:  StageINIT,
	}
	return NewBaseBallot(body)
}

func NewSIGNBallot(n node.Address, height Height, round Round, proposal, nextBlock hash.Hash) (BaseBallot, error) {
	body := BaseBallotBody{
		Node:      n,
		Height:    height,
		Round:     round,
		Stage:     StageSIGN,
		Proposal:  proposal,
		NextBlock: nextBlock,
	}
	return NewBaseBallot(body)
}

func NewACCEPTBallot(n node.Address, height Height, round Round, proposal, nextBlock hash.Hash) (BaseBallot, error) {
	body := BaseBallotBody{
		Node:      n,
		Height:    height,
		Round:     round,
		Stage:     StageACCEPT,
		Proposal:  proposal,
		NextBlock: nextBlock,
	}
	return NewBaseBallot(body)
}

func IsValidINITBallot(body BaseBallotBody) error {
	if body.Stage != StageINIT {
		return InvalidBallotError.Newf("(given)%q (expected)%q", body.Stage, StageINIT)
	}
	if !body.Proposal.Empty() {
		return InvalidBallotError.Newf("Proposal should empty for INIT")
	}
	if !body.NextBlock.Empty() {
		return InvalidBallotError.Newf("NextBlock should empty for INIT")
	}

	return nil
}

func IsValidSIGNBallot(body BaseBallotBody) error {
	if body.Stage != StageSIGN {
		return InvalidBallotError.Newf("(given)%q (expected)%q", body.Stage, StageSIGN)
	}
	if body.Proposal.Empty() {
		return InvalidBallotError.Newf("Proposal should not empty for SIGN")
	}
	if body.NextBlock.Empty() {
		return InvalidBallotError.Newf("NextBlock should not empty for SIGN")
	}

	return nil
}

func IsValidACCEPTBallot(body BaseBallotBody) error {
	if body.Stage != StageACCEPT {
		return InvalidBallotError.Newf("(given)%q (expected)%q", body.Stage, StageACCEPT)
	}
	if body.Proposal.Empty() {
		return InvalidBallotError.Newf("Proposal should not empty for ACCEPT")
	}
	if body.NextBlock.Empty() {
		return InvalidBallotError.Newf("NextBlock should not empty for ACCEPT")
	}

	return nil
}
