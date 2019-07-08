package isaac

import (
	"testing"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/stretchr/testify/suite"
)

type testStateTransition struct {
	suite.Suite
	suffrage     Suffrage
	homeState    *HomeState
	policy       Policy
	voteCompiler *VoteCompiler
}

func (t *testStateTransition) SetupTest() {
	t.homeState = NewRandomHomeState()
	t.homeState.SetState(node.StateBooting)

	t.policy = NewTestPolicy()

	nodes := []node.Node{t.homeState.Home()}
	t.suffrage = NewSuffrageTest(
		nodes,
		func(height Height, round Round, nodes []node.Node) []node.Node {
			return nodes
		},
	)

	threshold, err := NewThreshold(1, 66)
	t.NoError(err)

	ballotbox := NewBallotbox(threshold)
	t.voteCompiler = NewVoteCompiler(t.homeState, t.suffrage, ballotbox)
}

func (t *testStateTransition) TestNew() {
	st := NewStateTransition(t.homeState, t.voteCompiler)
	t.Nil(st.StateHandler())

	// register handler
	{ // booting
		err := st.SetStateHandler(
			NewBootingStateHandler(t.homeState, st.ChanState()),
		)
		t.NoError(err)

		// set again
		err = st.SetStateHandler(
			NewBootingStateHandler(t.homeState, st.ChanState()),
		)
		t.Contains(err.Error(), "already")
	}

	{ // sync
		err := st.SetStateHandler(
			NewSyncStateHandler(t.homeState, t.suffrage, t.policy, nil, st.ChanState()),
		)
		t.NoError(err)
	}

	{ // join
		err := st.SetStateHandler(
			NewJoinStateHandler(t.homeState, t.policy, nil, st.ChanState()),
		)
		t.NoError(err)
	}

	{ // consensus
		err := st.SetStateHandler(
			NewConsensusStateHandler(t.homeState, t.suffrage, t.policy, nil, st.ChanState()),
		)
		t.NoError(err)
	}

	{ // stopped
		err := st.SetStateHandler(
			NewStoppedStateHandler(t.homeState),
		)
		t.NoError(err)
	}
}

func (t *testStateTransition) TestTransition() {
	defer common.DebugPanic()

	st := NewStateTransition(t.homeState, t.voteCompiler)

	// register handler
	_ = st.SetStateHandler(NewBootingStateHandler(t.homeState, st.ChanState()))
	_ = st.SetStateHandler(NewSyncStateHandler(t.homeState, t.suffrage, t.policy, nil, st.ChanState()))
	_ = st.SetStateHandler(NewJoinStateHandler(t.homeState, t.policy, nil, st.ChanState()))
	_ = st.SetStateHandler(NewConsensusStateHandler(t.homeState, t.suffrage, t.policy, nil, st.ChanState()))
	_ = st.SetStateHandler(NewStoppedStateHandler(t.homeState))

	t.NoError(st.Start())
	defer st.Stop()

	<-time.After(time.Millisecond * 50)
	t.Nil(st.StateHandler())

	st.ChanState() <- common.SetContext(nil, "state", node.StateBooting)
	<-time.After(time.Millisecond * 50)
	t.Equal(t.homeState.State(), st.StateHandler().State())
}

func (t *testStateTransition) TestMissingState() {
	defer common.DebugPanic()

	st := NewStateTransition(t.homeState, t.voteCompiler)

	t.NoError(st.Start())
	defer st.Stop()

	st.ChanState() <- common.SetContext(nil, "state", node.StateBooting)
	<-time.After(time.Millisecond * 50)
	t.Nil(st.StateHandler())
}

func TestStateTransition(t *testing.T) {
	suite.Run(t, &testStateTransition{})
}
