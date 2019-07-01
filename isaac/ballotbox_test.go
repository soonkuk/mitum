package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
)

type testBallotbox struct {
	suite.Suite
}

func (t *testBallotbox) newBallotbox(total, threshold uint) *Ballotbox {
	return NewBallotbox(NewThreshold(total, threshold))
}

func (t *testBallotbox) newBallot(n node.Address, height Height, round Round, stage Stage, proposal hash.Hash, currentBlock hash.Hash, nextBlock hash.Hash) Ballot {
	ballot, _ := NewBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

	pk, _ := keypair.NewStellarPrivateKey()
	_ = ballot.Sign(pk, []byte{})

	return ballot
}

func (t *testBallotbox) TestNew() {
	bb := t.newBallotbox(10, 7)
	t.NotNil(bb)
}

func (t *testBallotbox) TestVote() {
	bb := t.newBallotbox(10, 7)

	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	currentBlock := NewRandomBlockHash()
	nextBlock := NewRandomBlockHash()

	ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

	vr, err := bb.Vote(ballot)

	t.NoError(err)
	t.NotEmpty(vr)
	t.True(vr.Records().IsNodeVoted(n))
}

func (t *testBallotbox) TestBasicVoteRecords() {
	var total, threshold uint = 5, 3

	bb := t.newBallotbox(total, threshold)

	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	currentBlock := NewRandomBlockHash()
	nextBlock := NewRandomBlockHash()

	// vote under threshold
	for i := uint(0); i < threshold-1; i++ {
		n := node.NewRandomAddress()

		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

		vr, err := bb.Vote(ballot)
		t.NoError(err)
		t.Equal(NotYetMajority, vr.Result())
	}

	{ // vote one more; it should be at least reached to threshold
		n := node.NewRandomAddress()

		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

		vr, err := bb.Vote(ballot)
		t.NoError(err)
		t.Equal(GotMajority, vr.Result())
		t.Equal(height, vr.Height())
		t.Equal(round, vr.Round())
		t.Equal(stage, vr.Stage())
		t.Equal(proposal, vr.Proposal())
		t.Equal(currentBlock, vr.CurrentBlock())
		t.Equal(nextBlock, vr.NextBlock())
	}
}

func (t *testBallotbox) TestClosedVoteRecords() {
	var total, threshold uint = 5, 3
	bb := t.newBallotbox(total, threshold)

	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	currentBlock := NewRandomBlockHash()
	nextBlock := NewRandomBlockHash()

	{
		n := node.NewRandomAddress()

		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

		_, err := bb.Vote(ballot)
		t.NoError(err)

		// close VoteRecords
		bh, _ := bb.boxHash(ballot)
		vrs, found := bb.voted[bh]
		t.True(found)
		vrs.Close()
	}

	{ // vote again
		n := node.NewRandomAddress()

		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)
		_, err := bb.Vote(ballot)
		t.NoError(err)

		// check closed
		bh, _ := bb.boxHash(ballot)
		vrs, found := bb.voted[bh]
		t.True(found)
		t.True(vrs.IsClosed())
	}

	{ // vote again, but the closed VoteRecords will not decide result
		n := node.NewRandomAddress()

		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)
		vr, err := bb.Vote(ballot)
		t.NoError(err)

		t.Equal(FinishedGotMajority, vr.Result())
	}
}

func (t *testBallotbox) TestVoteAgain() {
	bb := t.newBallotbox(10, 7)

	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	currentBlock := NewRandomBlockHash()
	nextBlock := NewRandomBlockHash()

	n := node.NewRandomAddress()

	{ // revoting with same seal; it will not be voted
		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

		_, err := bb.Vote(ballot)
		t.NoError(err)

		_, err = bb.Vote(ballot)
		t.True(xerrors.Is(err, AlreadyVotedError))
	}

	{ // revoting with different seal; it will be voted
		ballot := t.newBallot(n, height, round, stage, proposal, currentBlock, nextBlock)
		_, err := bb.Vote(ballot)
		t.NoError(err)
	}
}

func TestBallotbox(t *testing.T) {
	suite.Run(t, new(testBallotbox))
}
