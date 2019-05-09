package isaac

import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

// TODO state should be considered
type BlockStorage interface {
	NewBlock(Proposal) (Block, error)
	LatestBlock() (Block, error)
	BlockByProposal(common.Hash) (Block, error)
}

type DefaultBlockStorage struct {
	state *ConsensusState
	log   log15.Logger
}

func NewDefaultBlockStorage(state *ConsensusState) (*DefaultBlockStorage, error) {
	return &DefaultBlockStorage{
		state: state,
		log:   log.New(log15.Ctx{"node": state.Home().Name()}),
	}, nil
}

func (i *DefaultBlockStorage) NewBlock(proposal Proposal) (Block, error) {
	// TODO store block

	i.log.Debug("new block created", "proposal", proposal)

	return Block{}, nil
}

func (i *DefaultBlockStorage) LatestBlock() (Block, error) {
	return Block{}, nil
}

func (i *DefaultBlockStorage) BlockByProposal(phash common.Hash) (Block, error) {
	return Block{}, nil
}
