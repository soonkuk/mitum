package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type Consensus struct {
	sync.RWMutex

	state       *ConsensusState
	receiver    chan common.Seal
	sender      func(common.Node, common.Seal) error
	stopChan    chan bool
	sealHandler SealHandler
	voting      *RoundVoting
}

func NewConsensus(state *ConsensusState) (*Consensus, error) {
	return &Consensus{
		state:       state,
		stopChan:    make(chan bool),
		sealHandler: NewISAACSealHandler(),
		voting:      NewRoundVoting(),
	}, nil
}

func (c *Consensus) Name() string {
	return "isaac"
}

func (c *Consensus) Start() error {
	c.Lock()
	defer c.Unlock()

	if c.receiver != nil {
		close(c.receiver)
	}

	c.receiver = make(chan common.Seal)
	go c.receive()

	return nil
}

func (c *Consensus) Stop() error {
	c.Lock()
	defer c.Unlock()

	c.stopChan <- true

	if c.receiver != nil {
		close(c.receiver)
		c.receiver = nil
	}

	return nil
}

func (c *Consensus) Receiver() chan common.Seal {
	return c.receiver
}

func (c *Consensus) SealHandler() SealHandler {
	return c.sealHandler
}

func (c *Consensus) SetSealHandler(h SealHandler) error {
	c.Lock()
	defer c.Unlock()

	c.sealHandler = h

	return nil
}

func (c *Consensus) RegisterSendFunc(sender func(common.Node, common.Seal) error) error {
	c.Lock()
	defer c.Unlock()

	c.sender = sender

	return nil
}

func (c *Consensus) receive() {
	// these seal should be verified that is well-formed.
end:
	for {
		select {
		case seal, notClosed := <-c.receiver:
			if !notClosed {
				continue
			}

			if err := c.handleSeal(seal); err != nil {
				log.Error("failed to handle seal", "error", err)
				continue
			}
		case <-c.stopChan:
			break end
		}
	}
}

func (c *Consensus) handleSeal(seal common.Seal) error {
	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"seal-hash": sealHash, "type": seal.Type})
	log_.Debug("got new seal")

	switch seal.Type {
	case ProposeBallotSealType:
		vp, vs, err := c.voting.Open(seal)
		if err != nil {
			return err
		}
		log_.Debug("starting new round", "voting-proposal", vp, "voting-stage", vs)

		if err := c.voted(seal, vp, vs); err != nil {
			return err
		}
	case VoteBallotSealType:
		var voteBallot VoteBallot
		if err := seal.UnmarshalBody(&voteBallot); err != nil {
			return err
		}
		vp, vs, err := c.voting.Vote(voteBallot)
		if err != nil {
			return err
		}
		log_.Debug("voted", "ballot", voteBallot, "voting-proposal", vp, "voting-stage", vs)

		if err := c.voted(seal, vp, vs); err != nil {
			return err
		}
	case TransactionSealType:
		// TODO implement
	}

	if err := c.sealHandler.Add(seal); err != nil {
		return err
	}

	return nil
}

func (c *Consensus) voted(seal common.Seal, vp *VotingProposal, vs *VotingStage) error {
	if !vs.CanCount(c.state.Total(), c.state.Threshold()) {
		return nil
	}

	majority := vs.Majority(c.state.Total(), c.state.Threshold())
	switch majority {
	case VoteResultNotYet:
		return SomethingWrongVotingError.SetMessage(
			"something wrong; CanCount() but voting not yet finished",
		)
	}

	log.Debug(
		"consensus got majority",
		"majority", majority,
		"total", c.state.Total(),
		"threshold", c.state.Threshold(),
	)

	sealHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	// TODO store block & state data

	if err := c.voting.Finish(sealHash); err != nil {
		return err
	}

	return nil
}
