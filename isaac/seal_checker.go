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
		if KnownSealFoundError.Equal(err) {
			cstop, err := common.NewChainCheckerStop(
				err.(common.Error).Message(),
				"error", err,
			)
			if err != nil {
				return err
			} else {
				return cstop
			}
		}

		return err
	}

	c.Log().Debug("seal added", "seal", seal.Hash(), "seal-original", seal)

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
			CheckerProposalBlock,
			CheckerProposalState,
		)
	case BallotSealType:
		_ = c.SetContext("ballot", seal.(Ballot))

		return common.NewChainChecker(
			"ballot checker",
			c.Context(),
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

// CheckerSealSignedAtTimeIsValid checks `Seal.SignedAt` is not far from now
func CheckerSealSignedAtTimeIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}
