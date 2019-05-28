package isaac

import (
	"encoding/json"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type ConsensusState struct {
	sync.RWMutex
	*common.Logger
	home       common.HomeNode
	height     common.Big  // last Block.Height
	block      common.Hash // Block.Hash()
	state      []byte      // last State.Root.Hash()
	nodeState  NodeState
	validators map[common.Address]common.Validator
}

func NewConsensusState(home common.HomeNode) *ConsensusState {
	s := &ConsensusState{
		Logger:     common.NewLogger(log, "module", "consensus-state", "node", home.Name()),
		home:       home,
		validators: map[common.Address]common.Validator{},
	}

	_ = s.SetNodeState(NodeStateBooting)
	return s
}

func (c *ConsensusState) MarshalJSON() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()

	return json.Marshal(map[string]interface{}{
		"home":       c.home,
		"height":     c.height,
		"block":      c.block,
		"state":      c.state,
		"node-state": c.nodeState,
		"validators": c.Validators(),
	})
}

func (c *ConsensusState) String() string {
	c.RLock()
	defer c.RUnlock()

	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}

func (c *ConsensusState) Home() common.HomeNode {
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
	} else if c.nodeState == state {
		return nil
	}

	stateFrom := c.nodeState
	c.nodeState = state
	c.Log().Debug("node state transitted", "from", stateFrom, "to", c.nodeState)

	return nil
}

func (c *ConsensusState) Validators() []common.Validator {
	c.RLock()
	defer c.RUnlock()

	var validators []common.Validator
	for _, validator := range c.validators {
		validators = append(validators, validator)
	}

	return validators
}

func (c *ConsensusState) ExistsValidators(validator common.Address) bool {
	_, found := c.validators[validator]
	return found
}

func (c *ConsensusState) AddValidators(validators ...common.Validator) error {
	c.Lock()
	defer c.Unlock()

	for _, validator := range validators {
		if _, found := c.validators[validator.Address()]; found {
			continue
		} else if c.home.Equal(validator) { // NOTE validators does not contain home itself
			continue
		}

		c.validators[validator.Address()] = validator
	}

	return nil
}

func (c *ConsensusState) RemoveValidators(validators ...common.Address) error {
	c.Lock()
	defer c.Unlock()

	for _, validator := range validators {
		if _, found := c.validators[validator]; !found {
			continue
		}
		delete(c.validators, validator)
	}

	return nil
}
