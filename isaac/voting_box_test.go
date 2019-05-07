package isaac

import (
	"math"
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testVotingBox struct {
	suite.Suite
	home         *common.HomeNode
	votingBox    VotingBox
	seals        map[common.Hash]common.Seal
	proposeSeals map[common.Hash]Proposal
	ballotSeals  map[common.Hash]Ballot
	policy       ConsensusPolicy
}

func (t *testVotingBox) SetupTest() {
	t.home = common.NewRandomHome()
	t.policy = ConsensusPolicy{NetworkID: common.TestNetworkID, Total: 4, Threshold: 3}
	t.votingBox = NewDefaultVotingBox(t.policy)
	t.seals = map[common.Hash]common.Seal{}
	t.proposeSeals = map[common.Hash]Proposal{}
	t.ballotSeals = map[common.Hash]Ballot{}
}

func (t *testVotingBox) newNodes(n uint) []*common.HomeNode {
	var nodes []*common.HomeNode
	for i := uint(0); i < n; i++ {
		nodes = append(nodes, common.NewRandomHome())
	}

	return nodes
}

func (t *testVotingBox) newBallot(
	node *common.HomeNode,
	phash common.Hash,
	height common.Big,
	stage VoteStage,
	round Round,
	vote Vote,
) (Ballot, error) {
	ballot := NewTestSealBallot(
		phash,
		node.Address(),
		height,
		round,
		stage,
		vote,
	)

	err := ballot.Sign(common.TestNetworkID, node.Seed())
	if err != nil {
		return Ballot{}, err
	}

	t.seals[ballot.Hash()] = ballot

	return ballot, nil
}

func (t *testVotingBox) newBallotVote(
	node *common.HomeNode,
	phash common.Hash,
	height common.Big,
	stage VoteStage,
	round Round,
	vote Vote,
) (Ballot, VoteResultInfo, error) {
	ballot, err := t.newBallot(node, phash, height, stage, round, vote)
	if err != nil {
		return Ballot{}, VoteResultInfo{}, err
	}

	result, err := t.votingBox.Vote(ballot)
	if err != nil {
		return Ballot{}, result, err
	}

	return ballot, result, nil
}

func (t *testVotingBox) TestNew() {
	votingBox := NewDefaultVotingBox(t.policy)
	t.Nil(votingBox.current)
}

func (t *testVotingBox) TestOpen() {
	proposal := NewTestProposal(t.home.Address(), nil)
	proposal.Sign(common.TestNetworkID, t.home.Seed())

	_, err := t.votingBox.Open(proposal)
	t.NoError(err)

	votingBox := t.votingBox.(*DefaultVotingBox)
	vp := votingBox.Current()
	t.Equal(proposal.Hash(), vp.proposal)
	t.Equal(0, votingBox.unknown.Len())

	t.Equal(proposal.Block.Height, vp.height)
	t.Equal(proposal.Round, vp.round)
	t.Equal(VoteStageINIT, vp.stage)

	t.Equal(votingBox.current.proposal, vp.proposal)
	t.Equal(votingBox.current.height, vp.height)
	t.Equal(votingBox.current.round, vp.round)
	t.Equal(votingBox.current.stage, vp.stage)
	t.Equal(0, len(votingBox.current.stageINIT.voted))
	t.Equal(0, len(votingBox.current.stageSIGN.voted))
	t.Equal(0, len(votingBox.current.stageACCEPT.voted))
}

func (t *testVotingBox) TestClose() {
	err := t.votingBox.Close()
	t.Error(err, ProposalIsNotOpenedError)

	votingBox := t.votingBox.(*DefaultVotingBox)
	t.Nil(votingBox.current)
	t.Nil(votingBox.previous)

	// after open and then close
	proposal := NewTestProposal(t.home.Address(), nil)
	proposal.Sign(common.TestNetworkID, t.home.Seed())

	_, err = t.votingBox.Open(proposal)
	t.NoError(err)

	t.NotNil(votingBox.current)
	t.Nil(votingBox.previous)

	current := votingBox.Current()

	err = t.votingBox.Close()
	t.NoError(err)
	t.Nil(votingBox.current)
	t.NotNil(votingBox.previous)
	t.Equal(current, votingBox.previous)

	// current moves to previous; all the stages in previous should be closed
	t.True(votingBox.previous.Closed())
	t.True(votingBox.previous.Stage(VoteStageINIT).Closed())
	t.True(votingBox.previous.Stage(VoteStageSIGN).Closed())
	t.True(votingBox.previous.Stage(VoteStageACCEPT).Closed())
}

func (t *testVotingBox) TestVoteCurrent() {
	proposal := NewTestProposal(t.home.Address(), nil)
	_, err := t.votingBox.Open(proposal)
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHome()
		v0Vote := VoteYES

		_, _, err := t.newBallotVote(v0, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote the current proposal
		v1 := common.NewRandomHome()
		v1Vote := VoteNOP

		_, _, err := t.newBallotVote(v1, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v1.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v1.Address())
		t.True(found)
		t.Equal(v1Vote, sn.vote)
	}

	{ // v2 vote the current proposal
		v2 := common.NewRandomHome()
		v2Vote := VoteNOP

		_, _, err := t.newBallotVote(v2, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v2Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v2.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v2.Address())
		t.True(found)
		t.Equal(v2Vote, sn.vote)
	}
}

func (t *testVotingBox) TestVoteUnknown() {
	defer common.DebugPanic()

	proposal := NewTestProposal(t.home.Address(), nil)
	_, err := t.votingBox.Open(proposal)
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHome()
		v0Vote := VoteYES

		_, _, err := t.newBallotVote(v0, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote unknown
		phash1 := common.NewRandomHash("sl")

		v1 := common.NewRandomHome()
		v1Vote := VoteNOP

		sHash, _, err := t.newBallotVote(v1, phash1, proposal.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v1.Address())
		t.Equal(0, len(voted))

		current, unknown := votingBox.Voted(v1.Address())
		t.Equal(0, len(current))
		t.False(unknown.Empty())

		t.Equal(phash1, unknown.proposal)
		t.Equal(proposal.Block.Height, unknown.height)
		t.Equal(round, unknown.round)
		t.Equal(VoteStageSIGN, unknown.stage)
		t.Equal(v1Vote, unknown.vote)

		t.Equal(sHash, unknown.seal)
	}
}

func (t *testVotingBox) TestVoteUnknownCancel() {
	proposal := NewTestProposal(t.home.Address(), nil)
	_, err := t.votingBox.Open(proposal)
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHome()
		v0Vote := VoteYES

		_, _, err := t.newBallotVote(v0, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote unknown
		phash1 := common.NewRandomHash("sl")

		v1 := common.NewRandomHome()
		v1Vote := VoteNOP

		sHash, _, err := t.newBallotVote(v1, phash1, proposal.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v1.Address())
		t.Equal(0, len(voted))

		current, unknown := votingBox.Voted(v1.Address())
		t.Equal(0, len(current)) // not voted in current
		t.False(unknown.Empty()) // voted in unknown

		t.Equal(phash1, unknown.proposal)
		t.Equal(proposal.Block.Height, unknown.height)
		t.Equal(round, unknown.round)
		t.Equal(VoteStageSIGN, unknown.stage)
		t.Equal(v1Vote, unknown.vote)

		t.Equal(sHash, unknown.seal)
	}

	// 1. v1 voted in unknown
	// 1. v1 voted in current again
	// vote record in unknown will be removed

	{ // v1 vote the current proposal
		v1 := common.NewRandomHome()
		v1Vote := VoteYES

		_, _, err := t.newBallotVote(v1, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		current, unknown := votingBox.Voted(v1.Address())
		t.Equal(1, len(current)) // voted in current
		t.True(unknown.Empty())  // voted in unknown
	}
}

// VotingUnknown can contain all kind of vote of validators
func (t *testVotingBox) TestCanCountUnknownSameProposalAndStage() {
	// In unknown,
	// all votes has,
	// - same proposal
	// - same stage
	vo := NewVotingBoxUnknown()
	t.Equal(0, vo.Len())

	phash := common.NewRandomHash("sl")
	height := common.NewBig(200)
	stage := VoteStageACCEPT
	round := Round(99)

	var total uint = 4
	var threshold uint = 3
	nodes := t.newNodes(total)

	{ // vote under threshold
		for i, n := range nodes {
			if i == int(threshold)-1 {
				break
			}
			_, err := vo.Vote(
				phash,
				n.Address(),
				height,
				round,
				stage,
				VoteYES,
				common.NewRandomHash("sl"),
			)
			t.NoError(err)
			_, voted := vo.Voted(n.Address())
			t.True(voted)
		}

		canCount := vo.CanCount(total, threshold)
		t.False(canCount)

		r := vo.Majority(total, threshold)
		t.True(r.NotYet())
	}

	{ // vote over threshold
		for i, n := range nodes {
			if i < int(threshold)-1 {
				continue
			}
			_, err := vo.Vote(
				phash,
				n.Address(),
				height,
				round,
				stage,
				VoteYES,
				common.NewRandomHash("sl"),
			)
			t.NoError(err)
			_, voted := vo.Voted(n.Address())
			t.True(voted)
		}

		canCount := vo.CanCount(total, threshold)
		t.True(canCount)

		r := vo.Majority(total, threshold)
		t.False(r.NotYet())
		t.Equal(VoteResultYES, r.Result)
	}
}

func (t *testVotingBox) TestCanCountUnknownINITSameHeightAndRound() {
	// In unknown,
	// all votes has,
	// - same height
	// - same round

	vo := NewVotingBoxUnknown()
	t.Equal(0, vo.Len())

	height := common.NewBig(200)
	round := Round(99)

	var total uint = 4
	var threshold uint = 3
	nodes := t.newNodes(total)

	{ // vote under threshold
		for i, n := range nodes {
			if i == int(threshold)-1 {
				break
			}

			_, err := vo.Vote(
				common.NewRandomHash("sl"), // different proposal
				n.Address(),
				height,
				round,
				VoteStageINIT,
				VoteYES, // NOTE should be yes
				common.NewRandomHash("sl"),
			)
			t.NoError(err)
			_, voted := vo.Voted(n.Address())
			t.True(voted)
		}

		canCount := vo.CanCount(total, threshold)
		t.False(canCount)

		resultFromInit := vo.MajorityINIT(total, threshold)
		t.True(resultFromInit.NotYet())

		result := vo.Majority(total, threshold)
		t.True(result.NotYet())

		t.True(result.Proposal.Equal(resultFromInit.Proposal))
		t.True(result.Height.Equal(resultFromInit.Height))
		t.Equal(result.Round, resultFromInit.Round)
		t.Equal(result.Stage, resultFromInit.Stage)
	}

	{ // vote over threshold
		for i, n := range nodes {
			if i < int(threshold)-1 {
				continue
			}

			_, err := vo.Vote(
				common.NewRandomHash("sl"), // different proposal
				n.Address(),
				height,
				round,
				VoteStageINIT,
				VoteYES, // NOTE should be yes
				common.NewRandomHash("sl"),
			)
			t.NoError(err)
			_, voted := vo.Voted(n.Address())
			t.True(voted)
		}

		canCount := vo.CanCount(total, threshold)
		t.True(canCount)

		r := vo.MajorityINIT(total, threshold)
		t.False(r.NotYet())
		t.Equal(height, r.Height)
		t.Equal(round, r.Round)
	}
}

func (t *testVotingBox) TestCloseVotingBoxStage() {
	st := NewVotingBoxStage(common.NewRandomHash("sl"), common.NewBig(33), Round(0), VoteStageSIGN)
	t.False(st.Closed())

	st.Close()
	t.True(st.Closed())

	node := t.newNodes(1)[0]

	// even closed, vote can be possible
	st.Vote(node.Address(), VoteYES, common.NewRandomHash("sl"))
}

func (t *testVotingBox) TestCloseUnknown() {
	un := NewVotingBoxUnknown()

	var created []common.Time
	for i := 0; i < 5; i++ {
		address := common.NewRandomHome().Address()
		_, err := un.Vote(
			common.NewRandomHash("sl"),
			address,
			common.NewBig(1),
			Round(0),
			VoteStageSIGN,
			VoteYES,
			common.NewRandomHash("sl"),
		)
		t.NoError(err)

		unv, _ := un.Voted(address)
		created = append(created, unv.votedAt)
	}

	un.ClearBefore(created[3])
	t.Equal(2, len(un.voted))
}

func (t *testVotingBox) TestAlreadyVotedCurrent() {
	proposal := NewTestProposal(t.home.Address(), nil)
	_, err := t.votingBox.Open(proposal)
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	var votedSeal common.Seal
	{ // v0 vote the current proposal
		v0 := common.NewRandomHome()
		v0Vote := VoteYES

		var ballot Ballot
		ballot, _, err = t.newBallotVote(v0, proposal.Hash(), proposal.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)

		votedSeal = t.seals[ballot.Hash()]
	}

	{ // vote again
		_, err = votingBox.Vote(votedSeal.(Ballot))
		t.Error(err, SealAlreadyVotedError)
	}
}

func (t *testVotingBox) TestAlreadyVotedUnknown() {
	proposal := NewTestProposal(t.home.Address(), nil)
	_, err := t.votingBox.Open(proposal)
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	var votedSeal common.Seal
	{ // v0 vote the unknown proposal
		phashUnknown := common.NewRandomHash("sl")
		v0 := common.NewRandomHome()
		v0Vote := VoteYES

		var ballot Ballot
		ballot, _, err = t.newBallotVote(v0, phashUnknown, proposal.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		_, voted := votingBox.unknown.Voted(v0.Address())
		t.True(voted)

		votedSeal = t.seals[ballot.Hash()]
	}

	{ // vote again
		_, err = t.votingBox.Vote(votedSeal.(Ballot))
		t.Error(err, SealAlreadyVotedError)
	}
}

func (t *testVotingBox) TestOpenedVoteResultInfoStageINIT() {
	proposal := NewTestProposal(t.home.Address(), nil)
	result, err := t.votingBox.Open(proposal)
	t.NoError(err)

	t.Equal(VoteStageINIT, result.Stage)
	t.Equal(VoteResultYES, result.Result)
	t.Equal(proposal.Hash(), result.Proposal)
	t.Equal(proposal.Block.Height, result.Height)
	t.Equal(proposal.Round, result.Round)
}

// TestMajorityInUnknownCloseCurrent checks,
// - current is running
// - unknown reaches consensus
//    -  same block height
//    -  different round
// - current should be closed and open new current
func (t *testVotingBox) TestMajorityInUnknownCloseCurrent() {
	// open new current
	proposal := NewTestProposal(t.home.Address(), nil)

	t.votingBox.Open(proposal)

	{ // node votes the current proposal
		_, _, err := t.newBallotVote(
			common.NewRandomHome(),
			proposal.Hash(),
			proposal.Block.Height,
			VoteStageSIGN,
			proposal.Round,
			VoteYES,
		)
		t.NoError(err)
	}
	votingBox := t.votingBox.(*DefaultVotingBox)
	t.NotNil(votingBox.Current())
	t.Nil(votingBox.Previous())

	otherPhash := common.NewRandomHash("sl")
	otherRound := proposal.Round + 33
	otherStage := VoteStageACCEPT

	var result VoteResultInfo
	{ // others vote the unknown over threshold
		for i := 0; i < int(t.policy.Threshold); i++ {
			other := common.NewRandomHome()

			_, r, err := t.newBallotVote(
				other,
				otherPhash,
				proposal.Block.Height,
				otherStage,
				otherRound,
				VoteYES,
			)
			t.NoError(err)
			if r.NotYet() {
				continue
			}

			result = r
			break
		}
	}
	t.False(result.NotYet())

	t.Nil(votingBox.Current())                                   // current is closed
	t.True(proposal.Hash().Equal(votingBox.Previous().proposal)) // current moves to previous

	t.Equal(otherPhash, result.Proposal)
	t.Equal(otherRound, result.Round)
	t.Equal(VoteResultYES, result.Result)
	t.Equal(otherStage, result.Stage)
}

func (t *testVotingBox) TestNewRoundSignVote() {
	// open new current
	proposal := NewTestProposal(t.home.Address(), nil)

	t.votingBox.Open(proposal)

	votingBox := t.votingBox.(*DefaultVotingBox)

	t.False(votingBox.current.SealVoted(proposal.Hash()))
	t.False(votingBox.current.Stage(VoteStageSIGN).SealVoted(proposal.Hash()))

	_, found := votingBox.current.Stage(VoteStageSIGN).Voted(t.home.Address())
	t.False(found)
}

func TestVotingBox(t *testing.T) {
	suite.Run(t, new(testVotingBox))
}

type testVotingBoxStage struct {
	suite.Suite
}

func (t *testVotingBoxStage) newSeed() common.Seed {
	return common.RandomSeed()
}

func (t *testVotingBoxStage) makeNodes(c uint) []common.Seed {
	var nodes []common.Seed
	for i := 0; i < int(c); i++ {
		nodes = append(nodes, t.newSeed())
	}

	return nodes
}

func (t *testVotingBoxStage) newVotingBoxStage() *VotingBoxStage {
	return NewVotingBoxStage(common.NewRandomHash("sl"), common.NewBig(33), Round(0), VoteStageSIGN)
}

func (t *testVotingBoxStage) TestVote() {
	st := t.newVotingBoxStage()

	var nodeCount uint = 5
	nodes := t.makeNodes(nodeCount)

	{
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[2].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		yes, nop := st.VoteCount()
		t.Equal(3, yes)
		t.Equal(1, nop)
	}
}

func (t *testVotingBoxStage) TestMultipleVote() {
	st := t.newVotingBoxStage()

	var nodeCount uint = 5
	nodes := t.makeNodes(nodeCount)

	{ // node3 vote again with same vote
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[2].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		yes, nop := st.VoteCount()

		// result is not changed
		t.Equal(3, yes)
		t.Equal(1, nop)
	}

	{ // node3 overturns it's vote
		st.Vote(nodes[0].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[1].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[2].Address(), VoteYES, common.NewRandomHash("sl"))
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		yes, nop := st.VoteCount()
		t.Equal(3, yes)
		t.Equal(1, nop)

		st.Vote(nodes[3].Address(), VoteYES, common.NewRandomHash("sl"))

		yes, nop = st.VoteCount()

		// previous vote will be canceled
		t.Equal(4, yes)
		t.Equal(0, nop)
	}
}

func (t *testVotingBoxStage) TestCanCount() {
	st := t.newVotingBoxStage()

	var total uint = 4
	threshold := uint(math.Round(float64(4) * float64(0.67)))
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
		st.Vote(nodes[3].Address(), VoteNOP, common.NewRandomHash("sl"))

		t.Equal(4, st.Count())
		canCount := st.CanCount(total, threshold)
		t.True(canCount)
		ri := st.Majority(total, threshold)
		t.Equal(VoteResultDRAW, ri.Result)
	}

	{ // vote count is over threshold, and yes
		st.Vote(nodes[3].Address(), VoteYES, common.NewRandomHash("sl"))

		t.Equal(4, st.Count())
		canCount := st.CanCount(total, threshold)
		t.True(canCount)
		ri := st.Majority(total, threshold)
		t.Equal(VoteResultYES, ri.Result)
	}
}

func TestVotingBoxStage(t *testing.T) {
	suite.Run(t, new(testVotingBoxStage))
}
