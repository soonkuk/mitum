package isaac

import (
	"github.com/inconshreveable/log15"
	"golang.org/x/sync/syncmap"

	"github.com/spikeekips/mitum/common"
)

type SealPool interface {
	Exists(common.Hash) bool
	Get(common.Hash) (common.Seal, error)
	Add(common.Seal) error
}

type DefaultSealPool struct {
	home  *common.HomeNode
	log   log15.Logger
	seals *syncmap.Map // TODO should be stored in persistent storage
}

func NewDefaultSealPool(home *common.HomeNode) *DefaultSealPool {
	return &DefaultSealPool{
		seals: &syncmap.Map{},
		log:   log.New(log15.Ctx{"node": home.Name()}),
	}
}

func (s *DefaultSealPool) Exists(hash common.Hash) bool {
	_, found := s.seals.Load(hash)
	return found
}

func (s *DefaultSealPool) Get(hash common.Hash) (common.Seal, error) {
	n, found := s.seals.Load(hash)
	if !found {
		return nil, SealNotFoundError
	}

	return n.(common.Seal), nil
}

func (s *DefaultSealPool) Add(seal common.Seal) error {
	// NOTE seal should be checked well-formed already

	log_ := s.log.New(log15.Ctx{"seal": seal.Hash(), "type": seal.Type()})
	if _, found := s.seals.Load(seal.Hash()); found {
		log_.Debug("already received; it will be ignored")
		return KnownSealFoundError
	}

	s.seals.Store(seal.Hash(), seal)
	log_.Debug("seal added")

	return nil
}
