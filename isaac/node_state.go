package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type HomeState struct {
	sync.RWMutex
	home          node.Home
	currentBlock  Block
	previousBlock Block
	currentState  node.State
	previousState node.State
}

func NewHomeState(home node.Home, block Block) *HomeState {
	return &HomeState{
		home:         home,
		currentBlock: block,
		currentState: node.StateBooting, // by default, node state is booting
	}
}

func (hs *HomeState) Home() node.Home {
	return hs.home
}

func (hs *HomeState) Height() Height {
	hs.RLock()
	defer hs.RUnlock()

	return hs.currentBlock.Height()
}

func (hs *HomeState) PreviousHeight() Height {
	hs.RLock()
	defer hs.RUnlock()

	return hs.previousBlock.Height()
}

func (hs *HomeState) SetBlock(block Block) *HomeState {
	hs.Lock()
	defer hs.Unlock()

	if hs.currentBlock.Equal(block) {
		return hs
	}

	hs.previousBlock = hs.currentBlock
	hs.currentBlock = block
	return hs
}

func (hs *HomeState) Block() Block {
	hs.RLock()
	defer hs.RUnlock()

	return hs.currentBlock
}

func (hs *HomeState) PreviousBlock() Block {
	hs.RLock()
	defer hs.RUnlock()

	return hs.previousBlock
}

func (hs *HomeState) State() node.State {
	hs.RLock()
	defer hs.RUnlock()

	return hs.currentState
}

func (hs *HomeState) PreviousState() node.State {
	hs.RLock()
	defer hs.RUnlock()

	return hs.previousState
}

func (hs *HomeState) SetState(state node.State) *HomeState {
	hs.Lock()
	defer hs.Unlock()

	if hs.currentState == state {
		return hs
	}

	hs.previousState = hs.currentState
	hs.currentState = state
	return hs
}

func (hs *HomeState) Proposal() hash.Hash {
	return hs.currentBlock.Proposal()
}
