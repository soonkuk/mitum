package isaac

type ProposalValidator interface {
	NewBlock(Proposal) (Block, error)
}

type DefaultProposalValidator struct {
	policy Policy
}

func NewDefaultProposalValidator(policy Policy) *DefaultProposalValidator {
	return &DefaultProposalValidator{policy: policy}
}

func (dp *DefaultProposalValidator) isValid(proposal Proposal) error {
	if err := proposal.IsValid(); err != nil {
		return err
	}

	// TODO process transactions

	return nil
}

func (dp *DefaultProposalValidator) NewBlock(proposal Proposal) (Block, error) {
	if err := dp.isValid(proposal); err != nil {
		return Block{}, err
	}

	return NewBlock(proposal.Height().Add(1), proposal.Round(), proposal.Hash())
}
