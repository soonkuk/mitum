// +build test

package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

func NewTestPropose(proposer common.Address, transactions []common.Hash) (Propose, error) {
	currentBlockHash, err := common.NewHash("bk", []byte(common.RandomUUID()))
	if err != nil {
		return Propose{}, err
	}

	nextBlockHash, err := common.NewHash("bk", []byte(common.RandomUUID()))
	if err != nil {
		return Propose{}, err
	}

	return Propose{
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
	}, nil
}

func NewTestSealPropose(proposer common.Address, transactions []common.Hash) (Propose, common.Seal, error) {
	propose, err := NewTestPropose(proposer, transactions)
	if err != nil {
		return Propose{}, common.Seal{}, err
	}

	seal, err := common.NewSeal(ProposeSealType, propose)
	return propose, seal, err
}

func NewTestSealBallot(psHash common.Hash, source common.Address, height common.Big, round Round, stage VoteStage, vote Vote) (Ballot, common.Seal, error) {
	ballot, err := NewBallot(psHash, source, height, round, stage, vote)
	if err != nil {
		return Ballot{}, common.Seal{}, err
	}

	seal, err := common.NewSeal(BallotSealType, ballot)
	return ballot, seal, err
}

type TestSealBroadcaster struct {
	sync.RWMutex
	policy   ConsensusPolicy
	homeNode common.HomeNode
	seals    []common.Seal
	last     int
}

func NewTestSealBroadcaster(
	policy ConsensusPolicy,
	homeNode common.HomeNode,
) (*TestSealBroadcaster, error) {
	return &TestSealBroadcaster{
		policy:   policy,
		homeNode: homeNode,
	}, nil
}

func (i *TestSealBroadcaster) Send(sealType common.SealType, body common.Hasher, excludes ...common.Address) error {
	seal, err := common.NewSeal(sealType, body)
	if err != nil {
		return err
	}

	if err := seal.Sign(i.policy.NetworkID, i.homeNode.Seed()); err != nil {
		return err
	}

	i.Lock()
	defer i.Unlock()

	i.seals = append(i.seals, seal)

	return nil
}

func (i *TestSealBroadcaster) NewSeals() []common.Seal {
	i.Lock()
	defer i.Unlock()

	if len(i.seals) == i.last {
		return nil
	}

	l := i.seals[i.last:]
	i.last = len(i.seals)
	return l
}

func (i *TestSealBroadcaster) Clear() {
	i.Lock()
	defer i.Unlock()

	i.seals = nil
}
