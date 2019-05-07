package isaac

import (
	"github.com/spikeekips/mitum/common"
)

type BlockStorage interface {
	NewBlock(common.Seal /* Seal(Proposal) */) error
}

type DefaultBlockStorage struct {
	state *ConsensusState
}

func NewDefaultBlockStorage(state *ConsensusState) (*DefaultBlockStorage, error) {
	return &DefaultBlockStorage{
		state: state,
	}, nil
}

func (i *DefaultBlockStorage) NewBlock(seal common.Seal) error {
	if seal.Type() != ProposalSealType {
		return common.InvalidSealTypeError
	}

	var proposal Proposal
	if p, ok := seal.(Proposal); !ok {
		return common.UnknownSealTypeError.SetMessage("not Proposal")
	} else {
		proposal = p
	}

	// TODO store block

	// update state
	prevState := &ConsensusState{}
	prevState.SetHeight(i.state.Height())
	prevState.SetBlock(i.state.Block())
	prevState.SetState(i.state.State())

	i.state.SetHeight(proposal.Block.Height.Inc())
	i.state.SetBlock(proposal.Block.Next)
	i.state.SetState(proposal.State.Next)

	log.Debug(
		"allConfirmed",
		"proposal", seal.Hash(),
		"old-block-height", prevState.Height(),
		"old-block-hash", prevState.Block(),
		"old-state-hash", prevState.State(),
		"new-block-height", i.state.Height().String(),
		"new-block-hash", i.state.Block(),
		"new-state-hash", i.state.State(),
	)

	return nil
}
