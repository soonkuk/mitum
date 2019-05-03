package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerBallotIsValid(c *common.ChainChecker) error {
	// TODO test
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	if err := ballot.Wellformed(); err != nil {
		return err
	}

	// source must be same
	if seal.Source != ballot.Source {
		return BallotNotWellformedError.SetMessage(
			"Seal.Source does not match with Ballot.Source; '%s' != '%s'",
			seal.Source,
			ballot.Source,
		)
	}

	return nil
}

// CheckerSealBallotTimeIsValid checks `Ballot.VotedAt` is not
// far from now
func CheckerBallotTimeIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerBallotProposeSeal checks `Ballot.ProposeSeal`
// exists; if not, request to other nodes and then open new voting
func CheckerBallotProposeSeal(c *common.ChainChecker) error {
	// TODO test

	var sealPool SealPool
	if err := c.ContextValue("sealPool", &sealPool); err != nil {
		return err
	}

	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	seal, err := sealPool.Get(ballot.ProposeSeal)
	if SealNotFoundError.Equal(err) {
		// TODO unknown ProposeSeal found, request from other nodes
		return nil
	}

	_ = c.SetContext("proposeSeal", seal)

	return nil
}
