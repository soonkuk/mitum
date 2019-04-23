// +build test

package isaac

import "github.com/spikeekips/mitum/common"

func NewTestSealProposeBallot(proposer common.Address, transactions []common.Hash) (ProposeBallot, common.Seal, error) {
	currentBlockHash, err := common.NewHash("bh", []byte(common.RandomUUID()))
	if err != nil {
		return ProposeBallot{}, common.Seal{}, err
	}

	nextBlockHash, err := common.NewHash("bh", []byte(common.RandomUUID()))
	if err != nil {
		return ProposeBallot{}, common.Seal{}, err
	}

	ballot := ProposeBallot{
		Version: CurrentBallotVersion,
		Block: BallotBlock{
			Height:  common.NewBig(99),
			Current: currentBlockHash,
			Next:    nextBlockHash,
		},
		State: BallotBlockState{
			Current: []byte(common.RandomUUID()),
			Next:    []byte(common.RandomUUID()),
		},
		ProposedAt:   common.Now(),
		Transactions: transactions,
	}

	seal, err := common.NewSeal(ProposeBallotSealType, ballot)
	return ballot, seal, err
}

func NewTestSealVoteBallot(proposeBallotSealHash common.Hash, source common.Address, stage VoteStage, vote Vote) (VoteBallot, common.Seal, error) {

	ballot := VoteBallot{
		Version:           CurrentBallotVersion,
		ProposeBallotSeal: proposeBallotSealHash,
		Stage:             stage,
		Vote:              vote,
		Round:             0,
		VotedAt:           common.Now(),
	}

	seal, err := common.NewSeal(VoteBallotSealType, ballot)
	return ballot, seal, err
}
