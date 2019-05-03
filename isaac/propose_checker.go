package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerProposeIsValid(c *common.ChainChecker) error {
	// TODO test
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var propose Propose
	if err := c.ContextValue("propose", &propose); err != nil {
		return err
	}

	if err := propose.Wellformed(); err != nil {
		return err
	}

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	if len(propose.Transactions) > int(policy.MaxTransactionsInPropose) {
		return ProposeNotWellformedError.SetMessage(
			"max allowed number of transactions over; '%d' > '%d'",
			len(propose.Transactions),
			policy.MaxTransactionsInPropose,
		)
	}

	// source must be same
	if seal.Source != propose.Proposer {
		return ProposeNotWellformedError.SetMessage(
			"Seal.Source does not match with Propose.Proposer; '%s' != '%s'",
			seal.Source,
			propose.Proposer,
		)
	}

	return nil
}

// CheckerProposeProposerIsFromKnowns checks `Propose.Proposer` is
// in the known validators
func CheckerProposeProposerIsFromKnowns(c *common.ChainChecker) error {
	// TODO test

	var propose Propose
	if err := c.ContextValue("propose", &propose); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	isFromHomeNode := propose.Proposer == state.Node().Address()
	c.SetContext("isFromHomeNode", isFromHomeNode)

	if isFromHomeNode {
		var psHash common.Hash
		if err := c.ContextValue("sHash", &psHash); err != nil {
			return err
		}

		c.Log().Debug("Propose is from HomeNode", "seal", psHash)
	}

	return nil
}

// CheckerProposeProposerIsValid checks `Propose.Proposer` is the proper
// proposer with Height, Round and Validators
func CheckerProposeProposerIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerSealProposeTimeIsValid checks `Propose.ProposedAt` is not
// far from now
func CheckerProposeTimeIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerSealProposeBlock checks `Propose.Block` is correct
func CheckerProposeBlock(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerSealProposeState checks `Propose.State` is correct
func CheckerProposeState(c *common.ChainChecker) error {
	// TODO test
	return nil
}
