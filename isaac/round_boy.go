package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type stageTransitFunc func() (VoteStage, common.Seal, Vote)

type RoundBoy interface {
	common.StartStopper
	Transit(VoteStage, common.Seal, Vote)
}

type ISAACRoundBoy struct {
	sync.RWMutex
	policy   ConsensusPolicy
	state    *ConsensusState
	sealPool SealPool
	voting   *RoundVoting
	sender   func(common.Node, common.Seal) error
	channel  chan stageTransitFunc
	stopChan chan bool
}

func NewISAACRoundBoy(
	policy ConsensusPolicy,
	state *ConsensusState,
	sealPool SealPool,
	voting *RoundVoting,
) (*ISAACRoundBoy, error) {
	return &ISAACRoundBoy{
		policy:   policy,
		state:    state,
		sealPool: sealPool,
		voting:   voting,
		stopChan: make(chan bool),
		channel:  make(chan stageTransitFunc),
	}, nil
}

func (i *ISAACRoundBoy) SetSender(sender func(common.Node, common.Seal) error) error {
	i.Lock()
	defer i.Unlock()

	i.sender = sender

	return nil
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

func (i *ISAACRoundBoy) schedule() {
end:
	for {
		select {
		case <-i.stopChan:
			break end
		case f := <-i.channel:
			stage, seal, vote := f()
			if err := i.transit(stage, seal, vote); err != nil {
				log.Error("failed to transit", "error", err, "stage", stage, "vote", vote)
			}
		}
	}
}

func (i *ISAACRoundBoy) broadcast(sealType common.SealType, body common.Hasher, excludes ...common.Address) error {
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

func (i *ISAACRoundBoy) Transit(stage VoteStage, seal common.Seal, vote Vote) {
	go func() {
		i.channel <- func() (VoteStage, common.Seal, Vote) {
			return stage, seal, vote
		}
	}()
}

func (i *ISAACRoundBoy) transit(stage VoteStage, seal common.Seal, vote Vote) error {
	sHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"to": stage, "seal": sHash})
	log_.Debug("stage transitted")

	var psHash common.Hash
	switch seal.Type {
	case ProposeSealType:
		if stage != VoteStageSIGN {
			log_.Error("sign ballot should be created from Propose")
			return InvalidSealTypeError
		}

		if sHash, _, err := seal.Hash(); err != nil {
			return err
		} else {
			psHash = sHash
		}

		if err := i.transitToSIGN(psHash, vote); err != nil {
			return err
		}
	case BallotSealType:
		var ballot Ballot
		if err := seal.UnmarshalBody(&ballot); err != nil {
			return err
		}
		psHash = ballot.ProposeSeal

		switch stage {
		case VoteStageACCEPT:
			return i.transitToACCEPT(ballot, vote)
		case VoteStageALLCONFIRM:
			log_.Debug("consensus reached ALLCONFIRM")
			return i.transitToALLCONFIRM(ballot, vote)
		default:
			log_.Error("trying to weired stage")
		}
	default:
		return InvalidSealTypeError
	}

	if err := i.voting.Agreed(psHash); err != nil {
		return err
	}

	return nil
}

func (i *ISAACRoundBoy) transitToSIGN(psHash common.Hash, vote Vote) error {
	ballot, err := NewBallot(psHash, i.state.Node().Address(), vote)
	if err != nil {
		return err
	}

	if err := i.broadcast(BallotSealType, ballot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *ISAACRoundBoy) transitToACCEPT(ballot Ballot, vote Vote) error {
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

func (i *ISAACRoundBoy) transitToALLCONFIRM(ballot Ballot, _ Vote) error {
	psHash := ballot.ProposeSeal
	proposeSeal, err := i.sealPool.Get(psHash)
	if err != nil {
		return err
	}

	var propose Propose
	if err := proposeSeal.UnmarshalBody(&propose); err != nil {
		return err
	}

	// finish voting
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

func (i *ISAACRoundBoy) nextRound(seal common.Seal) error {
	return nil
}

func (i *ISAACRoundBoy) nextBlock(seal common.Seal) error {
	return nil
}
