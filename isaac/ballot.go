package isaac

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

var (
	CurrentBallotVersion common.Version = common.MustParseVersion("0.1.0-proto")
)

type Ballot struct {
	common.RawSeal
	Proposal common.Hash    `json:"proposal"` // NOTE Proposal.Hash()
	Proposer common.Address `json:"proposer"` // NOTE only for `INIT`
	// TODO
	// Validators   []common.Validator `json:validators`
	Height common.Big `json:"height"`
	Round  Round      `json:"round"`
	Stage  VoteStage  `json:"stage"`
	Vote   Vote       `json:"vote"`
}

func NewBallot(
	proposal common.Hash,
	proposer common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
) Ballot {
	b := Ballot{
		Proposal: proposal,
		Proposer: proposer,
		Height:   height,
		Round:    round,
		Stage:    stage,
		Vote:     vote,
	}

	raw := common.NewRawSeal(b, CurrentBallotVersion)
	b.RawSeal = raw

	return b
}

func (b Ballot) Type() common.SealType {
	return common.SealType("ballot")
}

func (b Ballot) Hint() string {
	return "bt"
}

func (b Ballot) SerializeRLP() ([]interface{}, error) {
	var proposal []byte
	if b.Proposal.IsValid() {
		h, err := b.Proposal.MarshalBinary()
		if err != nil {
			return nil, err
		}
		proposal = h
	}

	return []interface{}{
		proposal,
		b.Proposer,
		b.Height,
		b.Round,
		b.Stage,
		b.Vote,
	}, nil
}

func (b *Ballot) UnserializeRLP(m []rlp.RawValue) error {
	var proposal common.Hash
	{
		var vs []byte
		if err := common.Decode(m[6], &vs); err != nil {
			return err
		} else if len(vs) < 1 {
			//
		} else if err := proposal.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var proposer common.Address
	if err := common.Decode(m[7], &proposer); err != nil {
		return err
	}

	var height common.Big
	if err := common.Decode(m[8], &height); err != nil {
		return err
	}

	var round Round
	if err := common.Decode(m[9], &round); err != nil {
		return err
	}

	var stage VoteStage
	if err := common.Decode(m[10], &stage); err != nil {
		return err
	}

	var vote Vote
	if err := common.Decode(m[11], &vote); err != nil {
		return err
	}

	b.Proposal = proposal
	b.Proposer = proposer
	b.Height = height
	b.Round = round
	b.Stage = stage
	b.Vote = vote

	return nil
}

func (b Ballot) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"proposal": b.Proposal,
		"proposer": b.Proposer,
		"height":   b.Height,
		"round":    b.Round,
		"stage":    b.Stage,
		"vote":     b.Vote,
	}, nil
}

func (b Ballot) Wellformed() error {
	if b.RawSeal.Type() != BallotSealType {
		return common.InvalidSealTypeError.AppendMessage("not ballot")
	}

	if err := b.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if !b.Stage.IsValid() {
		return BallotNotWellformedError.SetMessage("Stage is invalid")
	}

	if !b.Stage.CanVote() {
		return BallotNotWellformedError.SetMessage("Stage is not for vote")
	}

	if !b.Vote.IsValid() {
		return BallotNotWellformedError.SetMessage("Vote is invalid")
	}

	if !b.Vote.CanVote() {
		return BallotNotWellformedError.SetMessage("Vote is not for vote")
	}

	if b.Stage == VoteStageINIT {
		if len(b.Proposer) < 1 {
			return BallotNotWellformedError.SetMessage("Proposer is empty; INIT")
		}

		if b.Proposal.IsValid() {
			return BallotNotWellformedError.SetMessage("Proposal is not empty")
		}
	} else {
		if len(b.Proposer) > 0 {
			return BallotNotWellformedError.SetMessage("Proposer is not empty; not INIT")
		}

		if !b.Proposal.IsValid() {
			return BallotNotWellformedError.SetMessage("Proposal is empty")
		}
	}

	if b.Vote != VoteYES && b.Stage != VoteStageSIGN {
		return BallotNotWellformedError.SetMessage("except sign stage, vote should be yes")
	}

	return nil
}
