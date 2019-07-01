package isaac

import (
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type Client interface {
	Home() node.Home
	Vote(Ballot) error
	RequestNodeInfo(...node.Address) ([]NodeInfo, error)
	RequestLatestBlockProof(...node.Address) error
	RequestBlockProof(hash.Hash, ...node.Address) error // request BlockProof by block hash
}
