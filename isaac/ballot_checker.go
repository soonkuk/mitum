package isaac

import (
	"github.com/spikeekips/mitum/common"
)

// CheckerBallotProposal checks `Ballot.Proposal` exists; if not, request to
// other nodes and then open new voting
func CheckerBallotProposal(c *common.ChainChecker) error {
	// TODO test

	var sealPool SealPool
	if err := c.ContextValue("sealPool", &sealPool); err != nil {
		return err
	}

	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	// NOTE INIT ballot does not have Proposal
	if !ballot.Proposal().Empty() {
		seal, err := sealPool.Get(ballot.Proposal())
		if SealNotFoundError.Equal(err) {
			// TODO unknown Proposal found, request from other nodes
			return nil
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
			"ballot has problem; height is not matched; ballot=%v proposal=%v",
			ballot.Height(),
			proposal.Block.Height,
		)
	}

	// check Round
	if ballot.Round() != proposal.Round {
		return BallotNotWellformedError.SetMessage(
			"ballot has problem; round is not matched; ballot=%v proposal=%v",
			ballot.Round(),
			proposal.Round,
		)
	}

	return nil
}

func CheckerBallotHasValidProposr(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	// NOTE INIT ballot should have Proposer
	if ballot.Stage() != VoteStageINIT {
		return nil
	}

	// TODO proposer should be checked
	return nil
}
