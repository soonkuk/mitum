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

func (i *DefaultBlockStorage) NewBlock(proposal Proposal) error {
	// TODO store block

	log.Debug(
		"new block created",
		"proposal", proposal,
	)

	return nil
}
