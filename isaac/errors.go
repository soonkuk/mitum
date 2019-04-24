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
	SomethingWrongVotingCode
	ProposeBallotNotWellformedCode
	VoteBallotNotWellformedCode
)

var (
	InvalidSealTypeError              common.Error = common.NewError("isaac", InvalidSealTypeCode, "invalid SealType")
	InvalidVoteError                  common.Error = common.NewError("isaac", InvalidVoteCode, "invalid vote found")
	InvalidVoteStageError             common.Error = common.NewError("isaac", InvalidVoteStageCode, "invalid vote stage found")
	VotingProposalAlreadyStartedError common.Error = common.NewError("isaac", VotingProposalAlreadyStartedCode, "VotingProposal already started")
	VotingProposalNotFoundError       common.Error = common.NewError("isaac", VotingProposalNotFoundCode, "VotingProposal not found")
	KnownSealFoundError               common.Error = common.NewError("isaac", KnownSealFoundCode, "know seal found")
	SomethingWrongVotingError         common.Error = common.NewError("isaac", SomethingWrongVotingCode, "")
	ProposeBallotNotWellformedError   common.Error = common.NewError("isaac", ProposeBallotNotWellformedCode, "")
	VoteBallotNotWellformedError      common.Error = common.NewError("isaac", VoteBallotNotWellformedCode, "")
)
