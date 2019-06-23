package isaac

import (
	"encoding/json"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

var (
	INITBallotType     common.DataType = common.NewDataType(uint(StageINIT), "init_ballot")
	INITBallotHashHint string          = "ib"
)

type INITBallot struct {
	seal.BaseSeal
	body INITBallotBody
}

func NewINITBallot(body INITBallotBody) (INITBallot, error) {
	b, err := rlp.EncodeToBytes(body)
	if err != nil {
		return INITBallot{}, err
	}

	hash, err := hash.NewArgon2Hash(INITBallotHashHint, b)
	if err != nil {
		return INITBallot{}, err
	}
	body.hash = hash

	return INITBallot{
		BaseSeal: seal.NewBaseSeal(body),
		body:     body,
	}, nil
}

func (ib INITBallot) MarshalJSON() ([]byte, error) {
	return json.Marshal(ib.BaseSeal)
}

func (ib INITBallot) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, ib.BaseSeal)
}

func (ib INITBallot) Type() common.DataType {
	return INITBallotType
}

func (ib INITBallot) Node() node.Address {
	return ib.body.Node
}

func (ib INITBallot) Height() Height {
	return ib.body.Height
}

func (ib INITBallot) Round() Round {
	return ib.body.Round
}

func (ib INITBallot) Stage() Stage {
	return ib.body.Stage
}

func (ib INITBallot) Proposal() hash.Hash {
	return ib.body.Proposal
}

func (ib INITBallot) NextBlock() hash.Hash {
	return ib.body.NextBlock
}

type INITBallotBody struct {
	hash      hash.Hash
	Node      node.Address
	Height    Height
	Round     Round
	Stage     Stage
	Proposal  hash.Hash
	NextBlock hash.Hash
}

func (ibb INITBallotBody) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, struct {
		Hash      hash.Hash
		Node      node.Address
		Height    Height
		Round     Round
		Stage     Stage
		Proposal  hash.Hash
		NextBlock hash.Hash
	}{
		Hash:      ibb.hash,
		Node:      ibb.Node,
		Height:    ibb.Height,
		Round:     ibb.Round,
		Stage:     ibb.Stage,
		Proposal:  ibb.Proposal,
		NextBlock: ibb.NextBlock,
	})
}

func (ibb *INITBallotBody) DecodeRLP(s *rlp.Stream) error {
	var n struct {
		Hash      hash.Hash
		Node      node.Address
		Height    Height
		Round     Round
		Stage     Stage
		Proposal  hash.Hash
		NextBlock hash.Hash
	}

	if err := s.Decode(&n); err != nil {
		return err
	}

	ibb.hash = n.Hash
	ibb.Node = n.Node
	ibb.Height = n.Height
	ibb.Round = n.Round
	ibb.Stage = n.Stage
	ibb.Proposal = n.Proposal
	ibb.NextBlock = n.NextBlock

	return nil
}

func (ibb INITBallotBody) Type() common.DataType {
	return INITBallotType
}

func (ibb INITBallotBody) Hash() hash.Hash {
	return ibb.hash
}

func (ibb INITBallotBody) IsValid() error {
	return nil
}

func (ibb INITBallotBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"node":      ibb.Node,
		"height":    ibb.Height,
		"round":     ibb.Round,
		"stage":     ibb.Stage,
		"proposal":  ibb.Proposal,
		"nextBlock": ibb.NextBlock,
	})
}
