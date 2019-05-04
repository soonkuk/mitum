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
	policy     ConsensusPolicy
	home       *common.HomeNode
	senderChan chan common.Seal
}

func NewTestSealBroadcaster(
	policy ConsensusPolicy,
	home *common.HomeNode,
) (*TestSealBroadcaster, error) {
	return &TestSealBroadcaster{
		policy: policy,
		home:   home,
	}, nil
}

func (i *TestSealBroadcaster) SetSenderChan(c chan common.Seal) {
	i.Lock()
	defer i.Unlock()

	if i.senderChan != nil {
		close(i.senderChan)
	}

	i.senderChan = c
}

func (i *TestSealBroadcaster) Send(
	sealType common.SealType,
	body common.Hasher,
	excludes ...common.Address,
) error {
	seal, err := common.NewSeal(sealType, body)
	if err != nil {
		return err
	}

	if err := seal.Sign(i.policy.NetworkID, i.home.Seed()); err != nil {
		return err
	}

	i.RLock()
	defer i.RUnlock()

	if i.senderChan == nil {
		return nil
	}

	i.senderChan <- seal

	return nil
}

type TestMockVotingBox struct {
	result VoteResultInfo
	err    error
}

func (t *TestMockVotingBox) SetResult(result VoteResultInfo, err error) {
	t.result = result
	t.err = err
}

func (t *TestMockVotingBox) Open(common.Seal) (VoteResultInfo, error) {
	return t.result, t.err
}

func (t *TestMockVotingBox) Vote(seal common.Seal) (VoteResultInfo, error) {
	return t.result, t.err
}

func (t *TestMockVotingBox) Close() error {
	return nil
}

type TProposerSelector struct {
	sync.RWMutex
	proposer common.Node
}

func NewTProposerSelector() *TProposerSelector {
	return &TProposerSelector{}
}

func (t *TProposerSelector) SetProposer(proposer common.Node) {
	t.Lock()
	defer t.Unlock()

	t.proposer = proposer
}

func (t *TProposerSelector) Select(block common.Hash, height common.Big, round Round) (common.Node, error) {
	t.RLock()
	defer t.RUnlock()

	return t.proposer, nil
}
