package common

const (
	_ uint = iota
	OverflowErrorCode
	SomethingWrongErrorCode
	UnknownSealTypeErrorCode
	JSONUnmarshalErrorCode
	NotImplementedErrorCode
	InvalidHashErrorCode
	EmptyHashErrorCode
	InvalidHashHintErrorCode
	InvalidBigStringErrorCode
	HashDoesNotMatchErrorCode
	SealNotWellformedErrorCode
	ContextValueNotFoundErrorCode
	SignatureVerificationFailedErrorCode
	InvalidAddressErrorCode
	InvalidNetAddrErrorCode
	StartStopperAlreadyStartedErrorCode
	InvalidSealTypeErrorCode
	InvalidSignatureErrorCode
	InvalidSeedErrorCode
)

var (
	OverflowError                    Error = NewError("common", OverflowErrorCode, "overflow number")
	SomethingWrongError              Error = NewError("common", SomethingWrongErrorCode, "something wrong")
	UnknownSealTypeError             Error = NewError("common", UnknownSealTypeErrorCode, "unknown seal type found")
	JSONUnmarshalError               Error = NewError("common", JSONUnmarshalErrorCode, "failed json unmarshal")
	NotImplementedError              Error = NewError("common", NotImplementedErrorCode, "not implemented")
	InvalidHashError                 Error = NewError("common", InvalidHashErrorCode, "invalid hash found")
	EmptyHashError                   Error = NewError("common", EmptyHashErrorCode, "hash is empty")
	InvalidHashHintError             Error = NewError("common", InvalidHashHintErrorCode, "invalid hash hint")
	InvalidBigStringError            Error = NewError("common", InvalidBigStringErrorCode, "invalid big string")
	HashDoesNotMatchError            Error = NewError("common", HashDoesNotMatchErrorCode, "hash does not match")
	SealNotWellformedError           Error = NewError("common", SealNotWellformedErrorCode, "seal is not wellformed")
	ContextValueNotFoundError        Error = NewError("common", ContextValueNotFoundErrorCode, "value not found in context")
	SignatureVerificationFailedError Error = NewError("common", SignatureVerificationFailedErrorCode, "signature verification failed")
	InvalidAddressError              Error = NewError("common", InvalidAddressErrorCode, "invalid Address")
	InvalidNetAddrError              Error = NewError("common", InvalidNetAddrErrorCode, "invalid NetAddr")
	StartStopperAlreadyStartedError  Error = NewError("common", StartStopperAlreadyStartedErrorCode, "StartStopper already started")
	InvalidSealTypeError             Error = NewError("isaac", InvalidSealTypeErrorCode, "invalid SealType")
	InvalidSignatureError            Error = NewError("isaac", InvalidSignatureErrorCode, "invalid Signature found")
	InvalidSeedError                 Error = NewError("isaac", InvalidSeedErrorCode, "invalid seed")
)
