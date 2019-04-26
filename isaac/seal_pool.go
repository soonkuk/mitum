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

type ISAACSealPool struct {
	seals *syncmap.Map // TODO should be stored in persistent storage
}

func NewISAACSealPool() *ISAACSealPool {
	return &ISAACSealPool{
		seals: &syncmap.Map{},
	}
}

func (s *ISAACSealPool) Exists(sealHash common.Hash) bool {
	_, found := s.seals.Load(sealHash)
	return found
}

func (s *ISAACSealPool) Get(sealHash common.Hash) (common.Seal, error) {
	n, found := s.seals.Load(sealHash)
	if !found {
		return common.Seal{}, SealNotFoundError
	}

	return n.(common.Seal), nil
}

func (s *ISAACSealPool) Add(seal common.Seal) error {
	// NOTE seal should be checked well-formed already

	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"seal-hash": sealHash, "type": seal.Type})
	if _, found := s.seals.Load(sealHash); found {
		log_.Debug("already received; it will be ignored")
		return KnownSealFoundError
	}

	s.seals.Store(sealHash, seal)
	log_.Debug("seal added")

	return nil
}
