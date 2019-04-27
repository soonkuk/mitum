package isaac

import (
	"encoding/json"
)

type VoteStage uint

const (
	VoteStageNONE VoteStage = iota
	VoteStageINIT
	VoteStageSIGN
	VoteStageACCEPT
	VoteStageALLCONFIRM
)

func (s VoteStage) String() string {
	switch s {
	case VoteStageINIT:
		return "INIT"
	case VoteStageSIGN:
		return "SIGN"
	case VoteStageACCEPT:
		return "ACCEPT"
	case VoteStageALLCONFIRM:
		return "ALLCONFIRM"
	default:
		return ""
	}
}

func (s VoteStage) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *VoteStage) UnmarshalJSON(b []byte) error {
	var n string
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}

	switch n {
	case "INIT":
		*s = VoteStageINIT
	case "SIGN":
		*s = VoteStageSIGN
	case "ACCEPT":
		*s = VoteStageACCEPT
	case "ALLCONFIRM":
		*s = VoteStageALLCONFIRM
	default:
		return InvalidVoteStageError
	}

	return nil
}

func (s VoteStage) IsValid() bool {
	switch s {
	case VoteStageINIT:
	case VoteStageSIGN:
	case VoteStageACCEPT:
	default:
		return false
	}

	return true
}

func (s VoteStage) Next() VoteStage {
	switch s {
	case VoteStageINIT:
		return VoteStageSIGN
	case VoteStageSIGN:
		return VoteStageACCEPT
	case VoteStageACCEPT:
		return VoteStageALLCONFIRM
	default:
		return VoteStageNONE
	}
}

func (s VoteStage) CanVote() bool {
	switch s {
	case VoteStageSIGN:
	case VoteStageACCEPT:
	default:
		return false
	}

	return true
}
