package isaac

import (
	"runtime/debug"
	"strconv"
	"sync"
	"testing"
	"unsafe"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/network"
	"github.com/stretchr/testify/suite"
)

type testHash struct {
	I uint64
}

func (t testHash) MarshalBinary() ([]byte, error) {
	return []byte(strconv.FormatUint(t.I, 10)), nil
}

func (t *testHash) UnmarshalBinary(b []byte) error {
	i, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}

	t.I = i
	return nil
}

func (t testHash) Hash() (common.Hash, []byte, error) {
	encoded, err := t.MarshalBinary()
	if err != nil {
		return common.Hash{}, nil, err
	}

	hash, _ := common.NewHash("th", encoded)
	return hash, encoded, nil
}

type nodeTestNetwork struct {
	sync.RWMutex
	chans map[int64]chan<- common.Seal
}

func newNodeTestNetwork() *nodeTestNetwork {
	return &nodeTestNetwork{
		chans: map[int64]chan<- common.Seal{},
	}
}

func (n *nodeTestNetwork) Start() error {
	return nil
}

func (n *nodeTestNetwork) Stop() error {
	return nil
}

func (n *nodeTestNetwork) addSeal(seal common.Seal) error {
	n.RLock()
	defer n.RUnlock()

	if len(n.chans) < 1 {
		return network.NoReceiversError
	}

	for _, c := range n.chans {
		c <- seal
	}

	return nil
}

func (n *nodeTestNetwork) RegisterReceiver(c chan<- common.Seal) error {
	n.Lock()
	defer n.Unlock()

	p := *(*int64)(unsafe.Pointer(&c))
	if _, found := n.chans[p]; found {
		return network.ReceiverAlreadyRegisteredError
	}

	n.chans[p] = c
	return nil
}

func (n *nodeTestNetwork) UnregisterReceiver(c chan<- common.Seal) error {
	n.Lock()
	defer n.Unlock()

	p := *(*int64)(unsafe.Pointer(&c))
	if _, found := n.chans[p]; !found {
		return network.ReceiverNotRegisteredError
	}

	delete(n.chans, p)
	return nil
}

func (n *nodeTestNetwork) Send(node common.Node, seal common.Seal) error {
	return nil
}

type testConsensus struct {
	suite.Suite
}

func (t *testConsensus) newSeal(c uint64) common.Seal {
	seal, err := common.NewSeal(VoteBallotSealType, testHash{I: c})
	t.NoError(err)

	return seal
}

func (t *testConsensus) TestNew() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	networkID := []byte("this-is-test-network")

	network := newNodeTestNetwork()
	consensus, err := NewConsensus()
	t.NoError(err)
	defer consensus.Stop()
	defer network.Stop()

	consensus.Start()

	network.RegisterReceiver(consensus.Receiver())
	consensus.RegisterSendFunc(network.Send)

	proposerSeed := common.RandomSeed()
	var proposeBallotSeal common.Seal
	{
		var err error
		_, proposeBallotSeal, err = NewTestSealProposeBallot(proposerSeed.Address(), nil)
		t.NoError(err)
		err = proposeBallotSeal.Sign(networkID, proposerSeed)
		t.NoError(err)

		network.addSeal(proposeBallotSeal)
	}

	voteSeed := common.RandomSeed()
	stage := VoteStageSIGN
	vote := VoteYES
	{
		sealHash, _, err := proposeBallotSeal.Hash()
		t.NoError(err)
		_, seal, err := NewTestSealVoteBallot(sealHash, voteSeed.Address(), stage, vote)
		t.NoError(err)
		err = seal.Sign(networkID, voteSeed)
		t.NoError(err)

		network.addSeal(seal)
	}
}

func TestConsensus(t *testing.T) {
	suite.Run(t, new(testConsensus))
}
