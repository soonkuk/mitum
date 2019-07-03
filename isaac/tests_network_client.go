// +build test

package isaac

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
)

type ClientTest struct {
	*common.Logger
	networkID    NetworkID
	nodesNetwork network.Nodes
}

func NewClientTest(nodesNetwork network.Nodes) ClientTest {
	return ClientTest{
		Logger:       common.NewLogger(Log(), "module", "network-client", "node", nodesNetwork.Home().Address()),
		nodesNetwork: nodesNetwork,
	}
}

func (ct ClientTest) Home() node.Home {
	return ct.nodesNetwork.Home()
}

func (ct ClientTest) RequestNodeInfo(address ...node.Address) ([]NodeInfo, error) {
	return nil, nil
}

func (ct ClientTest) Propose(proposal *Proposal) error {
	if err := proposal.Sign(ct.Home().PrivateKey(), ct.networkID); err != nil {
		return err
	}

	// TODO NOTE broadcast:
	// 1. send to the active suffrage members by Ballot.Stage()
	// 2. and then, broadcast to the suffrage network
	ct.Log().Debug("broadcast Proposal", "proposal", proposal)
	return ct.nodesNetwork.Broadcast(*proposal)
}

func (ct ClientTest) Vote(ballot Ballot) error {
	if err := ballot.(keypair.Signer).Sign(ct.Home().PrivateKey(), ct.networkID); err != nil {
		return err
	}

	// TODO NOTE broadcast:
	// 1. send to the active suffrage members by Ballot.Stage()
	// 2. and then, broadcast to the suffrage network
	return ct.nodesNetwork.Broadcast(ballot)
}

func (ct ClientTest) RequestLatestBlockProof(addresses ...node.Address) error {
	// TODO implement
	return nil
}

func (ct ClientTest) RequestBlockProof(block hash.Hash, addresses ...node.Address) error {
	// TODO implement
	return nil
}