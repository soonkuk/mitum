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
	VoteResultDRAW
)

func (v VoteResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v VoteResult) String() string {
	switch v {
	case VoteResultNotYet:
		return "not-yet"
	case VoteResultNOP:
		return "nop"
	case VoteResultYES:
		return "yes"
	case VoteResultDRAW:
		return "draw"
	default:
		return ""
	}
}

type VoteResultInfo struct {
	Result      VoteResult  `json:"result"`
	Proposal    common.Hash `json:"proposal"`
	Height      common.Big  `json:"height"`
	Round       Round       `json:"round"`
	Stage       VoteStage   `json:"stage"`
	Proposed    bool        `json:"proposed"`
	LastVotedAt common.Time `json:"last_voted_at"`
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

func (v VoteResultInfo) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
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
)

func (v Vote) String() string {
	switch v {
	case VoteNOP:
		return "nop"
	case VoteYES:
		return "yes"
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
