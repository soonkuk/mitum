package common

// TODO rename XXXCode  to XXXErrorCode
const (
	_ uint = iota
	OverflowErrorCode
	SomethingWrongCode
	UnknownSealTypeCode
	JSONUnmarshalCode
	NotImplementedCode
	InvalidHashCode
	EmptyHashCode
	InvalidHashHintCode
	InvalidBigStringCode
	HashDoesNotMatchCode
	SealNotWellformedCode
	ContextValueNotFoundCode
	SignatureVerificationFailedCode
	InvalidAddressCode
	InvalidNetAddrCode
	StartStopperAlreadyStartedCode
	InvalidSealTypeCode
	InvalidSignatureCode
)

var (
	OverflowError                    Error = NewError("common", OverflowErrorCode, "overflow number")
	SomethingWrongError              Error = NewError("common", SomethingWrongCode, "something wrong")
	UnknownSealTypeError             Error = NewError("common", UnknownSealTypeCode, "unknown seal type found")
	JSONUnmarshalError               Error = NewError("common", JSONUnmarshalCode, "failed json unmarshal")
	NotImplementedError              Error = NewError("common", NotImplementedCode, "not implemented")
	InvalidHashError                 Error = NewError("common", InvalidHashCode, "invalid hash found")
	EmptyHashError                   Error = NewError("common", EmptyHashCode, "hash is empty")
	InvalidHashHintError             Error = NewError("common", InvalidHashHintCode, "invalid hash hint")
	InvalidBigStringError            Error = NewError("common", InvalidBigStringCode, "invalid big string")
	HashDoesNotMatchError            Error = NewError("common", HashDoesNotMatchCode, "hash does not match")
	SealNotWellformedError           Error = NewError("common", SealNotWellformedCode, "seal is not wellformed")
	ContextValueNotFoundError        Error = NewError("common", ContextValueNotFoundCode, "value not found in context")
	SignatureVerificationFailedError Error = NewError("common", SignatureVerificationFailedCode, "signature verification failed")
	InvalidAddressError              Error = NewError("common", InvalidAddressCode, "invalid Address")
	InvalidNetAddrError              Error = NewError("common", InvalidNetAddrCode, "invalid NetAddr")
	StartStopperAlreadyStartedError  Error = NewError("common", StartStopperAlreadyStartedCode, "StartStopper already started")
	InvalidSealTypeError             Error = NewError("isaac", InvalidSealTypeCode, "invalid SealType")
	InvalidSignatureError            Error = NewError("isaac", InvalidSignatureCode, "invalid Signature found")
)
