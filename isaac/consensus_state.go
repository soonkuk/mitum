package isaac

import (
	"encoding/json"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type ConsensusState struct {
	sync.RWMutex
	home      *common.HomeNode
	height    common.Big  // last Block.Height
	block     common.Hash // Block.Hash()
	state     []byte      // last State.Root.Hash()
	nodeState NodeState
}

func NewConsensusState(home *common.HomeNode) *ConsensusState {
	return &ConsensusState{
		home:      home,
		nodeState: NodeStateBooting,
	}
}

func (c *ConsensusState) MarshalJSON() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()

	return json.Marshal(map[string]interface{}{
		"home":   c.home,
		"height": c.height,
		"block":  c.block,
		"state":  c.state,
	})
}

func (c *ConsensusState) String() string {
	c.RLock()
	defer c.RUnlock()

	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}

func (c *ConsensusState) Home() *common.HomeNode {
	c.RLock()
	defer c.RUnlock()

	return c.home
}

func (c *ConsensusState) Height() common.Big {
	c.RLock()
	defer c.RUnlock()

	return c.height
}

func (c *ConsensusState) SetHeight(height common.Big) error {
	c.Lock()
	defer c.Unlock()

	c.height = height

	return nil
}

func (c *ConsensusState) Block() common.Hash {
	c.RLock()
	defer c.RUnlock()

	return c.block
}

func (c *ConsensusState) SetBlock(block common.Hash) error {
	c.Lock()
	defer c.Unlock()

	c.block = block

	return nil
}

func (c *ConsensusState) State() []byte {
	c.RLock()
	defer c.RUnlock()

	return c.state
}

func (c *ConsensusState) SetState(state []byte) error {
	c.Lock()
	defer c.Unlock()

	c.state = state

	return nil
}

func (c *ConsensusState) NodeState() NodeState {
	c.RLock()
	defer c.RUnlock()

	return c.nodeState
}

func (c *ConsensusState) SetNodeState(state NodeState) error {
	c.Lock()
	defer c.Unlock()

	if err := state.IsValid(); err != nil {
		return err
	}

	c.nodeState = state

	return nil
}
