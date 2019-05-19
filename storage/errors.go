package storage

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	RecordNotFoundErrorCode
	RecordAlreadyExistsErrorCode
)

var (
	RecordNotFoundError      common.Error = common.NewError("storage", RecordNotFoundErrorCode, "record not found")
	RecordAlreadyExistsError common.Error = common.NewError("storage", RecordAlreadyExistsErrorCode, "record already exists")
)
