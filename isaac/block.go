package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/common"
)

var (
	CurrentBlockVersion common.Version = common.MustParseVersion("0.1.0-proto")
)

type Block struct {
	version      common.Version
	hash         common.Hash
	createdAt    common.Time // NOTE ignored for hash
	height       common.Big
	prevHash     common.Hash
	state        []byte
	prevState    []byte
	proposer     common.Address
	round        Round
	validators   []common.Validator
	transactions []common.Hash
	proposal     common.Hash
	proposedAt   common.Time
}

func NewBlockFromProposal(proposal Proposal) (Block, error) {
	block := Block{
		version:      CurrentBlockVersion,
		height:       proposal.Block.Height.Inc(),
		prevHash:     proposal.Block.Current,
		state:        proposal.State.Next,
		prevState:    proposal.State.Current,
		proposer:     proposal.Source(),
		round:        proposal.Round,
		transactions: proposal.Transactions,
		proposal:     proposal.Hash(),
		proposedAt:   proposal.SignedAt(),
		createdAt:    common.Now(),
	}

	hash, err := block.generateHash()
	if err != nil {
		return Block{}, err
	}

	block.hash = hash

	return block, nil
}

func (b Block) generateHash() (common.Hash, error) {
	s, err := b.serializeRLP(true /* skipHash */)
	if err != nil {
		return common.Hash{}, err
	}

	encoded, err := common.Encode(s[2:]) // NOTE ignore hash and createdAt
	if err != nil {
		return common.Hash{}, err
	}

	hash, err := common.NewHash("bk", encoded)
	if err != nil {
		return common.Hash{}, err
	}

	return hash, nil
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

func (b Block) Proposal() common.Hash {
	return b.proposal
}

func (b Block) Transactions() []common.Hash {
	return b.transactions
}

func (b Block) serializeRLP(skipHash bool) ([]interface{}, error) {
	version, err := b.version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	hash, err := b.hash.MarshalBinary()
	if !skipHash && err != nil {
		return nil, err
	}

	prevHash, err := b.prevHash.MarshalBinary()
	if err != nil {
		return nil, err
	}

	proposedAt, err := b.proposedAt.MarshalBinary()
	if err != nil {
		return nil, err
	}

	proposal, err := b.proposal.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var transactions [][]byte
	for _, t := range b.transactions {
		h, err := t.MarshalBinary()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, h)
	}

	return []interface{}{
		hash,
		b.createdAt,
		version,
		b.height,
		prevHash,
		b.state,
		b.prevState,
		b.proposer,
		b.round,
		proposedAt,
		proposal,
		transactions,
	}, nil
}

func (b Block) MarshalBinary() ([]byte, error) {
	s, err := b.serializeRLP(false)
	if err != nil {
		return nil, err
	}

	return common.Encode(s)
}

func (b Block) UnmarshalBinary(y []byte) error {
	// TODO
	return nil
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
		"created_at":   b.createdAt,
	}, false, false)
}

func (b Block) String() string {
	by, _ := json.Marshal(b)
	return common.TerminalLogString(string(by))
}
