package isaac

import "github.com/spikeekips/mitum/common"

const (
	InvalidStageErrorCode common.ErrorCode = iota + 1
	FailedToVoteErrorCode
	AlreadyVotedErrorCode
	InvalidPolicyValueErrorCode
	InvalidBallotErrorCode
	ChangeNodeStateToSyncErrorCode
)

var (
	InvalidStageError          = common.NewError("isaac", InvalidStageErrorCode, "invalid stage")
	FailedToVoteError          = common.NewError("isaac", FailedToVoteErrorCode, "failed to vote")
	AlreadyVotedError          = common.NewError("isaac", AlreadyVotedErrorCode, "node already voted")
	InvalidPolicyValueError    = common.NewError("isaac", InvalidPolicyValueErrorCode, "invalid policy value")
	InvalidBallotError         = common.NewError("isaac", InvalidBallotErrorCode, "invalid ballot")
	ChangeNodeStateToSyncError = common.NewError("isaac", ChangeNodeStateToSyncErrorCode, "state changes to sync")
)
