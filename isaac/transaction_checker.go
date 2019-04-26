package isaac

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/element"
)

func CheckTransactionIsValid(c *common.ChainChecker) error {
	// TODO test
	var tx element.Transaction
	if err := c.ContextValue("transaction", &tx); err != nil {
		return err
	}

	if err := tx.Wellformed(); err != nil {
		return err
	}

	var policy ConsensusPolicy
	if err := c.ContextValue("policy", &policy); err != nil {
		return err
	}

	if len(tx.Operations) > int(policy.MaxOperationsInTransaction) {
		return element.TransactionNotWellformedError.SetMessage(
			"max allowed number of operations over; '%d' > '%d'",
			len(tx.Operations),
			policy.MaxOperationsInTransaction,
		)
	}

	return nil
}
