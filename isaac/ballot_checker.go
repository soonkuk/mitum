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

// CheckerBallotIsFinished checks the vote is finished or not
func CheckerBallotIsFinished(c *common.ChainChecker) error {
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

	c.SetContext("proposeSeal", seal)

	return nil
}

// CheckerBallotVote votes
func CheckerBallotVote(c *common.ChainChecker) error {
	// TODO test
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	var roundVoting *RoundVoting
	if err := c.ContextValue("roundVoting", &roundVoting); err != nil {
		return err
	}

	var sHash common.Hash
	if err := c.ContextValue("sHash", &sHash); err != nil {
		return err
	}

	vp, vs, err := roundVoting.Vote(ballot)
	if err != nil {
		return err
	}

	c.Log().Debug("voted", "seal", sHash, "voting-proposal", vp, "voting-stage", vs)

	return nil
}

// CheckerBallotVoteResult checks voting result
func CheckerBallotVoteResult(c *common.ChainChecker) error {
	// TODO test

	var sHash common.Hash
	if err := c.ContextValue("sHash", &sHash); err != nil {
		return err
	}

	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	var roundVoting *RoundVoting
	if err := c.ContextValue("roundVoting", &roundVoting); err != nil {
		return err
	}

	vp := roundVoting.Proposal(ballot.ProposeSeal)
	if vp == nil {
		return VotingProposalNotFoundError
	}

	vs := vp.Stage(ballot.Stage)

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	if !vs.CanCount(policy.Total, policy.Threshold) {
		return common.ChainCheckerStop{}
	}

	majority := vs.Majority(policy.Total, policy.Threshold)

	c.Log().Debug(
		"consensus got majority",
		"ProposeSeal", ballot.ProposeSeal,
		"stage", ballot.Stage,
		"majority", majority,
		"total", policy.Total,
		"threshold", policy.Threshold,
	)

	switch majority {
	case VoteResultNotYet:
		return SomethingWrongVotingError.SetMessage(
			"something wrong; CanCount() but voting not yet finished",
		)
	}

	if err := roundVoting.Agreed(ballot.ProposeSeal); err != nil {
		return err
	}

	switch majority {
	case VoteResultDRAW:
		// TODO if voting result is draw, start new round
		return common.ChainCheckerStop{}
	case VoteResultNOP:
		// NOTE , start new round
		return common.ChainCheckerStop{}
	}

	// NOTE consensus agreed, move to next stage

	if ballot.Stage == VoteStageACCEPT { // it means, ALLCONFIRM reaches
		if err := roundVoting.Finish(ballot.ProposeSeal); err != nil {
			return err
		}

		var blockStorage BlockStorage
		if err := c.ContextValue("blockStorage", &blockStorage); err != nil {
			return err
		}

		var proposeSeal common.Seal
		if err := c.ContextValue("proposeSeal", &proposeSeal); err != nil {
			return err
		}

		if err := blockStorage.NewBlock(proposeSeal); err != nil {
			return nil
		}
	}

	roundBoy, ok := c.Context().Value("roundBoy").(RoundBoy)
	if !ok {
		return common.ContextValueNotFoundError.SetMessage("'roundBoy' not found")
	}

	// TODO set VoteXXX
	roundBoy.Transit(ballot.Stage, ballot.ProposeSeal, VoteYES)

	return nil
}
