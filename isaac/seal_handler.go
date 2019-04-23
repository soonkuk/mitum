package isaac

import (
	"github.com/inconshreveable/log15"
	"golang.org/x/sync/syncmap"

	"github.com/spikeekips/mitum/common"
)

type SealHandler interface {
	Receive(common.Seal) error
}

type ISAACSealHandler struct {
	seals  *syncmap.Map // TODO should be stored in persistent storage
	voting *RoundVotingManager
}

func NewISAACSealHandler() *ISAACSealHandler {
	return &ISAACSealHandler{
		seals:  &syncmap.Map{},
		voting: NewRoundVotingManager(),
	}
}

func (i *ISAACSealHandler) Receive(seal common.Seal) error {
	// NOTE seal should be checked well-formed already

	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"seal-hash": sealHash, "type": seal.Type})
	log_.Debug("got new seal")

	if _, found := i.seals.Load(sealHash); found {
		log_.Debug("already received; it will be ignored")
	}

	i.seals.Store(sealHash, seal)

	switch seal.Type {
	case ProposeBallotSealType:
		vp, err := i.voting.NewRound(seal)
		if err != nil {
			return err
		}
		log_.Debug("starting new round", "voting-proposal", vp)
	case VoteBallotSealType:
	case TransactionSealType:
	}

	return nil
}
