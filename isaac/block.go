package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/common"
)

var (
	CurrentBlockVersion common.Version = common.MustParseVersion("0.1.0-proto")
)

type Block struct {
	version common.Version

	height   common.Big
	hash     common.Hash
	prevHash common.Hash

	state     []byte
	prevState []byte

	proposer   common.Address
	round      Round
	proposedAt common.Time
	proposal   common.Hash // Seal(Proposal).Hash()
	validators []common.Validator

	transactions []common.Hash
}

func (b Block) Version() common.Version {
	return b.version
}

func (b Block) Height() common.Big {
	return b.height
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

func (b Block) MarshalJSON() ([]byte, error) {
	return common.EncodeJSON(map[string]interface{}{
		"version":      b.version,
		"height":       b.height,
		"hash":         b.hash,
		"prev_hash":    b.prevHash,
		"state":        b.state,
		"prev_state":   b.prevState,
		"proposer":     b.proposer,
		"round":        b.round,
		"proposed_at":  b.proposedAt,
		"proposal":     b.proposal,
		"validators":   b.validators,
		"transactions": b.transactions,
	}, false, false)
}

func (b Block) String() string {
	by, _ := json.Marshal(b)
	return common.TerminalLogString(string(by))
}
