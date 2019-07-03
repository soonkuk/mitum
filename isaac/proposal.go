package isaac

import (
	"encoding/json"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

var (
	ProposalType     common.DataType = common.NewDataType(3, "proposal")
	ProposalHashHint string          = "pp"
)

func NewProposalHash(b []byte) (hash.Hash, error) {
	return hash.NewArgon2Hash(ProposalHashHint, b)
}

type ProposalBody struct {
	hash         hash.Hash
	Height       Height
	Round        Round
	CurrentBlock hash.Hash
	Proposer     node.Address
	Transactions []hash.Hash
}

func NewProposalBody(
	height Height,
	round Round,
	currentBlock hash.Hash,
	proposer node.Address,
	transactions []hash.Hash,
) (ProposalBody, error) {
	body := ProposalBody{
		Height:       height,
		Round:        round,
		CurrentBlock: currentBlock,
		Proposer:     proposer,
		Transactions: transactions,
	}

	b, err := rlp.EncodeToBytes(body)
	if err != nil {
		return ProposalBody{}, err
	}

	hash, err := hash.NewArgon2Hash(ProposalHashHint, b)
	if err != nil {
		return ProposalBody{}, err
	}
	body.hash = hash

	return body, nil
}

func (pb ProposalBody) Type() common.DataType {
	return ProposalType
}

func (pb ProposalBody) Hash() hash.Hash {
	return pb.hash
}

func (pb ProposalBody) IsValid() error {
	// TODO create func to check proposal hash

	if pb.Height.IsZero() {
		return xerrors.Errorf("zero Height of ProposalBody")
	}

	if pb.CurrentBlock.Empty() {
		return xerrors.Errorf("empty CurrentBlock of ProposalBody")
	}

	if pb.hash.Empty() {
		return xerrors.Errorf("empty hash of ProposalBody")
	}

	if pb.Proposer.Empty() {
		return xerrors.Errorf("empty Proposer of ProposalBody")
	}

	return nil
}

func (pb ProposalBody) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, struct {
		Hash         hash.Hash
		Height       Height
		Round        Round
		CurrentBlock hash.Hash
		Proposer     node.Address
		Transactions []hash.Hash
	}{
		pb.hash,
		pb.Height,
		pb.Round,
		pb.CurrentBlock,
		pb.Proposer,
		pb.Transactions,
	})
}

func (pb *ProposalBody) DecodeRLP(s *rlp.Stream) error {
	var n struct {
		Hash         hash.Hash
		Height       Height
		Round        Round
		CurrentBlock hash.Hash
		Proposer     node.Address
		Transactions []hash.Hash
	}
	if err := s.Decode(&n); err != nil {
		return err
	}

	pb.hash = n.Hash
	pb.Height = n.Height
	pb.Round = n.Round
	pb.CurrentBlock = n.CurrentBlock
	pb.Proposer = n.Proposer
	pb.Transactions = n.Transactions

	return nil
}

func (pb ProposalBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":         pb.Type(),
		"hash":         pb.hash,
		"height":       pb.Height,
		"round":        pb.Round,
		"currentBlock": pb.CurrentBlock,
		"proposer":     pb.Proposer,
		"transactions": pb.Transactions,
	})
}

func (pb ProposalBody) String() string {
	b, _ := json.Marshal(pb)
	return string(b)
}

type Proposal struct {
	seal.BaseSeal
	body ProposalBody
}

func NewProposal(
	height Height,
	round Round,
	currentBlock hash.Hash,
	proposer node.Address,
	transactions []hash.Hash,
) (Proposal, error) {
	body, err := NewProposalBody(height, round, currentBlock, proposer, transactions)
	if err != nil {
		return Proposal{}, err
	}

	return Proposal{
		BaseSeal: seal.NewBaseSeal(body),
		body:     body,
	}, nil
}

func (pp Proposal) MarshalJSON() ([]byte, error) {
	return json.Marshal(pp.BaseSeal)
}

func (pp Proposal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, pp.BaseSeal)
}

func (pp *Proposal) DecodeRLP(s *rlp.Stream) error {
	var raw seal.RLPDecodeSeal
	if err := s.Decode(&raw); err != nil {
		return err
	}

	var body ProposalBody
	if err := rlp.DecodeBytes(raw.Body, &body); err != nil {
		return err
	}
	bsl := &seal.BaseSeal{}
	bsl = bsl.
		SetType(raw.Type).
		SetHash(raw.Hash).
		SetHeader(raw.Header).
		SetBody(body)

	*pp = Proposal{BaseSeal: *bsl, body: body}

	if err := pp.IsValid(); err != nil {
		return err
	}

	return nil
}

func (pp Proposal) Body() seal.Body {
	return pp.body
}

func (pp Proposal) Type() common.DataType {
	return ProposalType
}

func (pp Proposal) Height() Height {
	return pp.body.Height
}

func (pp Proposal) Round() Round {
	return pp.body.Round
}

func (pp Proposal) CurrentBlock() hash.Hash {
	return pp.body.CurrentBlock
}

func (pp Proposal) Proposer() node.Address {
	return pp.body.Proposer
}

func (pp Proposal) Transactions() []hash.Hash {
	return pp.body.Transactions
}
