package isaac

import (
	"time"

	"github.com/spikeekips/mitum/common"
)

type TestProposalValidator struct {
	*common.Logger
	policy   Policy
	duration time.Duration
}

func NewTestProposalValidator(policy Policy, duration time.Duration) *TestProposalValidator {
	return &TestProposalValidator{
		Logger:   common.NewLogger(log, "duration", duration),
		policy:   policy,
		duration: duration,
	}
}

func (dp *TestProposalValidator) isValid(proposal Proposal) error {
	if err := proposal.IsValid(); err != nil {
		return err
	}

	// TODO process transactions

	<-time.After(dp.duration)

	return nil
}

func (dp *TestProposalValidator) NewBlock(proposal Proposal) (block Block, err error) {
	dp.Log().Debug("trying to validate proposal", "proposal", proposal)
	if err = dp.isValid(proposal); err != nil {
		return
	}

	defer func() {
		dp.Log().Debug("proposal validated", "proposal", proposal, "block", block)
	}()

	block, err = NewBlock(proposal.Height().Add(1), proposal.Round(), proposal.Hash())

	return
}
