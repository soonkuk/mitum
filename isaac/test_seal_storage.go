// +build test

package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/seal"
	"golang.org/x/sync/syncmap"
	"golang.org/x/xerrors"
)

type TSealStorage struct {
	m *syncmap.Map
}

func NewTSealStorage() *TSealStorage {
	return &TSealStorage{
		m: &syncmap.Map{},
	}
}

func (tss *TSealStorage) Has(h hash.Hash) bool {
	_, found := tss.m.Load(h)
	return found
}

func (tss *TSealStorage) Get(h hash.Hash) seal.Seal {
	if s, found := tss.m.Load(h); !found {
		return nil
	} else if sl, ok := s.(seal.Seal); !ok {
		return nil
	} else {
		return sl
	}
}

func (tss *TSealStorage) Save(sl seal.Seal) error {
	if sl == nil {
		return xerrors.Errorf("seal should not be nil")
	}

	if tss.Has(sl.Hash()) {
		return xerrors.Errorf("already stored; %v", sl.Hash())
	}

	tss.m.Store(sl.Hash(), sl)

	return nil
}
