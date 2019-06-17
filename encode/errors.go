package encode

import "github.com/spikeekips/mitum/common"

const (
	EncoderAlreadyRegisteredErrorCode common.ErrorCode = iota + 1
	EncoderNotRegisteredErrorCode
	DecodeFailedErrorCode
)

var (
	EncoderAlreadyRegisteredError = common.NewError(
		"encode",
		EncoderAlreadyRegisteredErrorCode,
		"Encoder is already resitered in Encoders",
	)
	EncoderNotRegisteredError = common.NewError(
		"encode",
		EncoderNotRegisteredErrorCode,
		"Encoder is not resitered in Encoders",
	)
	DecodeFailedError = common.NewError("encode", DecodeFailedErrorCode, "failed to decode")
)
