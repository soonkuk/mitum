package isaac

import (
	"testing"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testRoundBoyCases struct {
	suite.Suite

	numberOfValidators uint
	threshold          uint

	homeNode    common.HomeNode
	validators  map[common.Address]common.HomeNode
	nodes       map[common.Address]common.HomeNode
	state       *ConsensusState
	policy      ConsensusPolicy
	roundBoy    *DefaultRoundBoy
	broadcaster *TestSealBroadcaster
}

func (t *testRoundBoyCases) SetupSuite() {
	t.numberOfValidators = 4
	t.threshold = 3

	t.policy = ConsensusPolicy{
		NetworkID: common.TestNetworkID,
		Total:     t.numberOfValidators,
		Threshold: t.threshold,
	}

	t.homeNode = common.NewHomeNode(common.RandomSeed(), common.NetAddr{})

	var validators []common.HomeNode
	for i := 0; i < int(t.numberOfValidators-1); i++ {
		node := common.NewHomeNode(common.RandomSeed(), common.NetAddr{})
		validators = append(validators, node)
	}
	t.validators = map[common.Address]common.HomeNode{}
	t.nodes = map[common.Address]common.HomeNode{}
	t.nodes[t.homeNode.Address()] = t.homeNode

	for _, node := range validators {
		t.validators[node.Address()] = node
		t.homeNode.AddValidators(node.AsValidator())
		t.nodes[node.Address()] = node
	}
}

func (t *testRoundBoyCases) SetupTest() {
	t.state = &ConsensusState{
		node:   t.homeNode,
		height: common.NewBig(99),
		block:  common.NewRandomHash("bk"),
		state:  []byte("1st state"),
	}

	t.broadcaster, _ = NewTestSealBroadcaster(t.policy, t.homeNode)
	t.roundBoy, _ = NewDefaultRoundBoy(t.homeNode)
	t.roundBoy.SetBroadcaster(t.broadcaster)
	t.roundBoy.Start()
}

func (t *testRoundBoyCases) TearDownTest() {
	t.roundBoy.Stop()
}

func (t *testRoundBoyCases) newProposeSeal(proposer common.Address) common.Seal {
	propose, err := NewTestPropose(proposer, nil)
	t.NoError(err)

	propose.Block.Height = t.state.height
	propose.Block.Current = t.state.block
	propose.State.Current = t.state.state

	seal, err := common.NewSeal(ProposeSealType, propose)
	err = seal.Sign(common.TestNetworkID, t.nodes[proposer].Seed())
	t.NoError(err)

	return seal
}

func (t *testRoundBoyCases) getNewSeal() []common.Seal {
	var seals []common.Seal

	ticker := time.NewTicker(time.Millisecond)
	stop := make(chan bool)
	go func() {
		for _ = range ticker.C {
			s := t.broadcaster.NewSeals()
			if len(s) < 1 {
				continue
			}
			seals = s
			break
		}
		stop <- true
	}()

	select {
	case <-stop:
		break
	case <-time.After(time.Second * 2):
		break
	}

	if len(seals) < 1 {
		return nil
	}

	return seals
}

func (t *testRoundBoyCases) TestBasic() {
	proposer := t.homeNode.Address()
	seal := t.newProposeSeal(proposer)
	psHash, _, err := seal.Hash()
	t.NoError(err)

	err = t.roundBoy.Transit(VoteStageINIT, psHash, VoteYES)
	t.NoError(err)

	seals := t.getNewSeal()
	t.NotNil(seals)
}

func TestRoundBoyCases(t *testing.T) {
	suite.Run(t, new(testRoundBoyCases))
}
