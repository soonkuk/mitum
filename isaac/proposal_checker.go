package isaac

import (
	"bytes"

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
func CheckerProposalCurrent(c *common.ChainChecker) error {
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
		return common.SealIgnoredError.AppendMessage(
			"proposal has different height: proposal=%v current=%v",
			proposal.Block.Height,
			state.Height(),
		)
	}

	if !proposal.Block.Current.Equal(state.Block()) {
		return common.SealIgnoredError.AppendMessage(
			"proposal has different block; proposal=%v current=%v",
			proposal.Block.Current,
			state.Block(),
		)
	}

	if bytes.Compare(proposal.State.Current, state.State()) != 0 {
		return common.SealIgnoredError.AppendMessage(
			"proposal has different state; proposal=%v current=%v",
			proposal.State.Current,
			state.State(),
		)
	}

	return nil
}
