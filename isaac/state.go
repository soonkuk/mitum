package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

type ConsensusState struct {
	sync.RWMutex
	height    common.Big  // last Block.Height
	block     common.Hash // Block.Hash()
	state     []byte      // last State.Root.Hash()
	total     uint        // total number of validators
	threshold uint        // consensus threshold
}

func (c *ConsensusState) Height() common.Big {
	return c.height
}

func (c *ConsensusState) Block() common.Hash {
	return c.block
}

func (c *ConsensusState) State() []byte {
	return c.state
}

func (c *ConsensusState) Total() uint {
	return c.total
}

func (c *ConsensusState) Threshold() uint {
	return c.threshold
}
