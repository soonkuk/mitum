package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type stageTransitFunc func() (VoteStage, common.Hash, Vote, chan<- error)

type RoundBoy interface {
	common.StartStopper
	Transit(VoteStage, common.Hash /* Seal(Propose).Hash() */, Vote) error
	Channel() chan stageTransitFunc
}

type DefaultRoundBoy struct {
	sync.RWMutex
	round       Round
	homeNode    common.HomeNode
	broadcaster SealBroadcaster
	channel     chan stageTransitFunc
	stopChan    chan bool
}

func NewDefaultRoundBoy(
	homeNode common.HomeNode,
) (*DefaultRoundBoy, error) {
	return &DefaultRoundBoy{
		homeNode: homeNode,
		stopChan: make(chan bool),
		channel:  make(chan stageTransitFunc),
	}, nil
}

func (i *DefaultRoundBoy) Start() error {
	go i.schedule()

	return nil
}

func (i *DefaultRoundBoy) Stop() error {
	if i.stopChan != nil {
		i.stopChan <- true
		close(i.stopChan)
		i.stopChan = nil
	}

	return nil
}

func (i *DefaultRoundBoy) Channel() chan stageTransitFunc {
	return i.channel
}

func (i *DefaultRoundBoy) SetBroadcaster(broadcaster SealBroadcaster) error {
	i.Lock()
	defer i.Unlock()

	i.broadcaster = broadcaster

	return nil
}

func (i *DefaultRoundBoy) schedule() {
end:
	for {
		select {
		case <-i.stopChan:
			break end
		case f := <-i.channel: // one stage at a time
			stage, psHash, vote, errChan := f()
			err := i.transit(stage, psHash, vote)
			if err != nil {
				log.Error("failed to transit", "error", err, "psHash", psHash, "stage", stage, "vote", vote)
			}

			errChan <- err
		}
	}
}

func (i *DefaultRoundBoy) Transit(stage VoteStage, psHash common.Hash, vote Vote) error {
	errChan := make(chan error)
	defer close(errChan)

	go func() {
		i.channel <- func() (VoteStage, common.Hash, Vote, chan<- error) {
			return stage, psHash, vote, errChan
		}
	}()

	return <-errChan
}

func (i *DefaultRoundBoy) transit(stage VoteStage, psHash common.Hash, vote Vote) error {
	log_ := log.New(log15.Ctx{"from": stage, "next": stage.Next(), "psHash": psHash})
	log_.Debug("stage transitted")

	switch stage {
	case VoteStageINIT:
		return i.transitToSIGN(psHash, vote)
	case VoteStageSIGN:
		return i.transitToACCEPT(psHash, vote)
	case VoteStageACCEPT:
		log_.Debug("consensus reached ALLCONFIRM")
		return i.transitToALLCONFIRM(psHash, vote)
	default:
		log_.Error("trying to weired stage")
	}

	return nil
}

func (i *DefaultRoundBoy) transitToSIGN(psHash common.Hash, vote Vote) error {
	/*
		ballot, err := NewBallot(psHash, i.homeNode.Address(), VoteStageSIGN, vote)
		if err != nil {
			return err
		}
	*/
	ballot := Ballot{}

	log.Debug("new Ballot will be broadcasted")

	if err := i.broadcaster.Send(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *DefaultRoundBoy) transitToACCEPT(psHash common.Hash, vote Vote) error {
	/*
		ballot, err := NewBallot(psHash, i.homeNode.Address(), VoteStageACCEPT, vote)
		if err != nil {
			return err
		}
	*/

	ballot := Ballot{}

	log.Debug("new Ballot will be broadcasted")

	if err := i.broadcaster.Send(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *DefaultRoundBoy) transitToALLCONFIRM(psHash common.Hash, _ Vote) error {
	i.Lock()
	defer i.Unlock()

	i.round = Round(0)

	return nil
}
