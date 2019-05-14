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

	if proposal.Block.Height.Cmp(state.Height().Inc()) != 0 {
		c.Log().Debug(
			"different height proposal received",
			"proposa", proposal.Block.Height,
			"current", state.Height(),
		)
		return common.ChainCheckerStop{}
	}

	if !proposal.Block.Current.Equal(state.Block()) {
		c.Log().Debug(
			"proposal block is not matched",
			"proposal", proposal.Block.Current,
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

	if ballot.Height.Cmp(state.Height()) < 1 {
		c.Log().Debug(
			"lower height ballot received",
			"ballot", ballot.Height,
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

	var last VoteResultInfo
	if err := c.ContextValue("lastVotingResult", &last); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	log_ := c.Log().New(log15.Ctx{"result": result, "current": state, "last": last})

	if result.Height.Cmp(state.Height()) < 1 {
		log_.Debug("lower height ballot received")
	}

	if last.NotYet() {
		return nil
	}

	if result.Height.Cmp(last.Height) < 0 {
		log_.Debug("height is lower than last result")
	}

	if result.Stage == VoteStageINIT {
		return nil
	}

	if result.Height.Equal(last.Height) && result.Round == last.Round && result.Stage <= last.Stage {
		log_.Debug("already finished; earlier stage found")
	}

	if result.Height.Cmp(state.Height().Inc()) > 0 {
		log_.Debug("different height found")
		return DifferentHeightConsensusError.AppendMessage(
			"height=%v, current height=%v",
			result.Height,
			state.Height(),
		)
	}

	return nil
}
