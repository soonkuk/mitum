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
		if !KnownSealFoundError.Equal(err) {
			return err
		}

		cstop, err := common.NewChainCheckerStop(
			err.(common.Error).Message(),
			"error", err,
		)
		if err != nil {
			return err
		} else {
			return cstop
		}

		return err
	}

	return nil
}

func CheckerSealTypes(c *common.ChainChecker) error {
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	switch seal.Type() {
	case ProposalSealType:
		_ = c.SetContext("proposal", seal.(Proposal))

		return common.NewChainChecker(
			"Proposal checker",
			c.Context(),
			CheckerProposalIsValid,
			CheckerProposalProposerIsFromKnowns,
			CheckerProposalProposerIsValid,
			CheckerProposalTimeIsValid,
			CheckerProposalBlock,
			CheckerProposalState,
		)
	case BallotSealType:
		_ = c.SetContext("ballot", seal.(Ballot))

		return common.NewChainChecker(
			"ballot checker",
			c.Context(),
			CheckerBallotIsValid,
			CheckerBallotTimeIsValid,
			CheckerBallotProposal,
		)
	case TransactionSealType:
		// TODO handle transaction
		cstop, err := common.NewChainCheckerStop("transaction seal found; this will be implemented")
		if err != nil {
			return err
		} else {
			return cstop
		}
	default:
		return common.UnknownSealTypeError.SetMessage("tyep=%v", seal.Type())
	}

	return nil
}
