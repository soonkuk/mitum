package isaac

import (
	"fmt"
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

func (t *testVotingStage) makeNodes(c int) []common.Seed {
	var nodes []common.Seed
	for i := 0; i < c; i++ {
		nodes = append(nodes, t.newSeed())
	}

	return nodes
}

func (t *testVotingStage) TestVote() {
	st := NewVotingStage()

	nodeCount := 5
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

	nodeCount := 5
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

	total := 5
	threshold := int(math.Round(float64(5) * float64(0.67)))
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

type testRoundVotingManager struct {
	suite.Suite
}

func (t *testRoundVotingManager) TestNew() {
	vm := NewRoundVotingManager()

	var proposeBallotSeal common.Seal
	{
		var err error
		proposerSeed := common.RandomSeed()
		_, proposeBallotSeal, err = NewTestSealProposeBallot(proposerSeed.Address(), nil)
		t.NoError(err)
		err = proposeBallotSeal.Sign(common.TestNetworkID, proposerSeed)
		t.NoError(err)
	}

	fmt.Println(proposeBallotSeal)

	vp, err := vm.NewRound(proposeBallotSeal)
	t.NoError(err)
	t.NotEmpty(vp)
}

func TestRoundVotingManager(t *testing.T) {
	suite.Run(t, new(testRoundVotingManager))
}
