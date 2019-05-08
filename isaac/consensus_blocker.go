package isaac

import (
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type ConsensusBlockerBlockingChanFunc func() (common.Seal, chan<- error)

type ConsensusBlocker struct {
	sync.RWMutex
	stopBlockingChan chan bool
	blockingChan     chan ConsensusBlockerBlockingChanFunc
	timer            *common.CallbackTimer

	policy           ConsensusPolicy
	state            *ConsensusState
	votingBox        VotingBox
	sealBroadcaster  SealBroadcaster
	sealPool         SealPool
	proposerSelector ProposerSelector
	blockStorage     BlockStorage
}

func NewConsensusBlocker(
	policy ConsensusPolicy,
	state *ConsensusState,
	votingBox VotingBox,
	sealBroadcaster SealBroadcaster,
	sealPool SealPool,
	proposerSelector ProposerSelector,
	blockStorage BlockStorage,
) *ConsensusBlocker {
	return &ConsensusBlocker{
		blockingChan:     make(chan ConsensusBlockerBlockingChanFunc),
		policy:           policy,
		state:            state,
		votingBox:        votingBox,
		sealBroadcaster:  sealBroadcaster,
		sealPool:         sealPool,
		proposerSelector: proposerSelector,
		blockStorage:     blockStorage,
	}
}

func (c *ConsensusBlocker) Start() error {
	if c.stopBlockingChan != nil {
		return common.StartStopperAlreadyStartedError
	}

	c.Lock()
	c.stopBlockingChan = make(chan bool)
	c.Unlock()

	go c.blocking()

	log.Debug("ConsensusBlocker started")

	return nil
}

func (c *ConsensusBlocker) Stop() error {
	c.Lock()
	defer c.Unlock()

	if c.stopBlockingChan == nil {
		return nil
	}

	c.stopBlockingChan <- true
	close(c.stopBlockingChan)
	c.stopBlockingChan = nil

	if c.timer != nil {
		c.timer.Stop()
	}

	log.Debug("ConsensusBlocker stopped")

	return nil
}

func (c *ConsensusBlocker) blocking() {
end:
	for {
		select {
		case <-c.stopBlockingChan:
			break end
		case f, notClosed := <-c.blockingChan:
			if !notClosed {
				continue
			}

			seal, errChan := f()

			err := c.vote(seal)
			if err != nil {
				log.Error("failed to vote", "error", err)
			}
			if errChan != nil {
				errChan <- err
			}
		}
	}
}

func (c *ConsensusBlocker) startTimer(keepRunning bool, callback func() error) error {
	c.Lock()
	defer c.Unlock()

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	c.timer = c.newTimer(callback, keepRunning)
	return c.timer.Start()
}

func (c *ConsensusBlocker) stopTimer() error {
	c.Lock()
	defer c.Unlock()

	if c.timer == nil {
		return nil
	}

	if err := c.timer.Stop(); err != nil {
		return err
	}
	c.timer = nil

	return nil
}

func (c *ConsensusBlocker) Vote(seal common.Seal, errChan chan<- error) {
	go func() {
		c.blockingChan <- func() (common.Seal, chan<- error) {
			return seal, errChan
		}
	}()
}

// vote votes and decides the next action.
func (c *ConsensusBlocker) vote(seal common.Seal) error {
	var votingResult VoteResultInfo
	{ // voting
		var err error

		switch seal.Type() {
		case ProposalSealType:
			var proposal Proposal
			if err := common.CheckSeal(seal, &proposal); err != nil {
				return err
			}

			votingResult, err = c.votingBox.Open(proposal)
		case BallotSealType:
			var ballot Ballot
			if err := common.CheckSeal(seal, &ballot); err != nil {
				return err
			}

			votingResult, err = c.votingBox.Vote(ballot)
		default:
			return common.InvalidSealTypeError
		}

		log.Debug("got voting result", "result", votingResult, "error", err)
		if err != nil {
			return err
		}

		if votingResult.NotYet() {
			return nil
		}
	}

	switch votingResult.Stage {
	case VoteStageINIT:
		if votingResult.Proposed {
			return c.doProposeAccepted(votingResult)
		}

		return c.runNewRound(votingResult.Height, votingResult.Round)
	case VoteStageSIGN:
		if votingResult.Result == VoteResultYES {
			return c.goToNextStage(
				votingResult.Proposal,
				votingResult.Height,
				votingResult.Round,
				votingResult.Stage.Next(),
			)
		}

		return c.startNewRound(votingResult.Height, votingResult.Round+1)
	case VoteStageACCEPT:
		return c.finishRound(votingResult.Proposal)
	}

	return nil
}

// doProposeAccepted will do,
// - validate proposal
// - decide YES/NOP
// - broadcast sign ballot
func (c *ConsensusBlocker) doProposeAccepted(votingResult VoteResultInfo) error {
	log.Debug("proposal accepted", "result", votingResult)

	err := c.startTimer(true, func() error {
		return c.broadcastINIT(votingResult.Height, votingResult.Round+1)
	})
	if err != nil {
		return err
	}

	// TODO validate proposal
	// TODO decide YES/NOP

	vote := VoteYES

	// broadcast sign ballot

	ballot := NewBallot(
		votingResult.Proposal,
		"",
		votingResult.Height,
		votingResult.Round,
		VoteStageSIGN,
		vote,
	)

	if err := c.sealBroadcaster.Send(&ballot); err != nil {
		return err
	}

	return nil
}

// goToNextStage goes to next stage
func (c *ConsensusBlocker) goToNextStage(
	phash common.Hash,
	height common.Big,
	round Round,
	stage VoteStage,
) error {
	log.Debug("go to next sage", "proposal", phash, "height", height, "round", round, "next", stage)

	err := c.startTimer(true, func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	// broadcast next stage ballot
	ballot := NewBallot(
		phash,
		"",
		height,
		round,
		stage,
		VoteYES,
	)

	if err := c.sealBroadcaster.Send(&ballot); err != nil {
		return err
	}

	return nil
}

// finishRound will do,
// - store block and state
// - update ConsensusBlockerState
// - ready to start new block
func (c *ConsensusBlocker) finishRound(phash common.Hash) error {
	log.Debug("finish round", "proposal", phash)

	seal, err := c.sealPool.Get(phash)
	if err != nil {
		return err
	}

	var proposal Proposal
	if err := common.CheckSeal(seal, &proposal); err != nil {
		return common.UnknownSealTypeError.SetMessage(err.Error())
	}

	// TODO store block and state
	if _, err := c.blockStorage.NewBlock(proposal); err != nil {
		return err
	}

	// update ConsensusBlockerState
	{
		prevState := *c.state

		if err := c.state.SetHeight(proposal.Block.Height.Inc()); err != nil {
			return err
		}
		if err := c.state.SetBlock(proposal.Block.Next); err != nil {
			return err
		}
		if err := c.state.SetState(proposal.State.Next); err != nil {
			return err
		}

		log.Debug(
			"round finished",
			"proposal", seal.Hash(),
			"old-block-height", prevState.Height().String(),
			"old-block-hash", prevState.Block(),
			"old-state-hash", prevState.State(),
			"new-block-height", c.state.Height().String(),
			"new-block-hash", c.state.Block(),
			"new-state-hash", c.state.State(),
		)
	}

	// propose or wait new proposal
	err = c.startTimer(true, func() error {
		return c.broadcastINIT(proposal.Block.Height.Inc(), Round(0))
	})
	if err != nil {
		return err
	}

	return nil
}

// startNewRound starts new round
func (c *ConsensusBlocker) startNewRound(height common.Big, round Round) error {
	log.Debug("start new round", "height", height, "round", round)

	err := c.startTimer(true, func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	return c.broadcastINIT(height, round)
}

// runNewRound starts new round and propose new proposal
func (c *ConsensusBlocker) runNewRound(height common.Big, round Round) error {
	log_ := log.New(log15.Ctx{"height": height, "round": round})
	log_.Debug("run new round")

	err := c.startTimer(true, func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	if !height.Equal(c.state.Height()) {
		log_.Debug("different height found", "init-height", height, "state-height", c.state.Height())
		return DifferentHeightConsensusError.AppendMessage(
			"init height=%v, state height=%v",
			height,
			c.state.Height(),
		)
	}

	delay := c.policy.AvgBlockRoundInterval
	log_.Debug("propose new proposal", "delay", delay)

	return c.propose(height, round, delay)
}

func (c *ConsensusBlocker) newTimer(callback func() error, keepRunning bool) *common.CallbackTimer {
	return common.NewCallbackTimer(
		"consensus_blocker_timer",
		c.policy.TimeoutWaitSeal,
		callback,
		keepRunning,
	)
}

func (c *ConsensusBlocker) broadcastINIT(height common.Big, round Round) error {
	log_ := log.New(log15.Ctx{"height": height, "round": round})
	log_.Debug("broadcast INIT ballot")

	proposer, err := c.proposerSelector.Select(c.state.Block(), height, round)
	if err != nil {
		return err
	}
	log_.Debug("proposer selected", "block", c.state.Block(), "proposer", proposer.Address())

	ballot := NewBallot(
		common.Hash{},
		proposer.Address(),
		height,
		round,
		VoteStageINIT,
		VoteYES,
	)

	// TODO self-signed ballot should not be needed to broadcast
	if err := c.sealBroadcaster.Send(&ballot); err != nil {
		return err
	}

	return nil
}

func (c *ConsensusBlocker) propose(height common.Big, round Round, delay time.Duration) error {
	log_ := log.New(log15.Ctx{"height": height, "round": round})

	proposer, err := c.proposerSelector.Select(c.state.Block(), height, round)
	if err != nil {
		return err
	}
	log_.Debug("proposer selected", "block", c.state.Block(), "proposer", proposer.Address())

	if !proposer.Equal(c.state.Home()) {
		log_.Debug("proposer is not home; will wait Proposal")
		return nil
	}

	// TODO validate transactions.
	proposal := NewProposal(
		round,
		ProposalBlock{
			Height:  height,
			Current: c.state.Block(),
			Next:    common.NewRandomHash("bk"),
		},
		ProposalState{
			Current: c.state.State(),
			Next:    []byte("next state"),
		},
		nil, // TODO transactions
	)

	go func(proposal Proposal, delay time.Duration) {
		if delay > 0 {
			<-time.After(delay)
		}

		err := c.sealBroadcaster.Send(&proposal)
		if err != nil {
			log_.Error("failed to broadcast", "proposal", proposal.Hash())
		}
		log_.Debug("proposal broadcasted", "proposal", proposal.Hash(), "delay", delay)
	}(proposal, delay)

	return nil
}
