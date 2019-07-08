package isaac

import "time"

type TestProposalValidator struct {
	policy   Policy
	duration time.Duration
}

func NewTestProposalValidator(policy Policy, duration time.Duration) *TestProposalValidator {
	return &TestProposalValidator{policy: policy, duration: duration}
}

func (dp *TestProposalValidator) isValid(proposal Proposal) error {
	if err := proposal.IsValid(); err != nil {
		return err
	}

	// TODO process transactions

	<-time.After(dp.duration)

	return nil
}

func (dp *TestProposalValidator) NewBlock(proposal Proposal) (Block, error) {
	if err := dp.isValid(proposal); err != nil {
		return Block{}, err
	}

	return NewBlock(proposal.Height().Add(1), proposal.Round(), proposal.Hash())
}
