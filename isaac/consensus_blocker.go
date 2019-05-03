package isaac

import (
	"sync"
	"time"

	"github.com/spikeekips/mitum/common"
)

type ConsensusBlocker struct {
	sync.RWMutex
	stop            chan bool
	homeNode        *common.HomeNode
	voteChan        chan common.Seal
	state           *ConsensusState
	voting          *Voting
	stageBlocker    *StageBlocker
	sealBroadcaster SealBroadcaster
	sealPool        SealPool
}

func NewConsensusBlocker(
	homeNode *common.HomeNode,
	state *ConsensusState,
	voting *Voting,
	stageBlocker *StageBlocker,
	sealBroadcaster SealBroadcaster,
	sealPool SealPool,
) *ConsensusBlocker {
	return &ConsensusBlocker{
		homeNode:        homeNode,
		voteChan:        make(chan common.Seal),
		state:           state,
		voting:          voting,
		stageBlocker:    stageBlocker,
		sealBroadcaster: sealBroadcaster,
		sealPool:        sealPool,
	}
}

func (c *ConsensusBlocker) Start() error {
	if c.stop != nil {
		return common.StartStopperAlreadyStartedError
	}

	c.stop = make(chan bool)

	go c.blocking()

	return nil
}

func (c *ConsensusBlocker) Stop() error {
	if c.stop == nil {
		return nil
	}

	c.stop <- true
	close(c.stop)
	c.stop = nil

	return nil
}

func (c *ConsensusBlocker) blocking() {
end:
	for {
		select {
		case <-c.stop:
			break end
		case seal, notClosed := <-c.voteChan:
			if !notClosed {
				continue
			}

			if err := c.vote(seal); err != nil {
				log.Error("failed to vote", "error", err)
			}
		}
	}
}

func (c *ConsensusBlocker) Vote(seal common.Seal) {
	go func() {
		c.voteChan <- seal
	}()
}

func (c *ConsensusBlocker) vote(seal common.Seal) error {
	var votingResult VoteResultInfo
	{ // voting
		rc := c.voting.Vote(seal)
		select {
		case <-time.After(time.Second): // TODO duration should be in ConsensusPolicy
			return VotingFailedError.SetMessage("failed to vote; timeouted: %v", time.Second)
		case r := <-rc:
			log.Debug("got voting result", "result", r.Result, "error", r.Err)
			if r.Err != nil {
				return r.Err
			}

			votingResult = r.Result
		}
	}

	var decision StageBlockerDecision
	{
		rc := c.stageBlocker.Check(votingResult)
		select {
		case <-time.After(time.Second): // TODO duration should be in ConsensusPolicy
			return VotingFailedError.SetMessage("failed to vote; timeouted: %v", time.Second)
		case r := <-rc:
			log.Debug("got stage blocker result", "decision", r.Decision, "error", r.Err)
			if r.Err != nil {
				return r.Err
			}

			decision = r.Decision
		}
	}

	switch decision {
	case ProposalAccepted:
		return c.doProposeAccepted(votingResult)
	case GoToNextStage:
		return c.doGoToNextStage(votingResult)
	case FinishRound:
		return c.doFinishRound(votingResult)
	case GoToNextRound:
		// TODO implement
	case StartNewRound:
		// TODO implement
	}

	return nil
}

// doProposeAccepted will do,
// - validate propose
// - decide YES/NOP
// - broadcast sign ballot
func (c *ConsensusBlocker) doProposeAccepted(votingResult VoteResultInfo) error {
	// TODO validate propose
	// TODO decide YES/NOP

	vote := VoteYES

	// broadcast sign ballot

	ballot, err := NewBallot(
		votingResult.Proposal,
		c.homeNode.Address(),
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

func (c *ConsensusBlocker) doGoToNextStage(votingResult VoteResultInfo) error {
	// broadcast next stage ballot
	ballot, err := NewBallot(
		votingResult.Proposal,
		c.homeNode.Address(),
		votingResult.Height,
		votingResult.Round,
		votingResult.Stage.Next(),
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

// doFinishRound will do,
// - store block and state
// - update ConsensusBlockerState
// - ready to start new block
func (c *ConsensusBlocker) doFinishRound(votingResult VoteResultInfo) error {
	c.Lock()
	defer c.Unlock()

	seal, err := c.sealPool.Get(votingResult.Proposal)
	if err != nil {
		return err
	}

	var propose Propose
	if err := seal.UnmarshalBody(&propose); err != nil {
		return err
	}

	// TODO store block and state

	// update ConsensusBlockerState
	c.state.SetHeight(votingResult.Height.Inc())
	c.state.SetBlock(propose.Block.Next)
	c.state.SetState(propose.State.Next)

	/*
		// TODO remove broadcast next stage ballot
		ballot, err := NewBallot(
			common.Hash{},
			c.homeNode.Address(),
			votingResult.Height.Inc(),
			Round(0),
			VoteStageINIT,
			VoteYES,
		)
		if err != nil {
			return err
		}

		ballot.Proposer = c.homeNode.Address()

		if err := c.sealBroadcaster.Send(BallotSealType, ballot); err != nil {
			return err
		}
	*/

	return nil
}
