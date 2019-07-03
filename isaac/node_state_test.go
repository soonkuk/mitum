package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/node"
)

type testHomeState struct {
	suite.Suite
}

func (t *testHomeState) TestNew() {
	home := node.NewRandomHome()

	block := NewRandomBlock()
	hs := NewHomeState(home, block)

	t.True(hs.Home().Equal(home))
	t.True(hs.Block().Equal(block))
	t.True(hs.Height().Equal(block.Height()))
	t.Equal(node.StateBooting, hs.State())

	t.True(hs.PreviousHeight().IsZero())
	t.True(hs.PreviousBlock().Empty())
	t.True(xerrors.Is(node.InvalidStateError, hs.PreviousState().IsValid()))
}

func (t *testHomeState) TestPrevious() {
	home := node.NewRandomHome()

	previousBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(previousBlock)

	hs := NewHomeState(home, previousBlock)

	previousState := hs.State()
	nextState := node.StateSync

	hs.
		SetBlock(nextBlock).
		SetState(nextState)

	t.True(hs.Home().Equal(home))
	t.True(hs.Block().Equal(nextBlock))
	t.True(hs.Height().Equal(nextBlock.Height()))
	t.Equal(nextState, hs.State())

	t.True(hs.PreviousHeight().Equal(previousBlock.Height()))
	t.True(hs.PreviousBlock().Equal(previousBlock))
	t.Equal(previousState, hs.PreviousState())
}

func TestHomeState(t *testing.T) {
	suite.Run(t, new(testHomeState))
}
