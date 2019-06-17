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
	KeypairAlreadyRegisteredError = common.NewError(
		"keypair",
		KeypairAlreadyRegisteredErrorCode,
		"Keypair is already resitered in Keypairs",
	)
	KeypairNotRegisteredError = common.NewError(
		"keypair",
		KeypairNotRegisteredErrorCode,
		"Keypair is not resitered in Keypairs",
	)
	FailedToUnmarshalKeypairError = common.NewError(
		"keypair",
		FailedToUnmarshalKeypairErrorCode,
		"Failed to unmarshal keypair",
	)
	UnknownKeyKindError = common.NewError(
		"keypair",
		UnknownKeyKindErrorCode,
		"unknown key kind found",
	)
	SignatureVerificationFailedError = common.NewError(
		"keypair",
		SignatureVerificationFailedErrorCode,
		"signature verification failed",
	)
)
