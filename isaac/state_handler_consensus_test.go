package isaac

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type testConsensusStateHandler struct {
	suite.Suite
	suffrage      Suffrage
	policy        Policy
	homeState     *HomeState
	threshold     *Threshold
	ballotbox     *Ballotbox
	networks      map[node.Address]*network.NodesTest
	clients       map[node.Address]ClientTest
	closeNetworks func()
}

func (t *testConsensusStateHandler) newBallot(n node.Address, height Height, round Round, stage Stage, proposal hash.Hash, currentBlock hash.Hash, nextBlock hash.Hash) Ballot {
	ballot, _ := NewBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

	pk, _ := keypair.NewStellarPrivateKey()
	_ = ballot.Sign(pk, []byte{})

	return ballot
}

func (t *testConsensusStateHandler) setupTest(total, thr uint) {
	t.homeState = NewRandomHomeState()

	nodes := []node.Node{t.homeState.Home()}
	for i := uint(0); i < total-1; i++ {
		n := node.NewRandomHome()
		nodes = append(nodes, n)
	}

	t.policy = NewTestPolicy()

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

	t.homeState.SetState(node.StateConsensus)

	t.threshold = NewThreshold(total, thr)
	t.ballotbox = NewBallotbox(t.threshold)
}

func (t *testConsensusStateHandler) TearDownTest() {
	if t.closeNetworks != nil {
		t.closeNetworks()
	}
}

func (t *testConsensusStateHandler) TestNew() {
	t.setupTest(4, 3)

	chanState := make(chan node.State)
	sc := NewConsensusStateHandler(
		t.homeState,
		t.suffrage,
		t.policy,
		t.ballotbox,
		t.clients[t.homeState.Home().Address()],
		chanState,
	)
	t.Equal(node.StateConsensus, sc.State())
}

func (t *testConsensusStateHandler) TestVoteBallot() {
	t.setupTest(4, 3)

	round := Round(0)

	chanState := make(chan node.State)
	sc := NewConsensusStateHandler(
		t.homeState,
		t.suffrage,
		t.policy,
		t.ballotbox,
		t.clients[t.homeState.Home().Address()],
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	nextBlock := NewRandomBlockHash()
	proposal := NewRandomProposalHash()

	for _, n := range t.suffrage.Nodes() {
		ballot := t.newBallot(
			n.Address(),
			t.homeState.Height(),
			round,
			StageSIGN,
			proposal,
			t.homeState.Block().Hash(),
			nextBlock,
		)
		t.True(sc.Write(ballot))
	}

	var wg sync.WaitGroup
	wg.Add(len(t.suffrage.Nodes()))

	go func() {
		for _, nt := range t.networks {
			message := <-nt.Reader()

			ballot, ok := message.(Ballot)
			t.True(ok)

			t.NoError(ballot.IsValid())
			t.Equal(t.homeState.Height(), ballot.Height())
			t.Equal(round, ballot.Round())
			t.Equal(StageACCEPT, ballot.Stage())
			//t.Equal(t.homeState.Home().Address(), ballot.Proposer())
			t.Equal(t.homeState.Block().Hash(), ballot.CurrentBlock())
			t.Equal(nextBlock, ballot.NextBlock())
			t.Equal(proposal, ballot.Proposal())
			t.Equal(t.homeState.Home().PublicKey(), ballot.Signer())

			wg.Done()
		}
	}()

	wg.Wait()
}

func (t *testConsensusStateHandler) TestPropose() {
	defer common.DebugPanic()

	t.setupTest(4, 3)

	round := Round(0)

	chanState := make(chan node.State)
	sc := NewConsensusStateHandler(
		t.homeState,
		t.suffrage,
		t.policy,
		t.ballotbox,
		t.clients[t.homeState.Home().Address()],
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	nextBlock := NewRandomBlockHash()
	proposalHash := NewRandomProposalHash()

	for _, n := range t.suffrage.Nodes() {
		ballot := t.newBallot(
			n.Address(),
			t.homeState.Height(),
			round,
			StageINIT,
			proposalHash,
			t.homeState.Block().Hash(),
			nextBlock,
		)
		t.True(sc.Write(ballot))
	}

	var wg sync.WaitGroup
	wg.Add(len(t.suffrage.Nodes()))

	go func() {
		for _, nt := range t.networks {
			message := <-nt.Reader()

			proposal, ok := message.(Proposal)
			t.True(ok)

			t.NoError(proposal.IsValid())

			t.Equal(t.homeState.Height(), proposal.Height())
			t.Equal(t.homeState.Block().Hash(), proposal.CurrentBlock())

			t.Equal(round, proposal.Round())
			t.Equal(t.homeState.Home().Address(), proposal.Proposer())
			t.Equal(t.homeState.Block().Hash(), proposal.CurrentBlock())
			t.Equal(t.homeState.Home().PublicKey(), proposal.Signer())

			wg.Done()
		}
	}()

	wg.Wait()
}

func (t *testConsensusStateHandler) TestVoteToINIT() {
	t.setupTest(1, 1)

	round := Round(3)

	chanState := make(chan node.State)
	sc := NewConsensusStateHandler(
		t.homeState,
		t.suffrage,
		t.policy,
		t.ballotbox,
		t.clients[t.homeState.Home().Address()],
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	nextBlock := NewRandomBlockHash()
	proposal := NewRandomProposalHash()

	for _, n := range t.suffrage.Nodes() {
		ballot := t.newBallot(
			n.Address(),
			t.homeState.Height(),
			round,
			StageINIT,
			proposal,
			t.homeState.Block().Hash(),
			nextBlock,
		)
		t.True(sc.Write(ballot))
	}

	var blocks []map[string]interface{}
	chanEnd := make(chan bool)

	startHeight := t.homeState.Height()
	endHeight := t.homeState.Height().Add(3)

	go func() {
		n := t.networks[t.homeState.Home().Address()]
	end:
		for message := range n.Reader() {
			sl, ok := message.(seal.Seal)
			if !ok {
				continue
			}

			if sl.Type().Equal(ProposalType) {
				proposal := sl.(Proposal)
				blocks = append(
					blocks,
					map[string]interface{}{
						"height": proposal.Height(),
						"round":  proposal.Round(),
					},
				)

				if t.homeState.Height().Equal(endHeight) {
					chanEnd <- true
					break end
				}
			}

			if err := sc.receiveSeal(sl); err != nil {
				Log().Error("error receiveSeal", "error", err, "seal", sl)
			}
		}
	}()

	select {
	case <-time.After(t.policy.TimeoutINITBallot + time.Millisecond*100):
		t.NoError(xerrors.Errorf("failed to get init ballot"))
	case <-chanEnd:
	}

	t.Equal(startHeight.Add(1), blocks[0]["height"])
	t.Equal(round, blocks[0]["round"])

	for i, d := range blocks[1:] {
		t.Equal(startHeight.Add(i+2), d["height"])
		t.Equal(Round(0), d["round"])
	}
}

func (t *testConsensusStateHandler) TestVoteToINITTimeout() {
	t.setupTest(1, 1)

	t.policy.TimeoutINITBallot = time.Millisecond * 700

	chanState := make(chan node.State)
	sc := NewConsensusStateHandler(
		t.homeState,
		t.suffrage,
		t.policy,
		t.ballotbox,
		t.clients[t.homeState.Home().Address()],
		chanState,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	chanEnd := make(chan Ballot)
	go func() {
		n := t.networks[t.homeState.Home().Address()]
		for message := range n.Reader() {
			ballot, ok := message.(Ballot)
			if !ok {
				continue
			}

			chanEnd <- ballot
		}
	}()

	var ballot Ballot
	select {
	case <-time.After(t.policy.TimeoutINITBallot + time.Millisecond*100):
		t.NoError(xerrors.Errorf("failed to get init ballot"))
	case ballot = <-chanEnd:
	}

	t.Equal(StageINIT, ballot.Stage())
	t.True(ballot.Height().Equal(t.homeState.PreviousBlock().Height()))
	t.Equal(Round(0), ballot.Round())
	t.True(ballot.CurrentBlock().Equal(t.homeState.PreviousBlock().Hash()))
	t.True(ballot.NextBlock().Equal(t.homeState.Block().Hash()))
	t.True(ballot.Proposal().Equal(t.homeState.Proposal()))
}

func TestConsensusStateHandler(t *testing.T) {
	suite.Run(t, &testConsensusStateHandler{})
}
