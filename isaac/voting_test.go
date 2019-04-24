package isaac

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
)

type testVotingStage struct {
	suite.Suite
}

func (t *testVotingStage) newSeed() common.Seed {
	return common.RandomSeed()
}

func (t *testVotingStage) newSealHash() common.Hash {
	hash, err := common.NewHashFromObject("sl", common.RandomUUID())
	t.NoError(err)

	return hash
}

func (t *testVotingStage) makeNodes(c uint) []common.Seed {
	var nodes []common.Seed
	for i := 0; i < int(c); i++ {
		nodes = append(nodes, t.newSeed())
	}

	return nodes
}

func (t *testVotingStage) TestVote() {
	st := NewVotingStage()

	var nodeCount uint = 5
	nodes := t.makeNodes(nodeCount)

	{
		st.Vote(t.newSealHash(), nodes[0].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[1].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[2].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[3].Address(), VoteNOP)

		yes := st.YES()
		nop := st.NOP()
		exp := st.EXP()
		t.Equal(3, len(yes))
		t.Equal(1, len(nop))
		t.Equal(0, len(exp))
	}
}

func (t *testVotingStage) TestMultipleVote() {
	st := NewVotingStage()

	var nodeCount uint = 5
	nodes := t.makeNodes(nodeCount)

	{ // node3 vote again with same vote
		st.Vote(t.newSealHash(), nodes[0].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[1].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[2].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[3].Address(), VoteNOP)

		st.Vote(t.newSealHash(), nodes[3].Address(), VoteNOP)

		yes := st.YES()
		nop := st.NOP()
		exp := st.EXP()

		// result is not changed
		t.Equal(3, len(yes))
		t.Equal(1, len(nop))
		t.Equal(0, len(exp))
	}

	{ // node3 overturns it's vote
		st.Vote(t.newSealHash(), nodes[0].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[1].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[2].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[3].Address(), VoteNOP)

		st.Vote(t.newSealHash(), nodes[3].Address(), VoteEXPIRE)

		yes := st.YES()
		nop := st.NOP()
		exp := st.EXP()

		// previous vote will be canceled
		t.Equal(3, len(yes))
		t.Equal(0, len(nop))
		t.Equal(1, len(exp))
	}
}

func (t *testVotingStage) TestCanCount() {
	st := NewVotingStage()

	var total uint = 5
	threshold := uint(math.Round(float64(5) * float64(0.67)))
	nodes := t.makeNodes(total)

	{ // under threshold
		st.Vote(t.newSealHash(), nodes[0].Address(), VoteYES)
		st.Vote(t.newSealHash(), nodes[1].Address(), VoteYES)

		t.Equal(2, st.Count())
		canCount := st.CanCount(total, threshold)
		t.False(canCount)
		majority := st.Majority(total, threshold)
		t.Equal(VoteResultNotYet, majority)
	}

	{ // vote count is over threshold, but draw
		st.Vote(t.newSealHash(), nodes[2].Address(), VoteNOP)

		t.Equal(3, st.Count())
		canCount := st.CanCount(total, threshold)
		t.False(canCount)
		majority := st.Majority(total, threshold)
		t.Equal(VoteResultNotYet, majority)
	}

	{ // vote count is over threshold, and yes
		st.Vote(t.newSealHash(), nodes[3].Address(), VoteYES)

		t.Equal(4, st.Count())
		canCount := st.CanCount(total, threshold)
		t.True(canCount)
		majority := st.Majority(total, threshold)
		t.Equal(VoteResultYES, majority)
	}

	{ // yes=2 nop=2 exp=1 draw
		st.Vote(t.newSealHash(), nodes[3].Address(), VoteNOP)
		st.Vote(t.newSealHash(), nodes[4].Address(), VoteEXPIRE)

		t.Equal(5, st.Count())
		canCount := st.CanCount(total, threshold)
		t.True(canCount)
		majority := st.Majority(total, threshold)
		t.Equal(VoteResultDRAW, majority)
	}
}

func TestVotingStage(t *testing.T) {
	suite.Run(t, new(testVotingStage))
}

type testRoundVoting struct {
	suite.Suite
}

func (t *testRoundVoting) newProposeBallotSeal(seed common.Seed) (common.Seal, ProposeBallot) {
	proposeBallot, proposeBallotSeal, err := NewTestSealProposeBallot(seed.Address(), nil)
	t.NoError(err)
	err = proposeBallotSeal.Sign(common.TestNetworkID, seed)
	t.NoError(err)

	return proposeBallotSeal, proposeBallot
}

func (t *testRoundVoting) TestNew() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	proposeBallotSeal, proposeBallot := t.newProposeBallotSeal(proposerSeed)
	t.Equal(1, proposeBallot.Block.Height.Cmp(common.NewBig(0)))

	vp, _, err := vm.Open(proposeBallotSeal)
	t.NoError(err)
	t.NotEmpty(vp)
	t.Equal(proposeBallot.Block.Height, vp.height)

	proposeBallotSealHash, _, err := proposeBallotSeal.Hash()
	t.True(vm.IsRunning(proposeBallotSealHash))
}

func (t *testRoundVoting) TestNewRoundSignVote() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	proposeBallotSeal, _ := t.newProposeBallotSeal(proposerSeed)

	vp, _, err := vm.Open(proposeBallotSeal)
	t.NoError(err)
	t.NotEmpty(vp)

	var vote Vote
	var voted bool

	vote, voted = vp.Stage(VoteStageINIT).Voted(proposerSeed.Address())
	t.Equal(VoteNONE, vote)
	t.False(voted)

	// ProposeBallot will be automatically voted in sign stage
	vote, voted = vp.Stage(VoteStageSIGN).Voted(proposerSeed.Address())
	t.Equal(VoteYES, vote)
	t.True(voted)

	vote, voted = vp.Stage(VoteStageACCEPT).Voted(proposerSeed.Address())
	t.Equal(VoteNONE, vote)
	t.False(voted)
}

func (t *testRoundVoting) TestVoteBeforePropose() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	proposeBallotSeal, _ := t.newProposeBallotSeal(proposerSeed)
	proposeBallotSealHash, _, err := proposeBallotSeal.Hash()
	t.NoError(err)

	voteSeed := common.RandomSeed()
	voteBallot, _, err := NewTestSealVoteBallot(
		proposeBallotSealHash,
		voteSeed.Address(),
		VoteStageSIGN,
		VoteYES,
	)
	t.NoError(err)

	_, _, err = vm.Vote(voteBallot)
	t.True(VotingProposalNotFoundError.Equal(err))
}

func (t *testRoundVoting) TestVote() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	proposeBallotSeal, _ := t.newProposeBallotSeal(proposerSeed)
	proposeBallotSealHash, _, err := proposeBallotSeal.Hash()
	t.NoError(err)

	vp, _, err := vm.Open(proposeBallotSeal)
	t.NoError(err)

	voteSeed := common.RandomSeed()
	voteBallot, _, err := NewTestSealVoteBallot(
		proposeBallotSealHash,
		voteSeed.Address(),
		VoteStageSIGN,
		VoteYES,
	)
	t.NoError(err)

	_, _, err = vm.Vote(voteBallot)
	t.NoError(err)

	stage := vp.Stage(voteBallot.Stage)
	var vote Vote
	var voted bool

	vote, voted = stage.Voted(proposerSeed.Address())
	t.Equal(VoteYES, vote)
	t.True(voted)

	vote, voted = stage.Voted(voteSeed.Address())
	t.Equal(VoteYES, vote)
	t.True(voted)

	unknownSeed := common.RandomSeed()
	vote, voted = stage.Voted(unknownSeed.Address())
	t.Equal(VoteNONE, vote)
	t.False(voted)
}

func TestRoundVoting(t *testing.T) {
	suite.Run(t, new(testRoundVoting))
}
