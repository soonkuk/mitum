package common

import (
	"fmt"
	"sync"

	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/seal"
	"github.com/spikeekips/mitum/valuehash"
)

type MapSealStorage struct {
	sm        *sync.Map
	proposals *sync.Map
}

func NewMapSealStorage() *MapSealStorage {
	return &MapSealStorage{
		sm:        &sync.Map{},
		proposals: &sync.Map{},
	}
}

func (ss *MapSealStorage) proposalKey(height isaac.Height, round isaac.Round) string {
	return fmt.Sprintf("%d-%d", height, round)
}

func (ss *MapSealStorage) Add(sl seal.Seal) error {
	ss.sm.Store(sl.Hash(), sl)

	if proposal, ok := sl.(isaac.Proposal); ok {
		ss.proposals.Store(ss.proposalKey(proposal.Height(), proposal.Round()), proposal.Hash())
	}

	return nil
}

func (ss *MapSealStorage) Delete(sh valuehash.Hash) error {
	if found, err := ss.Exists(sh); err != nil {
		return err
	} else if !found {
		return nil
	} else if sl, found, err := ss.Seal(sh); err != nil {
		return err
	} else if !found {
		return nil
	} else if proposal, ok := sl.(isaac.Proposal); ok {
		ss.proposals.Delete(ss.proposalKey(proposal.Height(), proposal.Round()))
	}

	ss.sm.Delete(sh)

	return nil
}

func (ss *MapSealStorage) Exists(sh valuehash.Hash) (bool, error) {
	_, found := ss.sm.Load(sh)
	return found, nil
}

func (ss *MapSealStorage) Seal(sh valuehash.Hash) (seal.Seal, bool, error) {
	i, found := ss.sm.Load(sh)
	if !found {
		return nil, false, nil
	}

	return i.(seal.Seal), true, nil
}

func (ss *MapSealStorage) Proposal(height isaac.Height, round isaac.Round) (isaac.Proposal, bool, error) {
	ph, found := ss.proposals.Load(ss.proposalKey(height, round))
	if !found {
		return nil, false, nil
	}

	sl, found, err := ss.Seal(ph.(valuehash.Hash))
	if err != nil || !found {
		return nil, found, err
	}

	return sl.(isaac.Proposal), true, nil
}