package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/node"
)

type testBallotBox struct {
	suite.Suite
}

func (t *testBallotBox) newBallotBox() *BallotBox {
	return NewBallotBox()
}

func (t *testBallotBox) TestNew() {
	bb := t.newBallotBox()
	t.NotNil(bb)
}

func (t *testBallotBox) TestVote() {
	bb := t.newBallotBox()

	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	nextBlock := NewRandomBlockHash()

	err := bb.Vote(
		n,
		height,
		round,
		stage,
		proposal,
		nextBlock,
	)

	t.NoError(err)
}

func TestBallotBox(t *testing.T) {
	suite.Run(t, new(testBallotBox))
}
