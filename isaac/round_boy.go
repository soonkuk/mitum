package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type stageTransitFunc func() (VoteStage, common.Hash, Vote)

type RoundBoy interface {
	common.StartStopper
	Transit(VoteStage, common.Hash /* Seal(Propose).Hash() */, Vote)
}

type ISAACRoundBoy struct {
	sync.RWMutex
	round       Round
	homeNode    common.HomeNode
	broadcaster SealBroadcaster
	channel     chan stageTransitFunc
	stopChan    chan bool
}

func NewISAACRoundBoy(
	homeNode common.HomeNode,
) (*ISAACRoundBoy, error) {
	return &ISAACRoundBoy{
		homeNode: homeNode,
		stopChan: make(chan bool),
		channel:  make(chan stageTransitFunc),
	}, nil
}

func (i *ISAACRoundBoy) Start() error {
	go i.schedule()

	return nil
}

func (i *ISAACRoundBoy) Stop() error {
	if i.stopChan != nil {
		i.stopChan <- true
		close(i.stopChan)
		i.stopChan = nil
	}

	return nil
}

func (i *ISAACRoundBoy) SetBroadcaster(broadcaster SealBroadcaster) error {
	i.Lock()
	defer i.Unlock()

	i.broadcaster = broadcaster

	return nil
}

func (i *ISAACRoundBoy) schedule() {
end:
	for {
		select {
		case <-i.stopChan:
			break end
		case f := <-i.channel: // one stage at a time
			stage, psHash, vote := f()
			if err := i.transit(stage, psHash, vote); err != nil {
				log.Error("failed to transit", "error", err, "psHash", psHash, "stage", stage, "vote", vote)
			}
		}
	}
}

func (i *ISAACRoundBoy) Transit(stage VoteStage, psHash common.Hash, vote Vote) {
	go func() {
		i.channel <- func() (VoteStage, common.Hash, Vote) {
			return stage, psHash, vote
		}
	}()
}

func (i *ISAACRoundBoy) transit(stage VoteStage, psHash common.Hash, vote Vote) error {
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

func (i *ISAACRoundBoy) transitToSIGN(psHash common.Hash, vote Vote) error {
	ballot, err := NewBallot(psHash, i.homeNode.Address(), VoteStageSIGN, vote)
	if err != nil {
		return err
	}

	log.Debug("new Ballot will be broadcasted")

	if err := i.broadcaster.Send(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *ISAACRoundBoy) transitToACCEPT(psHash common.Hash, vote Vote) error {
	ballot, err := NewBallot(psHash, i.homeNode.Address(), VoteStageACCEPT, vote)
	if err != nil {
		return err
	}

	log.Debug("new Ballot will be broadcasted")

	if err := i.broadcaster.Send(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *ISAACRoundBoy) transitToALLCONFIRM(psHash common.Hash, _ Vote) error {
	i.Lock()
	defer i.Unlock()

	i.round = Round(0)

	return nil
}
