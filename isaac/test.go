// +build test

package isaac

import (
	"reflect"
	"sync"

	"github.com/spikeekips/mitum/common"
)

func NewTestProposal(proposer common.Address, transactions []common.Hash) Proposal {
	currentBlockHash, _ := common.NewHash("bk", []byte(common.RandomUUID()))
	nextBlockHash, _ := common.NewHash("bk", []byte(common.RandomUUID()))

	return NewProposal(
		Round(0),
		ProposalBlock{
			Height:  common.NewBig(99),
			Current: currentBlockHash,
			Next:    nextBlockHash,
		},
		ProposalState{
			Current: []byte(common.RandomUUID()),
			Next:    []byte(common.RandomUUID()),
		},
		transactions,
	)
}

func NewTestSealBallot(
	phash common.Hash,
	proposer common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
) Ballot {
	return NewBallot(phash, proposer, height, round, stage, vote)
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
	seal common.Signer,
	excludes ...common.Address,
) error {
	if err := seal.Sign(i.policy.NetworkID, i.home.Seed()); err != nil {
		return err
	}

	i.RLock()
	defer i.RUnlock()

	if i.senderChan == nil {
		return nil
	}

	i.senderChan <- reflect.ValueOf(seal.(common.Seal)).Elem().Interface().(common.Seal)

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

func (t *TestMockVotingBox) Open(Proposal) (VoteResultInfo, error) {
	return t.result, t.err
}

func (t *TestMockVotingBox) Vote(Ballot) (VoteResultInfo, error) {
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

type TBlockStorage struct {
	sync.RWMutex
	proposals []Proposal
}

func NewTBlockStorage() *TBlockStorage {
	return &TBlockStorage{}
}

func (t *TBlockStorage) Proposals() []Proposal {
	t.RLock()
	defer t.RUnlock()

	return t.proposals
}

func (t *TBlockStorage) NewBlock(proposal Proposal) error {
	t.Lock()
	defer t.Unlock()

	t.proposals = append(t.proposals, proposal)

	log.Debug("new block created", "proposal", proposal.Hash())
	return nil
}
