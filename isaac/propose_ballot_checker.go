package isaac

import (
	"github.com/spikeekips/mitum/common"
)

func CheckerProposeBallotIsValid(c *common.ChainChecker) error {
	// TODO test
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var proposeBallot ProposeBallot
	if err := c.ContextValue("ballot", &proposeBallot); err != nil {
		return err
	}

	if err := proposeBallot.Wellformed(); err != nil {
		return err
	}

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	if len(proposeBallot.Transactions) > int(policy.MaxTransactionsInBallot) {
		return ProposeBallotNotWellformedError.SetMessage(
			"max allowed number of transactions over; '%d' > '%d'",
			len(proposeBallot.Transactions),
			policy.MaxTransactionsInBallot,
		)
	}

	// source must be same
	if seal.Source != proposeBallot.Proposer {
		return ProposeBallotNotWellformedError.SetMessage(
			"Seal.Source does not match with ProposeBallot.Proposer; '%s' != '%s'",
			seal.Source,
			proposeBallot.Proposer,
		)
	}

	return nil
}

// CheckerProposeBallotProposerIsFromKnowns checks `ProposeBallot.Proposer` is
// in the known validators
func CheckerProposeBallotProposerIsFromKnowns(c *common.ChainChecker) error {
	// TODO test

	var proposeBallot ProposeBallot
	if err := c.ContextValue("ballot", &proposeBallot); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	isFromHomeNode := proposeBallot.Proposer == state.Node().Address()
	c.SetContext("isFromHomeNode", isFromHomeNode)

	if isFromHomeNode {
		var sealHash common.Hash
		if err := c.ContextValue("sealHash", &sealHash); err != nil {
			return err
		}

		c.Log().Debug("proposeBallot is from HomeNode", "seal", sealHash)
	}

	return nil
}

// CheckerProposeBallotProposerIsValid checks `ProposeBallot.Proposer` is the proper
// proposer with Height and Round
func CheckerProposeBallotProposerIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerSealProposeBallotTimeIsValid checks `ProposeBallot.ProposedAt` is not
// far from now
func CheckerProposeBallotTimeIsValid(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerSealProposeBallotBlock checks `ProposeBallot.Block` is correct
func CheckerProposeBallotBlock(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerSealProposeBallotState checks `ProposeBallot.State` is correct
func CheckerProposeBallotState(c *common.ChainChecker) error {
	// TODO test
	return nil
}

// CheckerProposeBallotOpenVoting opens new voting
func CheckerProposeBallotOpenVoting(c *common.ChainChecker) error {
	// TODO test
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var sealHash common.Hash
	if err := c.ContextValue("sealHash", &sealHash); err != nil {
		return err
	}

	var roundVoting *RoundVoting
	if err := c.ContextValue("roundVoting", &roundVoting); err != nil {
		return err
	}

	vp, stage, err := roundVoting.Open(seal)
	if err != nil {
		return err
	}

	c.Log().Debug("starting new round", "seal", sealHash, "voting-proposal", vp, "voting-stage", stage)

	return nil
}

// CheckerProposeBallotValidate validates the received ProposeBallot; if
// validated and it looks good, `vote=VoteYES`.
func CheckerProposeBallotValidate(c *common.ChainChecker) error {
	// TODO test
	// TODO validate ProposeBallot

	var isFromHomeNode bool
	if err := c.ContextValue("isFromHomeNode", &isFromHomeNode); err != nil {
		return err
	}

	if isFromHomeNode {
		// NOTE should be YES
		c.SetContext("vote", VoteYES)
	}

	// TODO remove :) this is for testing
	c.SetContext("vote", VoteYES)

	return nil
}

func CheckerProposeBallotNextStageBroadcast(c *common.ChainChecker) error {
	var seal common.Seal
	if err := c.ContextValue("seal", &seal); err != nil {
		return err
	}

	var state *ConsensusState
	if err := c.ContextValue("state", &state); err != nil {
		return err
	}

	var vote Vote
	if err := c.ContextValue("vote", &vote); err != nil {
		return err
	}

	var sealHash common.Hash
	if err := c.ContextValue("sealHash", &sealHash); err != nil {
		return err
	}

	voteBallot, err := NewVoteBallot(sealHash, state.Node().Address(), vote)
	if err != nil {
		return err
	}

	c.Log().Debug("new VoteBallot will be broadcasted", "new-ballot", voteBallot)

	stageTransistor, ok := c.Context().Value("stageTransistor").(StageTransistor)
	if !ok {
		return common.ContextValueNotFoundError.SetMessage("'stageTransistor' not found")
	}

	err = stageTransistor.Transit(sealHash, VoteStageSIGN, seal, vote)
	if err != nil {
		c.Log().Error("failed to stage transition", "error", err)
		return err
	}

	return nil
}
