package isaac

import (
	"math"
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testVotingBox struct {
	suite.Suite
	homeNode     common.HomeNode
	votingBox    VotingBox
	seals        map[common.Hash]common.Seal
	proposeSeals map[common.Hash]Propose
	ballotSeals  map[common.Hash]Ballot
	policy       ConsensusPolicy
}

func (t *testVotingBox) SetupTest() {
	t.homeNode = common.NewRandomHomeNode()
	t.policy = ConsensusPolicy{NetworkID: common.TestNetworkID, Total: 4, Threshold: 3}
	t.votingBox = NewDefaultVotingBox(t.policy)
	t.seals = map[common.Hash]common.Seal{}
	t.proposeSeals = map[common.Hash]Propose{}
	t.ballotSeals = map[common.Hash]Ballot{}
}

func (t *testVotingBox) newNodes(n uint) []common.HomeNode {
	var nodes []common.HomeNode
	for i := uint(0); i < n; i++ {
		nodes = append(nodes, common.NewRandomHomeNode())
	}

	return nodes
}

func (t *testVotingBox) open(node common.HomeNode, round Round) (common.Hash, error) {
	propose, seal, _ := NewTestSealPropose(node.Address(), nil)
	_, err := t.votingBox.Open(seal)
	if err != nil {
		return common.Hash{}, err
	}

	psHash, _, err := seal.Hash()
	if err != nil {
		return common.Hash{}, err
	}

	t.seals[psHash] = seal
	t.proposeSeals[psHash] = propose

	return psHash, nil
}

func (t *testVotingBox) newBallot(
	node common.HomeNode,
	psHash common.Hash,
	height common.Big,
	stage VoteStage,
	round Round,
	vote Vote,
) (common.Hash, common.Seal, error) {
	ballot, ballotSeal, err := NewTestSealBallot(
		psHash,
		node.Address(),
		height,
		round,
		stage,
		vote,
	)
	if err != nil {
		return common.Hash{}, common.Seal{}, err
	}

	err = ballotSeal.Sign(common.TestNetworkID, node.Seed())
	if err != nil {
		return common.Hash{}, common.Seal{}, err
	}

	sHash, _, err := ballotSeal.Hash()
	if err != nil {
		return common.Hash{}, common.Seal{}, err
	}

	t.seals[sHash] = ballotSeal
	t.ballotSeals[sHash] = ballot

	return sHash, ballotSeal, nil
}

func (t *testVotingBox) newBallotVote(
	node common.HomeNode,
	psHash common.Hash,
	height common.Big,
	stage VoteStage,
	round Round,
	vote Vote,
) (common.Hash, VoteResultInfo, error) {
	sHash, seal, err := t.newBallot(node, psHash, height, stage, round, vote)
	if err != nil {
		return common.Hash{}, VoteResultInfo{}, err
	}

	result, err := t.votingBox.Vote(seal)
	if err != nil {
		return common.Hash{}, result, err
	}

	return sHash, result, nil
}

func (t *testVotingBox) TestNew() {
	votingBox := NewDefaultVotingBox(t.policy)
	t.Nil(votingBox.current)
}

func (t *testVotingBox) TestOpen() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	seal.Sign(common.TestNetworkID, t.homeNode.Seed())

	_, err := t.votingBox.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	votingBox := t.votingBox.(*DefaultVotingBox)
	vp := votingBox.Current()
	t.Equal(psHash, vp.psHash)
	t.Equal(0, votingBox.unknown.Len())

	t.Equal(propose.Block.Height, vp.height)
	t.Equal(propose.Round, vp.round)
	t.Equal(VoteStageINIT, vp.stage)

	t.Equal(votingBox.current.psHash, vp.psHash)
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
	_, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	seal.Sign(common.TestNetworkID, t.homeNode.Seed())

	_, err = t.votingBox.Open(seal)
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
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.votingBox.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		_, _, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote the current proposal
		v1 := common.NewRandomHomeNode()
		v1Vote := VoteNOP

		_, _, err := t.newBallotVote(v1, psHash, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v1.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v1.Address())
		t.True(found)
		t.Equal(v1Vote, sn.vote)
	}

	{ // v2 vote the current proposal
		v2 := common.NewRandomHomeNode()
		v2Vote := VoteNOP

		_, _, err := t.newBallotVote(v2, psHash, propose.Block.Height, VoteStageSIGN, round, v2Vote)
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

	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.votingBox.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		_, _, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote unknown
		psHash1 := common.NewRandomHash("sl")

		v1 := common.NewRandomHomeNode()
		v1Vote := VoteNOP

		sHash, _, err := t.newBallotVote(v1, psHash1, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v1.Address())
		t.Equal(0, len(voted))

		current, unknown := votingBox.Voted(v1.Address())
		t.Equal(0, len(current))
		t.False(unknown.Empty())

		t.Equal(psHash1, unknown.psHash)
		t.Equal(propose.Block.Height, unknown.height)
		t.Equal(round, unknown.round)
		t.Equal(VoteStageSIGN, unknown.stage)
		t.Equal(v1Vote, unknown.vote)

		t.Equal(sHash, unknown.seal)
	}
}

func (t *testVotingBox) TestVoteUnknownCancel() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.votingBox.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		_, _, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote unknown
		psHash1 := common.NewRandomHash("sl")

		v1 := common.NewRandomHomeNode()
		v1Vote := VoteNOP

		sHash, _, err := t.newBallotVote(v1, psHash1, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v1.Address())
		t.Equal(0, len(voted))

		current, unknown := votingBox.Voted(v1.Address())
		t.Equal(0, len(current)) // not voted in current
		t.False(unknown.Empty()) // voted in unknown

		t.Equal(psHash1, unknown.psHash)
		t.Equal(propose.Block.Height, unknown.height)
		t.Equal(round, unknown.round)
		t.Equal(VoteStageSIGN, unknown.stage)
		t.Equal(v1Vote, unknown.vote)

		t.Equal(sHash, unknown.seal)
	}

	// 1. v1 voted in unknown
	// 1. v1 voted in current again
	// vote record in unknown will be removed

	{ // v1 vote the current proposal
		v1 := common.NewRandomHomeNode()
		v1Vote := VoteYES

		_, _, err := t.newBallotVote(v1, psHash, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		current, unknown := votingBox.Voted(v1.Address())
		t.Equal(1, len(current)) // voted in current
		t.True(unknown.Empty())  // voted in unknown
	}
}

// VotingUnknown can contain all kind of vote of validators
func (t *testVotingBox) TestCanCountUnknownSamePSHashAndStage() {
	// In unknown,
	// all votes has,
	// - same psHash
	// - same stage
	vo := NewVotingBoxUnknown()
	t.Equal(0, vo.Len())

	psHash := common.NewRandomHash("sl")
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
				psHash,
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
				psHash,
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
				common.NewRandomHash("sl"), // different psHash
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
				common.NewRandomHash("sl"), // different psHash
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
		address := common.NewRandomHomeNode().Address()
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
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.votingBox.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	var votedSeal common.Seal
	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		sHash, _, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := votingBox.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)

		votedSeal = t.seals[sHash]
	}

	{ // vote again
		_, err = votingBox.Vote(votedSeal)
		t.Error(err, SealAlreadyVotedError)
	}
}

func (t *testVotingBox) TestAlreadyVotedUnknown() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.votingBox.Open(seal)
	t.NoError(err)

	round := Round(0)

	votingBox := t.votingBox.(*DefaultVotingBox)

	var votedSeal common.Seal
	{ // v0 vote the unknown proposal
		psHashUnknown := common.NewRandomHash("sl")
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		sHash, _, err := t.newBallotVote(v0, psHashUnknown, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		_, voted := votingBox.unknown.Voted(v0.Address())
		t.True(voted)

		votedSeal = t.seals[sHash]
	}

	{ // vote again
		_, err = t.votingBox.Vote(votedSeal)
		t.Error(err, SealAlreadyVotedError)
	}
}

func (t *testVotingBox) TestOpenedVoteResultInfoStageINIT() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	result, err := t.votingBox.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	t.Equal(VoteStageINIT, result.Stage)
	t.Equal(VoteResultYES, result.Result)
	t.Equal(psHash, result.Proposal)
	t.Equal(propose.Block.Height, result.Height)
	t.Equal(propose.Round, result.Round)
}

// TestMajorityInUnknownCloseCurrent checks,
// - current is running
// - unknown reaches consensus
//    -  same block height
//    -  different round
// - current should be closed and open new current
func (t *testVotingBox) TestMajorityInUnknownCloseCurrent() {
	// open new current
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	psHash, _, _ := seal.Hash()

	t.votingBox.Open(seal)

	{ // node votes the current proposal
		_, _, err := t.newBallotVote(
			common.NewRandomHomeNode(),
			psHash,
			propose.Block.Height,
			VoteStageSIGN,
			propose.Round,
			VoteYES,
		)
		t.NoError(err)
	}
	votingBox := t.votingBox.(*DefaultVotingBox)
	t.NotNil(votingBox.Current())
	t.Nil(votingBox.Previous())

	otherPSHash := common.NewRandomHash("sl")
	otherRound := propose.Round + 33
	otherStage := VoteStageACCEPT

	var result VoteResultInfo
	{ // others vote the unknown over threshold
		for i := 0; i < int(t.policy.Threshold); i++ {
			other := common.NewRandomHomeNode()

			_, r, err := t.newBallotVote(
				other,
				otherPSHash,
				propose.Block.Height,
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

	t.Nil(votingBox.Current())                        // current is closed
	t.True(psHash.Equal(votingBox.Previous().psHash)) // current moves to previous

	t.Equal(otherPSHash, result.Proposal)
	t.Equal(otherRound, result.Round)
	t.Equal(VoteResultYES, result.Result)
	t.Equal(otherStage, result.Stage)
}

func (t *testVotingBox) TestNewRoundSignVote() {
	// open new current
	_, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	psHash, _, _ := seal.Hash()

	t.votingBox.Open(seal)

	votingBox := t.votingBox.(*DefaultVotingBox)

	t.False(votingBox.current.SealVoted(psHash))
	t.False(votingBox.current.Stage(VoteStageSIGN).SealVoted(psHash))

	_, found := votingBox.current.Stage(VoteStageSIGN).Voted(t.homeNode.Address())
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
