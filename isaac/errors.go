package isaac

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	InvalidVoteCode
	InvalidVoteStageCode
	VotingBoxProposalAlreadyStartedCode
	VotingBoxProposalNotFoundCode
	KnownSealFoundCode
	SealNotFoundCode
	SomethingWrongVotingCode
	ProposalNotWellformedCode
	BallotNotWellformedCode
	ConsensusNotReadyCode
	AnotherProposalIsOpenedCode
	ProposalIsNotOpenedCode
	SealAlreadyVotedCode
	InvalidVoteResultInfoCode
	VotingFailedCode
	FailedToElectProposerCode
)

var (
	InvalidVoteError                     common.Error = common.NewError("isaac", InvalidVoteCode, "invalid vote found")
	InvalidVoteStageError                common.Error = common.NewError("isaac", InvalidVoteStageCode, "invalid vote stage found")
	VotingBoxProposalAlreadyStartedError common.Error = common.NewError("isaac", VotingBoxProposalAlreadyStartedCode, "VotingBoxProposal already started")
	VotingBoxProposalNotFoundError       common.Error = common.NewError("isaac", VotingBoxProposalNotFoundCode, "VotingBoxProposal not found")
	KnownSealFoundError                  common.Error = common.NewError("isaac", KnownSealFoundCode, "known seal found")
	SealNotFoundError                    common.Error = common.NewError("isaac", SealNotFoundCode, "seal not found")
	SomethingWrongVotingError            common.Error = common.NewError("isaac", SomethingWrongVotingCode, "")
	ProposalNotWellformedError           common.Error = common.NewError("isaac", ProposalNotWellformedCode, "")
	BallotNotWellformedError             common.Error = common.NewError("isaac", BallotNotWellformedCode, "")
	ConsensusNotReadyError               common.Error = common.NewError("isaac", ConsensusNotReadyCode, "consensus is not ready yet")
	AnotherProposalIsOpenedError         common.Error = common.NewError("isaac", AnotherProposalIsOpenedCode, "another opened proposal is running")
	ProposalIsNotOpenedError             common.Error = common.NewError("isaac", ProposalIsNotOpenedCode, "proposal is not opened")
	SealAlreadyVotedError                common.Error = common.NewError("isaac", SealAlreadyVotedCode, "seal is already voted")
	InvalidVoteResultInfoError           common.Error = common.NewError("isaac", InvalidVoteResultInfoCode, "invalid VoteResultInfo")
	VotingFailedError                    common.Error = common.NewError("isaac", VotingFailedCode, "voting failed")
	FailedToElectProposerError           common.Error = common.NewError("isaac", FailedToElectProposerCode, "failed to elect proposer")
)
