package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
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
	seal := seal.NewRandomSealHash()

	vrs, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)

	t.NoError(err)
	t.NotEmpty(vrs)
	t.True(vrs.IsNodeVoted(n))
}

func (t *testBallotBox) TestBasicVoteRecords() {
	bb := t.newBallotBox()

	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	nextBlock := NewRandomBlockHash()

	var total, threshold uint = 5, 3

	// vote under threshold
	for i := uint(0); i < threshold-1; i++ {
		n := node.NewRandomAddress()
		seal := seal.NewRandomSealHash()

		vrs, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)

		vr, err := vrs.CheckMajority(total, threshold)
		t.NoError(err)
		t.Equal(NotYetMajority, vr.Result())
	}

	{ // vote one more; it should be at least reached to threshold
		n := node.NewRandomAddress()
		seal := seal.NewRandomSealHash()
		vrs, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)
		vr, err := vrs.CheckMajority(total, threshold)
		t.NoError(err)
		t.Equal(GotMajority, vr.Result())
		t.Equal(height, vr.Height())
		t.Equal(round, vr.Round())
		t.Equal(stage, vr.Stage())
		t.Equal(proposal, vr.Proposal())
		t.Equal(nextBlock, vr.NextBlock())
	}
}

func (t *testBallotBox) TestClosedVoteRecords() {
	bb := t.newBallotBox()

	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	nextBlock := NewRandomBlockHash()

	{
		n := node.NewRandomAddress()
		seal := seal.NewRandomSealHash()

		vrs, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)

		// close VoteRecords
		err = bb.CloseVoteRecords(vrs.Hash())
		t.NoError(err)
	}

	{ // vote again
		n := node.NewRandomAddress()
		seal := seal.NewRandomSealHash()

		vrs, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)

		// check closed
		t.True(vrs.IsClosed())
	}

	var total, threshold uint = 5, 3

	{ // vote again, but the closed VoteRecords will not decide result
		n := node.NewRandomAddress()
		seal := seal.NewRandomSealHash()

		vrs, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)

		vr, err := vrs.CheckMajority(total, threshold)
		t.NoError(err)
		t.Equal(FinishedGotMajority, vr.Result())
	}
}

func (t *testBallotBox) TestVoteAgain() {
	bb := t.newBallotBox()

	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	nextBlock := NewRandomBlockHash()

	n := node.NewRandomAddress()

	{ // revoting with same seal; it will not be voted
		seal := seal.NewRandomSealHash()

		_, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)

		_, err = bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.True(xerrors.Is(err, AlreadyVotedError))
	}

	{ // revoting with different seal; it will be voted
		seal := seal.NewRandomSealHash()

		_, err := bb.Vote(n, height, round, stage, proposal, nextBlock, seal)
		t.NoError(err)
	}
}

func TestBallotBox(t *testing.T) {
	suite.Run(t, new(testBallotBox))
}
