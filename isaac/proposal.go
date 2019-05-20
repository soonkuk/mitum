package isaac

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

// TODO reshape to be similar to Ballot
type Proposal struct {
	common.RawSeal
	// TODO
	// Validators   []common.Validator `json:validators`
	Round        Round         `json:"round"`
	Block        ProposalBlock `json:"block"`
	State        ProposalState `json:"state"`
	Transactions []common.Hash `json:"transactions"` // NOTE check Hash.p is 'tx'
}

func NewProposal(
	round Round,
	block ProposalBlock,
	state ProposalState,
	transactions []common.Hash,
) Proposal {
	p := Proposal{
		Round:        round,
		Block:        block,
		State:        state,
		Transactions: transactions,
	}

	raw := common.NewRawSeal(p, CurrentBallotVersion)
	p.RawSeal = raw

	return p
}

func (p Proposal) Type() common.SealType {
	return ProposalSealType
}

func (p Proposal) Hint() string {
	return "pp"
}

func (p Proposal) SerializeRLP() ([]interface{}, error) {
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

	return []interface{}{
		p.Round,
		block,
		state,
		transactions,
	}, nil
}

func (p *Proposal) UnserializeRLP(m []rlp.RawValue) error {
	var round Round
	if err := common.Decode(m[6], &round); err != nil {
		return err
	}

	var block ProposalBlock
	{
		var vs []byte
		if err := common.Decode(m[7], &vs); err != nil {
			return err
		} else if err := block.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var state ProposalState
	{
		var vs []byte
		if err := common.Decode(m[8], &vs); err != nil {
			return err
		} else if err := state.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var transactions []common.Hash
	{
		var vs [][]byte
		if err := common.Decode(m[9], &vs); err != nil {
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

	p.Round = round
	p.Block = block
	p.State = state
	p.Transactions = transactions

	return nil
}
func (p Proposal) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"round":        p.Round,
		"block":        p.Block,
		"state":        p.State,
		"transactions": p.Transactions,
	}, nil
}

func (p Proposal) Wellformed() error {
	if p.RawSeal.Type() != ProposalSealType {
		return common.InvalidSealTypeError.AppendMessage("not Proposal")
	}

	if err := p.Block.Wellformed(); err != nil {
		return err
	}

	if err := p.State.Wellformed(); err != nil {
		return err
	}

	for _, th := range p.Transactions {
		if !th.IsValid() {
			return ProposalNotWellformedError.SetMessage(
				"empty Hash found in Proposal.Transactions",
			)
		}
	}

	return nil
}

type ProposalBlock struct {
	Height  common.Big  `json:"height"`
	Current common.Hash `json:"current"`
	Next    common.Hash `json:"next"`
}

func (bb ProposalBlock) MarshalBinary() ([]byte, error) {
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

func (bb *ProposalBlock) UnmarshalBinary(b []byte) error {
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

func (bb ProposalBlock) Wellformed() error {
	if !bb.Current.IsValid() {
		return ProposalNotWellformedError.SetMessage("Proposal.Block.Current is empty")
	}

	if !bb.Next.IsValid() {
		return ProposalNotWellformedError.SetMessage("Proposal.Block.Next is empty")
	}

	return nil
}

type ProposalState struct {
	Current []byte
	Next    []byte
}

func (bb ProposalState) MarshalBinary() ([]byte, error) {
	return common.Encode([]interface{}{
		bb.Current,
		bb.Next,
	})
}

func (bb *ProposalState) UnmarshalBinary(b []byte) error {
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

func (bb ProposalState) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"current": base64.StdEncoding.EncodeToString(bb.Current),
		"next":    base64.StdEncoding.EncodeToString(bb.Next),
	})
}

func (bb *ProposalState) UnmarshalJSON(b []byte) error {
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

func (bb ProposalState) Wellformed() error {
	if len(bb.Current) < 1 {
		return ProposalNotWellformedError.SetMessage("Proposal.State.Current is empty")
	}

	if len(bb.Next) < 1 {
		return ProposalNotWellformedError.SetMessage("Proposal.State.Next is empty")
	}

	return nil
}
