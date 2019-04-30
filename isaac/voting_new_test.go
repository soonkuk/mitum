package isaac

import (
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testVotingNew struct {
	suite.Suite
	homeNode     common.HomeNode
	rv           *RoundVoting
	seals        map[common.Hash]common.Seal
	proposeSeals map[common.Hash]Propose
	ballotSeals  map[common.Hash]Ballot
	policy       ConsensusPolicy
}

func (t *testVotingNew) SetupTest() {
	t.homeNode = common.NewRandomHomeNode()
	t.policy := ConsensusPolicy{NetworkID: common.TestNetworkID, Total: 4, Threshold: 3}
	t.rv = NewRoundVoting(policy)
	t.seals = map[common.Hash]common.Seal{}
	t.proposeSeals = map[common.Hash]Propose{}
	t.ballotSeals = map[common.Hash]Ballot{}
}

func (t *testVotingNew) newNodes(n uint) []common.HomeNode {
	var nodes []common.HomeNode
	for i := uint(0); i < n; i++ {
		nodes = append(nodes, common.NewRandomHomeNode())
	}

	return nodes
}

func (t *testVotingNew) open(node common.HomeNode, round Round) (common.Hash, error) {
	propose, seal, _ := NewTestSealPropose(node.Address(), nil)
	_, err := t.rv.Open(seal)
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

func (t *testVotingNew) newBallot(
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

func (t *testVotingNew) newBallotVote(
	node common.HomeNode,
	psHash common.Hash,
	height common.Big,
	stage VoteStage,
	round Round,
	vote Vote,
) (common.Hash, error) {
	sHash, seal, err := t.newBallot(node, psHash, height, stage, round, vote)
	if err != nil {
		return common.Hash{}, err
	}

	if _, err = t.rv.Vote(seal); err != nil {
		return common.Hash{}, err
	}

	return sHash, nil
}

func (t *testVotingNew) TestNew() {
	rv := NewRoundVoting()
	t.Nil(rv.current)
}

func (t *testVotingNew) TestOpen() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	seal.Sign(common.TestNetworkID, t.homeNode.Seed())

	vp, err := t.rv.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	t.Equal(psHash, vp.psHash)
	t.Equal(0, t.rv.unknown.Len())

	t.Equal(propose.Block.Height, vp.height)
	t.Equal(propose.Round, vp.round)
	t.Equal(VoteStageINIT, vp.stage)

	t.Equal(t.rv.current.psHash, vp.psHash)
	t.Equal(t.rv.current.height, vp.height)
	t.Equal(t.rv.current.round, vp.round)
	t.Equal(t.rv.current.stage, vp.stage)
	t.Equal(0, len(t.rv.current.stageINIT.voted))
	t.Equal(1, len(t.rv.current.stageSIGN.voted)) // proposer will be voted in sign
	t.Equal(0, len(t.rv.current.stageACCEPT.voted))
}

func (t *testVotingNew) TestClose() {
	err := t.rv.Close()
	t.Error(err, ProposalIsNotOpenedError)
	t.Nil(t.rv.current)
	t.Nil(t.rv.previous)

	// after open and then close
	_, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	seal.Sign(common.TestNetworkID, t.homeNode.Seed())

	_, err = t.rv.Open(seal)
	t.NoError(err)
	t.NotNil(t.rv.current)
	t.Nil(t.rv.previous)

	current := t.rv.Current()

	err = t.rv.Close()
	t.NoError(err)
	t.Nil(t.rv.current)
	t.NotNil(t.rv.previous)
	t.Equal(current, t.rv.previous)
}

func (t *testVotingNew) TestVoteCurrent() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.rv.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		_, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote the current proposal
		v1 := common.NewRandomHomeNode()
		v1Vote := VoteNOP

		_, err := t.newBallotVote(v1, psHash, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v1.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v1.Address())
		t.True(found)
		t.Equal(v1Vote, sn.vote)
	}

	{ // v2 vote the current proposal
		v2 := common.NewRandomHomeNode()
		v2Vote := VoteNOP

		_, err := t.newBallotVote(v2, psHash, propose.Block.Height, VoteStageSIGN, round, v2Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v2.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v2.Address())
		t.True(found)
		t.Equal(v2Vote, sn.vote)
	}
}

func (t *testVotingNew) TestVoteUnknown() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.rv.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		_, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote unknown
		psHash1 := common.NewRandomHash("sl")

		v1 := common.NewRandomHomeNode()
		v1Vote := VoteNOP

		sHash, err := t.newBallotVote(v1, psHash1, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v1.Address())
		t.Equal(0, len(voted))

		current, unknown := t.rv.Voted(v1.Address())
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

func (t *testVotingNew) TestVoteUnknownCancel() {
	propose, seal, _ := NewTestSealPropose(t.homeNode.Address(), nil)
	_, err := t.rv.Open(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	round := Round(0)

	{ // v0 vote the current proposal
		v0 := common.NewRandomHomeNode()
		v0Vote := VoteYES

		_, err := t.newBallotVote(v0, psHash, propose.Block.Height, VoteStageSIGN, round, v0Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v0.Address())
		t.NotNil(voted[VoteStageSIGN])

		sn, found := voted[VoteStageSIGN].Voted(v0.Address())
		t.True(found)
		t.Equal(v0Vote, sn.vote)
	}

	{ // v1 vote unknown
		psHash1 := common.NewRandomHash("sl")

		v1 := common.NewRandomHomeNode()
		v1Vote := VoteNOP

		sHash, err := t.newBallotVote(v1, psHash1, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		// check
		voted := t.rv.current.Voted(v1.Address())
		t.Equal(0, len(voted))

		current, unknown := t.rv.Voted(v1.Address())
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

		_, err := t.newBallotVote(v1, psHash, propose.Block.Height, VoteStageSIGN, round, v1Vote)
		t.NoError(err)

		current, unknown := t.rv.Voted(v1.Address())
		t.Equal(1, len(current)) // voted in current
		t.True(unknown.Empty())  // voted in unknown
	}
}

// VotingUnknown can contain all kind of vote of validators
func (t *testVotingNew) TestCanCountUnknownSamePSHashAndStage() {
	// In unknown,
	// all votes has,
	// - same psHash
	// - same stage
	vo := NewVotingUnknown()
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

	/*
		b, _ := json.Marshal(vo)
		fmt.Println(">>>>>", common.PrintJSON(b, true, false))
	*/
}

func (t *testVotingNew) TestCanCountUnknownINITSameHeightAndRound() {
	// In unknown,
	// all votes has,
	// - same height
	// - same round

	vo := NewVotingUnknown()
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

func (t *testVotingNew) TestCloseVotingStage() {
	st := NewVotingStage(common.NewRandomHash("sl"), common.NewBig(33), Round(0), VoteStageSIGN)
	t.False(st.Closed())

	st.Close()
	t.True(st.Closed())

	node := t.newNodes(1)[0]

	// even closed, vote can be possible
	st.Vote(node.Address(), VoteYES, common.NewRandomHash("sl"))
}

func (t *testVotingNew) TestCloseUnknown() {
	un := NewVotingUnknown()

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

func TestVotingNew(t *testing.T) {
	suite.Run(t, new(testVotingNew))
}
