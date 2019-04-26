package isaac

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type ConsensusState struct {
	sync.RWMutex
	node   common.HomeNode
	height common.Big  // last Block.Height
	block  common.Hash // Block.Hash()
	state  []byte      // last State.Root.Hash()
	round  Round       // currently running round
}

func (c ConsensusState) String() string {
	b, _ := json.Marshal(map[string]interface{}{
		"node":   c.node,
		"height": c.height,
		"block":  c.block,
		"state":  c.state,
		"round":  c.round,
	})
	return strings.Replace(string(b), "\"", "'", -1)
}

func (c *ConsensusState) Node() common.HomeNode {
	c.RLock()
	defer c.RUnlock()

	return c.node
}

func (c *ConsensusState) Height() common.Big {
	c.RLock()
	defer c.RUnlock()

	return c.height
}

func (c *ConsensusState) SetHeight(height common.Big) {
	c.Lock()
	defer c.Unlock()

	c.height = height
}

func (c *ConsensusState) Block() common.Hash {
	c.RLock()
	defer c.RUnlock()

	return c.block
}

func (c *ConsensusState) SetBlock(block common.Hash) {
	c.Lock()
	defer c.Unlock()

	c.block = block
}

func (c *ConsensusState) State() []byte {
	c.RLock()
	defer c.RUnlock()

	return c.state
}

func (c *ConsensusState) SetState(state []byte) {
	c.Lock()
	defer c.Unlock()

	c.state = state
}

func (c *ConsensusState) Round() Round {
	c.RLock()
	defer c.RUnlock()

	return c.round
}

func (c *ConsensusState) SetRound(round Round) {
	c.Lock()
	defer c.Unlock()

	c.round = round
}

type StageTransistor interface {
	common.StartStopper
	Transit( /* proposeBallotSealHash */ common.Hash, VoteStage, common.Seal, Vote) error
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

	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"seal": sealHash})
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

func (i *ISAACStageTransistor) Transit(proposeBallotSealHash common.Hash, stage VoteStage, seal common.Seal, vote Vote) error {
	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"to": stage, "seal": sealHash})
	log_.Debug("stage transitted")

	if err := i.voting.Agreed(proposeBallotSealHash); err != nil {
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
	if seal.Type != ProposeBallotSealType {
		log.Error("sign voteBallot should be created from ProposeBallot", "seal-type", seal.Type)
		return InvalidSealTypeError
	}

	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	voteBallot, err := NewVoteBallot(sealHash, i.state.Node().Address(), vote)
	if err != nil {
		return err
	}

	if err := i.broadcast(VoteBallotSealType, voteBallot); err != nil {
		log.Error("failed to broadcast", "error", err)
		return err
	}

	return nil
}

func (i *ISAACStageTransistor) transitToACCEPT(seal common.Seal, vote Vote) error {
	if seal.Type != VoteBallotSealType {
		log.Error("accept voteBallot should be created from VoteBallot", "seal-type", seal.Type, "seal", seal)
		return InvalidSealTypeError
	}

	var voteBallot VoteBallot
	if err := seal.UnmarshalBody(&voteBallot); err != nil {
		return err
	}

	ballot, err := voteBallot.NewForStage(VoteStageACCEPT, i.state.Node().Address(), vote)
	if err != nil {
		return err
	}

	log.Debug("new VoteBallot will be broadcasted", "new-ballot", ballot)

	if err := i.broadcast(VoteBallotSealType, ballot); err != nil {
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
	if seal.Type != VoteBallotSealType {
		log.Error("accept voteBallot should be created from VoteBallot", "seal-type", seal.Type)
		return InvalidSealTypeError
	}

	var voteBallot VoteBallot
	if err := seal.UnmarshalBody(&voteBallot); err != nil {
		return err
	}

	proposeSeal, err := i.sealPool.Get(voteBallot.ProposeBallotSeal)
	if err != nil {
		return err
	}

	var proposeBallot ProposeBallot
	if err := proposeSeal.UnmarshalBody(&proposeBallot); err != nil {
		return err
	}

	// finish voting
	sealHash, _, err := proposeSeal.Hash()
	if err != nil {
		return err
	}

	if err := i.voting.Finish(sealHash); err != nil {
		log.Error("failed to finish voting", "error", err)
		return err
	}

	// TODO store block

	// update state
	prevState := *i.state

	i.state.SetHeight(proposeBallot.Block.Height.Inc())
	i.state.SetBlock(proposeBallot.Block.Next)
	i.state.SetState(proposeBallot.State.Next)
	i.state.SetRound(Round(0))

	log.Debug(
		"allConfirmed",
		"proposeBallotSeal", sealHash,
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
