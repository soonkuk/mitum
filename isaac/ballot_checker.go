package isaac

import (
	"github.com/spikeekips/mitum/common"
)

// CheckerBallotHasValidState checks,
// - height is equal or higher than current
// - if height is same, block is same with current
// - if height is same, state is same with current
func CheckerBallotHasValidState(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	if ballot.Height().Cmp(state.Height()) < 0 {
		return common.SealIgnoredError.AppendMessage(
			"height is lower than current: ballot=%v current=%v",
			ballot.Height(), state.Height(),
		)
	}

	return nil
}

// CheckerBallotProposal checks `Ballot.Proposal` exists; if not, request to
// other nodes and then open new voting
func CheckerBallotProposal(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	if ballot.Stage() != VoteStageINIT {
		var sealPool SealPool
		if err := c.ContextValue("sealPool", &sealPool); err != nil {
			return err
		}

		seal, err := sealPool.Get(ballot.Proposal())
		if SealNotFoundError.Equal(err) {
			// TODO unknown Proposal found, request from other nodes
			// TODO check proposal is valid, including proposer
			return SealNotFoundError.SetError(err).AppendMessage(
				"failed to get proposal",
			)
		}

		_ = c.SetContext("proposal", seal)
	}

	return nil
}

func CheckerBallotHasValidProposal(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	// NOTE INIT ballot does not have Proposal
	if ballot.Stage() == VoteStageINIT {
		return nil
	}

	var proposal Proposal
	if err := c.ContextValue("proposal", &proposal); err != nil {
		return err
	}

	// check Height
	if !ballot.Height().Equal(proposal.Block.Height) {
		return BallotNotWellformedError.SetMessage(
			"ballot has problem; height does not match; ballot=%v in_ballot=%v proposal=%v",
			ballot.Hash(),
			ballot.Height(),
			proposal.Block.Height,
		)
	}

	// check Round
	if ballot.Round() != proposal.Round {
		return BallotNotWellformedError.SetMessage(
			"ballot has problem; round does not match; ballot=%v in_ballot=%v proposal=%v",
			ballot.Hash(),
			ballot.Round(),
			proposal.Round,
		)
	}

	// check Proposer
	if ballot.Proposer() != proposal.Source() {
		return BallotNotWellformedError.SetMessage(
			"ballot has problem; proposer deos not match; ballot=%v in_ballot=%v proposal=%v",
			ballot.Hash(),
			ballot.Proposer(),
			proposal.Source(),
		)
	}

	return nil
}

func CheckerBallotHasValidProposer(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	// NOTE INIT ballot does not have Proposal
	if ballot.Stage() == VoteStageINIT {
		return nil
	}

	var proposal Proposal
	if err := c.ContextValue("proposal", &proposal); err != nil {
		return err
	}

	var proposerSelector ProposerSelector
	if err := c.ContextValue("proposerSelector", &proposerSelector); err != nil {
		return err
	}
	proposer, err := proposerSelector.Select(
		proposal.Block.Current,
		ballot.Height(),
		ballot.Round(),
	)
	if err != nil {
		return err
	} else if ballot.Proposer() != proposer.Address() {
		return BallotHasInvalidProposerError.AppendMessage(
			"ballot has invalid proposer; ballot=%v selected=%v",
			ballot.Proposer(),
			proposer.Address(),
		)
	}

	return nil
}
