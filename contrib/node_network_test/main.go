package main

import (
	"time"

	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/storage/leveldbstorage"
)

var log log15.Logger = log15.New("module", "node-network-test")

func init() {
	common.InTest = false

	handler, _ := common.LogHandler(common.LogFormatter("json"), "")
	handler = log15.LvlFilterHandler(log15.LvlDebug, handler)
	handler = log15.CallerFileHandler(handler)

	loggers := []log15.Logger{
		log,
		common.Log(),
		isaac.Log(),
		network.Log(),
		storage.Log(),
	}
	for _, l := range loggers {
		l.SetHandler(handler)
	}

	{
		syncer, _ := common.NewTimeSyncer("zero.bora.net", time.Second*10)
		_ = syncer.Start()
		common.SetTimeSyncer(syncer)
	}
}

var (
	seeds []string = []string{
		"SACCEWPT5O5RW67F53M6TMVELMJ34DC5EDWQPSG4RRTQ72BJBSASJTWY", // GAEU.6BSA
		"SB4NCVQZQDX6NYVVQPIAJJGES3GOK5PUJI7DP6ZTJV62N3K7CS6JN5XN",
		"SAPIKULLJIWWSR3O3RKW3ZGIRAMJWFPDIR7B6H54ADSNWKMGHGU4HBQO",
		"SCVEZIL32YTRJBFQYJSZZDNAE2CZBRABMKG524LZPBQXHFLBP3OPYFLK",
		// "SAUR7Z7SPYCMAILQYMFWMGNFLPFQKI3DRPLNNEUBCHT23N2GSSHWW7QS",
		// "SDHMHAPZ3HCHK54MDTPYGTBN5YISVZNQP6U2ZYIXQNQINH7FLWU7IK3S",
		// "SAAMSDCC5WX6H3EMKN5FK2FRCMN5GA2HULXVBKJDPQZN4ZKHUZWSPQQM",
		// "SDGBJJQAWEE3JRNNXPFZPF4ESE3JAJVV5EM3YGVZVX7Q24VBIVGGDNLE",
		// "SBP3FD3YDKSJD36CWJF4KWYGMYVYS5ADRYKKNZ5YS2BKVOT5NTWPMZD7",
		// "SBJNPBE4RPLKV5C2PP47NPEWUNJ6TKKBYGDAWMXHSRBBZER6SEB255BW",
		// "SDKHEIABEUHJFDX5LXTQHSLR3DUB2TKLQIQPUN3NZGWQLWNORMTINHPI",
		// "SDTWLWDSIBYZUYF7VROKJ4GUQ4ABPUY5EKFJGOPMIX72P4EOEG2PTP7M",
		// "SDEPMOUYNUNWIEIGRFHR6JMUHRA35TFWRM4LDFCALW6H7XCWTXAVZRC7",
	}
	height     common.Big            = common.NewBig(33)
	block      common.Hash           = common.NewRandomHash("bk")
	blockState []byte                = []byte("initial state")
	policy     isaac.ConsensusPolicy = isaac.ConsensusPolicy{
		NetworkID:                 common.TestNetworkID,
		Total:                     uint(len(seeds)),
		Threshold:                 uint(len(seeds) - 1),
		TimeoutWaitSeal:           time.Second * 3,
		AvgBlockRoundInterval:     time.Second * 5,
		SealSignedAtAllowDuration: time.Second * 3,
	}
)

func createNode(seedString string) (*Node, error) {
	seed, err := common.SeedFromString(seedString)
	if err != nil {
		return nil, err
	}
	home := common.NewHome(seed, common.NetAddr{})
	state := isaac.NewConsensusState(home)
	_ = state.SetHeight(height)
	_ = state.SetBlock(block)
	_ = state.SetState(blockState)

	sealBroadcaster, err := isaac.NewDefaultSealBroadcaster(policy, state)
	if err != nil {
		return nil, err
	}

	blockStorage, _ := isaac.NewDefaultBlockStorage(leveldbstorage.NewMemStorage())
	node := &Node{
		home:             home,
		state:            state,
		nt:               network.NewNodeTestNetwork(),
		sealBroadcaster:  sealBroadcaster,
		sealPool:         isaac.NewDefaultSealPool(),
		proposerSelector: isaac.NewTProposerSelector(),
		blockStorage:     blockStorage,
	}
	_ = node.sealBroadcaster.SetSender(node.nt.Send)
	node.sealPool.SetLogContext("node", home.Name())

	votingBox := isaac.NewDefaultVotingBox(home, policy)

	blocker := isaac.NewConsensusBlocker(
		policy,
		node.state,
		votingBox,
		node.sealBroadcaster,
		node.sealPool,
		node.proposerSelector,
		isaac.NewDefaultProposalValidator(node.blockStorage, state),
	)
	consensus, err := isaac.NewConsensus(home, state, blocker)
	if err != nil {
		return nil, err
	}

	node.blocker = blocker
	node.consensus = consensus

	/* NOTE instead of registering channel, validator channel will be used
	if err := node.nt.AddReceiver(consensus.Receiver()); err != nil {
		return nil, err
	}
	*/
	node.nt.AddValidatorChan(home.AsValidator(), consensus.Receiver())

	node.consensus.SetContext(
		nil, // nolint
		"policy", policy,
		"state", node.state,
		"sealPool", node.sealPool,
		"proposerSelector", node.proposerSelector,
	)

	node.log = log.New(log15.Ctx{
		"node": node.Name(),
	})
	node.blockStorage.SetLogContext("node", home.Name())

	return node, nil
}

type Node struct {
	home             common.HomeNode
	log              log15.Logger
	state            *isaac.ConsensusState
	nt               *network.NodeTestNetwork
	sealBroadcaster  *isaac.DefaultSealBroadcaster
	blocker          *isaac.ConsensusBlocker
	sealPool         *isaac.DefaultSealPool
	proposerSelector *isaac.TProposerSelector
	blockStorage     *isaac.DefaultBlockStorage
	consensus        *isaac.Consensus
}

func (n *Node) Name() string {
	return n.home.Name()
}

func (n *Node) Start() error {
	if err := n.nt.Start(); err != nil {
		return err
	}

	if err := n.blocker.Start(); err != nil {
		return err
	}

	if err := n.consensus.Start(); err != nil {
		return err
	}

	return nil
}

func (n *Node) Stop() error {
	if err := n.nt.Stop(); err != nil {
		return err
	}

	if err := n.blocker.Stop(); err != nil {
		return err
	}

	if err := n.consensus.Stop(); err != nil {
		return err
	}

	return nil
}

func main() {
	log.Debug("starting", "policy", policy)

	var nodes []*Node
	for _, seed := range seeds {
		node, err := createNode(seed)
		if err != nil {
			panic(err)
		}
		node.log.Debug("node created")
		nodes = append(nodes, node)
	}

	// proposer is 1st node
	for _, node := range nodes {
		node.proposerSelector.SetProposer(nodes[0].home)
	}

	// connecting each others
	var validators []common.Validator
	for _, node := range nodes {
		validators = append(validators, node.home.AsValidator())
	}

	for _, node := range nodes {
		_ = node.state.AddValidators(validators...)

		for _, other := range nodes {
			if node.home.Equal(other.home) {
				continue
			}

			node.nt.AddValidatorChan(other.home.AsValidator(), other.consensus.Receiver())
		}
	}

	defer func(nodes []*Node) {
		for _, node := range nodes {
			if err := node.Stop(); err != nil {
				node.log.Error("failed to stop", "error", err)
				return
			}
			node.log.Debug("stopped")
		}
	}(nodes)

	// node state
	for _, node := range nodes {
		node.log.Info("node state", "node-state", node.state.NodeState())
	}

	// starting node
	for _, node := range nodes {
		if err := node.Start(); err != nil {
			node.log.Error("failed to start", "error", err)
			return
		}
		node.log.Debug("started")
	}

	for _, node := range nodes {
		_ = node.blocker.Join()
	}

	// node state
	for _, node := range nodes {
		node.log.Info("node state", "node-state", node.state.NodeState())
	}

	select {}
}
