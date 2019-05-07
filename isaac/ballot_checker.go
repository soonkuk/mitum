package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerBallotIsValid(c *common.ChainChecker) error {
	// TODO test
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	if err := ballot.Wellformed(); err != nil {
		return err
	}

	return nil
}

// CheckerSealBallotTimeIsValid checks `Ballot.VotedAt` is not
// far from now
func CheckerBallotTimeIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerBallotProposal checks `Ballot.Proposal`
// exists; if not, request to other nodes and then open new voting
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

	seal, err := sealPool.Get(ballot.Proposal)
	if SealNotFoundError.Equal(err) {
		// TODO unknown Proposal found, request from other nodes
		return nil
	}

	_ = c.SetContext("proposal", seal)

	return nil
}
