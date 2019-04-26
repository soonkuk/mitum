package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerSealTypes(c *common.ChainChecker) error {
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	switch seal.Type {
	case ProposeBallotSealType:
		var proposeBallot ProposeBallot
		if err := seal.UnmarshalBody(&proposeBallot); err != nil {
			return err
		}
		c.SetContext("ballot", proposeBallot)

		return common.NewChainChecker(
			"proposeBallot checker",
			c.Context(),
			CheckerProposeBallotIsValid,
			CheckerProposeBallotProposerIsFromKnowns,
			CheckerProposeBallotProposerIsValid,
			CheckerProposeBallotTimeIsValid,
			CheckerProposeBallotBlock,
			CheckerProposeBallotState,
			CheckerProposeBallotOpenVoting,
			CheckerProposeBallotValidate,
			CheckerProposeBallotNextStageBroadcast,
		)
	case VoteBallotSealType:
		var voteBallot VoteBallot
		if err := seal.UnmarshalBody(&voteBallot); err != nil {
			return err
		}
		c.SetContext("ballot", voteBallot)

		return common.NewChainChecker(
			"voteBallot checker",
			c.Context(),
			CheckerVoteBallotIsValid,
			CheckerVoteBallotTimeIsValid,
			CheckerVoteBallotIsFinished,
			CheckerVoteBallotProposeBallotSeal,
			CheckerVoteBallotVote,
			CheckerVoteBallotVoteResult,
		)
	case TransactionSealType:
		// TODO store transaction
		return common.ChainCheckerStop{}
	case common.SealedSealType:
		// TODO decapsule sealed seal
	default:
		return InvalidSealTypeError
	}

	return nil
}
