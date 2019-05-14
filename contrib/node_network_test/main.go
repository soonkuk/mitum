package main

import (
	"time"

	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
)

var log log15.Logger = log15.New("module", "♾")

func init() {
	common.InTest = false
	common.DEBUG = true

	//handler, _ := common.LogHandler(common.LogFormatter("terminal"), "")
	handler, _ := common.LogHandler(common.LogFormatter("json"), "")
	handler = log15.LvlFilterHandler(log15.LvlDebug, handler)
	handler = log15.CallerFileHandler(handler)

	loggers := []log15.Logger{
		log,
		common.Log(),
		isaac.Log(),
		network.Log(),
	}
	for _, l := range loggers {
		l.SetHandler(handler)
	}
}

var (
	seeds []string = []string{
		"SACCEWPT5O5RW67F53M6TMVELMJ34DC5EDWQPSG4RRTQ72BJBSASJTWY", // GAEU.6BSA
		"SB4NCVQZQDX6NYVVQPIAJJGES3GOK5PUJI7DP6ZTJV62N3K7CS6JN5XN",
		"SAPIKULLJIWWSR3O3RKW3ZGIRAMJWFPDIR7B6H54ADSNWKMGHGU4HBQO",
		"SCVEZIL32YTRJBFQYJSZZDNAE2CZBRABMKG524LZPBQXHFLBP3OPYFLK",
	}
	height     common.Big            = common.NewBig(33)
	block      common.Hash           = common.NewRandomHash("bk")
	blockState []byte                = []byte("initial state")
	policy     isaac.ConsensusPolicy = isaac.ConsensusPolicy{
		NetworkID:             common.TestNetworkID,
		Total:                 uint(len(seeds)),
		Threshold:             uint(len(seeds) - 1),
		TimeoutWaitSeal:       time.Second * 3,
		AvgBlockRoundInterval: time.Second * 5,
	}
)

func createNode(seedString string) (*Node, error) {
	seed, err := common.SeedFromString(seedString)
	if err != nil {
		return nil, err
	}
	home := common.NewHome(seed, common.NetAddr{})
	state := isaac.NewConsensusState(home)
	state.SetHeight(height)
	state.SetBlock(block)
	state.SetState(blockState)

	sealBroadcaster, err := isaac.NewDefaultSealBroadcaster(policy, home)
	if err != nil {
		return nil, err
	}

	node := &Node{
		home:             home,
		state:            state,
		nt:               network.NewNodeTestNetwork(),
		sealBroadcaster:  sealBroadcaster,
		sealPool:         isaac.NewDefaultSealPool(home),
		proposerSelector: isaac.NewTProposerSelector(),
		blockStorage:     isaac.NewTBlockStorage(),
	}
	node.sealBroadcaster.SetSender(node.nt.Send)

	votingBox := isaac.NewDefaultVotingBox(home, policy)

	blocker := isaac.NewConsensusBlocker(
		policy,
		node.state,
		votingBox,
		node.sealBroadcaster,
		node.sealPool,
		node.proposerSelector,
		node.blockStorage,
	)
	node.blocker = blocker

	consensus, err := isaac.NewConsensus(home, state, node.blocker)
	if err != nil {
		return nil, err
	}

	node.consensus = consensus

	if err := node.nt.AddReceiver(consensus.Receiver()); err != nil {
		return nil, err
	}

	node.consensus.SetContext(
		nil, // nolint
		"policy", policy,
		"state", node.state,
		"sealPool", node.sealPool,
	)

	node.log = log.New(log15.Ctx{
		"node": node.Name(),
	})

	return node, nil
}

type Node struct {
	home             *common.HomeNode
	log              log15.Logger
	state            *isaac.ConsensusState
	nt               *network.NodeTestNetwork
	sealBroadcaster  *isaac.DefaultSealBroadcaster
	blocker          *isaac.ConsensusBlocker
	sealPool         isaac.SealPool
	proposerSelector *isaac.TProposerSelector
	blockStorage     *isaac.TBlockStorage
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
	for i, seed := range seeds {
		node, err := createNode(seed)
		if err != nil {
			panic(err)
		}
		node.log.Debug("node created", "count", i)
		nodes = append(nodes, node)
	}

	// proposer is 1st node
	for _, node := range nodes {
		node.proposerSelector.SetProposer(nodes[0].home)
	}

	// connecting node's network
	for _, node := range nodes {
		for _, other := range nodes {
			if node.home.Equal(other.home) {
				continue
			}

			if err := node.nt.AddReceiver(other.consensus.Receiver()); err != nil {
				node.log.Error("failed to connect nodes", "other", other.Name())
				return
			}
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
