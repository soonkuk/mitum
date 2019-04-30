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

func (t *testVotingStage) makeNodes(c uint) []common.Seed {
	var nodes []common.Seed
	for i := 0; i < int(c); i++ {
		nodes = append(nodes, t.newSeed())
	}

	return nodes
}

func (t *testVotingStage) newVotingStage() *VotingStage {
	return NewVotingStage(common.NewRandomHash("sl"), common.NewBig(33), Round(0), VoteStageSIGN)
}

func (t *testVotingStage) TestVote() {
	st := t.newVotingStage()

	var nodeCount uint = 5
	nodes := t.makeNodes(nodeCount)

	{
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[2].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		yes, nop, exp := st.VoteCount()
		t.Equal(3, yes)
		t.Equal(1, nop)
		t.Equal(0, exp)
	}
}

func (t *testVotingStage) TestMultipleVote() {
	st := t.newVotingStage()

	var nodeCount uint = 5
	nodes := t.makeNodes(nodeCount)

	{ // node3 vote again with same vote
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[2].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		yes, nop, exp := st.VoteCount()

		// result is not changed
		t.Equal(3, yes)
		t.Equal(1, nop)
		t.Equal(0, exp)
	}

	{ // node3 overturns it's vote
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[2].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		st.Vote(nodes[3].Address(), VoteEXPIRE, common.NewRandomHash("sl"))

		yes, nop, exp := st.VoteCount()

		// previous vote will be canceled
		t.Equal(3, yes)
		t.Equal(0, nop)
		t.Equal(1, exp)
	}
}

func (t *testVotingStage) TestCanCount() {
	st := t.newVotingStage()

	var total uint = 5
	threshold := uint(math.Round(float64(5) * float64(0.67)))
	nodes := t.makeNodes(total)

	{ // under threshold
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))

		t.Equal(2, st.Count())
		canCount := st.CanCount(total, threshold)
		t.False(canCount)
		ri := st.Majority(total, threshold)
		t.Equal(VoteResultNotYet, ri.Result)
	}

	{ // vote count is over threshold, but draw
		st.Vote(nodes[2].Address(), VoteNOP, common.NewRandomHash("sl"))

		t.Equal(3, st.Count())
		canCount := st.CanCount(total, threshold)
		t.False(canCount)
		ri := st.Majority(total, threshold)
		t.Equal(VoteResultNotYet, ri.Result)
	}

	{ // vote count is over threshold, and yes
		st.Vote(nodes[3].Address(), VoteYES, common.NewRandomHash("sl"))

		t.Equal(4, st.Count())
		canCount := st.CanCount(total, threshold)
		t.True(canCount)
		ri := st.Majority(total, threshold)
		t.Equal(VoteResultYES, ri.Result)
	}

	{ // yes=2 nop=2 exp=1 draw
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))
		st.Vote(nodes[4].Address(), VoteEXPIRE, common.NewRandomHash("sl"))

		t.Equal(5, st.Count())
		canCount := st.CanCount(total, threshold)
		t.True(canCount)
		ri := st.Majority(total, threshold)
		t.Equal(VoteResultDRAW, ri.Result)
	}
}

func TestVotingStage(t *testing.T) {
	suite.Run(t, new(testVotingStage))
}

type testRoundVoting struct {
	suite.Suite
}

func (t *testRoundVoting) newProposeSeal(seed common.Seed) (common.Seal, Propose) {
	Propose, ProposeSeal, err := NewTestSealPropose(seed.Address(), nil)
	t.NoError(err)
	err = ProposeSeal.Sign(common.TestNetworkID, seed)
	t.NoError(err)

	return ProposeSeal, Propose
}

func (t *testRoundVoting) TestNew() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	ProposeSeal, Propose := t.newProposeSeal(proposerSeed)
	t.Equal(1, Propose.Block.Height.Cmp(common.NewBig(0)))

	vp, err := vm.Open(ProposeSeal)
	t.NoError(err)
	t.NotEmpty(vp)
	t.Equal(Propose.Block.Height, vp.height)

	psHash, _, err := ProposeSeal.Hash()
	t.Equal(psHash, vm.Current().psHash)
}

func (t *testRoundVoting) TestNewRoundSignVote() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	ProposeSeal, _ := t.newProposeSeal(proposerSeed)

	vp, err := vm.Open(ProposeSeal)
	t.NoError(err)
	t.NotEmpty(vp)

	var vote VotingStageNode
	var voted bool

	vote, voted = vp.Stage(VoteStageINIT).Voted(proposerSeed.Address())
	t.Equal(VoteNONE, vote.vote)
	t.False(voted)

	// Propose will be automatically voted in sign stage
	vote, voted = vp.Stage(VoteStageSIGN).Voted(proposerSeed.Address())
	t.Equal(VoteYES, vote.vote)
	t.True(voted)

	vote, voted = vp.Stage(VoteStageACCEPT).Voted(proposerSeed.Address())
	t.Equal(VoteNONE, vote.vote)
	t.False(voted)
}

func (t *testRoundVoting) TestVoteBeforePropose() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	ProposeSeal, _ := t.newProposeSeal(proposerSeed)
	psHash, _, err := ProposeSeal.Hash()
	t.NoError(err)

	voteSeed := common.RandomSeed()
	_, ballotSeal, err := NewTestSealBallot(
		psHash,
		voteSeed.Address(),
		common.NewBig(1),
		Round(1),
		VoteStageSIGN,
		VoteYES,
	)
	t.NoError(err)

	_, err = vm.Vote(ballotSeal)
	_, voted := vm.Unknown().Voted(voteSeed.Address())
	t.True(voted)
}

func (t *testRoundVoting) TestVote() {
	vm := NewRoundVoting()

	proposerSeed := common.RandomSeed()
	ProposeSeal, _ := t.newProposeSeal(proposerSeed)
	psHash, _, err := ProposeSeal.Hash()
	t.NoError(err)

	vp, err := vm.Open(ProposeSeal)
	t.NoError(err)

	voteSeed := common.RandomSeed()
	ballot, ballotSeal, err := NewTestSealBallot(
		psHash,
		voteSeed.Address(),
		common.NewBig(1),
		Round(1),
		VoteStageSIGN,
		VoteYES,
	)
	t.NoError(err)

	_, err = vm.Vote(ballotSeal)
	t.NoError(err)

	stage := vp.Stage(ballot.Stage)
	var vote VotingStageNode
	var voted bool

	vote, voted = stage.Voted(proposerSeed.Address())
	t.Equal(VoteYES, vote.vote)
	t.True(voted)

	vote, voted = stage.Voted(voteSeed.Address())
	t.Equal(VoteYES, vote.vote)
	t.True(voted)

	unknownSeed := common.RandomSeed()
	vote, voted = stage.Voted(unknownSeed.Address())
	t.Equal(VoteNONE, vote.vote)
	t.False(voted)
}

func TestRoundVoting(t *testing.T) {
	suite.Run(t, new(testRoundVoting))
}
