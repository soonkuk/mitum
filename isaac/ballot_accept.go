package isaac

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

type ACCEPTBallot struct {
	common.RawSeal
	stage      VoteStage        `json:"stage"`
	proposer   common.Address   `json:"proposer"`
	validators []common.Address `json:"validators"`
	height     common.Big       `json:"height"`
	round      Round            `json:"round"`
	proposal   common.Hash      `json:"proposal"`
	block      common.Hash      `json:"block"`
}

func NewACCEPTBallot(
	height common.Big,
	round Round,
	proposer common.Address,
	validators []common.Address,
	proposal common.Hash,
	block common.Hash,
) ACCEPTBallot {
	b := ACCEPTBallot{
		stage:      VoteStageACCEPT,
		height:     height,
		round:      round,
		proposer:   proposer,
		validators: validators,
		proposal:   proposal,
		block:      block,
	}

	raw := common.NewRawSeal(b, CurrentBallotVersion)
	b.RawSeal = raw

	return b
}

func (b ACCEPTBallot) Type() common.SealType {
	return ACCEPTBallotSealType
}

func (b ACCEPTBallot) Hint() string {
	return "ab"
}

func (b ACCEPTBallot) Stage() VoteStage {
	return b.stage
}

func (b ACCEPTBallot) Proposer() common.Address {
	return b.proposer
}

func (b ACCEPTBallot) Validators() []common.Address {
	return b.validators
}

func (b ACCEPTBallot) Height() common.Big {
	return b.height
}

func (b ACCEPTBallot) Round() Round {
	return b.round
}

func (b ACCEPTBallot) Proposal() common.Hash {
	return b.proposal
}

func (b ACCEPTBallot) Block() common.Hash {
	return b.block
}

func (b ACCEPTBallot) Vote() Vote {
	return VoteYES
}

func (b ACCEPTBallot) SerializeRLP() ([]interface{}, error) {
	proposal, err := b.proposal.MarshalBinary()
	if err != nil {
		return nil, err
	}

	block, err := b.block.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return []interface{}{
		b.stage,
		b.height,
		b.round,
		b.proposer,
		b.validators,
		proposal,
		block,
	}, nil
}

func (b *ACCEPTBallot) UnserializeRLP(m []rlp.RawValue) error {
	// TODO check length of `m`

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

	var proposal common.Hash
	{
		var vs []byte
		if err := common.Decode(m[11], &vs); err != nil {
			return err
		}
		if err := proposal.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var block common.Hash
	{
		var vs []byte
		if err := common.Decode(m[12], &vs); err != nil {
			return err
		}
		if err := block.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	b.stage = stage
	b.height = height
	b.round = round
	b.proposer = proposer
	b.validators = validators
	b.proposal = proposal
	b.block = block

	return nil
}

func (b ACCEPTBallot) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"stage":      b.stage,
		"proposer":   b.proposer,
		"validators": b.validators,
		"height":     b.height,
		"round":      b.round,
		"proposal":   b.proposal,
		"block":      b.block,
	}, nil
}

func (b ACCEPTBallot) Wellformed() error {
	if b.Type() != ACCEPTBallotSealType {
		return common.InvalidSealTypeError.AppendMessage("type=%v", b.Type())
	}

	if err := b.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if b.stage != VoteStageACCEPT {
		return BallotNotWellformedError.SetMessage("stage is not ACCEPT")
	}

	if len(b.proposer) < 1 {
		return BallotNotWellformedError.SetMessage("proposer is empty")
	}

	// TODO uncomment
	//if len(b.validators) < 1 {
	//	return BallotNotWellformedError.SetMessage("validators is empty")
	//}

	if !b.proposal.IsValid() {
		return BallotNotWellformedError.SetMessage("proposal is invalid")
	}

	if !b.block.IsValid() {
		return BallotNotWellformedError.SetMessage("block is invalid")
	}

	return nil
}
