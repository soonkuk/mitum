package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
)

type testSealCompiler struct {
	suite.Suite
	suffrage  Suffrage
	policy    Policy
	homeState *HomeState
	threshold *Threshold
	ballotbox *Ballotbox
}

func (t *testSealCompiler) newBallot(n node.Address, height Height, round Round, stage Stage, proposal hash.Hash, currentBlock hash.Hash, nextBlock hash.Hash) Ballot {
	ballot, _ := NewBallot(n, height, round, stage, proposal, currentBlock, nextBlock)

	pk, _ := keypair.NewStellarPrivateKey()
	_ = ballot.Sign(pk, []byte{})

	return ballot
}

func (t *testSealCompiler) setupTest(total, thr uint) {
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

	t.homeState.SetState(node.StateConsensus)

	t.threshold = NewThreshold(total, thr)
	t.ballotbox = NewBallotbox(t.threshold)
}

func (t *testSealCompiler) TestNew() {
	t.setupTest(4, 3)

	sc := NewSealCompiler(
		t.homeState,
		t.suffrage,
		t.ballotbox,
		nil,
	)
	t.Equal(Round(0), sc.LastRound())
}

func (t *testSealCompiler) TestVoteBallot() {
	t.setupTest(4, 3)

	round := Round(0)

	ch := make(chan interface{})

	sc := NewSealCompiler(
		t.homeState,
		t.suffrage,
		t.ballotbox,
		ch,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	nextBlock := NewRandomBlockHash()
	proposal := NewRandomProposalHash()

	go func() {
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
	}()

	var result VoteResult
	for v := range ch {
		vr, ok := v.(VoteResult)
		t.True(ok)

		if vr.Result() == GotMajority {
			result = vr
			break
		}
	}

	t.Equal(GotMajority, result.Result())
	t.Equal(t.homeState.Height(), result.Height())
	t.Equal(round, result.Round())
	t.Equal(StageSIGN, result.Stage())
	t.Equal(t.homeState.Block().Hash(), result.CurrentBlock())
	t.Equal(nextBlock, result.NextBlock())
	t.Equal(proposal, result.Proposal())
}

func (t *testSealCompiler) TestPropose() {
	defer common.DebugPanic()

	t.setupTest(4, 3)

	round := Round(0)

	ch := make(chan interface{})
	sc := NewSealCompiler(
		t.homeState,
		t.suffrage,
		t.ballotbox,
		ch,
	)
	err := sc.Start()
	t.NoError(err)
	defer sc.Stop()

	proposal, err := NewProposal(
		t.homeState.Height(),
		round,
		t.homeState.Block().Hash(),
		t.homeState.Home().Address(),
		nil,
	)
	t.NoError(err)

	_ = proposal.Sign(t.homeState.Home().PrivateKey(), []byte{})

	go func() {
		t.True(sc.Write(proposal))
	}()

	var received Proposal
	for v := range ch {
		pp, ok := v.(Proposal)
		t.True(ok)
		received = pp
		break
	}

	t.NoError(received.IsValid())

	t.Equal(t.homeState.Height(), received.Height())
	t.Equal(t.homeState.Block().Hash(), received.CurrentBlock())

	t.Equal(round, received.Round())
	t.Equal(t.homeState.Home().Address(), received.Proposer())
	t.Equal(t.homeState.Block().Hash(), received.CurrentBlock())
	t.Equal(t.homeState.Home().PublicKey(), received.Signer())
}

func TestSealCompiler(t *testing.T) {
	suite.Run(t, &testSealCompiler{})
}
