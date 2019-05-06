package isaac

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/spikeekips/mitum/common"
)

var (
	CurrentBallotVersion common.Version = common.MustParseVersion("0.1.0-proto")
)

type Ballot struct {
	Version     common.Version `json:"version"`
	Source      common.Address `json:"source"`
	ProposeSeal common.Hash    `json:"propose_seal"` // NOTE ProposeSeal.Hash() // TODO rename Proposal
	Proposer    common.Address `json:"proposer"`     // NOTE only for `INIT`
	Height      common.Big     `json:"height"`
	Round       Round          `json:"round"`
	Stage       VoteStage      `json:"stage"`
	Vote        Vote           `json:"vote"`
	VotedAt     common.Time    `json:"voted_at"`

	hash    common.Hash
	encoded []byte
}

func NewBallot(
	psHash common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
) (Ballot, error) {
	b := Ballot{
		Version:     CurrentBallotVersion,
		Source:      source,
		ProposeSeal: psHash,
		Height:      height,
		Round:       round,
		Stage:       stage,
		Vote:        vote,
		VotedAt:     common.Now(),
	}

	return b, nil
}

func (v Ballot) makeHash() (common.Hash, []byte, error) {
	encoded, err := v.MarshalBinary()
	if err != nil {
		return common.Hash{}, nil, err
	}

	hash, err := common.NewHash("vb", encoded)
	if err != nil {
		return common.Hash{}, nil, err
	}

	return hash, encoded, nil
}

func (v Ballot) Hash() (common.Hash, []byte, error) {
	if !v.hash.IsValid() {
		return v.makeHash()
	}

	return v.hash, v.encoded, nil
}

func (v Ballot) MarshalBinary() ([]byte, error) {
	version, err := v.Version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var psHash []byte
	if v.ProposeSeal.IsValid() {
		h, err := v.ProposeSeal.MarshalBinary()
		if err != nil {
			return nil, err
		}
		psHash = h
	}

	votedAt, err := v.VotedAt.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return common.Encode([]interface{}{
		version,
		v.Source,
		psHash,
		v.Proposer,
		v.Height,
		v.Round,
		v.Stage,
		v.Vote,
		votedAt,
	})
}

func (v *Ballot) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := common.Decode(b, &m); err != nil {
		return err
	}

	var version common.Version
	{
		var vs []byte
		if err := common.Decode(m[0], &vs); err != nil {
			return err
		}
		if err := version.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var source common.Address
	if err := common.Decode(m[1], &source); err != nil {
		return err
	}

	var psHash common.Hash
	{
		var vs []byte
		if err := common.Decode(m[2], &vs); err != nil {
			return err
		} else if len(vs) < 1 {
			//
		} else if err := psHash.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var proposer common.Address
	if err := common.Decode(m[3], &proposer); err != nil {
		return err
	}

	var height common.Big
	if err := common.Decode(m[4], &height); err != nil {
		return err
	}

	var round Round
	if err := common.Decode(m[5], &round); err != nil {
		return err
	}

	var stage VoteStage
	if err := common.Decode(m[6], &stage); err != nil {
		return err
	}

	var vote Vote
	if err := common.Decode(m[7], &vote); err != nil {
		return err
	}

	var votedAt common.Time
	{
		var vs []byte
		if err := common.Decode(m[8], &vs); err != nil {
			return err
		}
		if err := votedAt.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	v.Version = version
	v.Source = source
	v.ProposeSeal = psHash
	v.Proposer = proposer
	v.Height = height
	v.Round = round
	v.Stage = stage
	v.Vote = vote
	v.VotedAt = votedAt

	hash, encoded, err := v.makeHash()
	if err != nil {
		return err
	}

	v.hash = hash
	v.encoded = encoded

	return nil
}

func (v Ballot) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}

func (v Ballot) Wellformed() error {
	if _, err := v.Source.IsValid(); err != nil {
		return err
	}

	if !v.Stage.CanVote() {
		return BallotNotWellformedError.SetMessage("Stage is not for vote")
	}

	if v.Stage == VoteStageINIT {
		if len(v.Proposer) < 1 {
			return BallotNotWellformedError.SetMessage("Proposer is empty for INIT")
		}

		if v.ProposeSeal.IsValid() {
			return BallotNotWellformedError.SetMessage("ProposeSeal is not empty")
		}
	} else {
		if len(v.Proposer) > 0 {
			return BallotNotWellformedError.SetMessage("Proposer is not empty for not INIT")
		}

		if !v.ProposeSeal.IsValid() {
			return BallotNotWellformedError.SetMessage("ProposeSeal is empty")
		}
	}

	if !v.Stage.IsValid() {
		return BallotNotWellformedError.SetMessage("Stage is invalid")
	}

	if !v.Vote.IsValid() {
		return BallotNotWellformedError.SetMessage("Vote is invalid")
	}

	if !v.Vote.CanVote() {
		return BallotNotWellformedError.SetMessage("Vote is not for vote")
	}

	if v.Vote != VoteYES && v.Stage != VoteStageSIGN {
		return BallotNotWellformedError.SetMessage("except sign stage, vote should be yes")
	}

	return nil
}
