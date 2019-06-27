package transaction

import "github.com/spikeekips/mitum/common"

var (
	TransactionType     common.DataType = common.NewDataType(2, "transaction")
	TransactionHashHint string          = "transaction"
)
