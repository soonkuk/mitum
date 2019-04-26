package common

const (
	_ uint = iota
	OverflowErrorCode
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
)

var (
	OverflowError                    Error = NewError("common", OverflowErrorCode, "overflow number")
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
)
