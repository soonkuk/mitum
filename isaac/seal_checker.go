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
	case ProposeSealType:
		var propose Propose
		if err := seal.UnmarshalBody(&propose); err != nil {
			return err
		}
		c.SetContext("propose", propose)

		return common.NewChainChecker(
			"Propose checker",
			c.Context(),
			CheckerProposeIsValid,
			CheckerProposeProposerIsFromKnowns,
			CheckerProposeProposerIsValid,
			CheckerProposeTimeIsValid,
			CheckerProposeBlock,
			CheckerProposeState,
			CheckerProposeOpenVoting,
			CheckerProposeValidate,
			CheckerProposeNextStageBroadcast,
		)
	case BallotSealType:
		var ballot Ballot
		if err := seal.UnmarshalBody(&ballot); err != nil {
			return err
		}
		c.SetContext("ballot", ballot)

		return common.NewChainChecker(
			"ballot checker",
			c.Context(),
			CheckerBallotIsValid,
			CheckerBallotTimeIsValid,
			CheckerBallotIsFinished,
			CheckerBallotProposeSeal,
			CheckerBallotVote,
			CheckerBallotVoteResult,
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
