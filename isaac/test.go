// +build test

package isaac

import "github.com/spikeekips/mitum/common"

func NewTestSealPropose(proposer common.Address, transactions []common.Hash) (Propose, common.Seal, error) {
	currentBlockHash, err := common.NewHash("bh", []byte(common.RandomUUID()))
	if err != nil {
		return Propose{}, common.Seal{}, err
	}

	nextBlockHash, err := common.NewHash("bh", []byte(common.RandomUUID()))
	if err != nil {
		return Propose{}, common.Seal{}, err
	}

	ballot := Propose{
		Version:  CurrentBallotVersion,
		Proposer: proposer,
		Round:    0,
		Block: ProposeBlock{
			Height:  common.NewBig(99),
			Current: currentBlockHash,
			Next:    nextBlockHash,
		},
		State: ProposeState{
			Current: []byte(common.RandomUUID()),
			Next:    []byte(common.RandomUUID()),
		},
		ProposedAt:   common.Now(),
		Transactions: transactions,
	}

	seal, err := common.NewSeal(ProposeSealType, ballot)
	return ballot, seal, err
}

func NewTestSealBallot(psHash common.Hash, source common.Address, stage VoteStage, vote Vote) (Ballot, common.Seal, error) {
	ballot := Ballot{
		Version:     CurrentBallotVersion,
		Source:      source,
		ProposeSeal: psHash,
		Stage:       stage,
		Vote:        vote,
		VotedAt:     common.Now(),
	}

	seal, err := common.NewSeal(BallotSealType, ballot)
	return ballot, seal, err
}
