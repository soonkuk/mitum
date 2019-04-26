package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerVoteBallotIsValid(c *common.ChainChecker) error {
	// TODO test
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var voteBallot VoteBallot
	if err := c.ContextValue("ballot", &voteBallot); err != nil {
		return err
	}

	if err := voteBallot.Wellformed(); err != nil {
		return err
	}

	// source must be same
	if seal.Source != voteBallot.Source {
		return VoteBallotNotWellformedError.SetMessage(
			"Seal.Source does not match with VoteBallot.Source; '%s' != '%s'",
			seal.Source,
			voteBallot.Source,
		)
	}

	return nil
}

// CheckerSealVoteBallotTimeIsValid checks `VoteBallot.VotedAt` is not
// far from now
func CheckerVoteBallotTimeIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerVoteBallotIsFinished checks the vote is finished or not
func CheckerVoteBallotIsFinished(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerVoteBallotProposeBallotSeal checks `VoteBallot.ProposeBallotSeal`
// exists; if not, request to other nodes and then open new voting
func CheckerVoteBallotProposeBallotSeal(c *common.ChainChecker) error {
	// TODO test

	var sealPool SealPool
	if err := c.ContextValue("sealPool", &sealPool); err != nil {
		return err
	}

	var voteBallot VoteBallot
	if err := c.ContextValue("ballot", &voteBallot); err != nil {
		return err
	}

	seal, err := sealPool.Get(voteBallot.ProposeBallotSeal)
	if SealNotFoundError.Equal(err) {
		// TODO unknown proposeBallotSeal found, request from other nodes
		return nil
	}

	c.SetContext("proposeBallotSeal", seal)

	return nil
}

// CheckerVoteBallotVote votes
func CheckerVoteBallotVote(c *common.ChainChecker) error {
	// TODO test
	var voteBallot VoteBallot
	if err := c.ContextValue("ballot", &voteBallot); err != nil {
		return err
	}

	var roundVoting *RoundVoting
	if err := c.ContextValue("roundVoting", &roundVoting); err != nil {
		return err
	}

	var sealHash common.Hash
	if err := c.ContextValue("sealHash", &sealHash); err != nil {
		return err
	}

	vp, vs, err := roundVoting.Vote(voteBallot)
	if err != nil {
		return err
	}

	c.Log().Debug("voted", "seal", sealHash, "voting-proposal", vp, "voting-stage", vs)

	return nil
}

// CheckerVoteBallotVoteResult checks voting result
func CheckerVoteBallotVoteResult(c *common.ChainChecker) error {
	// TODO test

	var sealHash common.Hash
	if err := c.ContextValue("sealHash", &sealHash); err != nil {
		return err
	}

	var voteBallot VoteBallot
	if err := c.ContextValue("ballot", &voteBallot); err != nil {
		return err
	}

	var roundVoting *RoundVoting
	if err := c.ContextValue("roundVoting", &roundVoting); err != nil {
		return err
	}

	vp := roundVoting.Proposal(voteBallot.ProposeBallotSeal)
	if vp == nil {
		return VotingProposalNotFoundError
	}

	vs := vp.Stage(voteBallot.Stage)

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	if !vs.CanCount(policy.Total, policy.Threshold) {
		return common.ChainCheckerStop{}
	}

	majority := vs.Majority(policy.Total, policy.Threshold)
	switch majority {
	case VoteResultNotYet:
		return SomethingWrongVotingError.SetMessage(
			"something wrong; CanCount() but voting not yet finished",
		)
	}

	c.Log().Debug(
		"consensus got majority",
		"proposeBallotSeal", voteBallot.ProposeBallotSeal,
		"stage", VoteStageSIGN,
		"majority", majority,
		"total", policy.Total,
		"threshold", policy.Threshold,
	)

	switch majority {
	case VoteResultNotYet:
		return common.ChainCheckerStop{}
	case VoteResultNOP, VoteResultDRAW:
		// NOTE back to propose
		return common.ChainCheckerStop{}
	}

	// NOTE consensus agreed, move to next stage

	stageTransistor, ok := c.Context().Value("stageTransistor").(StageTransistor)
	if !ok {
		return common.ContextValueNotFoundError.SetMessage("'stageTransistor' not found")
	}

	nextStage := voteBallot.Stage.Next()
	c.Log().Debug("stage will be changed", "current-stage", voteBallot.Stage, "next-stage", nextStage)

	// TODO set VoteXXX

	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	err := stageTransistor.Transit(voteBallot.ProposeBallotSeal, nextStage, seal, VoteYES)
	if err != nil {
		c.Log().Error("failed to stage transition", "error", err)
		return err
	}

	return nil
}
