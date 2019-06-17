package isaac

import "github.com/spikeekips/mitum/common"

const (
	InvalidStageErrorCode common.ErrorCode = iota + 1
	FailedToVoteErrorCode
)

var (
	InvalidStageError = common.NewError("isaac", InvalidStageErrorCode, "invalid stage")
	FailedToVoteError = common.NewError("isaac", FailedToVoteErrorCode, "failed to vote")
)
