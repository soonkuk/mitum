package lib

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/storage/leveldbstorage"
)

type Node struct {
	*common.Logger
	Home             common.HomeNode
	Policy           isaac.ConsensusPolicy
	State            *isaac.ConsensusState
	NT               *network.NodeTestNetwork
	SealBroadcaster  *WrongSealBroadcaster
	Blocker          *isaac.ConsensusBlocker
	sealPool         *isaac.DefaultSealPool
	ProposerSelector isaac.ProposerSelector
	blockStorage     *isaac.DefaultBlockStorage
	Consensus        *isaac.Consensus
}

func (n *Node) Name() string {
	return n.Home.Name()
}

func (n *Node) Start() error {
	if err := n.NT.Start(); err != nil {
		return err
	}

	if err := n.Blocker.Start(); err != nil {
		return err
	}

	if err := n.Consensus.Start(); err != nil {
		return err
	}

	return nil
}

func (n *Node) Stop() error {
	if err := n.NT.Stop(); err != nil {
		return err
	}

	if err := n.Blocker.Stop(); err != nil {
		return err
	}

	if err := n.Consensus.Stop(); err != nil {
		return err
	}

	return nil
}

func CreateNode(
	seedString string,
	height common.Big,
	block common.Hash,
	blockState []byte,
	policy isaac.ConsensusPolicy,
) (*Node, error) {
	blockStorage, _ := isaac.NewDefaultBlockStorage(leveldbstorage.NewMemStorage())
	seed, err := common.SeedFromString(seedString)
	if err != nil {
		log.Error("failed to parse seedString", "error", err)
		return nil, err
	}

	home := common.NewHome(seed, common.NetAddr{})
	state := isaac.NewConsensusState(home)
	state.SetLogContext("node", home.Name())
	_ = state.SetHeight(height)
	_ = state.SetBlock(block)
	_ = state.SetState(blockState)

	sealBroadcaster, _ := NewWrongSealBroadcaster(policy, state)

	node := &Node{
		Logger:          common.NewLogger(log),
		Home:            state.Home(),
		Policy:          policy,
		State:           state,
		NT:              network.NewNodeTestNetwork(),
		SealBroadcaster: sealBroadcaster,
		sealPool:        isaac.NewDefaultSealPool(),
		//ProposerSelector: isaac.NewTProposerSelector(),
		blockStorage: blockStorage,
	}
	return node, nil
}
