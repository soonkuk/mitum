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

func NewTestSealBallot(
	psHash common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
) (Ballot, common.Seal, error) {
	ballot, err := NewBallot(psHash, source, height, round, stage, vote)
	if err != nil {
		return Ballot{}, common.Seal{}, err
	}

	seal, err := common.NewSeal(BallotSealType, ballot)
	return ballot, seal, err
}

// TODO remove if unused
type TestSealBroadcaster struct {
	sync.RWMutex
	policy   ConsensusPolicy
	homeNode common.HomeNode
	sendChan chan common.Seal
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

func (i *TestSealBroadcaster) SetSendChan(c chan common.Seal) {
	i.sendChan = c
}

func (i *TestSealBroadcaster) Send(sealType common.SealType, body common.Hasher, excludes ...common.Address) error {
	seal, err := common.NewSeal(sealType, body)
	if err != nil {
		return err
	}

	if err := seal.Sign(i.policy.NetworkID, i.homeNode.Seed()); err != nil {
		return err
	}

	if i.sendChan == nil {
		return nil
	}

	i.sendChan <- seal

	return nil
}

type TestMockVoting struct {
	result VoteResultInfo
	err    error
}

func (t *TestMockVoting) SetResult(result VoteResultInfo, err error) {
	t.result = result
	t.err = err
}

func (t *TestMockVoting) Open(common.Seal) (VoteResultInfo, error) {
	return t.result, t.err
}

func (t *TestMockVoting) Vote(seal common.Seal) (VoteResultInfo, error) {
	return t.result, t.err
}

func (t *TestMockVoting) Close() error {
	return nil
}
