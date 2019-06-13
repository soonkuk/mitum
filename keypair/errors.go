package keypair

import "github.com/spikeekips/mitum/common"

const (
	KeypairAlreadyRegisteredErrorCode common.ErrorCode = iota + 1
	KeypairNotRegisteredErrorCode
	FailedToUnmarshalKeypairErrorCode
	UnknownKeyKindErrorCode
	SignatureVerificationFailedErrorCode
)

var (
	KeypairAlreadyRegisteredError = common.NewErrorType(
		"keypair",
		KeypairAlreadyRegisteredErrorCode,
		"Keypair is already resitered in Keypairs",
	)
	KeypairNotRegisteredError = common.NewErrorType(
		"keypair",
		KeypairNotRegisteredErrorCode,
		"Keypair is not resitered in Keypairs",
	)
	FailedToUnmarshalKeypairError = common.NewErrorType(
		"keypair",
		FailedToUnmarshalKeypairErrorCode,
		"Failed to unmarshal keypair",
	)
	UnknownKeyKindError = common.NewErrorType(
		"keypair",
		UnknownKeyKindErrorCode,
		"unknown key kind found",
	)
	SignatureVerificationFailedError = common.NewErrorType(
		"keypair",
		SignatureVerificationFailedErrorCode,
		"signature verification failed",
	)
)
