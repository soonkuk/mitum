package leveldbstorage

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	LevelDBErrorCode
	DBNotClosedErrorCode
	TransactionAlreadyOpenedErrorCode
	TransactionNotOpenedErrorCode
	WrongBatchErrorCode
)

var (
	LevelDBError                  common.Error = common.NewError("leveldb", LevelDBErrorCode, "leveldb error")
	DBNotClosedError              common.Error = common.NewError("leveldb", DBNotClosedErrorCode, "db not closed")
	TransactionAlreadyOpenedError common.Error = common.NewError("leveldb", TransactionAlreadyOpenedErrorCode, "transaction already opened")
	TransactionNotOpenedError     common.Error = common.NewError("leveldb", TransactionNotOpenedErrorCode, "transaction not opened")
	WrongBatchError               common.Error = common.NewError("leveldb", WrongBatchErrorCode, "wrong batch")
)
