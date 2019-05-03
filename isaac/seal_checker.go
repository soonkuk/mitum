package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerSealPool(c *common.ChainChecker) error {
	var sealPool SealPool
	if err := c.ContextValue("sealPool", &sealPool); err != nil {
		return err
	}

	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	if err := sealPool.Add(seal); err != nil {
		return err
	}

	return nil
}

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
		_ = c.SetContext("propose", propose)

		return common.NewChainChecker(
			"Propose checker",
			c.Context(),
			CheckerProposeIsValid,
			CheckerProposeProposerIsFromKnowns,
			CheckerProposeProposerIsValid,
			CheckerProposeTimeIsValid,
			CheckerProposeBlock,
			CheckerProposeState,
		)
	case BallotSealType:
		var ballot Ballot
		if err := seal.UnmarshalBody(&ballot); err != nil {
			return err
		}
		_ = c.SetContext("ballot", ballot)

		return common.NewChainChecker(
			"ballot checker",
			c.Context(),
			CheckerBallotIsValid,
			CheckerBallotTimeIsValid,
			CheckerBallotProposeSeal,
		)
	case TransactionSealType:
		// TODO store transaction
		return common.ChainCheckerStop{}
	case common.SealedSealType:
		// TODO decapsule sealed seal
	default:
		return common.InvalidSealTypeError
	}

	return nil
}
