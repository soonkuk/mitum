package isaac

import (
	"context"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/network"
)

type Consensus struct {
	sync.RWMutex
	*common.Logger
	receiver chan common.Seal
	stop     chan bool
	voteChan chan common.Seal
	ctx      context.Context
	home     common.HomeNode
	blocker  *ConsensusBlocker
	state    *ConsensusState
}

func NewConsensus(
	home common.HomeNode,
	state *ConsensusState,
	blocker *ConsensusBlocker,
) (*Consensus, error) {
	c := &Consensus{
		Logger:   common.NewLogger(log, "module", "consensus", "node", home.Name()),
		receiver: make(chan common.Seal),
		voteChan: make(chan common.Seal),
		ctx:      context.Background(),
		home:     home,
		state:    state,
		blocker:  blocker,
	}

	return c, nil
}

func (c *Consensus) Name() string {
	return "isaac"
}

func (c *Consensus) Start() error {
	if c.stop != nil {
		return common.StartStopperAlreadyStartedError
	}

	c.stop = make(chan bool)

	// check context values, which should exist
	if _, ok := c.Context().Value("policy").(ConsensusPolicy); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"policy",
		)
	}

	if _, ok := c.Context().Value("state").(*ConsensusState); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"state",
		)
	}

	if _, ok := c.Context().Value("proposerSelector").(ProposerSelector); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"proposerSelector",
		)
	}

	if _, ok := c.Context().Value("sealPool").(SealPool); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"sealPool",
		)
	}

	go c.schedule()

	return nil
}

func (c *Consensus) Stop() error {
	c.Lock()
	defer c.Unlock()

	if c.stop == nil {
		return nil
	}

	c.stop <- true
	close(c.stop)
	c.stop = nil

	close(c.receiver)
	c.receiver = nil

	return nil
}

func (c *Consensus) Context() context.Context {
	c.RLock()
	defer c.RUnlock()

	return c.ctx
}

func (c *Consensus) SetContext(ctx context.Context, args ...interface{}) {
	c.Lock()
	defer c.Unlock()

	if ctx == nil {
		ctx = c.ctx
	}

	if len(args) < 1 {
		c.ctx = ctx
		return
	}

	c.ctx = common.ContextWithValues(ctx, args...)
}

func (c *Consensus) Receiver() network.ReceiverFunc {
	return c.receiverFunc
}

func (c *Consensus) receiverFunc(seal common.Seal) error {
	checker := common.NewChainChecker(
		"receive-seal-checker",
		common.ContextWithValues(
			c.ctx,
			"seal", seal,
		),
		CheckerSealFromKnowValidator,
		CheckerSealIsValid,
		CheckerSealPool,
		CheckerSealTypes,
	)
	checker.SetLogContext("node", c.home.Name())
	if err := checker.Check(); err != nil {
		checker.Log().Error("failed to check", "error", err)
		return err
	}

	go func() {
		c.receiver <- seal
	}()

	return nil
}

// TODO Please correct this boring method name, `schedule` :(
func (c *Consensus) schedule() {
end:
	for {
		select {
		case <-c.stop:
			break end
		case seal, notClosed := <-c.receiver:
			if !notClosed {
				continue
			}

			go func() {
				if err := c.receiveSeal(seal); err != nil {
					c.Log().Error("failed to receive seal", "error", err)
				}
			}()
		}
	}
}

func (c *Consensus) receiveSeal(seal common.Seal) error {
	// NOTE these seal should be verified that is wellformed before.
	log_ := c.Log().New(log15.Ctx{"seal": seal.Hash(), "seal-type": seal.Type()})

	if !c.state.NodeState().CanVote() {
		log_.Error("node cannot vote", "state", c.state.NodeState())
		return nil
	}

	go func(seal common.Seal) {
		errChan := make(chan error)
		c.blocker.Vote(seal, errChan)

		if err := <-errChan; err != nil {
			c.Log().Error("failed to vote", "seal", seal.Hash(), "error", err)
		}
	}(seal)

	return nil
}
