package element

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	TransactionNotWellformedErrorCode
)

var (
	TransactionNotWellformedError common.Error = common.NewError("element", TransactionNotWellformedErrorCode, "")
)
