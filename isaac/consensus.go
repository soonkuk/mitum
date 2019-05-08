package isaac

import (
	"context"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type Consensus struct {
	sync.RWMutex
	receiver chan common.Seal
	stop     chan bool
	voteChan chan common.Seal
	ctx      context.Context
	blocker  *ConsensusBlocker
}

func NewConsensus(
	blocker *ConsensusBlocker,
) (*Consensus, error) {
	c := &Consensus{
		receiver: make(chan common.Seal),
		voteChan: make(chan common.Seal),
		ctx:      context.Background(),
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

func (c *Consensus) Receiver() chan common.Seal {
	return c.receiver
}

// TODO Please correct this boring method name, `schedule` :(
func (c *Consensus) schedule() {
	// NOTE these seal should be verified that is wellformed.
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
					log.Error("failed to receive seal")
				}
			}()
		}
	}
}

func (c *Consensus) receiveSeal(seal common.Seal) error {
	log_ := log.New(log15.Ctx{"seal": seal.Hash(), "seal-type": seal.Type()})

	checker := common.NewChainChecker(
		"receiveSeal",
		common.ContextWithValues(
			c.ctx,
			"seal", seal,
		),
		CheckerSealPool,
		CheckerSealTypes,
	)
	if err := checker.Check(); err != nil {
		log_.Error("failed to checker for seal", "error", err)
		return err
	}

	go func(seal common.Seal) {
		errChan := make(chan error)
		c.blocker.Vote(seal, errChan)

		select {
		case err := <-errChan:
			if err == nil {
				return
			}

			log.Error("failed to vote", "error", err)

			cerr, ok := err.(common.Error)
			if !ok {
				return
			}

			switch cerr.Code() {
			case DifferentHeightConsensusError.Code():
				// TODO detect sync
				//c.state.SetNodeState(NodeStateSync)
			}
		}
	}(seal)

	return nil
}
