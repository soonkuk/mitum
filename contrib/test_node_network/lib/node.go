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
	ProposerSelector *isaac.TProposerSelector
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
	seed common.Seed,
	height common.Big,
	block common.Hash,
	blockState []byte,
	policy isaac.ConsensusPolicy,
	state *isaac.ConsensusState,
	sealBroadcaster *WrongSealBroadcaster,
) (*Node, error) {
	blockStorage, _ := isaac.NewDefaultBlockStorage(leveldbstorage.NewMemStorage())
	node := &Node{
		Logger:           common.NewLogger(log),
		Home:             state.Home(),
		Policy:           policy,
		State:            state,
		NT:               network.NewNodeTestNetwork(),
		SealBroadcaster:  sealBroadcaster,
		sealPool:         isaac.NewDefaultSealPool(),
		ProposerSelector: isaac.NewTProposerSelector(),
		blockStorage:     blockStorage,
	}
	node.NT.SetLogContext("node", state.Home().Name())
	node.sealPool.SetLogContext("node", state.Home().Name())

	votingBox := isaac.NewDefaultVotingBox(state.Home(), policy)

	blocker := isaac.NewConsensusBlocker(
		policy,
		node.State,
		votingBox,
		node.SealBroadcaster,
		node.sealPool,
		node.ProposerSelector,
		isaac.NewDefaultProposalValidator(node.blockStorage, state),
	)
	consensus, err := isaac.NewConsensus(state.Home(), state, blocker)
	if err != nil {
		return nil, err
	}

	node.Blocker = blocker
	node.Consensus = consensus

	/* NOTE instead of registering channel, validator channel will be used
	if err := node.NT.AddReceiver(consensus.Receiver()); err != nil {
		return nil, err
	}
	*/
	node.NT.AddValidatorChan(state.Home().AsValidator(), consensus.Receiver())

	node.Consensus.SetContext(
		nil, // nolint
		"policy", policy,
		"state", node.State,
		"sealPool", node.sealPool,
		"proposerSelector", node.ProposerSelector,
	)

	node.blockStorage.SetLogContext("node", state.Home().Name())

	return node, nil
}
