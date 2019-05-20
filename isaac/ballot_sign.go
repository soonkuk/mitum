package isaac

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

type SIGNBallot struct {
	common.RawSeal
	stage      VoteStage
	proposer   common.Address
	validators []common.Address
	height     common.Big
	round      Round
	proposal   common.Hash
	block      common.Hash
	vote       Vote
}

func NewSIGNBallot(
	height common.Big,
	round Round,
	proposer common.Address,
	validators []common.Address,
	proposal common.Hash,
	block common.Hash,
	vote Vote,
) SIGNBallot {
	b := SIGNBallot{
		stage:      VoteStageSIGN,
		height:     height,
		round:      round,
		proposer:   proposer,
		validators: validators,
		proposal:   proposal,
		block:      block,
		vote:       vote,
	}

	raw := common.NewRawSeal(b, CurrentBallotVersion)
	b.RawSeal = raw

	return b
}

func (b SIGNBallot) Type() common.SealType {
	return SIGNBallotSealType
}

func (b SIGNBallot) Hint() string {
	return "sb"
}

func (b SIGNBallot) Stage() VoteStage {
	return b.stage
}

func (b SIGNBallot) Proposer() common.Address {
	return b.proposer
}

func (b SIGNBallot) Validators() []common.Address {
	return b.validators
}

func (b SIGNBallot) Height() common.Big {
	return b.height
}

func (b SIGNBallot) Round() Round {
	return b.round
}

func (b SIGNBallot) Proposal() common.Hash {
	return b.proposal
}

func (b SIGNBallot) Block() common.Hash {
	return b.block
}

func (b SIGNBallot) Vote() Vote {
	return b.vote
}

func (b SIGNBallot) SerializeRLP() ([]interface{}, error) {
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
		b.vote,
	}, nil
}

func (b *SIGNBallot) UnserializeRLP(m []rlp.RawValue) error {
	if len(m) != 14 {
		return BallotNotWellformedError.SetMessage("insufficient rlp.RawValue for SIGNBallot")
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

	var vote Vote
	if err := common.Decode(m[13], &vote); err != nil {
		return err
	}

	b.stage = stage
	b.height = height
	b.round = round
	b.proposer = proposer
	b.validators = validators
	b.proposal = proposal
	b.block = block
	b.vote = vote

	return nil
}

func (b SIGNBallot) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"stage":      b.stage,
		"proposer":   b.proposer,
		"validators": b.validators,
		"height":     b.height,
		"round":      b.round,
		"proposal":   b.proposal,
		"block":      b.block,
		"vote":       b.vote,
	}, nil
}

func (b SIGNBallot) Wellformed() error {
	if b.Type() != SIGNBallotSealType {
		return common.InvalidSealTypeError.AppendMessage("type=%v", b.Type())
	}

	if err := b.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if b.stage != VoteStageSIGN {
		return BallotNotWellformedError.SetMessage("stage is not SIGN")
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

	if !b.vote.CanVote() {
		return BallotNotWellformedError.SetMessage("vote is invalid; %v", b.vote)
	}

	return nil
}
