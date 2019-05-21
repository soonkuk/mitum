package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/element"
	"github.com/spikeekips/mitum/storage"
)

type ProposalValidator interface {
	Validate(Proposal) (Block, Vote, error)
	Store(Proposal) error
}

type DefaultProposalValidator struct {
	sync.RWMutex
	*common.Logger
	bst             BlockStorage
	runningProposal common.Hash
	block           Block
	batch           storage.Batch
	// TODO prepare state
}

func NewDefaultProposalValidator(bst BlockStorage, state *ConsensusState) *DefaultProposalValidator {
	return &DefaultProposalValidator{
		Logger: common.NewLogger(log),
		bst:    bst,
	}
}

func (v *DefaultProposalValidator) Validate(proposal Proposal) (Block, Vote, error) {
	v.Lock()
	defer v.Unlock()

	log_ := v.Log().New(log15.Ctx{"proposal": proposal.Hash()})

	if !v.runningProposal.Empty() {
		log_.Debug("another proposal is validating", "another", v.runningProposal)
		return Block{}, VoteNOP, ValidationIsRunningError.AppendMessage(
			"proposal=%v another=%v",
			proposal.Hash(), v.runningProposal,
		)
	}

	log_.Debug("starting to validate proposal")

	v.runningProposal = proposal.Hash()
	v.block = Block{}
	v.batch = nil

	defer func() {
		log_.Debug("finish to validate proposal")
		v.runningProposal = common.Hash{}
	}()

	// TODO implement

	// NOTE create new block
	block, batch, err := v.bst.NewBlock(proposal)
	if err != nil {
		return Block{}, VoteNONE, err
	}

	v.block = block
	v.batch = batch

	return block, VoteYES, nil
}

func (v *DefaultProposalValidator) Validated(proposal common.Hash) bool {
	v.RLock()
	defer v.RUnlock()

	return v.block.proposal.Equal(proposal)
}

func (v *DefaultProposalValidator) Store(proposal Proposal) error {
	v.RLock()
	defer v.RUnlock()

	if v.block.Hash().Empty() {
		return ValidationIsNotDoneError.AppendMessage("block is nil")
	}

	if !v.Validated(proposal.Hash()) {
		return ValidationIsNotDoneError.AppendMessage(
			"proposal is not validated; validated=%v proposal=%v",
			v.block.Proposal(),
			proposal.Hash(),
		)
	}

	if v.batch == nil {
		return ValidationIsNotDoneError.AppendMessage("batch is nil")
	}

	if err := v.bst.Storage().WriteBatch(v.batch); err != nil {
		return err
	}

	return nil
}

type TransactionValidation struct {
	sync.RWMutex
	*common.Logger
	st storage.Storage
}

func NewTransactionValidation() *TransactionValidation {
	return &TransactionValidation{
		Logger: common.NewLogger(log),
	}
}

func (t *TransactionValidation) Validate(transactions []element.Transaction) ([]common.Hash /* valid */, []common.Hash /* invalid */, error) {
	// TODO implement
	var valids, invalids []common.Hash
	for _, t := range transactions {
		valids = append(valids, t.Hash())
	}

	return valids, invalids, nil
}
