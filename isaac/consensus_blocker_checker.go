package isaac

import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

func CheckerBlockerProposalBlock(c *common.ChainChecker) error {
	var proposal Proposal
	if err := c.ContextValue("proposal", &proposal); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	if proposal.Block.Height.Cmp(state.Height()) != 0 {
		c.Log().Debug(
			"different height proposal received",
			"in_proposal", proposal.Block.Height,
			"current", state.Height(),
		)
		return common.NewChainCheckerStop(
			"different height proposal received",
			"in_proposal", proposal.Block.Height,
			"current", state.Height(),
		)
	}

	if !proposal.Block.Current.Equal(state.Block()) {
		c.Log().Debug(
			"proposal block is not matched",
			"in_proposal", proposal.Block.Current,
			"current", state.Block(),
		)
	}

	return nil
}

func CheckerBlockerBallot(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	if ballot.Height().Cmp(state.Height()) < 0 {
		c.Log().Debug(
			"lower height ballot received",
			"in_ballot", ballot.Height(),
			"current", state.Height(),
		)
	}

	return nil
}

func CheckerBlockerBallotVotingResult(c *common.ChainChecker) error {
	var result VoteResultInfo
	if err := c.ContextValue("votingResult", &result); err != nil {
		return err
	}
	if result.NotYet() {
		return common.NewChainCheckerStop("voting result, not yet", "result", result)
	}

	var last VoteResultInfo
	if err := c.ContextValue("lastVotingResult", &last); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	log_ := c.Log().New(log15.Ctx{"result": result, "current": state, "last": last})

	if result.Height.Cmp(state.Height()) < 0 {
		log_.Debug(
			"height, lower than current state",
		)
		return nil
	}

	if result.Height.Cmp(last.Height) < 0 {
		log_.Debug(
			"height, lower than last result",
		)
		return nil
	}

	if result.Stage == VoteStageINIT {
		return nil
	}

	if result.Height.Equal(last.Height) && result.Round == last.Round && result.Stage <= last.Stage {
		log_.Debug("already finished; earlier stage found")
		return nil
	}

	if result.Height.Cmp(state.Height()) > 0 {
		log_.Debug("higher height found")
		return DifferentHeightConsensusError.AppendMessage(
			"height=%v, current height=%v",
			result.Height,
			state.Height(),
		)
	}

	return nil
}
