package element

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	TransactionNotWellformedCode
)

var (
	TransactionNotWellformedError common.Error = common.NewError("element", TransactionNotWellformedCode, "")
)
