package isaac

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	InvalidSealTypeCode
	InvalidVoteCode
	InvalidVoteStageCode
	RunningRoundAlreadyExistsCode
	RunningRoundNotFoundCode
	VotingProposalAlreadyStartedCode
	VotingRoundAlreadyStartedCode
)

var (
	InvalidSealTypeError              common.Error = common.NewError("isaac", InvalidSealTypeCode, "invalid SealType")
	InvalidVoteError                  common.Error = common.NewError("isaac", InvalidVoteCode, "invalid vote found")
	InvalidVoteStageError             common.Error = common.NewError("isaac", InvalidVoteStageCode, "invalid vote stage found")
	RunningRoundAlreadyExistsError    common.Error = common.NewError("isaac", RunningRoundAlreadyExistsCode, "RunningRound already exists")
	RunningRoundNotFoundError         common.Error = common.NewError("isaac", RunningRoundNotFoundCode, "RunningRound not found")
	VotingProposalAlreadyStartedError common.Error = common.NewError("isaac", VotingProposalAlreadyStartedCode, "VotingProposal already started")
	VotingRoundAlreadyStartedError    common.Error = common.NewError("isaac", VotingRoundAlreadyStartedCode, "VotingRound already started")
)
