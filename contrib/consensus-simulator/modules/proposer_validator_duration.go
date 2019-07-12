package modules

import (
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
)

type DurationProposalValidator struct {
	*common.Logger
	duration time.Duration
}

func NewDurationProposalValidator(duration time.Duration) *DurationProposalValidator {
	return &DurationProposalValidator{
		Logger:   common.NewLogger(log, "duration", duration),
		duration: duration,
	}
}

func (dp *DurationProposalValidator) isValid(proposal isaac.Proposal) error {
	if err := proposal.IsValid(); err != nil {
		return err
	}

	<-time.After(dp.duration)

	return nil
}

func (dp *DurationProposalValidator) NewBlock(proposal isaac.Proposal) (block isaac.Block, err error) {
	dp.Log().Debug("trying to validate proposal", "proposal", proposal)
	if err = dp.isValid(proposal); err != nil {
		return
	}

	defer func() {
		dp.Log().Debug("proposal validated", "proposal", proposal, "block", block)
	}()

	block, err = isaac.NewBlock(proposal.Height().Add(1), proposal.Round(), proposal.Hash())

	return
}
