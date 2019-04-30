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
	stopChan chan bool
	ctx      context.Context
}

func NewConsensus() (*Consensus, error) {
	c := &Consensus{
		stopChan: make(chan bool),
		receiver: make(chan common.Seal),
		ctx:      context.Background(),
	}

	return c, nil
}

func (c *Consensus) Name() string {
	return "isaac"
}

func (c *Consensus) Start() error {
	// check context values, which should exist
	if _, ok := c.Context().Value("state").(*ConsensusState); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"state",
		)
	}

	if _, ok := c.Context().Value("policy").(ConsensusPolicy); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"policy",
		)
	}

	if _, ok := c.Context().Value("blockStorage").(BlockStorage); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"blockStorage",
		)
	}

	if _, ok := c.Context().Value("roundVoting").(*RoundVoting); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"roundVoting",
		)
	}

	if _, ok := c.Context().Value("roundBoy").(RoundBoy); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"roundBoy",
		)
	}

	if _, ok := c.Context().Value("sealPool").(SealPool); !ok {
		return ConsensusNotReadyError.SetMessage(
			"%s; '%v' is missing in context",
			ConsensusNotReadyError.Message(),
			"sealPool",
		)
	}

	go c.doLoop()

	return nil
}

func (c *Consensus) Stop() error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		c.stopChan <- true
		close(c.stopChan)
		c.stopChan = nil
	}

	if c.receiver != nil {
		close(c.receiver)
		c.receiver = nil
	}

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

// TODO Please correct this boring method name, `doLoop` :(
func (c *Consensus) doLoop() {
	// NOTE these seal should be verified that is well-formed.
end:
	for {
		select {
		case <-c.stopChan:
			break end
		case seal, notClosed := <-c.receiver:
			if !notClosed {
				continue
			}

			go c.receiveSeal(seal)
		}
	}

}

func (c *Consensus) receiveSeal(seal common.Seal) error {
	sHash, _, err := seal.Hash()
	if err != nil {
		log.Error("failed to get Seal.Hash()", "error", err, "seal", seal)
		return err
	}

	log_ := log.New(log15.Ctx{"seal": sHash, "seal-type": seal.Type})

	ctx := common.ContextWithValues(
		c.ctx,
		"seal", seal,
		"sHash", sHash,
	)

	checker := common.NewChainChecker(
		"received-seal-checker",
		ctx,
		CheckerSealPool,
		CheckerSealTypes,
	)
	if err := checker.Check(); err != nil {
		log_.Error("failed to checker for seal", "error", err, "seal", sHash)
		return err
	}

	return nil
}
