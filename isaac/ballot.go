package isaac

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/spikeekips/mitum/common"
)

var (
	CurrentBallotVersion common.Version = common.MustParseVersion("0.1.0-proto")
)

type BallotBlock struct {
	Height  common.Big  `json:"height"`
	Current common.Hash `json:"current"`
	Next    common.Hash `json:"next"`
}

func (bb BallotBlock) MarshalBinary() ([]byte, error) {
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

func (bb *BallotBlock) UnmarshalBinary(b []byte) error {
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

func (bb BallotBlock) Wellformed() error {
	if bb.Current.Empty() {
		return ProposeBallotNotWellformedError.SetMessage("ProposeBallot.Block.Current is empty")
	}

	if bb.Next.Empty() {
		return ProposeBallotNotWellformedError.SetMessage("ProposeBallot.Block.Next is empty")
	}

	return nil
}

type BallotBlockState struct {
	Current []byte
	Next    []byte
}

func (bb BallotBlockState) MarshalBinary() ([]byte, error) {
	return common.Encode([]interface{}{
		bb.Current,
		bb.Next,
	})
}

func (bb *BallotBlockState) UnmarshalBinary(b []byte) error {
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

func (bb BallotBlockState) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"current": base64.StdEncoding.EncodeToString(bb.Current),
		"next":    base64.StdEncoding.EncodeToString(bb.Next),
	})
}

func (bb *BallotBlockState) UnmarshalJSON(b []byte) error {
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

func (bb BallotBlockState) Wellformed() error {
	if len(bb.Current) < 1 {
		return ProposeBallotNotWellformedError.SetMessage("ProposeBallot.State.Current is empty")
	}

	if len(bb.Next) < 1 {
		return ProposeBallotNotWellformedError.SetMessage("ProposeBallot.State.Next is empty")
	}

	return nil
}

type ProposeBallot struct {
	Version      common.Version   `json:"version"`
	Proposer     common.Address   `json:"proposer"`
	Round        Round            `json:"round"`
	Block        BallotBlock      `json:"block"`
	State        BallotBlockState `json:"state"`
	Transactions []common.Hash    `json:"transactions"` // NOTE check Hash.p is 'tx'
	ProposedAt   common.Time      `json:"proposed_at"`

	hash    common.Hash
	encoded []byte
}

func (p ProposeBallot) makeHash() (common.Hash, []byte, error) {
	encoded, err := p.MarshalBinary()
	if err != nil {
		return common.Hash{}, nil, err
	}

	hash, err := common.NewHash("pb", encoded)
	if err != nil {
		return common.Hash{}, nil, err
	}

	return hash, encoded, nil
}

func (p ProposeBallot) Hash() (common.Hash, []byte, error) {
	if p.hash.Empty() {
		return p.makeHash()
	}

	return p.hash, p.encoded, nil
}

func (p ProposeBallot) MarshalBinary() ([]byte, error) {
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
		hash, err := t.MarshalBinary()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, hash)
	}

	version, err := p.Version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return common.Encode([]interface{}{
		version,
		p.Proposer,
		p.Round,
		block,
		state,
		transactions,
		p.ProposedAt,
	})
}

func (p *ProposeBallot) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := common.Decode(b, &m); err != nil {
		return err
	}

	var version common.Version
	{
		var vs []byte
		if err := common.Decode(m[0], &vs); err != nil {
			return err
		} else if err := version.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var proposer common.Address
	if err := common.Decode(m[1], &proposer); err != nil {
		return err
	}

	var round Round
	if err := common.Decode(m[2], &round); err != nil {
		return err
	}

	var block BallotBlock
	{
		var vs []byte
		if err := common.Decode(m[3], &vs); err != nil {
			return err
		} else if err := block.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var state BallotBlockState
	{
		var vs []byte
		if err := common.Decode(m[4], &vs); err != nil {
			return err
		} else if err := state.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var transactions []common.Hash
	{
		var vs [][]byte
		if err := common.Decode(m[5], &vs); err != nil {
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

	var proposedAt common.Time
	if err := common.Decode(m[6], &proposedAt); err != nil {
		return err
	}

	p.Version = version
	p.Proposer = proposer
	p.Round = round
	p.Block = block
	p.State = state
	p.Transactions = transactions
	p.ProposedAt = proposedAt

	hash, encoded, err := p.makeHash()
	if err != nil {
		return err
	}

	p.hash = hash
	p.encoded = encoded

	return nil
}

func (p ProposeBallot) String() string {
	b, _ := json.Marshal(p)
	return strings.Replace(string(b), "\"", "'", -1)
}

func (p ProposeBallot) Wellformed() error {
	if _, err := p.Proposer.IsValid(); err != nil {
		return err
	}

	if err := p.Block.Wellformed(); err != nil {
		return err
	}

	if err := p.State.Wellformed(); err != nil {
		return err
	}

	for _, th := range p.Transactions {
		if th.Empty() {
			return ProposeBallotNotWellformedError.SetMessage(
				"empty Hash found in ProposeBallot.Transactions",
			)
		}
	}

	return nil
}

type VoteBallot struct {
	Version           common.Version `json:"version"`
	Source            common.Address `json:"source"`
	ProposeBallotSeal common.Hash    `json:"propose_ballot_seal"` // NOTE Seal.Hash() of ProposeBallot
	Stage             VoteStage      `json:"stage"`
	Vote              Vote           `json:"vote"`
	VotedAt           common.Time    `json:"voted_at"`

	hash    common.Hash
	encoded []byte
}

func NewVoteBallot(sealHash common.Hash, source common.Address, vote Vote) (VoteBallot, error) {
	b := VoteBallot{
		Version:           CurrentBallotVersion,
		Source:            source,
		ProposeBallotSeal: sealHash,
		Stage:             VoteStageSIGN,
		Vote:              vote,
		VotedAt:           common.Now(),
	}

	if err := b.Wellformed(); err != nil {
		return VoteBallot{}, err
	}

	return b, nil
}

func (v VoteBallot) makeHash() (common.Hash, []byte, error) {
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

func (v VoteBallot) Hash() (common.Hash, []byte, error) {
	if v.hash.Empty() {
		return v.makeHash()
	}

	return v.hash, v.encoded, nil
}

func (v VoteBallot) MarshalBinary() ([]byte, error) {
	proposeBallotSeal, err := v.ProposeBallotSeal.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return common.Encode([]interface{}{
		v.Version,
		v.Source,
		proposeBallotSeal,
		v.Stage,
		v.Vote,
		v.VotedAt,
	})
}

func (v *VoteBallot) UnmarshalBinary(b []byte) error {
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

	var proposeBallotSeal common.Hash
	{
		var vs []byte
		if err := common.Decode(m[2], &vs); err != nil {
			return err
		} else if err := proposeBallotSeal.UnmarshalBinary(vs); err != nil {
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
	v.ProposeBallotSeal = proposeBallotSeal
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

func (v VoteBallot) String() string {
	b, _ := json.Marshal(v)
	return strings.Replace(string(b), "\"", "'", -1)
}

func (v VoteBallot) Wellformed() error {
	if _, err := v.Source.IsValid(); err != nil {
		return err
	}

	if v.ProposeBallotSeal.Empty() {
		return VoteBallotNotWellformedError.SetMessage("VoteBallot.ProposeBallotSeal is empty")
	}

	if !v.Stage.IsValid() {
		return VoteBallotNotWellformedError.SetMessage("VoteBallot.Stage is invalid")
	}

	if !v.Stage.CanVote() {
		return VoteBallotNotWellformedError.SetMessage("VoteBallot.Stage is not for vote")
	}

	if !v.Vote.IsValid() {
		return VoteBallotNotWellformedError.SetMessage("VoteBallot.Vote is invalid")
	}

	if !v.Vote.CanVote() {
		return VoteBallotNotWellformedError.SetMessage("VoteBallot.Vote is not for vote")
	}

	return nil
}

func (v VoteBallot) NewForStage(stage VoteStage, source common.Address, vote Vote) (VoteBallot, error) {
	if err := v.Wellformed(); err != nil {
		return VoteBallot{}, err
	}

	if !stage.IsValid() || !stage.CanVote() {
		return VoteBallot{}, InvalidVoteStageError
	}

	newBallot := VoteBallot{
		Version:           CurrentBallotVersion,
		Source:            source,
		ProposeBallotSeal: v.ProposeBallotSeal,
		Stage:             stage,
		Vote:              vote,
		VotedAt:           common.Now(),
	}

	if err := newBallot.Wellformed(); err != nil {
		return VoteBallot{}, err
	}

	return newBallot, nil
}
