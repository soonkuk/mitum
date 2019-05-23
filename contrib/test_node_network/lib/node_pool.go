package lib

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
)

func PrepareNodePool(nodes []*Node) error {
	// proposer selector

	// validators
	var validators []common.Validator
	for _, node := range nodes {
		_ = node.SealBroadcaster.SetSender(node.NT.Send)
		validators = append(validators, node.Home.AsValidator())
	}

	for _, node := range nodes {
		node.NT.SetLogContext("node", node.State.Home().Name())

		// seal pool
		node.sealPool.SetLogContext("node", node.State.Home().Name())

		votingBox := isaac.NewDefaultVotingBox(node.State.Home(), node.Policy)

		blocker := isaac.NewConsensusBlocker(
			node.Policy,
			node.State,
			votingBox,
			node.SealBroadcaster,
			node.sealPool,
			node.ProposerSelector,
			isaac.NewDefaultProposalValidator(node.blockStorage, node.State),
		)
		consensus, err := isaac.NewConsensus(node.State.Home(), node.State, blocker)
		if err != nil {
			return err
		}

		node.Blocker = blocker
		node.Consensus = consensus

		/* NOTE instead of registering channel, validator channel will be used
		if err := node.NT.AddReceiver(consensus.Receiver()); err != nil {
			return nil, err
		}
		*/
		node.NT.AddValidatorChan(node.State.Home().AsValidator(), consensus.Receiver())

		node.Consensus.SetContext(
			nil, // nolint
			"policy", node.Policy,
			"state", node.State,
			"sealPool", node.sealPool,
			"proposerSelector", node.ProposerSelector,
		)

		node.blockStorage.SetLogContext("node", node.State.Home().Name())
	}

	for _, node := range nodes {
		_ = node.State.AddValidators(validators...)

		// network
		for _, other := range nodes {
			if node.Home.Equal(other.Home) {
				continue
			}

			node.NT.AddValidatorChan(other.Home.AsValidator(), other.Consensus.Receiver())
		}
	}

	return nil
}

func StartNodes(nodes []*Node) error {
	// node state
	for _, node := range nodes {
		node.Log().Info("node info", "state", node.State, "policy", node.Policy)
	}

	// starting node
	for _, node := range nodes {
		if err := node.Start(); err != nil {
			node.Log().Error("failed to start", "error", err)
			return err
		}
		node.Log().Debug("started")
	}

	// join
	for _, node := range nodes {
		_ = node.Blocker.Join()
	}

	// node state
	for _, node := range nodes {
		node.Log().Info("node state", "node-state", node.State.NodeState())
	}

	return nil
}

func StopNodes(nodes []*Node) error {
	for _, node := range nodes {
		if err := node.Stop(); err != nil {
			node.Log().Error("failed to stop", "error", err)
			return err
		}
		node.Log().Debug("stopped")
	}

	return nil
}
