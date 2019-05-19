package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerSealFromKnowValidator(c *common.ChainChecker) error {
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	isFromHome := seal.Source() == state.Home().Address()
	_ = c.SetContext("isFromHome", isFromHome)
	if isFromHome {
		c.Log().Debug("seal is from home", "seal", seal.Hash())
	}

	if !isFromHome && !state.ExistsValidators(seal.Source()) {
		return SealNotFromValidatorsError
	}

	return nil
}

func CheckerSealIsValid(c *common.ChainChecker) error {
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	// NOTE checks `Seal.SignedAt` is not far from now
	if !seal.SignedAt().Between(common.Now(), policy.SealSignedAtAllowDuration) {
		return OverSealSignedAtAllowDurationError.AppendMessage(
			"duration=%v", policy.SealSignedAtAllowDuration,
		)
	}

	if err := seal.Wellformed(); err != nil {
		return err
	}

	if err := seal.CheckSignature(policy.NetworkID); err != nil {
		return err
	}

	return nil
}

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
			return common.NewChainCheckerStop(err.(common.Error).Message(), "error", err)
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
			"proposal-checker",
			c.Context(),
			CheckerProposalIsValid,
			CheckerProposalBlock,
			CheckerProposalState,
		)
	case BallotSealType:
		_ = c.SetContext("ballot", seal.(Ballot))

		return common.NewChainChecker(
			"ballot-checker",
			c.Context(),
			CheckerBallotProposal,
			CheckerBallotHasValidProposal,
			CheckerBallotHasValidProposr,
		)
	case TransactionSealType:
		// TODO handle transaction
		return common.NewChainCheckerStop("transaction seal found; this will be implemented")
	default:
		return common.UnknownSealTypeError.SetMessage("tyep=%v", seal.Type())
	}
}
