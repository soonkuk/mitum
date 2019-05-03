package isaac

import (
	"encoding/json"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type ConsensusState struct {
	sync.RWMutex
	node   *common.HomeNode
	height common.Big  // last Block.Height
	block  common.Hash // Block.Hash()
	state  []byte      // last State.Root.Hash()
}

func (c *ConsensusState) String() string {
	c.RLock()
	defer c.RUnlock()

	b, _ := json.Marshal(map[string]interface{}{
		"node":   c.node,
		"height": c.height,
		"block":  c.block,
		"state":  c.state,
	})
	return common.TerminalLogString(string(b))
}

func (c *ConsensusState) Node() *common.HomeNode {
	c.RLock()
	defer c.RUnlock()

	return c.node
}

func (c *ConsensusState) Height() common.Big {
	c.RLock()
	defer c.RUnlock()

	return c.height
}

func (c *ConsensusState) SetHeight(height common.Big) {
	c.Lock()
	defer c.Unlock()

	c.height = height
}

func (c *ConsensusState) Block() common.Hash {
	c.RLock()
	defer c.RUnlock()

	return c.block
}

func (c *ConsensusState) SetBlock(block common.Hash) {
	c.Lock()
	defer c.Unlock()

	c.block = block
}

func (c *ConsensusState) State() []byte {
	c.RLock()
	defer c.RUnlock()

	return c.state
}

func (c *ConsensusState) SetState(state []byte) {
	c.Lock()
	defer c.Unlock()

	c.state = state
}
