package isaac

import (
	"github.com/spikeekips/mitum/common"
)

var (
	CurrentBlockVersion common.Version = common.MustParseVersion("0.1.0-proto")
)

type Block struct {
	version common.Version

	hash     common.Hash
	prevHash common.Hash

	state     []byte
	prevState []byte

	proposer       common.Address
	proposedAt     common.Time
	proposedBallot common.Hash

	transactions []common.Hash
}

func (b Block) Version() common.Version {
	return b.version
}

func (b Block) Hash() common.Hash {
	return b.hash
}

func (b Block) PrevHash() common.Hash {
	return b.prevHash
}

func (b Block) State() []byte {
	return b.state
}

func (b Block) PrevState() []byte {
	return b.prevState
}

func (b Block) Transactions() []common.Hash {
	return b.transactions
}
