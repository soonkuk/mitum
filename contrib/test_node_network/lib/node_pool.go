package lib

import (
	"github.com/spikeekips/mitum/common"
)

func PrepareNodePool(nodes []*Node) error {
	// proposer selector

	// validators
	var validators []common.Validator
	for _, node := range nodes {
		_ = node.SealBroadcaster.SetSender(node.NT.Send)
		validators = append(validators, node.Home.AsValidator())
		node.ProposerSelector.SetProposer(nodes[0].Home)
	}

	for _, node := range nodes {
		_ = node.State.AddValidators(validators...)

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
