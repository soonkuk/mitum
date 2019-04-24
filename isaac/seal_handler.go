package isaac

import (
	"github.com/inconshreveable/log15"
	"golang.org/x/sync/syncmap"

	"github.com/spikeekips/mitum/common"
)

type SealHandler interface {
	Add(common.Seal) error
}

type ISAACSealHandler struct {
	seals *syncmap.Map // TODO should be stored in persistent storage
}

func NewISAACSealHandler() *ISAACSealHandler {
	return &ISAACSealHandler{
		seals: &syncmap.Map{},
	}
}

func (s *ISAACSealHandler) Add(seal common.Seal) error {
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
