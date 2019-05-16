package isaac

import (
	"github.com/spikeekips/mitum/common"
)

// TODO state should be considered
type BlockStorage interface {
	NewBlock(Proposal) (Block, error)
	LatestBlock() (Block, error)
	BlockByProposal(common.Hash) (Block, error)
}

type DefaultBlockStorage struct {
	*common.Logger
}

func NewDefaultBlockStorage() (*DefaultBlockStorage, error) {
	return &DefaultBlockStorage{
		Logger: common.NewLogger(log),
	}, nil
}

func (d *DefaultBlockStorage) NewBlock(proposal Proposal) (Block, error) {
	// TODO store block

	d.Log().Debug("new block created", "proposal", proposal)

	return Block{}, nil
}

func (d *DefaultBlockStorage) LatestBlock() (Block, error) {
	return Block{}, nil
}

func (d *DefaultBlockStorage) BlockByProposal(proposal common.Hash) (Block, error) {
	return Block{}, nil
}
