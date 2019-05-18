package isaac

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/network"
)

type testConsensus struct {
	suite.Suite
	height    common.Big
	block     common.Hash
	state     []byte
	total     uint
	threshold uint

	home             *common.HomeNode
	nt               *network.NodeTestNetwork
	sealBroadcaster  *DefaultSealBroadcaster
	blocker          *ConsensusBlocker
	sealPool         *DefaultSealPool
	consensus        *Consensus
	cstate           *ConsensusState
	proposerSelector *TProposerSelector
	blockStorage     *TBlockStorage
}

func (t *testConsensus) SetupSuite() {
	t.home = common.NewRandomHome()
	t.height = common.NewBig(1)
	t.block = common.NewRandomHash("bk")
	t.state = []byte("sl")
	t.total = 1
	t.threshold = 1
}

func (t *testConsensus) SetupTest() {
	t.cstate = NewConsensusState(t.home)
	t.cstate.SetHeight(t.height)
	t.cstate.SetBlock(t.block)
	t.cstate.SetState(t.state)

	policy := ConsensusPolicy{
		NetworkID:       common.TestNetworkID,
		Total:           t.total,
		Threshold:       t.threshold,
		TimeoutWaitSeal: time.Second * 3,
	}

	votingBox := NewDefaultVotingBox(t.home, policy)

	t.nt = network.NewNodeTestNetwork()

	t.sealBroadcaster, _ = NewDefaultSealBroadcaster(policy, t.cstate)
	t.sealBroadcaster.SetSender(t.nt.Send)

	t.sealPool = NewDefaultSealPool()
	t.sealPool.SetLogContext("node", t.home.Name())

	t.proposerSelector = NewTProposerSelector()
	t.proposerSelector.SetProposer(t.home)

	t.blockStorage = NewTBlockStorage()
	t.blocker = NewConsensusBlocker(
		policy,
		t.cstate,
		votingBox,
		t.sealBroadcaster,
		t.sealPool,
		t.proposerSelector,
		t.blockStorage,
	)
	t.blocker.Start()

	consensus, err := NewConsensus(t.home, t.cstate, t.blocker)
	t.NoError(err)
	t.consensus = consensus

	err = t.nt.AddReceiver(t.consensus.Receiver())
	t.NoError(err)

	t.consensus.SetContext(
		nil, // nolint
		"policy", policy,
		"state", t.cstate,
		"sealPool", t.sealPool,
	)

	t.cstate.SetNodeState(NodeStateJoin)
	t.consensus.Start()
}

func (t *testConsensus) TeardownTest() {
	t.consensus.Stop()
	t.nt.Stop()
	t.blocker.Stop()
}

func (t *testConsensus) TestNew() {
	defer common.DebugPanic()

	var proposal Proposal
	{
		var err error
		proposal = NewTestProposal(t.cstate.Home().Address(), nil)
		proposal.Block.Height = t.height
		proposal.Block.Current = t.block
		proposal.State.Current = t.state
		t.NoError(err)
		err = proposal.Sign(common.TestNetworkID, t.cstate.Home().Seed())
		t.NoError(err)

		t.nt.Send(t.cstate.Home(), proposal)
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	for range ticker.C {
		if t.cstate.Height().Equal(proposal.Block.Height.Inc()) {
			break
		}
	}
	ticker.Stop()

	t.True(proposal.Block.Height.Inc().Equal(t.cstate.Height()))
	t.True(proposal.Block.Next.Equal(t.cstate.Block()))
	t.Equal(proposal.State.Next, t.cstate.State())

	t.Equal(NodeStateConsensus, t.cstate.NodeState())
}

func TestConsensus(t *testing.T) {
	suite.Run(t, new(testConsensus))
}
