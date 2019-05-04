package isaac

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
)

type testConsensusBlocker struct {
	suite.Suite

	// set current height, block, state
	height    common.Big
	block     common.Hash
	state     []byte
	total     uint
	threshold uint

	home             *common.HomeNode
	cstate           *ConsensusState
	policy           ConsensusPolicy
	votingBox        *TestMockVotingBox
	sealBroadcaster  *TestSealBroadcaster
	sealPool         SealPool
	blocker          *ConsensusBlocker
	proposerSelector *TProposerSelector
}

func (t *testConsensusBlocker) SetupSuite() {
	t.home = common.NewRandomHome()
}

func (t *testConsensusBlocker) SetupTest() {
	t.policy = ConsensusPolicy{
		NetworkID:       common.TestNetworkID,
		Total:           t.total,
		Threshold:       t.threshold,
		TimeoutWaitSeal: time.Second * 3,
	}
	t.cstate = &ConsensusState{home: t.home, height: t.height, block: t.block, state: t.state}

	t.votingBox = &TestMockVotingBox{}
	t.sealBroadcaster, _ = NewTestSealBroadcaster(t.policy, t.home)
	t.sealPool = NewDefaultSealPool()
	t.proposerSelector = NewTProposerSelector()
	t.proposerSelector.SetProposer(t.home)
}

func (t *testConsensusBlocker) newBlocker() *ConsensusBlocker {
	b := NewConsensusBlocker(
		t.policy,
		t.cstate,
		t.votingBox,
		t.sealBroadcaster,
		t.sealPool,
		t.proposerSelector,
	)
	b.Start()

	return b
}

// TestFreshNewProposal simulates, new proposal is received,
// - blocker will broadcast sign ballot for proposal
func (t *testConsensusBlocker) TestFreshNewProposal() {
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	propose, err := NewTestPropose(t.home.Address(), nil)
	t.NoError(err)

	{ // correcting propose
		propose.Block.Height = t.height
		propose.Block.Current = t.block
		propose.Block.Next = common.NewRandomHash("bk")
		propose.State.Current = t.state
		propose.State.Next = []byte("showme")
		propose.Round = round
	}

	seal, err := common.NewSeal(ProposeSealType, propose)
	t.NoError(err)
	err = t.sealPool.Add(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: true,
		Proposal: psHash,
		Result:   VoteResultYES,
		Height:   t.height,
		Round:    round,
		Stage:    VoteStageINIT,
	}

	t.votingBox.SetResult(votingResult, nil)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	errChan := make(chan error)
	blocker.Vote(seal, errChan)
	t.NoError(<-errChan)

	var receivedSeal common.Seal
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
	}

	// sign ballot is received
	t.Equal(BallotSealType, receivedSeal.Type)
	t.NoError(receivedSeal.Wellformed())

	var receivedBallot Ballot
	err = receivedSeal.UnmarshalBody(&receivedBallot)
	t.NoError(err)

	t.True(votingResult.Height.Equal(receivedBallot.Height))
	t.Equal(round, receivedBallot.Round)
	t.Equal(t.home.Address(), receivedBallot.Source)
	t.Equal(VoteStageSIGN, receivedBallot.Stage)
	t.True(receivedBallot.Vote.IsValid())
}

// TestSIGN simulates, sign ballots consensused,
// - blocker will broadcast accept ballot for proposal
func (t *testConsensusBlocker) TestSIGN() {
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	psHash := common.NewRandomHash("sl")

	ballot, err := NewBallot(
		psHash,
		t.home.Address(),
		t.height,
		round,
		VoteStageSIGN,
		VoteYES,
	)
	t.NoError(err)

	seal, err := common.NewSeal(BallotSealType, ballot)
	t.NoError(err)
	err = t.sealPool.Add(seal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: psHash,
		Result:   VoteResultYES,
		Height:   t.height,
		Round:    round,
		Stage:    VoteStageSIGN,
	}

	t.votingBox.SetResult(votingResult, nil)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	errChan := make(chan error)
	blocker.Vote(seal, errChan)
	t.NoError(<-errChan)

	var receivedSeal common.Seal
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
	}

	// sign ballot is received
	t.Equal(BallotSealType, receivedSeal.Type)
	t.NoError(receivedSeal.Wellformed())

	var receivedBallot Ballot
	err = receivedSeal.UnmarshalBody(&receivedBallot)
	t.NoError(err)

	t.True(votingResult.Height.Equal(receivedBallot.Height))
	t.Equal(round, receivedBallot.Round)
	t.Equal(t.home.Address(), receivedBallot.Source)
	t.Equal(VoteStageACCEPT, receivedBallot.Stage)
	t.True(receivedBallot.Vote.IsValid())
}

// TestACCEPT simulates, accept ballots consensused, blocker will,
// - finishes round
// - store block and state
// - ready to start next block
func (t *testConsensusBlocker) TestACCEPT() {
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	{ // timer is started after propose accepted
		err := blocker.startTimer(false, func() error {
			return blocker.broadcastINIT(t.height, round)
		})
		t.NoError(err)
	}

	var psHash common.Hash
	var propose Propose
	{ // store proposal seal first
		var err error
		propose, err = NewTestPropose(t.home.Address(), nil)
		t.NoError(err)

		{ // correcting propose
			propose.Block.Height = t.height
			propose.Block.Current = t.block
			propose.Block.Next = common.NewRandomHash("bk")
			propose.State.Current = t.state
			propose.State.Next = []byte("showme")
			propose.Round = round
		}

		seal, err := common.NewSeal(ProposeSealType, propose)
		t.NoError(err)
		err = t.sealPool.Add(seal)
		t.NoError(err)

		psHash, _, err = seal.Hash()
		t.NoError(err)
	}

	ballot, err := NewBallot(
		psHash,
		t.home.Address(),
		t.height,
		round,
		VoteStageACCEPT,
		VoteYES,
	)
	t.NoError(err)

	seal, err := common.NewSeal(BallotSealType, ballot)
	t.NoError(err)
	err = t.sealPool.Add(seal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: psHash,
		Result:   VoteResultYES,
		Height:   t.height,
		Round:    round,
		Stage:    VoteStageACCEPT,
	}

	t.votingBox.SetResult(votingResult, nil)

	currentTimer := fmt.Sprintf("%p", blocker.timer)

	errChan := make(chan error)
	blocker.Vote(seal, errChan)
	t.NoError(<-errChan)

	{ //check state
		t.True(propose.Block.Height.Inc().Equal(t.cstate.Height()))
		t.True(propose.Block.Next.Equal(t.cstate.Block()))
		t.Equal(propose.State.Next, t.cstate.State())
	}

	// new timer is started
	t.NotEqual(currentTimer, fmt.Sprintf("%p", blocker.timer))
}

// TestSIGNButNOP simulates, SIGN ballots consensused with NOP vote, blocker will,
// - ready to start next round with same height
// - broadcast INIT ballot with same height and increased round
func (t *testConsensusBlocker) TestSIGNButNOP() {
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)
	psHash := common.NewRandomHash("sl")

	ballot, err := NewBallot(
		psHash,
		t.home.Address(),
		t.height,
		round,
		VoteStageSIGN,
		VoteNOP,
	)
	t.NoError(err)

	seal, err := common.NewSeal(BallotSealType, ballot)
	t.NoError(err)
	err = t.sealPool.Add(seal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: psHash,
		Result:   VoteResultNOP,
		Height:   t.height,
		Round:    round,
		Stage:    VoteStageSIGN,
	}

	t.votingBox.SetResult(votingResult, nil)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	errChan := make(chan error)
	blocker.Vote(seal, errChan)
	t.NoError(<-errChan)

	var receivedSeal common.Seal
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
	}

	t.Equal(BallotSealType, receivedSeal.Type)
	t.NoError(receivedSeal.Wellformed())

	var receivedBallot Ballot
	err = receivedSeal.UnmarshalBody(&receivedBallot)
	t.NoError(err)

	// should be round + 1
	t.Equal(round+1, receivedBallot.Round)

	t.True(votingResult.Height.Equal(receivedBallot.Height))
	t.Equal(t.home.Address(), receivedBallot.Source)
	t.Equal(VoteStageINIT, receivedBallot.Stage)
	t.True(receivedBallot.Vote.IsValid())
}

// TestFreshNewProposalButExpired simulates, new proposal is received, but next
// expecting ballot is not received.
// - blocker will broadcast INIT ballot for next round
func (t *testConsensusBlocker) TestFreshNewProposalButExpired() {
	t.policy.TimeoutWaitSeal = time.Millisecond * 300
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	propose, err := NewTestPropose(t.home.Address(), nil)
	t.NoError(err)

	{ // correcting propose
		propose.Block.Height = t.height
		propose.Block.Current = t.block
		propose.Block.Next = common.NewRandomHash("bk")
		propose.State.Current = t.state
		propose.State.Next = []byte("showme")
		propose.Round = round
	}

	seal, err := common.NewSeal(ProposeSealType, propose)
	t.NoError(err)
	err = t.sealPool.Add(seal)
	t.NoError(err)

	psHash, _, err := seal.Hash()
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: true,
		Proposal: psHash,
		Result:   VoteResultYES,
		Height:   t.height,
		Round:    round,
		Stage:    VoteStageINIT,
	}

	t.votingBox.SetResult(votingResult, nil)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	errChan := make(chan error)
	blocker.Vote(seal, errChan)
	t.NoError(<-errChan)

	<-bChan // this ballot is SIGN ballot
	<-time.After(t.policy.TimeoutWaitSeal)

	var receivedSeal common.Seal
	select {
	case <-time.After(time.Millisecond * 300):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		// this ballot is INIT ballot for next round after timeout
	}

	var receivedBallot Ballot
	err = receivedSeal.UnmarshalBody(&receivedBallot)
	t.NoError(err)

	t.True(votingResult.Height.Equal(receivedBallot.Height))
	t.Equal(round+1, receivedBallot.Round)
	t.Equal(t.home.Address(), receivedBallot.Source)
	t.Equal(VoteStageINIT, receivedBallot.Stage)
	t.Equal(VoteYES, receivedBallot.Vote)
}

// TestWaitingBallotButExpired simulates that proposal accepted, but next
// expecting ballot is not received.
// - blocker will broadcast INIT ballot for next round
func (t *testConsensusBlocker) TestWaitingBallotButExpired() {
	t.policy.TimeoutWaitSeal = time.Millisecond * 300
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	{ // timer is started after propose accepted
		err := blocker.startTimer(false, func() error {
			return blocker.broadcastINIT(t.height, round+1)
		})
		t.NoError(err)
	}

	var receivedSeal common.Seal
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		// this ballot is INIT ballot for next round after timeout
	}

	var receivedBallot Ballot
	err := receivedSeal.UnmarshalBody(&receivedBallot)
	t.NoError(err)

	t.True(t.height.Equal(receivedBallot.Height))
	t.Equal(round+1, receivedBallot.Round)
	t.Equal(t.home.Address(), receivedBallot.Source)
	t.Equal(VoteStageINIT, receivedBallot.Stage)
	t.Equal(VoteYES, receivedBallot.Vote)
}

func TestConsensusBlockerTotal4Threshold3(t *testing.T) {
	suite.Run(
		t,
		&testConsensusBlocker{
			height:    common.NewBig(33),
			block:     common.NewRandomHash("bk"),
			state:     []byte("showme"),
			total:     4,
			threshold: 3,
		},
	)
}
