package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

type ConsensusBlocker struct {
	sync.RWMutex
	stop            chan bool
	voteChan        chan common.Seal
	state           *ConsensusState
	votingBox       VotingBox
	sealBroadcaster SealBroadcaster
	sealPool        SealPool
}

func NewConsensusBlocker(
	state *ConsensusState,
	votingBox VotingBox,
	sealBroadcaster SealBroadcaster,
	sealPool SealPool,
) *ConsensusBlocker {
	return &ConsensusBlocker{
		voteChan:        make(chan common.Seal),
		state:           state,
		votingBox:       votingBox,
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

		return c.doNewRound(votingResult)
	case VoteStageSIGN:
		if votingResult.Result == VoteResultYES {
			return c.doGoToNextStage(votingResult)
		}

		return c.doNextRound(votingResult)
	case VoteStageACCEPT:
		return c.doFinishRound(votingResult)
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

	log.Debug("proposal accepted", "result", votingResult)

	return nil
}

// doGoToNextStage goes to next stage
func (c *ConsensusBlocker) doGoToNextStage(votingResult VoteResultInfo) error {
	// broadcast next stage ballot
	ballot, err := NewBallot(
		votingResult.Proposal,
		c.state.Home().Address(),
		votingResult.Height,
		votingResult.Round,
		votingResult.Stage.Next(),
		VoteYES,
	)
	if err != nil {
		return err
	}

	log.Debug("move to next sage", "result", votingResult, "next", votingResult.Stage.Next())

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

	log.Debug("finish round", "result", votingResult, "propose", propose)

	return nil
}

// doNewRound starts new round
func (c *ConsensusBlocker) doNewRound(votingResult VoteResultInfo) error {
	log.Debug("start new round", "result", votingResult)

	ballot, err := NewBallot(
		common.Hash{},
		c.state.Home().Address(),
		votingResult.Height,
		votingResult.Round,
		VoteStageINIT,
		VoteYES, // should be yes
	)
	if err != nil {
		return err
	}

	// TODO Proposer should be selected
	ballot.Proposer = c.state.Home().Address()

	if err := c.sealBroadcaster.Send(BallotSealType, ballot); err != nil {
		return err
	}

	return nil
}

// doNextRound starts next round
func (c *ConsensusBlocker) doNextRound(votingResult VoteResultInfo) error {
	log.Debug("next round", "result", votingResult)

	ballot, err := NewBallot(
		common.Hash{},
		c.state.Home().Address(),
		votingResult.Height,
		votingResult.Round+1, // next round
		VoteStageINIT,
		VoteYES, // should be yes
	)
	if err != nil {
		return err
	}

	// TODO Proposer should be selected
	ballot.Proposer = c.state.Home().Address()

	if err := c.sealBroadcaster.Send(BallotSealType, ballot); err != nil {
		return err
	}

	return nil
}
