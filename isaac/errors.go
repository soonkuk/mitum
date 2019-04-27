package isaac

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	InvalidSealTypeCode
	InvalidVoteCode
	InvalidVoteStageCode
	VotingProposalAlreadyStartedCode
	VotingProposalNotFoundCode
	KnownSealFoundCode
	SealNotFoundCode
	SomethingWrongVotingCode
	ProposeNotWellformedCode
	BallotNotWellformedCode
	ConsensusNotReadyCode
)

var (
	InvalidSealTypeError              common.Error = common.NewError("isaac", InvalidSealTypeCode, "invalid SealType")
	InvalidVoteError                  common.Error = common.NewError("isaac", InvalidVoteCode, "invalid vote found")
	InvalidVoteStageError             common.Error = common.NewError("isaac", InvalidVoteStageCode, "invalid vote stage found")
	VotingProposalAlreadyStartedError common.Error = common.NewError("isaac", VotingProposalAlreadyStartedCode, "VotingProposal already started")
	VotingProposalNotFoundError       common.Error = common.NewError("isaac", VotingProposalNotFoundCode, "VotingProposal not found")
	KnownSealFoundError               common.Error = common.NewError("isaac", KnownSealFoundCode, "known seal found")
	SealNotFoundError                 common.Error = common.NewError("isaac", SealNotFoundCode, "seal not found")
	SomethingWrongVotingError         common.Error = common.NewError("isaac", SomethingWrongVotingCode, "")
	ProposeNotWellformedError         common.Error = common.NewError("isaac", ProposeNotWellformedCode, "")
	BallotNotWellformedError          common.Error = common.NewError("isaac", BallotNotWellformedCode, "")
	ConsensusNotReadyError            common.Error = common.NewError("isaac", ConsensusNotReadyCode, "consensus is not ready yet")
)
