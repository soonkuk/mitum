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
		return common.UnknownSealTypeError.SetMessage("not Proposal; type=%v", seal.Type())
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

	// NOTE checks `Proposal.Proposer` is the proper proposer with Height, Round
	// and Validators
	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	var proposerSelector ProposerSelector
	if err := c.ContextValue("proposerSelector", &proposerSelector); err != nil {
		return err
	}
	proposer, err := proposerSelector.Select(
		proposal.Block.Current,
		proposal.Block.Height,
		proposal.Round,
	)
	if err != nil {
		return err
	} else if proposal.Source() != proposer.Address() {
		return ProposalHasInvalidProposerError.AppendMessage(
			"proposal=%v selected=%v",
			proposal.Source(),
			proposer,
		)
	}

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
		return common.NewChainCheckerStop(
			"proposal has different height",
			"proposal", proposal.Block.Height,
			"current", state.Height(),
		)
	}

	if !proposal.Block.Current.Equal(state.Block()) {
		return common.NewChainCheckerStop(
			"proposal has different block",
			"proposal", proposal.Block.Current,
			"current", state.Block(),
		)
	}

	return nil
}

// CheckerProposalState checks `Proposal.State` is correct
func CheckerProposalState(c *common.ChainChecker) error {
	// TODO test
	return nil
}
