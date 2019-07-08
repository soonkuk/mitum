package isaac

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
)

type testJoinStateHandler struct {
	suite.Suite
	policy    Policy
	homeState *HomeState
	network   *network.NodesTest
	client    ClientTest
}

func (t *testJoinStateHandler) SetupTest() {
	t.homeState = NewRandomHomeState()
	t.network = network.NewNodesTest(t.homeState.Home())

	t.client = NewClientTest(t.network)
	t.policy = NewTestPolicy()
	t.homeState.SetState(node.StateJoin)

	_ = t.network.Start()
}

func (t *testJoinStateHandler) TearDownTest() {
	_ = t.network.Stop()
}

func (t *testJoinStateHandler) TestNew() {
	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.client,
		make(chan context.Context),
	)
	t.Equal(node.StateJoin, sc.State())
}

func (t *testJoinStateHandler) TestINITVoteResultNewBlockCreated() {
	// condition:
	// - height: same with homeState

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.client,
		make(chan context.Context),
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	// next VoteResult
	vr := NewVoteResult(
		t.homeState.Height(),
		Round(0),
		StageINIT,
		NewRandomProposalHash(),
		VoteRecords{},
	)

	nextBlock, err := NewBlock(vr.Height().Add(1), vr.Round(), vr.Proposal())
	t.NoError(err)

	vr.currentBlock = t.homeState.Block().Hash()
	vr.nextBlock = nextBlock.Hash()
	vr.result = GotMajority

	t.True(sc.Write(vr))

	<-time.After(time.Millisecond * 100)
	t.True(vr.Height().Add(1).Equal(t.homeState.Height()))
	t.True(vr.NextBlock().Equal(t.homeState.Block().Hash()))
	t.True(vr.Proposal().Equal(t.homeState.Block().Proposal()))
}

func (t *testJoinStateHandler) TestINITACCEPTVoteResult() {
	// condition:
	// - height: same with homeState
	// - init VoteResult is finished
	// - expected accept VoteResult is received
	// - state will be changed

	chanState := make(chan context.Context)

	go func() {
		for ctx := range chanState {
			newState := ctx.Value("state").(node.State)
			t.homeState.SetState(newState)
		}
	}()

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.client,
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	// init VoteResult
	initVR := NewVoteResult(
		t.homeState.Height(),
		Round(0),
		StageINIT,
		NewRandomProposalHash(),
		VoteRecords{},
	)

	{
		nextBlock, err := NewBlock(initVR.Height().Add(1), initVR.Round(), initVR.Proposal())
		t.NoError(err)

		initVR.currentBlock = t.homeState.Block().Hash()
		initVR.nextBlock = nextBlock.Hash()
		initVR.result = GotMajority

		t.True(sc.Write(initVR))

		<-time.After(time.Millisecond * 100)
		t.True(initVR.Height().Add(1).Equal(t.homeState.Height()))
		t.True(initVR.NextBlock().Equal(t.homeState.Block().Hash()))
		t.True(initVR.Proposal().Equal(t.homeState.Block().Proposal()))
	}

	// and then ACCEPT VoteResult
	acceptVR := NewVoteResult(
		t.homeState.Height(),
		initVR.Round(),
		StageACCEPT,
		NewRandomProposalHash(), // new proposal
		VoteRecords{},
	)

	{
		nextBlock, err := NewBlock(t.homeState.Height().Add(1), acceptVR.Round(), acceptVR.Proposal())
		t.NoError(err)

		acceptVR.currentBlock = t.homeState.Block().Hash()
		acceptVR.nextBlock = nextBlock.Hash()
		acceptVR.result = GotMajority

		t.True(sc.Write(acceptVR))
	}

	<-time.After(time.Millisecond * 100)
	t.Equal(node.StateConsensus, t.homeState.State())
}

func (t *testJoinStateHandler) TestINITButInvalidACCEPTVoteResult() {
	// condition:
	// - height: same with homeState
	// - init VoteResult is finished
	// - invalid accept VoteResult is received
	// - it will be ignored

	chanState := make(chan context.Context)

	go func() {
		for ctx := range chanState {
			newState := ctx.Value("state").(node.State)
			t.homeState.SetState(newState)
		}
	}()

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.client,
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	// init VoteResult
	initVR := NewVoteResult(
		t.homeState.Height(),
		Round(0),
		StageINIT,
		NewRandomProposalHash(),
		VoteRecords{},
	)

	{
		nextBlock, err := NewBlock(initVR.Height().Add(1), initVR.Round(), initVR.Proposal())
		t.NoError(err)

		initVR.currentBlock = t.homeState.Block().Hash()
		initVR.nextBlock = nextBlock.Hash()
		initVR.result = GotMajority

		t.True(sc.Write(initVR))

		<-time.After(time.Millisecond * 100)
		t.True(initVR.Height().Add(1).Equal(t.homeState.Height()))
		t.True(initVR.NextBlock().Equal(t.homeState.Block().Hash()))
		t.True(initVR.Proposal().Equal(t.homeState.Block().Proposal()))
	}

	// and then invalid ACCEPT VoteResult; wrong height
	acceptVR := NewVoteResult(
		t.homeState.Height().Add(10),
		initVR.Round(),
		StageACCEPT,
		NewRandomProposalHash(), // new proposal
		VoteRecords{},
	)

	{
		nextBlock, err := NewBlock(t.homeState.Height().Add(1), acceptVR.Round(), acceptVR.Proposal())
		t.NoError(err)

		acceptVR.currentBlock = t.homeState.Block().Hash()
		acceptVR.nextBlock = nextBlock.Hash()
		acceptVR.result = GotMajority

		t.True(sc.Write(acceptVR))
	}

	// wrong accept VoteResult will be ignored
	<-time.After(time.Millisecond * 100)
	t.Equal(node.StateJoin, t.homeState.State())
}

func (t *testJoinStateHandler) TestBroadcastINITBallot() {
	chanState := make(chan context.Context)

	go func() {
		for ctx := range chanState {
			newState := ctx.Value("state").(node.State)
			t.homeState.SetState(newState)
		}
	}()

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.client,
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	// check init ballot
	<-time.After(time.Millisecond * 50)

	message := <-t.network.Reader()

	ballot, ok := message.(Ballot)
	t.True(ok)

	t.True(t.homeState.PreviousBlock().Height().Equal(ballot.Height()))
	t.Equal(StageINIT, ballot.Stage())
	t.Equal(t.homeState.Block().Round()+1, ballot.Round())
	t.Equal(t.homeState.Home().PublicKey(), ballot.Signer())
	t.True(t.homeState.Block().Proposal().Equal(ballot.Proposal()))
	t.True(t.homeState.Block().Hash().Equal(ballot.NextBlock()))
}

func TestJoinStateHandler(t *testing.T) {
	suite.Run(t, &testJoinStateHandler{})
}
