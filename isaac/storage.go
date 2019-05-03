package isaac

import (
	"github.com/spikeekips/mitum/common"
)

type BlockStorage interface {
	NewBlock(common.Seal /* Seal(Propose) */) error
}

type DefaultBlockStorage struct {
	state *ConsensusState
}

func NewDefaultBlockStorage(state *ConsensusState) (*DefaultBlockStorage, error) {
	return &DefaultBlockStorage{
		state: state,
	}, nil
}

func (i *DefaultBlockStorage) NewBlock(proposeSeal common.Seal) error {
	if proposeSeal.Type != ProposeSealType {
		return common.InvalidSealTypeError
	}

	psHash, _, err := proposeSeal.Hash()
	if err != nil {
		return err
	}

	var propose Propose
	if err := proposeSeal.UnmarshalBody(&propose); err != nil {
		return err
	}

	// TODO store block

	// update state
	prevState := &ConsensusState{}
	prevState.SetHeight(i.state.Height())
	prevState.SetBlock(i.state.Block())
	prevState.SetState(i.state.State())

	i.state.SetHeight(propose.Block.Height.Inc())
	i.state.SetBlock(propose.Block.Next)
	i.state.SetState(propose.State.Next)

	log.Debug(
		"allConfirmed",
		"psHash", psHash,
		"old-block-height", prevState.Height(),
		"old-block-hash", prevState.Block(),
		"old-state-hash", prevState.State(),
		"new-block-height", i.state.Height().String(),
		"new-block-hash", i.state.Block(),
		"new-state-hash", i.state.State(),
	)

	return nil
}
