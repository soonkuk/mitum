package isaac

import (
	"context"
	"testing"
	"time"

	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
	"github.com/stretchr/testify/suite"
)

type testCountRound struct {
	suite.Suite
	homeState         *HomeState
	network           *network.NodesTest
	client            ClientTest
	suffrage          Suffrage
	policy            Policy
	proposalValidator ProposalValidator
	sc                *ConsensusStateHandler
}

func (t *testCountRound) SetupTest() {
	t.homeState = NewRandomHomeState()
	other := NewRandomHomeState()

	t.policy = NewTestPolicy()
	t.proposalValidator = NewTestProposalValidator(t.policy, time.Millisecond*1)

	t.suffrage = NewSuffrageTest(
		[]node.Node{
			other.Home(),
			t.homeState.Home(), // home is not proposer
		},
		func(height Height, round Round, nodes []node.Node) []node.Node {
			return nodes
		},
	)

	t.network = network.NewNodesTest(t.homeState.Home())
	t.NoError(t.network.Start())
	t.client = NewClientTest(t.network)
}

func (t *testCountRound) TearDownTest() {
	t.NoError(t.network.Stop())
}

func (t *testCountRound) TestTimeout() {
	// broadcast timeout init ballot right now
	t.policy.TimeoutINITBallot = time.Millisecond * 1

	t.sc = NewConsensusStateHandler(
		t.homeState,
		t.suffrage,
		t.policy,
		t.client,
		t.proposalValidator,
		make(chan context.Context),
	)
	defer t.sc.Stop()

	// initial VoteResult
	vr := NewVoteResult(
		t.homeState.PreviousBlock().Height(),
		t.homeState.Block().Round()+1,
		StageINIT,
		t.homeState.Block().Proposal(),
		VoteRecords{},
	)
	vr.currentBlock = t.homeState.PreviousBlock().Hash()
	vr.nextBlock = t.homeState.Block().Hash()
	vr.result = GotMajority
	vr.lastRound = t.homeState.Block().Round() + 1

	err := t.sc.StartWithContext(context.WithValue(context.Background(), "vr", vr))
	t.NoError(err)

	lastRound := t.homeState.Block().Round() + 1
	expectedHeight := t.homeState.PreviousBlock().Height()
	t.Equal(lastRound, t.sc.LastVoteResult().LastRound())

	message := <-t.network.Reader()

	ballot, ok := message.(Ballot)
	t.True(ok)

	t.Equal(expectedHeight, ballot.Height())
	t.Equal(lastRound+1, ballot.Round())
}

func TestCountRound(t *testing.T) {
	suite.Run(t, &testCountRound{})
}
