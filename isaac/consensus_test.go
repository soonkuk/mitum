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
	sealPool         SealPool
	consensus        *Consensus
	proposerSelector *TProposerSelector
	blockStorage     *TBlockStorage
}

func (t *testConsensus) SetupSuite() {
	t.home = common.NewRandomHome()
	t.height = common.NewBig(0)
	t.block = common.NewRandomHash("bk")
	t.state = []byte("sl")
	t.total = 1
	t.threshold = 1
}

func (t *testConsensus) SetupTest() {
	cstate := &ConsensusState{home: t.home, height: t.height, block: t.block, state: t.state}
	policy := ConsensusPolicy{
		NetworkID:       common.TestNetworkID,
		Total:           t.total,
		Threshold:       t.threshold,
		TimeoutWaitSeal: time.Second * 3,
	}

	votingBox := NewDefaultVotingBox(policy)

	t.nt = network.NewNodeTestNetwork()

	t.sealBroadcaster, _ = NewDefaultSealBroadcaster(policy, t.home)
	t.sealBroadcaster.SetSender(t.nt.Send)

	t.sealPool = NewDefaultSealPool()

	t.proposerSelector = NewTProposerSelector()
	t.proposerSelector.SetProposer(t.home)

	t.blockStorage = NewTBlockStorage()
	t.blocker = NewConsensusBlocker(
		policy,
		cstate,
		votingBox,
		t.sealBroadcaster,
		t.sealPool,
		t.proposerSelector,
		t.blockStorage,
	)
	t.blocker.Start()

	consensus, err := NewConsensus(t.blocker)
	t.NoError(err)
	t.consensus = consensus

	err = t.nt.AddReceiver(t.consensus.Receiver())
	t.NoError(err)

	t.consensus.SetContext(
		nil, // nolint
		"policy", policy,
		"state", cstate,
		"sealPool", t.sealPool,
	)

	t.consensus.Start()
}

func (t *testConsensus) TeardownTest() {
	t.consensus.Stop()
	t.nt.Stop()
	t.blocker.Stop()
}

func (t *testConsensus) TestNew() {
	defer common.DebugPanic()

	state := t.consensus.Context().Value("state").(*ConsensusState)

	proposerSeed := state.Home().Seed()
	var proposal Proposal
	{
		var err error
		proposal = NewTestProposal(proposerSeed.Address(), nil)
		t.NoError(err)
		err = proposal.Sign(common.TestNetworkID, proposerSeed)
		t.NoError(err)

		t.nt.Send(state.Home(), proposal)
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	for range ticker.C {
		if state.Height().Equal(proposal.Block.Height.Inc()) {
			break
		}
	}
	ticker.Stop()

	t.True(proposal.Block.Height.Inc().Equal(state.Height()))
	t.True(proposal.Block.Next.Equal(state.Block()))
	t.Equal(proposal.State.Next, state.State())
}

func TestConsensus(t *testing.T) {
	suite.Run(t, new(testConsensus))
}
