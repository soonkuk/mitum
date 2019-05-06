package isaac

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

// TODO rename to Proposal
type Propose struct {
	Version  common.Version `json:"version"`
	Proposer common.Address `json:"proposer"`
	// TODO // Validators   []common.Validator `json:validators`
	Round        Round         `json:"round"`
	Block        ProposeBlock  `json:"block"`
	State        ProposeState  `json:"state"`
	Transactions []common.Hash `json:"transactions"` // NOTE check Hash.p is 'tx'
	ProposedAt   common.Time   `json:"proposed_at"`

	hash    common.Hash
	encoded []byte
}

func (p Propose) makeHash() (common.Hash, []byte, error) {
	encoded, err := p.MarshalBinary()
	if err != nil {
		return common.Hash{}, nil, err
	}

	hash, err := common.NewHash("pb", encoded)
	if err != nil {
		return common.Hash{}, nil, err
	}

	return hash, encoded, nil
}

func (p Propose) Hash() (common.Hash, []byte, error) {
	if !p.hash.IsValid() {
		return p.makeHash()
	}

	return p.hash, p.encoded, nil
}

func (p Propose) MarshalBinary() ([]byte, error) {
	block, err := p.Block.MarshalBinary()
	if err != nil {
		return nil, err
	}

	state, err := p.State.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var transactions [][]byte
	for _, t := range p.Transactions {
		var hash []byte
		if hash, err = t.MarshalBinary(); err != nil {
			return nil, err
		}
		transactions = append(transactions, hash)
	}

	version, err := p.Version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return common.Encode([]interface{}{
		version,
		p.Proposer,
		p.Round,
		block,
		state,
		transactions,
		p.ProposedAt,
	})
}

func (p *Propose) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := common.Decode(b, &m); err != nil {
		return err
	}

	var version common.Version
	{
		var vs []byte
		if err := common.Decode(m[0], &vs); err != nil {
			return err
		} else if err := version.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var proposer common.Address
	if err := common.Decode(m[1], &proposer); err != nil {
		return err
	}

	var round Round
	if err := common.Decode(m[2], &round); err != nil {
		return err
	}

	var block ProposeBlock
	{
		var vs []byte
		if err := common.Decode(m[3], &vs); err != nil {
			return err
		} else if err := block.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var state ProposeState
	{
		var vs []byte
		if err := common.Decode(m[4], &vs); err != nil {
			return err
		} else if err := state.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var transactions []common.Hash
	{
		var vs [][]byte
		if err := common.Decode(m[5], &vs); err != nil {
			return err
		}

		for _, e := range vs {
			var hash common.Hash
			if err := hash.UnmarshalBinary(e); err != nil {
				return err
			}

			transactions = append(transactions, hash)
		}
	}

	var proposedAt common.Time
	if err := common.Decode(m[6], &proposedAt); err != nil {
		return err
	}

	p.Version = version
	p.Proposer = proposer
	p.Round = round
	p.Block = block
	p.State = state
	p.Transactions = transactions
	p.ProposedAt = proposedAt

	hash, encoded, err := p.makeHash()
	if err != nil {
		return err
	}

	p.hash = hash
	p.encoded = encoded

	return nil
}

func (p Propose) String() string {
	b, _ := json.Marshal(p)
	return common.TerminalLogString(string(b))
}

func (p Propose) Wellformed() error {
	if _, err := p.Proposer.IsValid(); err != nil {
		return err
	}

	if err := p.Block.Wellformed(); err != nil {
		return err
	}

	if err := p.State.Wellformed(); err != nil {
		return err
	}

	for _, th := range p.Transactions {
		if !th.IsValid() {
			return ProposeNotWellformedError.SetMessage(
				"empty Hash found in Propose.Transactions",
			)
		}
	}

	return nil
}

type ProposeBlock struct {
	Height  common.Big  `json:"height"`
	Current common.Hash `json:"current"`
	Next    common.Hash `json:"next"`
}

func (bb ProposeBlock) MarshalBinary() ([]byte, error) {
	currentHash, err := bb.Current.MarshalBinary()
	if err != nil {
		return nil, err
	}

	nextHash, err := bb.Next.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return common.Encode([]interface{}{
		bb.Height,
		currentHash,
		nextHash,
	})
}

func (bb *ProposeBlock) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := common.Decode(b, &m); err != nil {
		return err
	}

	var height common.Big
	if err := common.Decode(m[0], &height); err != nil {
		return err
	}

	var current, next common.Hash
	var currentByte, nextByte []byte

	if err := common.Decode(m[1], &currentByte); err != nil {
		return err
	} else if err := current.UnmarshalBinary(currentByte); err != nil {
		return err
	}

	if err := common.Decode(m[2], &nextByte); err != nil {
		return err
	} else if err := next.UnmarshalBinary(nextByte); err != nil {
		return err
	}

	bb.Height = height
	bb.Current = current
	bb.Next = next

	return nil
}

func (bb ProposeBlock) Wellformed() error {
	if !bb.Current.IsValid() {
		return ProposeNotWellformedError.SetMessage("Propose.Block.Current is empty")
	}

	if !bb.Next.IsValid() {
		return ProposeNotWellformedError.SetMessage("Propose.Block.Next is empty")
	}

	return nil
}

type ProposeState struct {
	Current []byte
	Next    []byte
}

func (bb ProposeState) MarshalBinary() ([]byte, error) {
	return common.Encode([]interface{}{
		bb.Current,
		bb.Next,
	})
}

func (bb *ProposeState) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := common.Decode(b, &m); err != nil {
		return err
	}

	var current, next []byte

	if err := common.Decode(m[0], &current); err != nil {
		return err
	}

	if err := common.Decode(m[1], &next); err != nil {
		return err
	}

	bb.Current = current
	bb.Next = next

	return nil
}

func (bb ProposeState) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"current": base64.StdEncoding.EncodeToString(bb.Current),
		"next":    base64.StdEncoding.EncodeToString(bb.Next),
	})
}

func (bb *ProposeState) UnmarshalJSON(b []byte) error {
	var a map[string]string
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	var current, next []byte
	if d, err := base64.StdEncoding.DecodeString(a["current"]); err != nil {
		return err
	} else {
		current = d
	}

	if d, err := base64.StdEncoding.DecodeString(a["next"]); err != nil {
		return err
	} else {
		next = d
	}

	bb.Current = current
	bb.Next = next

	return nil
}

func (bb ProposeState) Wellformed() error {
	if len(bb.Current) < 1 {
		return ProposeNotWellformedError.SetMessage("Propose.State.Current is empty")
	}

	if len(bb.Next) < 1 {
		return ProposeNotWellformedError.SetMessage("Propose.State.Next is empty")
	}

	return nil
}
