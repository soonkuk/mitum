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

	// TODO rename to proposal
	proposal := NewTestProposal(t.home.Address(), nil)

	{ // correcting proposal
		proposal.Block.Height = t.height
		proposal.Block.Current = t.block
		proposal.Block.Next = common.NewRandomHash("bk")
		proposal.State.Current = t.state
		proposal.State.Next = []byte("showme")
		proposal.Round = round
	}

	err := t.sealPool.Add(proposal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: true,
		Proposal: proposal.Hash(),
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
	blocker.Vote(proposal, errChan)
	t.NoError(<-errChan)

	var receivedSeal common.Seal
	var receivedBallot Ballot
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		b, ok := receivedSeal.(Ballot)
		t.True(ok)
		receivedBallot = b
	}

	// sign ballot is received
	t.Equal(BallotSealType, receivedSeal.Type)
	t.NoError(receivedSeal.Wellformed())

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

	phash := common.NewRandomHash("sl")

	ballot := NewBallot(
		phash,
		t.home.Address(),
		t.height,
		round,
		VoteStageSIGN,
		VoteYES,
	)

	err := t.sealPool.Add(ballot)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: phash,
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
	blocker.Vote(ballot, errChan)
	t.NoError(<-errChan)

	var receivedBallot Ballot
	var receivedSeal common.Seal
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		b, ok := receivedSeal.(Ballot)
		t.True(ok)
		receivedBallot = b
	}

	// sign ballot is received
	t.Equal(BallotSealType, receivedSeal.Type)
	t.NoError(receivedSeal.Wellformed())

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

	{ // timer is started after proposal accepted
		err := blocker.startTimer(false, func() error {
			return blocker.broadcastINIT(t.height, round)
		})
		t.NoError(err)
	}

	var proposal Proposal
	{ // store proposal seal first
		var err error
		proposal = NewTestProposal(t.home.Address(), nil)

		{ // correcting proposal
			proposal.Block.Height = t.height
			proposal.Block.Current = t.block
			proposal.Block.Next = common.NewRandomHash("bk")
			proposal.State.Current = t.state
			proposal.State.Next = []byte("showme")
			proposal.Round = round
		}

		err = t.sealPool.Add(proposal)
		t.NoError(err)
	}

	ballot := NewBallot(
		proposal.Hash(),
		t.home.Address(),
		t.height,
		round,
		VoteStageACCEPT,
		VoteYES,
	)

	err := t.sealPool.Add(ballot)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: proposal.Hash(),
		Result:   VoteResultYES,
		Height:   t.height,
		Round:    round,
		Stage:    VoteStageACCEPT,
	}

	t.votingBox.SetResult(votingResult, nil)

	currentTimer := fmt.Sprintf("%p", blocker.timer)

	errChan := make(chan error)
	blocker.Vote(ballot, errChan)
	t.NoError(<-errChan)

	{ //check state
		t.True(proposal.Block.Height.Inc().Equal(t.cstate.Height()))
		t.True(proposal.Block.Next.Equal(t.cstate.Block()))
		t.Equal(proposal.State.Next, t.cstate.State())
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
	phash := common.NewRandomHash("sl")

	ballot := NewBallot(
		phash,
		t.home.Address(),
		t.height,
		round,
		VoteStageSIGN,
		VoteNOP,
	)

	err := t.sealPool.Add(ballot)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: phash,
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
	blocker.Vote(ballot, errChan)
	t.NoError(<-errChan)

	var receivedSeal common.Seal
	var receivedBallot Ballot
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		b, ok := receivedSeal.(Ballot)
		t.True(ok)
		receivedBallot = b
	}

	t.Equal(BallotSealType, receivedSeal.Type)
	t.NoError(receivedSeal.Wellformed())

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

	proposal := NewTestProposal(t.home.Address(), nil)

	{ // correcting proposal
		proposal.Block.Height = t.height
		proposal.Block.Current = t.block
		proposal.Block.Next = common.NewRandomHash("bk")
		proposal.State.Current = t.state
		proposal.State.Next = []byte("showme")
		proposal.Round = round
	}

	err := t.sealPool.Add(proposal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: true,
		Proposal: proposal.Hash(),
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
	blocker.Vote(proposal, errChan)
	t.NoError(<-errChan)

	<-bChan // this ballot is SIGN ballot
	<-time.After(t.policy.TimeoutWaitSeal)

	var receivedSeal common.Seal
	var receivedBallot Ballot
	select {
	case <-time.After(time.Millisecond * 300):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		// this ballot is INIT ballot for next round after timeout
		b, ok := receivedSeal.(Ballot)
		t.True(ok)
		receivedBallot = b
	}

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

	{ // timer is started after proposal accepted
		err := blocker.startTimer(false, func() error {
			return blocker.broadcastINIT(t.height, round+1)
		})
		t.NoError(err)
	}

	var receivedSeal common.Seal
	var receivedBallot Ballot
	select {
	case <-time.After(time.Second):
		t.Empty("timeout to wait receivedSeal")
		return
	case receivedSeal = <-bChan:
		// this ballot is INIT ballot for next round after timeout
		b, ok := receivedSeal.(Ballot)
		t.True(ok)
		receivedBallot = b
	}

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
