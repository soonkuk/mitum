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

	switch seal.Type() {
	case ProposalSealType:
		var proposal Proposal
		if p, ok := seal.(Proposal); !ok {
			return common.UnknownSealTypeError.SetMessage("not Proposal")
		} else {
			proposal = p
		}
		_ = c.SetContext("proposal", proposal)

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
		var ballot Ballot
		if b, ok := seal.(Ballot); !ok {
			return common.UnknownSealTypeError.SetMessage("not Ballot")
		} else {
			ballot = b
		}
		_ = c.SetContext("ballot", ballot)

		return common.NewChainChecker(
			"ballot checker",
			c.Context(),
			CheckerBallotIsValid,
			CheckerBallotTimeIsValid,
			CheckerBallotProposal,
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
