package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type NetworkClient interface {
	Home() node.Home
	Propose(*Proposal) error
	Vote(Ballot) error
	RequestNodeInfo(...node.Address) ([]NodeInfo, error)
	RequestLatestBlockProof(...node.Address) error
	RequestBlockProof(hash.Hash, ...node.Address) error // request BlockProof by block hash
}
