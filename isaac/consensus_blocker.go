package isaac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type ConsensusBlockerBlockingChanFunc func() (common.Seal, chan<- error)

type ConsensusBlocker struct {
	sync.RWMutex
	*common.Logger
	stopBlockingChan       chan bool
	blockingChan           chan ConsensusBlockerBlockingChanFunc
	timer                  *common.CallbackTimer
	lastVotingResult       VoteResultInfo
	lastFinishedProposalAt common.Time

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
		Logger:           common.NewLogger(log, "node", state.Home().Name(), "module", "blocker"),
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

	c.Log().Debug(
		"blocker started",
		"policy", c.policy,
		"state", c.state,
		"seal-broadcastor", fmt.Sprintf("%T", c.sealBroadcaster),
		"seal-pool", fmt.Sprintf("%T", c.sealPool),
		"proposer-selector", fmt.Sprintf("%T", c.proposerSelector),
		"block-storage", fmt.Sprintf("%T", c.blockStorage),
	)

	return nil
}

func (c *ConsensusBlocker) Join() error {
	if c.state.NodeState() != NodeStateJoin {
		_ = c.state.SetNodeState(NodeStateJoin)
	}

	go func() {
		if err := c.joinConsensus(c.state.Height()); err != nil {
			c.Log().Error("failed to joinConsensus", "error", err)
		}
	}()

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
		_ = c.timer.Stop()
	}

	c.Log().Debug("ConsensusBlocker stopped")

	// votingBox also be cleared automatically
	if err := c.votingBox.Clear(); err != nil {
		return err
	}

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
			if errChan != nil {
				errChan <- err
			}
		}
	}
}

func (c *ConsensusBlocker) startTimer(
	name string,
	timeout time.Duration,
	keepRunning bool,
	callback func() error,
) error {
	c.Lock()
	defer c.Unlock()

	if c.timer != nil {
		_ = c.timer.Stop()
		c.timer = nil
	}

	// TODO initial timeout
	c.timer = common.NewCallbackTimer(
		timeout,
		callback,
		keepRunning,
	)
	c.timer.SetLogger(log)
	c.timer.SetLogContext(
		"module", name,
		"node", c.state.Home().Name(),
	)

	return c.timer.Start()
}

func (c *ConsensusBlocker) stopTimer() error { // nolint
	c.Lock()
	defer c.Unlock()

	if c.timer == nil {
		return nil
	}

	_ = c.timer.Stop()
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
	log_ := c.Log().New(log15.Ctx{
		"seal":      seal.Hash(),
		"seal_type": seal.Type(),
	})

	var err error
	var votingResult VoteResultInfo
	{ // voting
		switch seal.Type() {
		case ProposalSealType:
			var proposal Proposal
			if err = common.CheckSeal(seal, &proposal); err != nil {
				return err
			}

			votingResult, err = c.voteProposal(proposal)
		case INITBallotSealType:
			var ballot INITBallot
			if err = common.CheckSeal(seal, &ballot); err != nil {
				return err
			}

			votingResult, err = c.voteBallot(ballot)
		case SIGNBallotSealType:
			var ballot SIGNBallot
			if err = common.CheckSeal(seal, &ballot); err != nil {
				return err
			}

			votingResult, err = c.voteBallot(ballot)
		case ACCEPTBallotSealType:
			var ballot ACCEPTBallot
			if err = common.CheckSeal(seal, &ballot); err != nil {
				return err
			}

			votingResult, err = c.voteBallot(ballot)
		default:
			return common.InvalidSealTypeError
		}
	}

	if err != nil {
		cerr, ok := err.(common.Error)
		if !ok {
			return err
		}

		switch cerr.Code() {
		case DifferentHeightConsensusError.Code():
			// TODO go to sync
			log_.Debug("go to sync", "error", err)
			_ = c.state.SetNodeState(NodeStateSync)
			if err = c.Stop(); err != nil {
				return err
			}
		case DifferentBlockHashConsensusError.Code():
			// TODO go to sync
			log_.Debug("go to sync", "error", err)
			_ = c.state.SetNodeState(NodeStateSync)
			if err = c.Stop(); err != nil {
				return err
			}
		case ConsensusButBlockDoesNotMatchError.Code():
			log_.Error("something wrong", "erorr", cerr)
		}

		return err
	}

	if votingResult.NotYet() {
		return nil
	}

	log_.Debug("got votingResult", "result", votingResult)

	if c.state.NodeState() != NodeStateConsensus {
		_ = c.state.SetNodeState(NodeStateConsensus)
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
				votingResult.Proposer,
				votingResult.Block,
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

func (c *ConsensusBlocker) voteProposal(proposal Proposal) (VoteResultInfo, error) {
	checker := common.NewChainChecker(
		"blocker-vote-proposal-checker",
		common.ContextWithValues(
			context.Background(),
			"proposal", proposal,
			"state", c.state,
		),
		CheckerBlockerProposalBlock,
	)
	checker.SetLogContext(
		"node", c.state.Home().Name(),
		"seal", proposal.Hash(),
		"seal_type", proposal.Type(),
	)
	if err := checker.Check(); err != nil {
		checker.Log().Error("failed to check", "error", err)
		return VoteResultInfo{}, err
	}

	return c.votingBox.Open(proposal)
}

func (c *ConsensusBlocker) voteBallot(ballot Ballot) (VoteResultInfo, error) {
	logContext := []interface{}{"node", c.state.Home().Name(), "seal", ballot.Hash(), "seal_type", ballot.Type()}

	ballotChecker := common.NewChainChecker(
		"blocker-vote-ballot-checker",
		common.ContextWithValues(
			context.Background(),
			"ballot", ballot,
			"state", c.state,
		),
		CheckerBlockerBallot,
	)
	ballotChecker.SetLogContext(logContext...)
	if err := ballotChecker.Check(); err != nil {
		ballotChecker.Log().Error("failed to check", "error", err)
		return VoteResultInfo{}, err
	}

	votingResult, err := c.votingBox.Vote(ballot)
	if err != nil {
		return VoteResultInfo{}, err
	}

	if votingResult.NotYet() {
		return votingResult, nil
	}

	{
		c.RLock()
		resultChecker := common.NewChainChecker(
			"blocker-vote-ballot-result-checker",
			common.ContextWithValues(
				context.Background(),
				"votingResult", votingResult,
				"lastVotingResult", c.lastVotingResult,
				"state", c.state,
			),
			CheckerBlockerBallotVotingResult,
		)
		resultChecker.SetLogContext(ballotChecker.LogContext()...)
		c.RUnlock()
		if err := resultChecker.Check(); err != nil {
			resultChecker.Log().Error("failed to check", "error", err)
			return votingResult, err
		}
	}

	c.Lock()
	c.lastVotingResult = votingResult
	c.Unlock()

	return votingResult, nil
}

// joinConsensus tries to join network; broadcasts INIT ballot with
// - latest known height block
// - round 0
func (c *ConsensusBlocker) joinConsensus(height common.Big) error {
	c.Log().Debug("trying to join consensus", "height", height)

	err := c.startTimer("join-consensus-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
		return c.broadcastINIT(height, Round(0))
	})
	if err != nil {
		return err
	}

	return c.broadcastINIT(height, Round(0))
}

// doProposeAccepted will do,
// - validate proposal
// - decide YES/NOP
// - broadcast sign ballot
func (c *ConsensusBlocker) doProposeAccepted(votingResult VoteResultInfo) error {
	c.Log().Debug("proposal accepted", "result", votingResult)

	err := c.startTimer("proposal-accepted-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
		return c.broadcastINIT(votingResult.Height, votingResult.Round+1)
	})
	if err != nil {
		return err
	}

	// TODO validate proposal
	// TODO decide YES/NOP

	vote := VoteYES

	// NOTE broadcast sign ballot
	ballot := NewSIGNBallot(
		votingResult.Height,
		votingResult.Round,
		votingResult.Proposer,
		nil, // TODO set validators
		votingResult.Proposal,
		votingResult.Block,
		vote,
	)

	if err := c.sealBroadcaster.Send(&ballot); err != nil {
		return err
	}

	return nil
}

// goToNextStage goes to next stage; the next stage should be ACCEPT
func (c *ConsensusBlocker) goToNextStage(
	proposal common.Hash,
	proposer common.Address,
	block common.Hash,
	height common.Big,
	round Round,
	stage VoteStage,
) error {
	c.Log().Debug("go to next stage", "proposal", proposal, "height", height, "round", round, "next", stage)

	// TODO if proposer is different from current, check
	// - current Block is same
	// - current Height is same
	// - current Height is same

	err := c.startTimer("next-stage-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	// NOTE broadcast next stage ballot
	ballot := NewACCEPTBallot(
		height,
		round,
		proposer,
		nil, // TODO set validators
		proposal,
		block,
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
func (c *ConsensusBlocker) finishRound(proposal common.Hash) error {
	c.Log().Debug("finishing round", "proposal", proposal)

	seal, err := c.sealPool.Get(proposal)
	if err != nil {
		return err
	}

	var p Proposal
	if err = common.CheckSeal(seal, &p); err != nil {
		return common.UnknownSealTypeError.SetMessage(err.Error())
	}

	// TODO store block and state
	if _, _, err = c.blockStorage.NewBlock(p); err != nil {
		return err
	}

	// update ConsensusBlockerState
	{
		prevState := &ConsensusState{}
		*prevState = *c.state // nolint

		if err = c.state.SetHeight(p.Block.Height.Inc()); err != nil {
			return err
		}
		if err = c.state.SetBlock(p.Block.Next); err != nil {
			return err
		}
		if err = c.state.SetState(p.State.Next); err != nil {
			return err
		}

		c.Log().Debug(
			"finished round",
			"old-block-height", prevState.Height().String(),
			"old-block-hash", prevState.Block(),
			"old-state-hash", prevState.State(),
			"new-block-height", c.state.Height().String(),
			"new-block-hash", c.state.Block(),
			"new-state-hash", c.state.State(),
		)

		c.lastFinishedProposalAt = p.SignedAt()
	}

	c.Lock()
	c.lastVotingResult = VoteResultInfo{}
	c.Unlock()

	// propose or wait new proposal
	err = c.startTimer("finish-round-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
		return c.broadcastINIT(p.Block.Height.Inc(), Round(0))
	})
	if err != nil {
		return err
	}

	return nil
}

// startNewRound starts new round
func (c *ConsensusBlocker) startNewRound(height common.Big, round Round) error {
	c.Log().Debug("start new round", "height", height, "round", round)

	err := c.startTimer("new-round-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
		return c.broadcastINIT(height, round+1)
	})
	if err != nil {
		return err
	}

	return c.broadcastINIT(height, round)
}

// runNewRound starts new round and propose new proposal
func (c *ConsensusBlocker) runNewRound(height common.Big, round Round) error {
	log_ := c.Log().New(log15.Ctx{"height": height, "round": round})
	log_.Debug("run new round")

	err := c.startTimer(
		"run-round-broadcast-init",
		c.policy.TimeoutWaitSeal,
		true,
		func() error {
			return c.broadcastINIT(height, round+1)
		},
	)
	if err != nil {
		return err
	}

	return c.propose(height, round)
}

func (c *ConsensusBlocker) broadcastINIT(height common.Big, round Round) error {
	log_ := c.Log().New(log15.Ctx{"height": height, "round": round})

	proposer, err := c.proposerSelector.Select(c.state.Block(), height, round)
	if err != nil {
		return err
	}
	ballot := NewINITBallot(
		height,
		round,
		proposer.Address(),
		nil, // TODO set validators
	)

	// TODO self-signed ballot should not be needed to broadcast
	if err := c.sealBroadcaster.Send(&ballot); err != nil {
		return err
	}

	log_.Debug(
		"INIT ballot broadcasted",
		"block", c.state.Block(),
		"proposer", proposer.Address(),
		"seal", ballot.Hash(),
	)

	return nil
}

func (c *ConsensusBlocker) propose(height common.Big, round Round) error {
	log_ := c.Log().New(log15.Ctx{"height": height, "round": round})
	log_.Debug("new proposal will be proposed")

	proposer, err := c.proposerSelector.Select(c.state.Block(), height, round)
	if err != nil {
		return err
	}
	log_.Debug("proposer selected", "block", c.state.Block(), "proposer", proposer.Address())

	if !proposer.Equal(c.state.Home()) {
		log_.Debug("proposer is not home; will wait new proposal")
		return nil
	}

	log_.Debug("proposer is home; new proposal will be proposed")
	// TODO validate transactions.
	proposal := NewProposal(
		round,
		ProposalBlock{
			Height:  height,
			Current: c.state.Block(),
			Next:    common.NewRandomHash("bk"), // TODO should be determined by validation
		},
		ProposalState{
			Current: c.state.State(),
			Next:    []byte("next state"), // TODO should be determined by validation
		},
		nil, // TODO transactions
	)

	// NOTE will propose after,
	// duration = ConsensusPolicy.AvgBlockRoundInterval - (<Latest block's Proposal.SignedAt> - <Now()>)
	// if duration is under 0, after 300 millisecond
	var delay = time.Millisecond * 300
	if !c.lastFinishedProposalAt.IsZero() {
		delay = c.policy.AvgBlockRoundInterval - common.Now().Sub(c.lastFinishedProposalAt)
		if delay < 0 {
			delay = time.Second
		}
	}

	err = c.startTimer("propose-new-proposal", delay, false, func() error {
		return c.broadcastProposal(proposal)
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsensusBlocker) broadcastProposal(proposal Proposal) error {
	err := c.sealBroadcaster.Send(&proposal)
	if err != nil {
		c.Log().Error("failed to broadcast", "proposal", proposal.Hash())
		return err
	}

	c.Log().Debug("proposal broadcasted", "proposal", proposal.Hash())

	return nil
}
