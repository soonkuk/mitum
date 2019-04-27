package isaac

import (
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/network"
)

type testConsensus struct {
	suite.Suite
}

func (t *testConsensus) newConsensus(height common.Big, block common.Hash, state []byte) (*Consensus, network.NodeNetwork) {
	node := common.NewRandomHomeNode()
	cstate := &ConsensusState{node: node, height: height, block: block, state: state}
	policy := ConsensusPolicy{NetworkID: common.TestNetworkID, Total: 1, Threshold: 1}

	consensus, err := NewConsensus()
	t.NoError(err)

	nt := network.NewNodeTestNetwork()
	nt.AddReceiver(consensus.Receiver())

	sb, _ := NewISAACSealBroadcaster(policy, cstate)
	sb.SetSender(nt.Send)

	rv := NewRoundVoting()
	sealPool := NewISAACSealPool()
	roundBoy, _ := NewISAACRoundBoy(policy, cstate, sealPool, rv)
	roundBoy.SetBroadcaster(sb)

	bs, _ := NewISAACBlockStorage(cstate)

	consensus.SetContext(
		nil,
		"state", cstate,
		"blockStorage", bs,
		"policy", policy,
		"roundVoting", rv,
		"roundBoy", roundBoy,
		"sealPool", sealPool,
	)

	consensus.Start()
	roundBoy.Start()

	return consensus, nt
}

func (t *testConsensus) TestNew() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}
	}()

	consensus, nt := t.newConsensus(common.NewBig(0), common.NewRandomHash("bk"), []byte("sl"))
	defer consensus.Stop()
	defer nt.Stop()

	roundBoy := consensus.Context().Value("roundBoy").(RoundBoy)
	defer roundBoy.Stop()

	state := consensus.Context().Value("state").(*ConsensusState)

	proposerSeed := state.Node().Seed()
	var proposeSeal common.Seal
	var propose Propose
	{
		var err error
		propose, proposeSeal, err = NewTestSealPropose(proposerSeed.Address(), nil)
		t.NoError(err)
		err = proposeSeal.Sign(common.TestNetworkID, proposerSeed)
		t.NoError(err)

		nt.Send(state.Node(), proposeSeal)
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	for _ = range ticker.C {
		if state.Height().Equal(propose.Block.Height.Inc()) {
			break
		}
	}
	ticker.Stop()
	consensus.Stop()

	t.True(propose.Block.Height.Inc().Equal(state.Height()))
	t.True(propose.Block.Next.Equal(state.Block()))
	t.Equal(propose.State.Next, state.State())
}

func TestConsensus(t *testing.T) {
	suite.Run(t, new(testConsensus))
}
