package isaac

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

type INITBallot struct {
	common.RawSeal
	stage      VoteStage        `json:"stage"`
	proposer   common.Address   `json:"proposer"`
	validators []common.Address `json:"validators"`
	height     common.Big       `json:"height"`
	round      Round            `json:"round"`
}

func NewINITBallot(
	height common.Big,
	round Round,
	proposer common.Address,
	validators []common.Address,
) INITBallot {
	b := INITBallot{
		stage:      VoteStageINIT,
		height:     height,
		round:      round,
		proposer:   proposer,
		validators: validators,
	}

	raw := common.NewRawSeal(b, CurrentBallotVersion)
	b.RawSeal = raw

	return b
}

func (b INITBallot) Type() common.SealType {
	return INITBallotSealType
}

func (b INITBallot) Hint() string {
	return "ib"
}

func (b INITBallot) Stage() VoteStage {
	return b.stage
}

func (b INITBallot) Proposer() common.Address {
	return b.proposer
}

func (b INITBallot) Validators() []common.Address {
	return b.validators
}

func (b INITBallot) Height() common.Big {
	return b.height
}

func (b INITBallot) Round() Round {
	return b.round
}

func (b INITBallot) Vote() Vote {
	return VoteYES
}

func (b INITBallot) Proposal() common.Hash {
	return common.Hash{}
}

func (b INITBallot) Block() common.Hash {
	return common.Hash{}
}

func (b INITBallot) SerializeRLP() ([]interface{}, error) {
	return []interface{}{
		b.stage,
		b.height,
		b.round,
		b.proposer,
		b.validators,
	}, nil
}

func (b *INITBallot) UnserializeRLP(m []rlp.RawValue) error {
	if len(m) != 11 {
		return BallotNotWellformedError.SetMessage("insufficient rlp.RawValue for INITBallot")
	}

	var stage VoteStage
	if err := common.Decode(m[6], &stage); err != nil {
		return err
	}

	var height common.Big
	if err := common.Decode(m[7], &height); err != nil {
		return err
	}

	var round Round
	if err := common.Decode(m[8], &round); err != nil {
		return err
	}

	var proposer common.Address
	if err := common.Decode(m[9], &proposer); err != nil {
		return err
	}

	var validators []common.Address
	if err := common.Decode(m[10], &validators); err != nil {
		return err
	}

	b.stage = stage
	b.height = height
	b.round = round
	b.proposer = proposer
	b.validators = validators

	return nil
}

func (b INITBallot) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"stage":      b.stage,
		"proposer":   b.proposer,
		"validators": b.validators,
		"height":     b.height,
		"round":      b.round,
	}, nil
}

func (b INITBallot) Wellformed() error {
	if b.Type() != INITBallotSealType {
		return common.InvalidSealTypeError.AppendMessage("type=%v", b.Type())
	}

	if err := b.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if b.stage != VoteStageINIT {
		return BallotNotWellformedError.SetMessage("stage is not INIT")
	}

	if len(b.proposer) < 1 {
		return BallotNotWellformedError.SetMessage("proposer is empty")
	}

	// TODO uncomment
	//if len(b.validators) < 1 {
	//	return BallotNotWellformedError.SetMessage("validators is empty")
	//}

	return nil
}
