package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type Consensus struct {
	sync.RWMutex

	policy   ConsensusPolicy
	state    *ConsensusState
	receiver chan common.Seal
	sender   func(common.Node, common.Seal) error
	stopChan chan bool
	sealPool SealPool
	voting   *RoundVoting
	roundboy Roundboy
}

func NewConsensus(policy ConsensusPolicy, state *ConsensusState) (*Consensus, error) {
	c := &Consensus{
		policy:   policy,
		state:    state,
		stopChan: make(chan bool),
		sealPool: NewISAACSealPool(),
		voting:   NewRoundVoting(),
		receiver: make(chan common.Seal),
	}

	return c, nil
}

func (c *Consensus) Name() string {
	return "isaac"
}

func (c *Consensus) Start() error {
	c.Lock()
	defer c.Unlock()

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

func (c *Consensus) Policy() ConsensusPolicy {
	return c.policy
}

func (c *Consensus) State() *ConsensusState {
	return c.state
}

func (c *Consensus) Receiver() chan common.Seal {
	return c.receiver
}

func (c *Consensus) SealPool() SealPool {
	return c.sealPool
}

func (c *Consensus) SetSealPool(h SealPool) error {
	c.Lock()
	defer c.Unlock()

	c.sealPool = h

	return nil
}

func (c *Consensus) Roundboy() Roundboy {
	c.RLock()
	defer c.RUnlock()

	return c.roundboy
}

func (c *Consensus) SetRoundboy(s Roundboy) {
	c.Lock()
	defer c.Unlock()

	c.roundboy = s
}

func (c *Consensus) SetSender(sender func(common.Node, common.Seal) error) error {
	c.Lock()
	defer c.Unlock()

	c.sender = sender

	return nil
}

func (c *Consensus) Voting() *RoundVoting {
	c.RLock()
	defer c.RUnlock()

	return c.voting
}

// TODO Please correct this boring method name, `doLoop` :(
func (c *Consensus) doLoop() {
	// NOTE these seal should be verified that is well-formed.
end:
	for {
		select {
		case seal, notClosed := <-c.receiver:
			if !notClosed {
				continue
			}

			go c.receiveSeal(seal)
		case <-c.stopChan:
			break end
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

	if err := c.sealPool.Add(seal); err != nil {
		log_.Error("failed to SealPool", "error", err)
		return err
	}

	ctx := common.ContextWithValues(
		nil,
		"policy", c.policy,
		"state", c.state,
		"seal", seal,
		"sHash", sHash,
		"sealPool", c.sealPool,
		"roundVoting", c.voting,
		"roundboy", c.roundboy,
	)

	checker := common.NewChainChecker("received-seal-checker", ctx, CheckerSealTypes)
	if err := checker.Check(); err != nil {
		log_.Error("failed to checker for seal", "error", err, "seal", sHash)
		return err
	}

	return nil
}
