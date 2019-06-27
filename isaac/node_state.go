package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type HomeState struct {
	sync.RWMutex
	home   node.Home
	height Height
	block  hash.Hash
	state  node.State
}

func NewHomeState(home node.Home, height Height, block hash.Hash) *HomeState {
	return &HomeState{
		home:   home,
		height: height,
		block:  block,
		state:  node.StateBooting, // by default, node state is booting
	}
}

func (hs *HomeState) Home() node.Home {
	return hs.home
}

func (hs *HomeState) Height() Height {
	hs.RLock()
	defer hs.RUnlock()

	return hs.height
}

func (hs *HomeState) SetHeight(height Height) *HomeState {
	hs.Lock()
	defer hs.Unlock()

	if hs.height.Equal(height) {
		return hs
	}

	hs.height = height
	return hs
}

func (hs *HomeState) Block() hash.Hash {
	hs.RLock()
	defer hs.RUnlock()

	return hs.block
}

func (hs *HomeState) SetBlock(block hash.Hash) *HomeState {
	hs.Lock()
	defer hs.Unlock()

	if hs.block.Equal(block) {
		return hs
	}

	hs.block = block
	return hs
}

func (hs *HomeState) State() node.State {
	hs.RLock()
	defer hs.RUnlock()

	return hs.state
}

func (hs *HomeState) SetState(state node.State) *HomeState {
	hs.Lock()
	defer hs.Unlock()

	if hs.state == state {
		return hs
	}

	hs.state = state
	return hs
}
