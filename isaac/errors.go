package isaac

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	InvalidVoteErrorCode
	InvalidVoteStageErrorCode
	VotingBoxProposalAlreadyStartedErrorCode
	VotingBoxProposalNotFoundErrorCode
	KnownSealFoundErrorCode
	SealNotFoundErrorCode
	SomethingWrongVotingErrorCode
	ProposalNotWellformedErrorCode
	BallotNotWellformedErrorCode
	ConsensusNotReadyErrorCode
	AnotherProposalIsOpenedErrorCode
	ProposalIsNotOpenedErrorCode
	SealAlreadyVotedErrorCode
	InvalidVoteResultInfoErrorCode
	VotingFailedErrorCode
	FailedToElectProposerErrorCode
	DifferentHeightConsensusErrorCode
	DifferentBlockHashConsensusErrorCode
	InvalidNodeStateErrorCode
	BlockNotFoundErrorCode
	IgnoreVotingResultErrorCode
	BallotIsTooOldErrorCode
	SealNotFromValidatorsErrorCode
	OverSealSignedAtAllowDurationErrorCode
	ProposalHasInvalidProposerErrorCode
	ConsensusButBlockDoesNotMatchErrorCode
	ValidationIsRunningErrorCode
	ValidationIsNotDoneErrorCode
)

var (
	InvalidVoteError                     common.Error = common.NewError("isaac", InvalidVoteErrorCode, "invalid vote found")
	InvalidVoteStageError                common.Error = common.NewError("isaac", InvalidVoteStageErrorCode, "invalid vote stage found")
	VotingBoxProposalAlreadyStartedError common.Error = common.NewError("isaac", VotingBoxProposalAlreadyStartedErrorCode, "VotingBoxProposal already started")
	VotingBoxProposalNotFoundError       common.Error = common.NewError("isaac", VotingBoxProposalNotFoundErrorCode, "VotingBoxProposal not found")
	KnownSealFoundError                  common.Error = common.NewError("isaac", KnownSealFoundErrorCode, "known seal found")
	SealNotFoundError                    common.Error = common.NewError("isaac", SealNotFoundErrorCode, "seal not found")
	SomethingWrongVotingError            common.Error = common.NewError("isaac", SomethingWrongVotingErrorCode, "")
	ProposalNotWellformedError           common.Error = common.NewError("isaac", ProposalNotWellformedErrorCode, "")
	BallotNotWellformedError             common.Error = common.NewError("isaac", BallotNotWellformedErrorCode, "")
	ConsensusNotReadyError               common.Error = common.NewError("isaac", ConsensusNotReadyErrorCode, "consensus is not ready yet")
	AnotherProposalIsOpenedError         common.Error = common.NewError("isaac", AnotherProposalIsOpenedErrorCode, "another opened proposal is running")
	ProposalIsNotOpenedError             common.Error = common.NewError("isaac", ProposalIsNotOpenedErrorCode, "proposal is not opened")
	SealAlreadyVotedError                common.Error = common.NewError("isaac", SealAlreadyVotedErrorCode, "seal is already voted")
	InvalidVoteResultInfoError           common.Error = common.NewError("isaac", InvalidVoteResultInfoErrorCode, "invalid VoteResultInfo")
	VotingFailedError                    common.Error = common.NewError("isaac", VotingFailedErrorCode, "voting failed")
	FailedToElectProposerError           common.Error = common.NewError("isaac", FailedToElectProposerErrorCode, "failed to elect proposer")
	DifferentHeightConsensusError        common.Error = common.NewError("isaac", DifferentHeightConsensusErrorCode, "consensused, but different height found")
	DifferentBlockHashConsensusError     common.Error = common.NewError("isaac", DifferentBlockHashConsensusErrorCode, "consensused, but different block hash found")
	InvalidNodeStateError                common.Error = common.NewError("isaac", InvalidNodeStateErrorCode, "invalid NodeState")
	BlockNotFoundError                   common.Error = common.NewError("isaac", BlockNotFoundErrorCode, "block not found")
	IgnoreVotingResultError              common.Error = common.NewError("isaac", IgnoreVotingResultErrorCode, "ignore votingResult")
	BallotIsTooOldError                  common.Error = common.NewError("isaac", BallotIsTooOldErrorCode, "ballot is too old")
	SealNotFromValidatorsError           common.Error = common.NewError("isaac", SealNotFromValidatorsErrorCode, "seal is not from validators")
	OverSealSignedAtAllowDurationError   common.Error = common.NewError("isaac", OverSealSignedAtAllowDurationErrorCode, "Seal.SignedAt() is not within SealSignedAtAllowDuration")
	ProposalHasInvalidProposerError      common.Error = common.NewError("isaac", ProposalHasInvalidProposerErrorCode, "Proposal has invalid proposer")
	ConsensusButBlockDoesNotMatchError   common.Error = common.NewError("isaac", ConsensusButBlockDoesNotMatchErrorCode, "consensus, but VoteResultInfo.Block does not match with previous result")
	ValidationIsRunningError             common.Error = common.NewError("isaac", ValidationIsRunningErrorCode, "validation is running")
	ValidationIsNotDoneError             common.Error = common.NewError("isaac", ValidationIsNotDoneErrorCode, "validation is not done")
)
