// +build test

package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
)

type ClientTest struct {
	networkID    NetworkID
	home         node.Home
	nodesNetwork network.Nodes
}

func (ct ClientTest) Home() node.Home {
	return ct.home
}

func (ct ClientTest) RequestNodeInfo(address ...node.Address) ([]NodeInfo, error) {
	return nil, nil
}

func (ct ClientTest) Vote(ballot Ballot) error {
	if err := ballot.(keypair.Signer).Sign(ct.home.PrivateKey(), ct.networkID); err != nil {
		return err
	}

	// NOTE broadcast:
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
