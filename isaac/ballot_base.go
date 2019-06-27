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

type BaseBallot struct {
	seal.BaseSeal
	body BaseBallotBody
}

func NewBaseBallot(body BaseBallotBody) (BaseBallot, error) {
	b, err := rlp.EncodeToBytes(body)
	if err != nil {
		return BaseBallot{}, err
	}

	hash, err := hash.NewArgon2Hash(BallotHashHint, b)
	if err != nil {
		return BaseBallot{}, err
	}
	body.hash = hash

	return BaseBallot{
		BaseSeal: seal.NewBaseSeal(body),
		body:     body,
	}, nil
}

func NewRawBaseBallot(baseSeal seal.BaseSeal, body BaseBallotBody) (BaseBallot, error) {
	return BaseBallot{BaseSeal: baseSeal, body: body}, nil
}

func (ib BaseBallot) MarshalJSON() ([]byte, error) {
	return json.Marshal(ib.BaseSeal)
}

func (ib BaseBallot) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, ib.BaseSeal)
}

func (ib *BaseBallot) DecodeRLP(s *rlp.Stream) error {
	var raw seal.RLPDecodeSeal
	if err := s.Decode(&raw); err != nil {
		return err
	}

	var body BaseBallotBody
	if err := rlp.DecodeBytes(raw.Body, &body); err != nil {
		return err
	}
	bsl := &seal.BaseSeal{}
	bsl = bsl.
		SetType(raw.Type).
		SetHash(raw.Hash).
		SetHeader(raw.Header).
		SetBody(body)

	*ib = BaseBallot{BaseSeal: *bsl, body: body}

	if err := ib.IsValid(); err != nil {
		return err
	}

	return nil
}

func (ib BaseBallot) Body() BaseBallotBody {
	return ib.body
}

func (ib BaseBallot) Type() common.DataType {
	return BallotType
}

func (ib BaseBallot) Node() node.Address {
	return ib.body.Node
}

func (ib BaseBallot) Height() Height {
	return ib.body.Height
}

func (ib BaseBallot) Round() Round {
	return ib.body.Round
}

func (ib BaseBallot) Stage() Stage {
	return ib.body.Stage
}

func (ib BaseBallot) Proposal() hash.Hash {
	return ib.body.Proposal
}

func (ib BaseBallot) NextBlock() hash.Hash {
	return ib.body.NextBlock
}

type BaseBallotBody struct {
	hash         hash.Hash
	Node         node.Address
	Height       Height
	Round        Round
	Stage        Stage
	Proposal     hash.Hash
	CurrentBlock hash.Hash
	NextBlock    hash.Hash
}

func (ibb BaseBallotBody) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, struct {
		Hash         hash.Hash
		Node         node.Address
		Height       Height
		Round        Round
		Stage        Stage
		Proposal     hash.Hash
		CurrentBlock hash.Hash
		NextBlock    hash.Hash
	}{
		Hash:         ibb.hash,
		Node:         ibb.Node,
		Height:       ibb.Height,
		Round:        ibb.Round,
		Stage:        ibb.Stage,
		Proposal:     ibb.Proposal,
		CurrentBlock: ibb.CurrentBlock,
		NextBlock:    ibb.NextBlock,
	})
}

func (ibb *BaseBallotBody) DecodeRLP(s *rlp.Stream) error {
	var n struct {
		Hash         hash.Hash
		Node         node.Address
		Height       Height
		Round        Round
		Stage        Stage
		Proposal     hash.Hash
		CurrentBlock hash.Hash
		NextBlock    hash.Hash
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
	ibb.CurrentBlock = n.CurrentBlock
	ibb.NextBlock = n.NextBlock

	return nil
}

func (ibb BaseBallotBody) Type() common.DataType {
	return BallotType
}

func (ibb BaseBallotBody) Hash() hash.Hash {
	return ibb.hash
}

func (ibb BaseBallotBody) IsValid() error {
	if ibb.Type() != BallotType {
		return InvalidBallotError.Newf("type=%q", ibb.Type())
	}

	if err := ibb.Node.IsValid(); err != nil {
		return InvalidBallotError.Newf("invalid node; node=%q", ibb.Node)
	}

	switch ibb.Stage {
	case StageINIT:
		return IsValidINITBallot(ibb)
	case StageSIGN:
		return IsValidSIGNBallot(ibb)
	case StageACCEPT:
		return IsValidACCEPTBallot(ibb)
	default:
		return InvalidStageError.Newf("stage=%q", ibb.Stage)
	}
}

func (ibb BaseBallotBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"node":         ibb.Node,
		"height":       ibb.Height,
		"round":        ibb.Round,
		"stage":        ibb.Stage,
		"proposal":     ibb.Proposal,
		"currentBlock": ibb.CurrentBlock,
		"nextBlock":    ibb.NextBlock,
	})
}
