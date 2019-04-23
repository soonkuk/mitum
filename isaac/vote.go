package isaac

import "encoding/json"

type Vote uint

const (
	VoteNONE Vote = iota
	VoteYES
	VoteNOP
	VoteEXPIRE
)

type VoteResult uint

const (
	VoteResultNotYet VoteResult = iota
	VoteResultYES
	VoteResultNOP
	VoteResultEXPIRE
	VoteResultDRAW
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
