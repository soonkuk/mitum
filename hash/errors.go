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
	HashFailedError                     = common.NewErrorType("hash", HashFailedErrorCode, "failed to make hash")
	EmptyHashError                      = common.NewErrorType("hash", EmptyHashErrorCode, "hash is empty")
	InvalidHashInputError               = common.NewErrorType("hash", InvalidHashInputErrorCode, "invalid hash input value")
	HashAlgorithmAlreadyRegisteredError = common.NewErrorType(
		"hash",
		HashAlgorithmAlreadyRegisteredErrorCode,
		"HashAlgorithm is already resitered in Hashes",
	)
	HashAlgorithmNotRegisteredError = common.NewErrorType(
		"hash",
		HashAlgorithmNotRegisteredErrorCode,
		"HashAlgorithm is not resitered in Hashes",
	)
)
