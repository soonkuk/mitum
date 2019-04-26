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
}

func (t *testConsensus) newConsensus(height common.Big, block common.Hash, state []byte) (*Consensus, network.NodeNetwork) {
	node := common.NewRandomHomeNode()
	cstate := &ConsensusState{node: node, height: height, block: block, state: state}
	policy := ConsensusPolicy{NetworkID: common.TestNetworkID, Total: 1, Threshold: 1}

	consensus, err := NewConsensus(policy, cstate)
	t.NoError(err)

	nt := network.NewNodeTestNetwork()

	nt.AddReceiver(consensus.Receiver())
	consensus.SetSender(nt.Send)

	stageTransistor, _ := NewISAACStageTransistor(policy, cstate, consensus.SealPool(), consensus.Voting())
	stageTransistor.SetSender(nt.Send)
	consensus.SetStageTransistor(stageTransistor)

	consensus.Start()
	stageTransistor.Start()

	return consensus, nt
}

func (t *testConsensus) TestNew() {
	consensus, nt := t.newConsensus(common.NewBig(0), common.NewRandomHash("bk"), []byte("sl"))
	defer consensus.Stop()
	defer consensus.StageTransistor().Stop()
	defer nt.Stop()

	proposerSeed := consensus.State().Node().Seed()
	var proposeBallotSeal common.Seal
	var proposeBallot ProposeBallot
	{
		var err error
		proposeBallot, proposeBallotSeal, err = NewTestSealProposeBallot(proposerSeed.Address(), nil)
		t.NoError(err)
		err = proposeBallotSeal.Sign(common.TestNetworkID, proposerSeed)
		t.NoError(err)

		nt.Send(consensus.State().Node(), proposeBallotSeal)
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	for _ = range ticker.C {
		if consensus.State().Height().Equal(proposeBallot.Block.Height.Inc()) {
			break
		}
	}
	ticker.Stop()
	consensus.Stop()

	t.True(proposeBallot.Block.Height.Inc().Equal(consensus.State().Height()))
	t.True(proposeBallot.Block.Next.Equal(consensus.State().Block()))
	t.Equal(proposeBallot.State.Next, consensus.State().State())
	t.Equal(Round(0), consensus.State().Round())
}

func TestConsensus(t *testing.T) {
	suite.Run(t, new(testConsensus))
}
