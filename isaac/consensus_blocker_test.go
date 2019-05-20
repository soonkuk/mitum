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

	home             common.HomeNode
	cstate           *ConsensusState
	policy           ConsensusPolicy
	votingBox        *TestMockVotingBox
	sealBroadcaster  *TestSealBroadcaster
	sealPool         *DefaultSealPool
	proposerSelector *TProposerSelector
	blockStorage     *TBlockStorage
}

func (t *testConsensusBlocker) SetupSuite() {
	t.home = common.NewRandomHome()
}

func (t *testConsensusBlocker) SetupTest() {
	t.policy = DefaultConsensusPolicy()
	t.policy.NetworkID = common.TestNetworkID
	t.policy.Total = t.total
	t.policy.Threshold = t.threshold
	t.policy.TimeoutWaitSeal = time.Second * 3
	t.policy.AvgBlockRoundInterval = time.Millisecond * 300

	t.cstate = NewConsensusState(t.home)
	t.cstate.SetHeight(t.height)
	t.cstate.SetBlock(t.block)
	t.cstate.SetState(t.state)

	t.blockStorage = NewTBlockStorage()

	t.votingBox = &TestMockVotingBox{}
	t.sealBroadcaster, _ = NewTestSealBroadcaster(t.policy, t.home)
	t.sealPool = NewDefaultSealPool()
	t.sealPool.SetLogContext("node", t.home.Name())
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
		t.blockStorage,
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

	proposal := NewTestProposal(t.home.Address(), nil)

	{ // correcting proposal
		proposal.Block.Height = t.height
		proposal.Block.Current = t.block
		proposal.Block.Next = common.NewRandomHash("bk")
		proposal.State.Current = t.state
		proposal.State.Next = []byte("showme")
		proposal.Round = round
	}

	err := proposal.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)
	err = t.sealPool.Add(proposal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: true,
		Proposal: proposal.Hash(),
		Proposer: proposal.Source(),
		Block:    proposal.Block.Next,
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
	t.Equal(SIGNBallotSealType, receivedSeal.Type())
	t.NoError(receivedSeal.Wellformed())

	t.True(votingResult.Height.Equal(receivedBallot.Height()))
	t.Equal(round, receivedBallot.Round())
	t.Equal(t.home.Address(), receivedBallot.Source())
	t.Equal(VoteStageSIGN, receivedBallot.Stage())
	t.True(receivedBallot.Vote().IsValid())
}

// TestSIGN simulates, sign ballots consensused,
// - blocker will broadcast accept ballot for proposal
func (t *testConsensusBlocker) TestSIGN() {
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	proposal := common.NewRandomHash("sl")

	ballot := NewSIGNBallot(
		t.height,
		round,
		t.home.Address(),
		nil, // TODO set validators
		proposal,
		common.NewRandomHash("bk"),
		VoteYES,
	)

	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)
	err = t.sealPool.Add(ballot)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposer: t.home.Address(),
		Block:    ballot.Block(),
		Proposal: proposal,
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

	// accept ballot is received
	t.Equal(ACCEPTBallotSealType, receivedSeal.Type())
	t.NoError(receivedSeal.Wellformed())

	t.True(votingResult.Height.Equal(receivedBallot.Height()))
	t.Equal(round, receivedBallot.Round())
	t.Equal(t.home.Address(), receivedBallot.Source())
	t.Equal(VoteStageACCEPT, receivedBallot.Stage())
	t.True(receivedBallot.Vote().IsValid())
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
		err := blocker.startTimer("", blocker.policy.TimeoutWaitSeal, false, func() error {
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
		err = proposal.Sign(common.TestNetworkID, t.home.Seed())
		t.NoError(err)

		err = t.sealPool.Add(proposal)
		t.NoError(err)
	}

	ballot := NewACCEPTBallot(
		t.height,
		round,
		t.home.Address(),
		nil, // TODO set validators
		proposal.Hash(),
		common.NewRandomHash("bk"),
	)
	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)

	err = t.sealPool.Add(ballot)
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
		fmt.Println(proposal.Block.Height.Inc(), t.cstate.Height())
		t.True(proposal.Block.Height.Inc().Equal(t.cstate.Height()))
		t.True(proposal.Block.Next.Equal(t.cstate.Block()))
		t.Equal(proposal.State.Next, t.cstate.State())
		t.NotEmpty(t.blockStorage.Blocks())
		t.True(t.blockStorage.Blocks()[0].proposal.Equal(proposal.Hash()))
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
	proposal := common.NewRandomHash("sl")

	ballot := NewSIGNBallot(
		t.height,
		round,
		t.home.Address(),
		nil, // TODO set validators
		proposal,
		common.NewRandomHash("bk"),
		VoteNOP,
	)

	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)
	err = t.sealPool.Add(ballot)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: proposal,
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

	t.Equal(INITBallotSealType, receivedSeal.Type())
	t.NoError(receivedSeal.Wellformed())

	// should be round + 1
	t.Equal(round+1, receivedBallot.Round())

	t.True(votingResult.Height.Equal(receivedBallot.Height()))
	t.Equal(t.home.Address(), receivedBallot.Source())
	t.Equal(VoteStageINIT, receivedBallot.Stage())
	t.True(receivedBallot.Vote().IsValid())
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

	err := proposal.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)
	err = t.sealPool.Add(proposal)
	t.NoError(err)

	votingResult := VoteResultInfo{
		Proposed: true,
		Proposal: proposal.Hash(),
		Proposer: proposal.Source(),
		Block:    proposal.Block.Next,
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

	t.True(votingResult.Height.Equal(receivedBallot.Height()))
	t.Equal(round+1, receivedBallot.Round())
	t.Equal(t.home.Address(), receivedBallot.Source())
	t.Equal(VoteStageINIT, receivedBallot.Stage())
	t.Equal(VoteYES, receivedBallot.Vote())
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
		err := blocker.startTimer("", blocker.policy.TimeoutWaitSeal, false, func() error {
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

	t.True(t.height.Equal(receivedBallot.Height()))
	t.Equal(round+1, receivedBallot.Round())
	t.Equal(t.home.Address(), receivedBallot.Source())
	t.Equal(VoteStageINIT, receivedBallot.Stage())
	t.Equal(VoteYES, receivedBallot.Vote())
}

// TestProposeNewProposal simulates, init ballots consensused, blocker will,
// - start next round after given time
func (t *testConsensusBlocker) TestProposeNewProposalNextRound() {
	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)

	votingResult := VoteResultInfo{
		Proposed: false,
		Proposal: common.NewRandomHash("pp"),
		Result:   VoteResultYES,
		Height:   t.height,
		Round:    round + 1,
		Stage:    VoteStageINIT,
	}
	t.votingBox.SetResult(votingResult, nil)

	ballot := NewINITBallot(
		t.height,
		votingResult.Round,
		t.home.Address(),
		nil, // TODO set validators
	)
	_ = ballot.Sign(common.TestNetworkID, t.home.Seed())
	_ = t.sealPool.Add(ballot)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	errChan := make(chan error)
	blocker.Vote(ballot, errChan)
	t.NoError(<-errChan)

	var receivedSeal common.Seal
	var receivedProposal Proposal
end:
	for {
		select {
		case <-time.After(time.Second):
			t.Empty("timeout to wait receivedSeal")
			blocker.Stop()
			return
		case receivedSeal = <-bChan:
			if receivedSeal.Type() != ProposalSealType {
				continue
			}

			receivedProposal = receivedSeal.(Proposal)
			break end
		}
	}

	proposer, _ := t.proposerSelector.Select(t.cstate.block, t.cstate.height, round+1)

	t.Equal(t.cstate.height, receivedProposal.Block.Height)
	t.Equal(t.cstate.block, receivedProposal.Block.Current)
	t.Equal(t.cstate.state, receivedProposal.State.Current)
	t.Equal(votingResult.Round, receivedProposal.Round)
	t.Equal(proposer.Address(), receivedProposal.Source())
}

// TestACCEPTedButBlockDoesNotMatch simulates, ACCEPT ballots gets consensus,
// but VoteResultInfo.Block does not match with previous SIGNed
// VoteResultInfo.Block
func (t *testConsensusBlocker) TestACCEPTedButBlockDoesNotMatch() {
	defer common.DebugPanic()

	blocker := t.newBlocker()
	defer blocker.Stop()

	round := Round(1)
	proposer, _ := t.proposerSelector.Select(t.cstate.block, t.cstate.height, round+1)
	nextBlock := common.NewRandomHash("bk")

	proposal := NewTestProposal(t.home.Address(), nil)

	{ // correcting proposal
		proposal.Block.Height = t.cstate.height
		proposal.Block.Current = t.cstate.block
		proposal.Block.Next = nextBlock
		proposal.State.Current = t.state
		proposal.State.Next = []byte("showme")
		proposal.Round = round
	}

	_ = proposal.Sign(common.TestNetworkID, t.home.Seed())
	_ = t.sealPool.Add(proposal)

	bChan := make(chan common.Seal, 1)
	defer close(bChan)
	t.sealBroadcaster.SetSenderChan(bChan)

	{ // signed
		votingResult := VoteResultInfo{
			Proposed: false,
			Proposal: proposal.Hash(),
			Proposer: proposer.Address(),
			Block:    proposal.Block.Next,
			Result:   VoteResultYES,
			Height:   t.cstate.height,
			Round:    round,
			Stage:    VoteStageSIGN,
		}
		t.votingBox.SetResult(votingResult, nil)

		ballot := NewSIGNBallot(
			t.cstate.height,
			round,
			t.home.Address(),
			nil, // TODO set validators
			proposal.Hash(),
			proposal.Block.Next,
			VoteYES,
		)

		_ = ballot.Sign(common.TestNetworkID, t.home.Seed())
		_ = t.sealPool.Add(ballot)

		errChan := make(chan error)
		blocker.Vote(ballot, errChan)
		t.NoError(<-errChan)
		defer close(errChan)

		var receivedSeal common.Seal
		var receivedBallot ACCEPTBallot
	end0:
		for {
			select {
			case <-time.After(time.Second):
				t.Empty("timeout to wait receivedSeal")
				return
			case receivedSeal = <-bChan:
				if receivedSeal.Type() != ACCEPTBallotSealType {
					continue
				}

				receivedBallot = receivedSeal.(ACCEPTBallot)
				break end0
			}
		}

		t.Equal(t.cstate.height, receivedBallot.Height())
		t.Equal(proposal.Block.Next, receivedBallot.Block())
		t.Equal(votingResult.Round, receivedBallot.Round())
		t.Equal(proposer.Address(), receivedBallot.Source())
	}

	{ // accepted, but block does not matched
		unknownBlock := common.NewRandomHash("bk")
		votingResult := VoteResultInfo{
			Proposed: false,
			Proposal: proposal.Hash(),
			Proposer: proposer.Address(),
			Block:    unknownBlock,
			Result:   VoteResultYES,
			Height:   t.cstate.height,
			Round:    round,
			Stage:    VoteStageACCEPT,
		}
		t.votingBox.SetResult(votingResult, nil)

		ballot := NewACCEPTBallot(
			t.cstate.height,
			round,
			t.home.Address(),
			nil, // TODO set validators
			proposal.Hash(),
			unknownBlock, // set unknown block
		)

		_ = ballot.Sign(common.TestNetworkID, t.home.Seed())
		_ = t.sealPool.Add(ballot)

		errChan := make(chan error)
		defer close(errChan)
		blocker.Vote(ballot, errChan)

		err := <-errChan
		t.True(ConsensusButBlockDoesNotMatchError.Equal(err))

	end1:
		for {
			select {
			case <-time.After(time.Second):
				break end1
			}
		}
	}
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
