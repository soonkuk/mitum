package encode

import "github.com/spikeekips/mitum/common"

const (
	EncoderAlreadyRegisteredErrorCode common.ErrorCode = iota + 1
	EncoderNotRegisteredErrorCode
	DecodeFailedErrorCode
)

var (
	EncoderAlreadyRegisteredError = common.NewErrorType(
		"encode",
		EncoderAlreadyRegisteredErrorCode,
		"Encoder is already resitered in Encoders",
	)
	EncoderNotRegisteredError = common.NewErrorType(
		"encode",
		EncoderNotRegisteredErrorCode,
		"Encoder is not resitered in Encoders",
	)
	DecodeFailedError = common.NewErrorType("encode", DecodeFailedErrorCode, "failed to decode")
)
