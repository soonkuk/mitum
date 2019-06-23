package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/hash"
)

type CheckMajorityResult uint

const (
	NotYetMajority CheckMajorityResult = iota
	GotMajority
	JustDraw
	FinishedGotMajority
)

func (c CheckMajorityResult) String() string {
	switch c {
	case NotYetMajority:
		return "not-yet-majority"
	case GotMajority:
		return "got-majority"
	case JustDraw:
		return "draw"
	case FinishedGotMajority:
		return "finished"
	}

	return ""
}

func (c CheckMajorityResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

type VoteResult struct {
	height    Height
	round     Round
	stage     Stage
	proposal  hash.Hash
	nextBlock hash.Hash
	records   VoteRecords
	result    CheckMajorityResult
}

func NewVoteResult(
	height Height,
	round Round,
	stage Stage,
	proposal hash.Hash,
	records VoteRecords,
) VoteResult {
	return VoteResult{
		height:   height,
		round:    round,
		stage:    stage,
		proposal: proposal,
		records:  records,
		result:   NotYetMajority,
	}
}

func (vr VoteResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"height":    vr.height,
		"round":     vr.round,
		"stage":     vr.stage,
		"proposal":  vr.proposal,
		"nextBlock": vr.nextBlock,
		"result":    vr.result,
		"records":   vr.records,
	})
}

func (vr VoteResult) Height() Height {
	return vr.height
}

func (vr VoteResult) Round() Round {
	return vr.round
}

func (vr VoteResult) Stage() Stage {
	return vr.stage
}

func (vr VoteResult) Proposal() hash.Hash {
	return vr.proposal
}

func (vr VoteResult) Result() CheckMajorityResult {
	return vr.result
}

func (vr VoteResult) Records() VoteRecords {
	return vr.records
}

// NextBlock should be set by one of majority
func (vr VoteResult) NextBlock() hash.Hash {
	return vr.nextBlock
}

func (vr VoteResult) String() string {
	b, _ := json.Marshal(vr)
	return string(b)
}
