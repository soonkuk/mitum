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
	ProposeSeal common.Hash    `json:"propose_seal"` // NOTE ProposeSeal.Hash()
	Stage       VoteStage      `json:"stage"`
	Vote        Vote           `json:"vote"`
	VotedAt     common.Time    `json:"voted_at"`

	hash    common.Hash
	encoded []byte
}

// NewBallot creates new Ballot
//  - psHash: Seal(Propose).Hash()
func NewBallot(psHash common.Hash, source common.Address, vote Vote) (Ballot, error) {
	b := Ballot{
		Version:     CurrentBallotVersion,
		Source:      source,
		ProposeSeal: psHash,
		Stage:       VoteStageSIGN,
		Vote:        vote,
		VotedAt:     common.Now(),
	}

	if err := b.Wellformed(); err != nil {
		return Ballot{}, err
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
	if v.hash.Empty() {
		return v.makeHash()
	}

	return v.hash, v.encoded, nil
}

func (v Ballot) MarshalBinary() ([]byte, error) {
	psHash, err := v.ProposeSeal.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return common.Encode([]interface{}{
		v.Version,
		v.Source,
		psHash,
		v.Stage,
		v.Vote,
		v.VotedAt,
	})
}

func (v *Ballot) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := common.Decode(b, &m); err != nil {
		return err
	}

	var version common.Version
	if err := common.Decode(m[0], &version); err != nil {
		return err
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
		} else if err := psHash.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var stage VoteStage
	if err := common.Decode(m[3], &stage); err != nil {
		return err
	}

	var vote Vote
	if err := common.Decode(m[4], &vote); err != nil {
		return err
	}

	var votedAt common.Time
	if err := common.Decode(m[5], &votedAt); err != nil {
		return err
	}

	v.Version = version
	v.Source = source
	v.ProposeSeal = psHash
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

	if v.ProposeSeal.Empty() {
		return BallotNotWellformedError.SetMessage("ProposeSeal is empty")
	}

	if !v.Stage.IsValid() {
		return BallotNotWellformedError.SetMessage("Stage is invalid")
	}

	if !v.Stage.CanVote() {
		return BallotNotWellformedError.SetMessage("Stage is not for vote")
	}

	if !v.Vote.IsValid() {
		return BallotNotWellformedError.SetMessage("Vote is invalid")
	}

	if !v.Vote.CanVote() {
		return BallotNotWellformedError.SetMessage("Vote is not for vote")
	}

	return nil
}

func (v Ballot) NewForStage(stage VoteStage, source common.Address, vote Vote) (Ballot, error) {
	if err := v.Wellformed(); err != nil {
		return Ballot{}, err
	}

	if !stage.IsValid() || !stage.CanVote() {
		return Ballot{}, InvalidVoteStageError
	}

	newBallot := Ballot{
		Version:     CurrentBallotVersion,
		Source:      source,
		ProposeSeal: v.ProposeSeal,
		Stage:       stage,
		Vote:        vote,
		VotedAt:     common.Now(),
	}

	if err := newBallot.Wellformed(); err != nil {
		return Ballot{}, err
	}

	return newBallot, nil
}
