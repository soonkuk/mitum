package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerProposalIsValid(c *common.ChainChecker) error {
	// TODO test
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var proposal Proposal
	if p, ok := seal.(Proposal); !ok {
		return common.UnknownSealTypeError.SetMessage("not Proposal")
	} else {
		proposal = p
	}

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	if len(proposal.Transactions) > int(policy.MaxTransactionsInProposal) {
		return ProposalNotWellformedError.SetMessage(
			"max allowed number of transactions over; '%d' > '%d'",
			len(proposal.Transactions),
			policy.MaxTransactionsInProposal,
		)
	}

	return nil
}

// CheckerProposalProposerIsFromKnowns checks `Proposal.Proposer` is
// in the known validators
func CheckerProposalProposerIsFromKnowns(c *common.ChainChecker) error {
	// TODO test

	var proposal Proposal
	if err := c.ContextValue("proposal", &proposal); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	isFromHome := proposal.Source() == state.Home().Address()
	_ = c.SetContext("isFromHome", isFromHome)

	if isFromHome {
		c.Log().Debug("Proposal is from home", "seal", proposal.Hash())
	}

	return nil
}

// CheckerProposalProposerIsValid checks `Proposal.Proposer` is the proper
// proposer with Height, Round and Validators
func CheckerProposalProposerIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerProposalBlock checks `Proposal.Block` is correct,
// - Proposal.Block.Height is same
// - Proposal.Block.Current is same
func CheckerProposalBlock(c *common.ChainChecker) error {
	// TODO test
	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	var proposal Proposal
	if err := c.ContextValue("proposal", &proposal); err != nil {
		return err
	}

	if !proposal.Block.Height.Equal(state.Height()) {
		cstop, err := common.NewChainCheckerStop(
			"proposal has different height",
			"proposal", proposal.Block.Height,
			"current", state.Height(),
		)
		if err != nil {
			return err
		}

		return cstop
	}

	if !proposal.Block.Current.Equal(state.Block()) {
		cstop, err := common.NewChainCheckerStop(
			"proposal has different block",
			"proposal", proposal.Block.Current,
			"current", state.Block(),
		)
		if err != nil {
			return err
		}

		return cstop
	}

	return nil
}

// CheckerProposalState checks `Proposal.State` is correct
func CheckerProposalState(c *common.ChainChecker) error {
	// TODO test
	return nil
}
