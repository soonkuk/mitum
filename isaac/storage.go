package isaac

import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/storage"
)

// TODO state should be considered
type BlockStorage interface {
	Storage() storage.Storage
	NewBlock(Proposal) (Block, storage.Batch, error)
	LatestBlock() (Block, error)
	BlockByProposal(common.Hash) (Block, error)
}

type DefaultBlockStorage struct {
	*common.Logger
	st storage.Storage
}

func NewDefaultBlockStorage(st storage.Storage) (*DefaultBlockStorage, error) {
	return &DefaultBlockStorage{
		Logger: common.NewLogger(log, "module", "block-storage"),
		st:     st,
	}, nil
}

func (d *DefaultBlockStorage) Storage() storage.Storage {
	return d.st
}

func (d *DefaultBlockStorage) NewBlock(proposal Proposal) (Block, storage.Batch, error) {
	// TODO store block with Batch

	log_ := d.Log().New(log15.Ctx{"proposal": proposal.Hash()})

	log_.Debug("new block will be prepared")
	batch := d.st.Batch()

	block, err := NewBlockFromProposal(proposal)
	if err != nil {
		return Block{}, nil, err
	}

	bytes, err := block.MarshalBinary()
	if err != nil {
		return Block{}, nil, err
	}

	// TODO needs storage key
	batch.Put(block.Hash().Bytes(), bytes)

	log_.Debug("new block prepared", "block", block)

	return block, batch, nil
}

func (d *DefaultBlockStorage) LatestBlock() (Block, error) {
	return Block{}, nil
}

func (d *DefaultBlockStorage) BlockByProposal(proposal common.Hash) (Block, error) {
	return Block{}, nil
}
