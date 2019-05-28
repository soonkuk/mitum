package isaac

import (
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
		return common.SealIgnoredError.AppendMessage(
			"different height proposal received; seal=%v in_proposal=%v current=%v",
			proposal.Hash(),
			proposal.Block.Height,
			state.Height(),
		)
	}

	if !proposal.Block.Current.Equal(state.Block()) {
		c.Log().Debug(
			"different block proposal received",
			"in_proposal", proposal.Block.Current,
			"current", state.Block(),
		)
		return common.SealIgnoredError.AppendMessage(
			"different block proposal received; seal=%v in_proposal=%v current=%v",
			proposal.Hash(),
			proposal.Block.Current,
			state.Block(),
		)
	}

	// TODO check votingbox is opened or not

	return nil
}

// CheckerBlockerProposalLastVotingResult checks proposal is from last voting result
func CheckerBlockerProposalLastVotingResult(c *common.ChainChecker) error {
	// TODO test
	var last VoteResultInfo
	if err := c.ContextValue("lastVotingResult", &last); err != nil {
		return err
	}

	if last.NotYet() {
		return nil
	}

	var proposal Proposal
	if err := c.ContextValue("proposal", &proposal); err != nil {
		return err
	}

	// NOTE check last voting result is INIT
	if last.Stage != VoteStageINIT {
		return common.SealIgnoredError.AppendMessage(
			"last voting result is not INIT; last=%v",
			last.Stage,
		)
	}

	// NOTE check height
	if !proposal.Block.Height.Equal(last.Height) {
		return common.SealIgnoredError.AppendMessage(
			"height is different from last voting result; proposal=%v last=%v",
			proposal.Block.Height,
			last.Height,
		)
	}

	// NOTE check round
	if proposal.Round != last.Round {
		return common.SealIgnoredError.AppendMessage(
			"round is different from last voting result; proposal=%v last=%v",
			proposal.Round,
			last.Round,
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

	c.Log().Debug("checking votingResult", "result", result, "current", state, "last", last)

	if result.Height.Cmp(state.Height()) > 0 {
		c.Log().Debug("higher height found")
		return DifferentHeightConsensusError.AppendMessage(
			"height=%v, current height=%v",
			result.Height,
			state.Height(),
		)
	}

	if result.Height.Cmp(state.Height()) < 0 {
		c.Log().Debug("height, lower than current state")

		return common.SealIgnoredError.AppendMessage(
			"height, lower than current state; height=%v current=%v",
			result.Height,
			state.Height(),
		)
	}

	if result.Stage == VoteStageINIT {
		return nil
	}

	if result.Height.Cmp(state.Height()) > 0 {
		c.Log().Debug("higher height found")
		return DifferentHeightConsensusError.AppendMessage(
			"height=%v, current height=%v",
			result.Height,
			state.Height(),
		)
	}

	if !last.NotYet() {
		if result.Height.Equal(last.Height) && result.Round == last.Round && result.Stage <= last.Stage {
			c.Log().Debug("already finished; ")
			return common.SealIgnoredError.AppendMessage(
				"earlier stage found; stage already got result; result=%v last=%v",
				result.Stage,
				last.Stage,
			)
			return nil
		}

		if !last.Block.Empty() && !result.Block.Equal(last.Block) {
			return ConsensusButBlockDoesNotMatchError.AppendMessage(
				"result block=%v, last block=%v",
				result.Block,
				last.Block,
			)
		}
	}

	return nil
}

func CheckerBlockerVotingBallotResult(c *common.ChainChecker) error {
	var result VoteResultInfo
	if err := c.ContextValue("votingResult", &result); err != nil {
		return err
	}

	if result.Stage == VoteStageINIT {
		return nil
	}

	var last VoteResultInfo
	if err := c.ContextValue("lastVotingResult", &last); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	if result.Height.Equal(last.Height) && result.Round == last.Round && result.Stage <= last.Stage {
		c.Log().Debug("already finished; earlier stage found")
		// TODO this ballot should be ignored in blocker
		return nil
	}

	if result.Height.Cmp(state.Height()) > 0 {
		c.Log().Debug("higher height found")
		return DifferentHeightConsensusError.AppendMessage(
			"height=%v, current height=%v",
			result.Height,
			state.Height(),
		)
	}

	if !last.Block.Empty() && !result.Block.Equal(last.Block) {
		return ConsensusButBlockDoesNotMatchError.AppendMessage(
			"result block=%v, last block=%v",
			result.Block,
			last.Block,
		)
	}

	return nil
}
