package isaac

import (
	"encoding/json"
	"time"

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
	Result      VoteResult                        `json:"result"`
	Proposal    common.Hash                       `json:"proposal"`
	Proposer    common.Address                    `json:"proposer"`
	Block       common.Hash                       `json:"block"`
	Height      common.Big                        `json:"height"`
	Round       Round                             `json:"round"`
	Stage       VoteStage                         `json:"stage"`
	Proposed    bool                              `json:"proposed"`
	LastVotedAt common.Time                       `json:"last_voted_at"`
	Voted       map[common.Address]VotingBoxVoted `json:"voted"` // NOTE for helping to check
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

func (v VoteResultInfo) JSONLog() {}

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

// TODO rename `VotingBoxStageNode`
type VotingBoxStageNode struct {
	vote    Vote
	block   common.Hash
	seal    common.Hash
	votedAt common.Time
}

func NewVotingBoxStageNode(vote Vote, hash common.Hash, block common.Hash) VotingBoxStageNode {
	return VotingBoxStageNode{
		vote:    vote,
		block:   block,
		seal:    hash,
		votedAt: common.Now(),
	}
}

func (v VotingBoxStageNode) Expired(d time.Duration) bool {
	if d == 0 {
		return false
	}

	if d > 0 {
		d = d * -1
	}

	return v.votedAt.Before(common.Now().Add(d))
}

func (v VotingBoxStageNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"vote":    v.vote,
		"block":   v.block,
		"seal":    v.seal,
		"votedAt": v.votedAt,
	})
}

func (v VotingBoxStageNode) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}

// TODO rename `VotingBoxVoted`
type VotingBoxVoted struct {
	VotingBoxStageNode
	height   common.Big
	round    Round
	proposal common.Hash
	stage    VoteStage
}

func NewVotingBoxVoted(
	voteNode VotingBoxStageNode,
	height common.Big,
	round Round,
	proposal common.Hash,
	stage VoteStage,
) VotingBoxVoted {
	return VotingBoxVoted{
		VotingBoxStageNode: voteNode,
		height:             height,
		round:              round,
		proposal:           proposal,
		stage:              stage,
	}
}

func (v VotingBoxVoted) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"vote":     v.vote,
		"block":    v.block,
		"seal":     v.seal,
		"votedAt":  v.votedAt,
		"height":   v.height,
		"round":    v.round,
		"proposal": v.proposal,
		"stage":    v.stage,
	})
}
