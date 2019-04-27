package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type StageTransistor interface {
	common.StartStopper
	Transit( /* Seal(Propose).Hash() */ common.Hash, VoteStage, common.Seal, Vote) error
}

type ISAACStageTransistor struct {
	sync.RWMutex
	policy         ConsensusPolicy
	state          *ConsensusState
	sealPool       SealPool
	voting         *RoundVoting
	sender         func(common.Node, common.Seal) error
	stopChan       chan bool
	allconfirmChan chan common.Seal
}

func NewISAACStageTransistor(
	policy ConsensusPolicy,
	state *ConsensusState,
	sealPool SealPool,
	voting *RoundVoting,
) (*ISAACStageTransistor, error) {
	return &ISAACStageTransistor{
		policy:         policy,
		state:          state,
		sealPool:       sealPool,
		voting:         voting,
		stopChan:       make(chan bool),
		allconfirmChan: make(chan common.Seal),
	}, nil
}

func (i *ISAACStageTransistor) SetSender(sender func(common.Node, common.Seal) error) error {
	i.Lock()
	defer i.Unlock()

	i.sender = sender

	return nil
}

func (i *ISAACStageTransistor) Start() error {
	go i.schedule()

	return nil
}

func (i *ISAACStageTransistor) Stop() error {
	if i.stopChan != nil {
		i.stopChan <- true
		close(i.stopChan)
		i.stopChan = nil
	}

	return nil
}

func (i *ISAACStageTransistor) schedule() {
end:
	for {
		select {
		case <-i.stopChan:
			break end
		case seal := <-i.allconfirmChan:
			if err := i.doALLCONFIRM(seal); err != nil {
				log.Error("failed to ALLCONFIRM", "error", err)
			}
		}
	}
}

func (i *ISAACStageTransistor) broadcast(sealType common.SealType, body common.Hasher, excludes ...common.Address) error {
	seal, err := common.NewSeal(sealType, body)
	if err != nil {
		return err
	}

	if err := seal.Sign(i.policy.NetworkID, i.state.Node().Seed()); err != nil {
		return err
	}

	sHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"seal": sHash})
	log_.Debug("seal will be broadcasted")

	var targets = []common.Node{i.state.Node()}
	for _, node := range i.state.Node().Validators() {
		var exclude bool
		for _, a := range excludes {
			if a == node.Address() {
				exclude = true
				break
			}
		}
		if exclude {
			continue
		}
		targets = append(targets, node)
	}

	for _, node := range targets {
		if err := i.sender(node, seal); err != nil {
			return err
		}
	}

	return nil
}

func (i *ISAACStageTransistor) Transit(psHash common.Hash, stage VoteStage, seal common.Seal, vote Vote) error {
	sHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"to": stage, "seal": sHash, "psHash": psHash})
	log_.Debug("stage transitted")

	if err := i.voting.Agreed(psHash); err != nil {
		return err
	}

	switch stage {
	case VoteStageSIGN:
		return i.transitToSIGN(seal, vote)
	case VoteStageACCEPT:
		return i.transitToACCEPT(seal, vote)
	case VoteStageALLCONFIRM:
		log_.Debug("consensus reached ALLCONFIRM")
		return i.transitToALLCONFIRM(seal, vote)
	default:
		log_.Error("trying to weired stage")
	}

	return nil
}

func (i *ISAACStageTransistor) transitToSIGN(seal common.Seal, vote Vote) error {
	if seal.Type != ProposeSealType {
		log.Error("sign ballot should be created from Propose", "seal-type", seal.Type)
		return InvalidSealTypeError
	}

	sHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	ballot, err := NewBallot(sHash, i.state.Node().Address(), vote)
	if err != nil {
		return err
	}

	if err := i.broadcast(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *ISAACStageTransistor) transitToACCEPT(seal common.Seal, vote Vote) error {
	if seal.Type != BallotSealType {
		log.Error("accept ballot should be created from Ballot", "seal-type", seal.Type, "seal", seal)
		return InvalidSealTypeError
	}

	var ballot Ballot
	if err := seal.UnmarshalBody(&ballot); err != nil {
		return err
	}

	ballot, err := ballot.NewForStage(VoteStageACCEPT, i.state.Node().Address(), vote)
	if err != nil {
		return err
	}

	log.Debug("new Ballot will be broadcasted", "new-ballot", ballot)

	if err := i.broadcast(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *ISAACStageTransistor) transitToALLCONFIRM(seal common.Seal, vote Vote) error {
	go func() {
		i.allconfirmChan <- seal
	}()

	return nil
}

func (i *ISAACStageTransistor) doALLCONFIRM(seal common.Seal) error {
	if seal.Type != BallotSealType {
		log.Error("accept ballot should be created from Ballot", "seal-type", seal.Type)
		return InvalidSealTypeError
	}

	var ballot Ballot
	if err := seal.UnmarshalBody(&ballot); err != nil {
		return err
	}

	proposeSeal, err := i.sealPool.Get(ballot.ProposeSeal)
	if err != nil {
		return err
	}

	var propose Propose
	if err := proposeSeal.UnmarshalBody(&propose); err != nil {
		return err
	}

	// finish voting
	psHash, _, err := proposeSeal.Hash()
	if err != nil {
		return err
	}

	if err := i.voting.Finish(psHash); err != nil {
		log.Error("failed to finish voting", "error", err)
		return err
	}

	// TODO store block

	// update state
	prevState := *i.state

	i.state.SetHeight(propose.Block.Height.Inc())
	i.state.SetBlock(propose.Block.Next)
	i.state.SetState(propose.State.Next)
	i.state.SetRound(Round(0))

	log.Debug(
		"allConfirmed",
		"psHash", psHash,
		"old-block-height", prevState.Height(),
		"old-block-hash", prevState.Block(),
		"old-state-hash", prevState.State(),
		"old-round", prevState.Round(),
		"new-block-height", i.state.Height().String(),
		"new-block-hash", i.state.Block(),
		"new-state-hash", i.state.State(),
		"new-round", i.state.Round(),
	)

	return nil
}

func (i *ISAACStageTransistor) nextRound(seal common.Seal) error {
	return nil
}

func (i *ISAACStageTransistor) nextBlock(seal common.Seal) error {
	return nil
}
