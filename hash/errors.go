package hash

import "github.com/spikeekips/mitum/common"

const (
	HashFailedErrorCode common.ErrorCode = iota + 1
	EmptyHashErrorCode
	InvalidHashInputErrorCode
	HashAlgorithmAlreadyRegisteredErrorCode
	HashAlgorithmNotRegisteredErrorCode
)

var (
	HashFailedError                     = common.NewError("hash", HashFailedErrorCode, "failed to make hash")
	EmptyHashError                      = common.NewError("hash", EmptyHashErrorCode, "hash is empty")
	InvalidHashInputError               = common.NewError("hash", InvalidHashInputErrorCode, "invalid hash input value")
	HashAlgorithmAlreadyRegisteredError = common.NewError(
		"hash",
		HashAlgorithmAlreadyRegisteredErrorCode,
		"HashAlgorithm is already resitered in Hashes",
	)
	HashAlgorithmNotRegisteredError = common.NewError(
		"hash",
		HashAlgorithmNotRegisteredErrorCode,
		"HashAlgorithm is not resitered in Hashes",
	)
)
