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
	policy           ConsensusPolicy
	state            *ConsensusState
	votingBox        VotingBox
	sealBroadcaster  SealBroadcaster
	sealPool         SealPool
	timer            *ConsensusBlockerTimer
}

func NewConsensusBlocker(
	policy ConsensusPolicy,
	state *ConsensusState,
	votingBox VotingBox,
	sealBroadcaster SealBroadcaster,
	sealPool SealPool,
) *ConsensusBlocker {
	return &ConsensusBlocker{
		blockingChan:    make(chan ConsensusBlockerBlockingChanFunc),
		policy:          policy,
		state:           state,
		votingBox:       votingBox,
		sealBroadcaster: sealBroadcaster,
		sealPool:        sealPool,
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

func (c *ConsensusBlocker) startTimer(callback func() error) error {
	c.Lock()
	defer c.Unlock()

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	c.timer = c.newTimer(callback)
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

		switch seal.Type {
		case ProposeSealType:
			votingResult, err = c.votingBox.Open(seal)
		case BallotSealType:
			votingResult, err = c.votingBox.Vote(seal)
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
// - validate propose
// - decide YES/NOP
// - broadcast sign ballot
func (c *ConsensusBlocker) doProposeAccepted(votingResult VoteResultInfo) error {
	log.Debug("proposal accepted", "result", votingResult)

	err := c.startTimer(func() error {
		return c.broadcastINIT(votingResult.Height, votingResult.Round+1)
	})
	if err != nil {
		return err
	}

	// TODO validate propose
	// TODO decide YES/NOP

	vote := VoteYES

	// broadcast sign ballot

	ballot, err := NewBallot(
		votingResult.Proposal,
		c.state.Home().Address(),
		votingResult.Height,
		votingResult.Round,
		VoteStageSIGN,
		vote,
	)
	if err != nil {
		return err
	}

	if err := c.sealBroadcaster.Send(BallotSealType, ballot); err != nil {
		return err
	}

	return nil
}

// goToNextStage goes to next stage
func (c *ConsensusBlocker) goToNextStage(
	proposal common.Hash,
	height common.Big,
	round Round,
	stage VoteStage,
) error {
	log.Debug("go to next sage", "proposal", proposal, "height", height, "round", round, "next", stage)

	err := c.startTimer(func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	// broadcast next stage ballot
	ballot, err := NewBallot(
		proposal,
		c.state.Home().Address(),
		height,
		round,
		stage,
		VoteYES,
	)
	if err != nil {
		return err
	}

	if err := c.sealBroadcaster.Send(BallotSealType, ballot); err != nil {
		return err
	}

	return nil
}

// finishRound will do,
// - store block and state
// - update ConsensusBlockerState
// - ready to start new block
func (c *ConsensusBlocker) finishRound(proposal common.Hash) error {
	log.Debug("finish round", "proposal", proposal)

	seal, err := c.sealPool.Get(proposal)
	if err != nil {
		return err
	}

	var propose Propose
	if err := seal.UnmarshalBody(&propose); err != nil {
		return err
	}

	// TODO store block and state

	// update ConsensusBlockerState
	c.state.SetHeight(propose.Block.Height.Inc())
	c.state.SetBlock(propose.Block.Next)
	c.state.SetState(propose.State.Next)

	// propose or wait new proposal
	err = c.startTimer(func() error {
		return c.broadcastINIT(propose.Block.Height.Inc(), Round(0))
	})
	if err != nil {
		return err
	}

	return nil
}

// startNewRound starts new round
func (c *ConsensusBlocker) startNewRound(height common.Big, round Round) error {
	log.Debug("start new round", "height", height, "round", round)

	err := c.startTimer(func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	return c.broadcastINIT(height, round)
}

// runNewRound starts new round and propose new proposal
func (c *ConsensusBlocker) runNewRound(height common.Big, round Round) error {
	log.Debug("run new round", "height", height, "round", round)

	err := c.startTimer(func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	// TODO broadcast propsal
	log.Debug("propose new proposal")

	return nil
}

func (c *ConsensusBlocker) newTimer(callback func() error) *ConsensusBlockerTimer {
	return &ConsensusBlockerTimer{
		timeout:  c.policy.TimeoutWaitSeal,
		callback: callback,
		log: log.New(log15.Ctx{
			"timeout": c.policy.TimeoutWaitSeal,
		}),
	}
}

func (c *ConsensusBlocker) broadcastINIT(height common.Big, round Round) error {
	log.Debug("expired; we go to next round", "next", round)

	ballot, err := NewBallot(
		common.Hash{},
		c.state.Home().Address(),
		height,
		round,
		VoteStageINIT,
		VoteYES,
	)
	if err != nil {
		return err
	}

	// TODO Proposer should be selected
	ballot.Proposer = c.state.Home().Address()

	// TODO self-signed ballot should not be needed to broadcast
	if err := c.sealBroadcaster.Send(BallotSealType, ballot); err != nil {
		return err
	}

	return nil
}

type ConsensusBlockerTimer struct {
	timeout  time.Duration
	stopChan chan bool
	log      log15.Logger
	callback func() error
}

func (c *ConsensusBlockerTimer) Start() error {
	if c.stopChan != nil {
		return common.StartStopperAlreadyStartedError
	}

	c.stopChan = make(chan bool)

	go c.waiting()

	c.log.Debug("ConsensusBlockerTimer started")

	return nil
}

func (c *ConsensusBlockerTimer) Stop() error {
	if c.stopChan == nil {
		return nil
	}

	c.stopChan <- true
	close(c.stopChan)
	c.stopChan = nil

	log.Debug("ConsensusBlockerTimer stopped")

	return nil
}

func (c *ConsensusBlockerTimer) waiting() {
end:
	for {
		select {
		case <-c.stopChan:
			c.log.Debug("timer is stopped")
			break end
		case <-time.After(c.timeout):
			c.log.Debug("wating seal expired")
			if err := c.callback(); err != nil {
				log.Error("failed to doTimeout", "error", err)
			}
		}
	}
}
