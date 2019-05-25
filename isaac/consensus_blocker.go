package isaac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/element"
)

type ConsensusBlockerBlockingChanFunc func() (common.Seal, chan<- error)

type ConsensusBlocker struct {
	sync.RWMutex
	*common.Logger
	stopBlockingChan       chan bool
	blockingChan           chan ConsensusBlockerBlockingChanFunc
	timer                  common.TimerCallback
	proposalTimer          common.TimerCallback
	lastVotingResult       VoteResultInfo
	lastFinishedProposalAt common.Time

	policy                ConsensusPolicy
	state                 *ConsensusState
	votingBox             VotingBox
	sealBroadcaster       SealBroadcaster
	sealPool              SealPool
	proposerSelector      ProposerSelector
	transactionValidation *TransactionValidation
	proposalValidator     ProposalValidator
}

func NewConsensusBlocker(
	policy ConsensusPolicy,
	state *ConsensusState,
	votingBox VotingBox,
	sealBroadcaster SealBroadcaster,
	sealPool SealPool,
	proposerSelector ProposerSelector,
	proposalValidator ProposalValidator,
) *ConsensusBlocker {
	return &ConsensusBlocker{
		Logger:                common.NewLogger(log, "node", state.Home().Name(), "module", "blocker"),
		blockingChan:          make(chan ConsensusBlockerBlockingChanFunc),
		policy:                policy,
		state:                 state,
		votingBox:             votingBox,
		sealBroadcaster:       sealBroadcaster,
		sealPool:              sealPool,
		proposerSelector:      proposerSelector,
		transactionValidation: NewTransactionValidation(),
		proposalValidator:     proposalValidator,
	}
}

func (c *ConsensusBlocker) Start() error {
	c.Log().Debug("trying to start blocker")
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
	)

	return nil
}

func (c *ConsensusBlocker) Join() error {
	c.Log().Debug("trying to join consensus")
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

	c.Log().Debug("trying to stop blocker")
	if c.stopBlockingChan == nil {
		return nil
	}

	c.stopBlockingChan <- true
	close(c.stopBlockingChan)
	c.stopBlockingChan = nil

	if c.timer != nil {
		_ = c.timer.Stop()
	}

	// NOTE votingBox also be cleared automatically
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
			c.Log().Debug("blocker stopped")
			break end
		case f, notClosed := <-c.blockingChan:
			if !notClosed {
				continue
			}

			seal, errChan := f()

			err := c.handle(seal)
			if errChan != nil {
				errChan <- err
			}
		}
	}
}

func (c *ConsensusBlocker) startTimerWithFunc(
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

	// TODO timer initial timeout
	timer := common.NewDefaultTimerCallback(timeout, callback)
	timer.SetKeepRunning(keepRunning)
	timer.SetLogger(log)
	timer.SetLogContext(
		"module", name,
		"timer-id", common.RandomUUID(),
		"node", c.state.Home().Name(),
	)

	c.timer = timer

	return c.timer.Start()
}

func (c *ConsensusBlocker) startTimerWithTimer(timer common.TimerCallback) error {
	c.Lock()
	defer c.Unlock()

	if c.timer != nil {
		_ = c.timer.Stop()
		c.timer = nil
	}

	c.timer = timer

	return c.timer.Start()
}

func (c *ConsensusBlocker) startProposalTimer(proposal Proposal, delay time.Duration) error {
	c.Lock()
	defer c.Unlock()

	if c.proposalTimer != nil {
		_ = c.proposalTimer.Stop()
		c.proposalTimer = nil
	}

	timer := common.NewDefaultTimerCallback(delay, func() error {
		return c.broadcastProposal(proposal)
	})
	timer.SetLogger(c.Log())
	timer.SetLogContext(
		"module", "propose-new-proposal",
		"timer-id", common.RandomUUID(),
	)

	c.proposalTimer = timer

	return c.proposalTimer.Start()
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

// NOTE vote votes and decides the next action.
func (c *ConsensusBlocker) handle(seal common.Seal) error {
	log_ := c.Log().New(log15.Ctx{
		"seal":      seal.Hash(),
		"seal_type": seal.Type(),
	})

	log_.Debug(fmt.Sprintf("`%v` seal reached to blocker", seal.Type()))

	err := c.vote(seal)
	if err == nil {
		return nil
	}

	switch {
	case DifferentHeightConsensusError.Equal(err),
		DifferentBlockHashConsensusError.Equal(err),
		ValidationIsNotDoneError.Equal(err),
		ConsensusButBlockDoesNotMatchError.Equal(err):
		// TODO go to sync
		log_.Debug("blocker can not handle seal; go to sync", "error", err)
		_ = c.state.SetNodeState(NodeStateSync)
		go func() {
			if e := c.Stop(); e != nil {
				c.Log().Error("failed to stop", "error", e)
			}
		}()
	default:
		log_.Error("something wrong", "error", err)
	}

	return err
}

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
		return err
	}

	if votingResult.NotYet() {
		return nil
	}

	log_.Debug("got votingResult", "votingResult", votingResult)

	if c.state.NodeState() != NodeStateConsensus {
		_ = c.state.SetNodeState(NodeStateConsensus)
	}

	switch votingResult.Stage {
	case VoteStageINIT:
		if votingResult.Proposed {
			return c.doProposeAccepted(seal, votingResult)
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

	ballotChecker.Log().Debug("got votingResult of ballot", "votingResult", votingResult)

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
			CheckerBlockerVotingBallotResult,
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
	err := c.startTimerWithFunc("join-consensus-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
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
func (c *ConsensusBlocker) doProposeAccepted(seal common.Seal, votingResult VoteResultInfo) error {
	c.Log().Debug("proposal accepted", "result", votingResult)

	var proposal Proposal
	if err := common.CheckSeal(seal, &proposal); err != nil {
		return err
	}

	err := c.startTimerWithFunc("proposal-accepted-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
		return c.broadcastINIT(votingResult.Height, votingResult.Round+1)
	})
	if err != nil {
		return err
	}

	block, vote, err := c.proposalValidator.Validate(proposal)
	if err != nil {
		return err
	}

	// NOTE broadcast sign ballot
	ballot := NewSIGNBallot(
		votingResult.Height,
		votingResult.Round,
		votingResult.Proposer,
		nil, // TODO set validators
		votingResult.Proposal,
		block.Hash(),
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

	err := c.startTimerWithFunc("next-stage-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
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
	log_ := c.Log().New(log15.Ctx{"proposal": proposal})
	log_.Debug("finishing round")

	seal, err := c.sealPool.Get(proposal)
	if err != nil {
		return err
	}

	var p Proposal
	if err = common.CheckSeal(seal, &p); err != nil {
		return common.UnknownSealTypeError.SetMessage(err.Error())
	}

	// NOTE if failed to store,
	// - err is ValidationIsNotDoneError; validate it
	//   - if failed to validate or VoteNOP, go to sync
	// TODO test
	if err = c.proposalValidator.Store(p); err != nil {
		log_.Error("failed to store", "error", err)
		switch {
		case ValidationIsNotDoneError.Equal(err):
			// NOTE validate proposal
			var vote Vote
			_, vote, err = c.proposalValidator.Validate(p)
			if err != nil {
				return FailedToStoreBlockError.SetError(err)
			} else if vote == VoteNOP {
				return FailedToStoreBlockError.AppendMessage(
					"proposal validated, but VoteNOP; proposal=%v",
					proposal,
				)
			}
			if err = c.proposalValidator.Store(p); err != nil {
				return FailedToStoreBlockError.SetError(err)
			}
		default:
			return FailedToStoreBlockError.SetError(err)
		}
	}

	// update ConsensusBlockerState
	{
		prevState := &ConsensusState{}
		*prevState = *c.state // nolint

		if err = c.state.SetHeight(p.Block.Height.Inc()); err != nil {
			return err
		}
		// TODO set new block's hash
		// if err = c.state.SetBlock(<new block hash>); err != nil {
		// 	return err
		// }
		if err = c.state.SetState(p.State.Next); err != nil {
			return err
		}

		log_.Debug(
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
	err = c.startTimerWithFunc("finish-round-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
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

	err := c.startTimerWithFunc("new-round-broadcast-init", c.policy.TimeoutWaitSeal, true, func() error {
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

	err := c.startTimerWithFunc(
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

	// TODO these transactions are from transaction pool
	transactions := []element.Transaction{}

	var valids, invalids []common.Hash
	if len(transactions) > 0 {
		valids, invalids, err = c.transactionValidation.Validate(transactions)
		if err != nil {
			log_.Error("failed to validate transactions", "error", err)
			return err
		}
	}
	log_.Debug(
		"transactons validated",
		"transactions", len(transactions),
		"valids", valids,
		"invalids", invalids,
	)

	// TODO remove valids and invalids transactions
	// - remove transaction pool

	proposal := NewProposal(
		round,
		ProposalBlock{
			Height:  height,
			Current: c.state.Block(),
		},
		ProposalState{
			Current: c.state.State(),
			Next:    []byte("next state"), // TODO should be determined by validation
		},
		valids, // TODO transactions
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

	if err := c.startProposalTimer(proposal, delay); err != nil {
		c.Log().Error("failed to start proposal timer", "error", err)
	}

	return nil
}

func (c *ConsensusBlocker) broadcastProposal(proposal Proposal) error {
	if err := c.sealBroadcaster.Send(&proposal); err != nil {
		c.Log().Error("failed to broadcast", "proposal", proposal.Hash())
		return err
	}

	c.Log().Debug("proposal broadcasted", "proposal", proposal.Hash())

	return nil
}
