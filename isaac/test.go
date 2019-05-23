// +build test

package isaac

import (
	"reflect"
	"sync"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/storage/leveldbstorage"
)

func NewTestProposal(proposer common.Address, transactions []common.Hash) Proposal {
	currentBlockHash, _ := common.NewHash("bk", []byte(common.RandomUUID()))

	return NewProposal(
		Round(0),
		ProposalBlock{
			Height:  common.NewBig(99),
			Current: currentBlockHash,
		},
		ProposalState{
			Current: []byte(common.RandomUUID()),
			Next:    []byte(common.RandomUUID()),
		},
		transactions,
	)
}

func NewTestSealBallot(
	proposal common.Hash,
	proposer common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	block common.Hash,
) Ballot {
	var ballot Ballot
	switch stage {
	case VoteStageINIT:
		ballot = NewINITBallot(
			height,
			round,
			proposer,
			nil, // TODO set validators
		)
	case VoteStageSIGN:
		ballot = NewSIGNBallot(
			height,
			round,
			proposer,
			nil, // TODO set validators
			proposal,
			block,
			vote,
		)
	case VoteStageACCEPT:
		ballot = NewACCEPTBallot(
			height,
			round,
			proposer,
			nil, // TODO set validators
			proposal,
			block,
		)
	}

	return ballot
}

type TestSealBroadcaster struct {
	sync.RWMutex
	policy     ConsensusPolicy
	home       common.HomeNode
	senderChan chan common.Seal
}

func NewTestSealBroadcaster(
	policy ConsensusPolicy,
	home common.HomeNode,
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

func (t *TestMockVotingBox) Clear() error {
	return nil
}

type TBlockStorage struct {
	sync.RWMutex
	*common.Logger
	blocks           []Block
	blocksbyProposal map[common.Hash]Block
	st               *leveldbstorage.Storage
}

func NewTBlockStorage() *TBlockStorage {
	return &TBlockStorage{
		Logger:           common.NewLogger(log),
		blocksbyProposal: map[common.Hash]Block{},
		st:               leveldbstorage.NewMemStorage(),
	}
}

func (t *TBlockStorage) Blocks() []Block {
	t.RLock()
	defer t.RUnlock()

	return t.blocks
}

func (t *TBlockStorage) Storage() storage.Storage {
	return t.st
}

func (t *TBlockStorage) NewBlock(proposal Proposal) (Block, storage.Batch, error) {
	t.Lock()
	defer t.Unlock()

	block := Block{
		version:      CurrentBlockVersion,
		height:       proposal.Block.Height.Inc(),
		hash:         common.NewRandomHash("bk"), // TODO set new block hash
		prevHash:     proposal.Block.Current,
		state:        proposal.State.Next,
		prevState:    proposal.State.Current,
		proposer:     proposal.Source(),
		round:        proposal.Round,
		proposedAt:   proposal.SignedAt(),
		proposal:     proposal.Hash(),
		transactions: proposal.Transactions,
	}

	t.blocks = append(t.blocks, block)
	t.blocksbyProposal[proposal.Hash()] = block

	t.Log().Debug("new block prepared", "proposal", proposal.Hash(), "block", block)
	return block, &storage.TBatch{}, nil
}

func (t *TBlockStorage) LatestBlock() (Block, error) {
	if len(t.blocks) < 1 {
		return Block{}, BlockNotFoundError.SetMessage("no blocks")
	}

	return t.blocks[len(t.blocks)-1], nil
}

func (t *TBlockStorage) BlockByProposal(proposal common.Hash) (Block, error) {
	block, ok := t.blocksbyProposal[proposal]
	if !ok {
		return Block{}, BlockNotFoundError
	}

	return block, nil
}

type NullProposalValidator struct {
	bst BlockStorage
}

func NewNullProposalValidator(bst BlockStorage) *NullProposalValidator {
	return &NullProposalValidator{
		bst: bst,
	}
}

func (n *NullProposalValidator) Validate(proposal Proposal) (Block, Vote, error) {
	block, _, err := n.bst.NewBlock(proposal)
	if err != nil {
		return Block{}, VoteNONE, err
	}

	return block, VoteYES, nil
}

func (n *NullProposalValidator) Store(proposal Proposal) error {
	_, _, err := n.bst.NewBlock(proposal)
	if err != nil {
		return err
	}

	return nil
}
