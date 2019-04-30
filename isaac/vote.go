package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/common"
)

type VoteResult uint

const (
	VoteResultNotYet VoteResult = iota
	VoteResultYES
	VoteResultNOP
	VoteResultEXPIRE
	VoteResultDRAW
)

func (v VoteResult) String() string {
	switch v {
	case VoteResultNotYet:
		return "not-yet"
	case VoteResultNOP:
		return "nop"
	case VoteResultYES:
		return "yes"
	case VoteResultEXPIRE:
		return "exp"
	case VoteResultDRAW:
		return "draw"
	default:
		return ""
	}
}

type VoteResultInfo struct {
	Result      VoteResult
	Proposal    common.Hash
	Height      common.Big
	Round       Round
	Stage       VoteStage
	LastVotedAt common.Time
}

func NewVoteResultInfo() VoteResultInfo {
	return VoteResultInfo{}
}

func (v VoteResultInfo) NotYet() bool {
	return v.Result == VoteResultNotYet
}

func (v VoteResultInfo) Draw() bool {
	return v.Result == VoteResultDRAW
}

func (v VoteResultInfo) Vote() Vote {
	if v.Stage == VoteStageSIGN {
		switch v.Result {
		case VoteResultYES:
			return VoteYES
		}
		return VoteNOP
	}

	return VoteYES
}

type Majoritier interface {
	CanCount(uint, uint) bool
	Majority(uint, uint) VoteResultInfo
}

type Vote uint

const (
	VoteNONE Vote = iota
	VoteYES
	VoteNOP
	VoteEXPIRE
)

func (v Vote) String() string {
	switch v {
	case VoteNOP:
		return "nop"
	case VoteYES:
		return "yes"
	case VoteEXPIRE:
		return "exp"
	default:
		return ""
	}
}

func (v Vote) MarshalJSON() ([]byte, error) {
	if v == VoteNONE {
		return nil, InvalidVoteError
	}

	return json.Marshal(v.String())
}

func (v *Vote) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "nop":
		*v = VoteNOP
	case "yes":
		*v = VoteYES
	case "exp":
		*v = VoteEXPIRE
	default:
		return InvalidVoteError
	}

	return nil
}

func (s Vote) IsValid() bool {
	switch s {
	case VoteNONE:
	case VoteYES:
	case VoteNOP:
	case VoteEXPIRE:
	default:
		return false
	}

	return true
}

func (s Vote) CanVote() bool {
	switch s {
	case VoteNONE:
	default:
		return true
	}

	return false
}
