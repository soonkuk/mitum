package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

type Consensus struct {
	sync.RWMutex

	receiver    chan common.Seal
	sender      func(common.Node, common.Seal) error
	stopChan    chan bool
	sealHandler SealHandler
}

func NewConsensus() (*Consensus, error) {
	return &Consensus{
		stopChan:    make(chan bool),
		sealHandler: NewISAACSealHandler(),
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

			if err := c.sealHandler.Receive(seal); err != nil {
				// TODO error occurred
				log.Error("failed to handle seal", "error", err)
			}
		case <-c.stopChan:
			break end
		}
	}
}
