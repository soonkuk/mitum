package common

const (
	NotFoundErrorCode ErrorCode = iota + 1
)

var (
	NotFoundError = NewError("common", NotFoundErrorCode, "not found")
)
