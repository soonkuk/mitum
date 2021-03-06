package isaac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/node"
)

type testCompilerBallotChecker struct {
	suite.Suite
}

func (t *testCompilerBallotChecker) TestEmptyLastVoteResult() {
	home := node.NewRandomHome()
	lastBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(lastBlock)

	homeState := NewHomeState(home, lastBlock)

	ballot, _ := NewINITBallot(
		home.Address(),
		lastBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Height().Add(1),
		nextBlock.Hash(),
		Round(1),
		nextBlock.Proposal(),
	)

	suffrage := NewFixedProposerSuffrage(home, home)
	checker := NewCompilerBallotChecker(homeState, suffrage)
	err := checker.
		New(context.TODO()).
		SetContext("ballot", ballot).
		SetContext("lastINITVoteResult", VoteResult{}).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.NoError(err)
}

func (t *testCompilerBallotChecker) TestINITBallotHeightNotHigherThanHomeState() {
	home := node.NewRandomHome()
	lastBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(lastBlock)

	homeState := NewHomeState(home, lastBlock)

	ballot, _ := NewINITBallot(
		home.Address(),
		lastBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Height().Sub(1),
		nextBlock.Hash(),
		Round(0),
		nextBlock.Proposal(),
	)

	suffrage := NewFixedProposerSuffrage(home, home)
	checker := NewCompilerBallotChecker(homeState, suffrage)
	err := checker.
		New(context.TODO()).
		SetContext("ballot", ballot).
		SetContext("lastINITVoteResult", VoteResult{}).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.Contains(err.Error(), "lower ballot height")
}

func (t *testCompilerBallotChecker) TestINITBallotHeightLowerThanLastINITVoteResult() {
	home := node.NewRandomHome()
	lastBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(lastBlock)

	homeState := NewHomeState(home, lastBlock)

	lastINITVoteResult := NewVoteResult(
		lastBlock.Height(),
		lastBlock.Round(),
		StageINIT,
	)
	lastINITVoteResult = lastINITVoteResult.
		SetAgreement(Majority)

	ballot, _ := NewINITBallot(
		home.Address(),
		lastBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Height().Sub(1),
		nextBlock.Hash(),
		Round(0),
		nextBlock.Proposal(),
	)

	suffrage := NewFixedProposerSuffrage(home, home)
	checker := NewCompilerBallotChecker(homeState, suffrage)
	err := checker.
		New(context.TODO()).
		SetContext("ballot", ballot).
		SetContext("lastINITVoteResult", lastINITVoteResult).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.Contains(err.Error(), "lower ballot height")
}

func (t *testCompilerBallotChecker) TestSIGNBallotHeightNotSameWithLastINITVoteResult() {
	home := node.NewRandomHome()
	lastBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(lastBlock)

	homeState := NewHomeState(home, lastBlock)

	ballot, _ := NewSIGNBallot(
		home.Address(),
		lastBlock.Hash(),
		lastBlock.Round(),
		nextBlock.Height(),
		nextBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Proposal(),
	)

	lastINITVoteResult := NewVoteResult(
		lastBlock.Height(),
		nextBlock.Round(),
		StageINIT,
	)
	lastINITVoteResult = lastINITVoteResult.
		SetAgreement(Majority)

	suffrage := NewFixedProposerSuffrage(home, home)
	checker := NewCompilerBallotChecker(homeState, suffrage)
	err := checker.
		New(context.TODO()).
		SetContext("ballot", ballot).
		SetContext("lastINITVoteResult", lastINITVoteResult).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.Contains(err.Error(), "lower ballot height")
}

func (t *testCompilerBallotChecker) TestSIGNBallotRoundNotSameWithLastINITVoteResult() {
	home := node.NewRandomHome()
	lastBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(lastBlock)

	homeState := NewHomeState(home, lastBlock)

	ballot, _ := NewSIGNBallot(
		home.Address(),
		lastBlock.Hash(),
		lastBlock.Round(),
		nextBlock.Height(),
		nextBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Proposal(),
	)

	lastINITVoteResult := NewVoteResult(
		nextBlock.Height(),
		nextBlock.Round()-1,
		StageINIT,
	)
	lastINITVoteResult = lastINITVoteResult.
		SetAgreement(Majority)

	suffrage := NewFixedProposerSuffrage(home, home)
	checker := NewCompilerBallotChecker(homeState, suffrage)
	err := checker.
		New(context.TODO()).
		SetContext("ballot", ballot).
		SetContext("lastINITVoteResult", lastINITVoteResult).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.Contains(err.Error(), "lower ballot round")
}

func (t *testCompilerBallotChecker) TestBallotFromNotActingNode() {
	home := node.NewRandomHome()
	lastBlock := NewRandomBlock()
	nextBlock := NewRandomNextBlock(lastBlock)

	homeState := NewHomeState(home, lastBlock)

	lastINITVoteResult := NewVoteResult(
		nextBlock.Height(),
		nextBlock.Round(),
		StageINIT,
	)
	lastINITVoteResult = lastINITVoteResult.
		SetAgreement(Majority)

	suffrage := NewFixedProposerSuffrage(home, home)
	checker := NewCompilerBallotChecker(homeState, suffrage)

	// ballot from in acting
	ballotInActing, _ := NewSIGNBallot(
		home.Address(),
		lastBlock.Hash(),
		lastBlock.Round(),
		nextBlock.Height(),
		nextBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Proposal(),
	)

	err := checker.
		New(context.TODO()).
		SetContext("ballot", ballotInActing).
		SetContext("lastINITVoteResult", lastINITVoteResult).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.NoError(err)

	// ballot from not in acting
	other := node.NewRandomHome()
	ballotOther, _ := NewSIGNBallot(
		other.Address(),
		lastBlock.Hash(),
		lastBlock.Round(),
		nextBlock.Height(),
		nextBlock.Hash(),
		nextBlock.Round(),
		nextBlock.Proposal(),
	)

	err = checker.
		New(context.TODO()).
		SetContext("ballot", ballotOther).
		SetContext("lastINITVoteResult", lastINITVoteResult).
		SetContext("lastStagesVoteResult", VoteResult{}).
		Check()
	t.Contains(err.Error(), "not in acting suffrage")
}

func TestCompilerBallotChecker(t *testing.T) {
	suite.Run(t, new(testCompilerBallotChecker))
}
