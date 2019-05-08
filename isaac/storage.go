package isaac

// TODO state should be considered
type BlockStorage interface {
	NewBlock(Proposal) (Block, error)
	LatestBlock() (Block, error)
}

type DefaultBlockStorage struct {
	state *ConsensusState
}

func NewDefaultBlockStorage(state *ConsensusState) (*DefaultBlockStorage, error) {
	return &DefaultBlockStorage{
		state: state,
	}, nil
}

func (i *DefaultBlockStorage) NewBlock(proposal Proposal) (Block, error) {
	// TODO store block

	log.Debug("new block created", "proposal", proposal)

	return Block{}, nil
}

func (i *DefaultBlockStorage) LatestBlock() (Block, error) {
	return Block{}, nil
}
