package isaac

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
)

type testJoinStateHandler struct {
	suite.Suite
	policy        Policy
	suffrage      Suffrage
	homeState     *HomeState
	networks      map[node.Address]*network.NodesTest
	clients       map[node.Address]ClientTest
	closeNetworks func()
}

func (t *testJoinStateHandler) setupTest(total uint) {
	t.homeState = NewRandomHomeState()

	nodes := []node.Node{t.homeState.Home()}
	for i := uint(0); i < total-1; i++ {
		n := node.NewRandomHome()
		nodes = append(nodes, n)
	}

	t.suffrage = NewSuffrageTest(
		nodes,
		func(height Height, round Round, nodes []node.Node) []node.Node {
			return nodes
		},
	)

	networks := map[node.Address]*network.NodesTest{}
	for _, n := range t.suffrage.Nodes() {
		networks[n.Address()] = network.NewNodesTest(n.(node.Home))
	}
	t.networks = networks

	for _, nt := range networks {
		for _, ot := range networks {
			nt.AddReceiver(ot.Home().Address(), ot.ReceiveFunc)
		}
		err := nt.Start()
		t.NoError(err)
	}

	clients := map[node.Address]ClientTest{}
	for _, nt := range networks {
		clients[nt.Home().Address()] = NewClientTest(nt)
	}

	t.clients = clients

	t.closeNetworks = func() {
		for _, nt := range networks {
			err := nt.Stop()
			t.NoError(err)
		}
	}

	t.policy = NewTestPolicy()

	t.homeState.SetState(node.StateJoin)
}

func (t *testJoinStateHandler) TearDownTest() {
	if t.closeNetworks != nil {
		t.closeNetworks()
	}
}

func (t *testJoinStateHandler) TestNew() {
	t.setupTest(4)

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.clients[t.homeState.Home().Address()],
		make(chan node.State),
	)
	t.Equal(node.StateJoin, sc.State())
}

func (t *testJoinStateHandler) TestINITVoteResultNewBlockCreated() {
	t.setupTest(4)

	// condition:
	// - height: same with homeState

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.clients[t.homeState.Home().Address()],
		make(chan node.State),
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

	nextBlock, err := NewBlock(vr.Height().Add(1), vr.Proposal())
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
	t.setupTest(4)

	// condition:
	// - height: same with homeState
	// - init VoteResult is finished
	// - expected accept VoteResult is received
	// - state will be changed

	chanState := make(chan node.State)

	go func() {
		for {
			select {
			case newState := <-chanState:
				t.homeState.SetState(newState)
			}
		}
	}()

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.clients[t.homeState.Home().Address()],
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
		nextBlock, err := NewBlock(initVR.Height().Add(1), initVR.Proposal())
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
		nextBlock, err := NewBlock(t.homeState.Height().Add(1), acceptVR.Proposal())
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
	t.setupTest(4)

	// condition:
	// - height: same with homeState
	// - init VoteResult is finished
	// - invalid accept VoteResult is received
	// - it will be ignored

	chanState := make(chan node.State)

	go func() {
		for {
			select {
			case newState := <-chanState:
				t.homeState.SetState(newState)
			}
		}
	}()

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.clients[t.homeState.Home().Address()],
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
		nextBlock, err := NewBlock(initVR.Height().Add(1), initVR.Proposal())
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
		nextBlock, err := NewBlock(t.homeState.Height().Add(1), acceptVR.Proposal())
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
	t.setupTest(4)

	chanState := make(chan node.State)

	go func() {
		for {
			select {
			case newState := <-chanState:
				t.homeState.SetState(newState)
			}
		}
	}()

	sc := NewJoinStateHandler(
		t.homeState,
		t.policy,
		t.clients[t.homeState.Home().Address()],
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	// check init ballot
	<-time.After(time.Millisecond * 50)
	for _, nt := range t.networks {
		message := <-nt.Reader()

		ballot, ok := message.(Ballot)
		t.True(ok)

		t.True(t.homeState.PreviousBlock().Height().Equal(ballot.Height()))
		t.Equal(StageINIT, ballot.Stage())
		t.Equal(Round(0), ballot.Round())
		t.Equal(t.homeState.Home().PublicKey(), ballot.Signer())
		t.True(t.homeState.Block().Proposal().Equal(ballot.Proposal()))
		t.True(t.homeState.Block().Hash().Equal(ballot.NextBlock()))
	}
}

func TestJoinStateHandler(t *testing.T) {
	suite.Run(t, &testJoinStateHandler{})
}
