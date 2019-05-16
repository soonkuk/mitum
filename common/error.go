package common

import (
	"fmt"
)

type Error struct {
	code    string
	message string
}

func (e Error) MarshalJSON() ([]byte, error) {
	return EncodeJSON(map[string]string{
		"code":    e.code,
		"message": e.message,
	}, false, false)
}

func (e Error) Error() string {
	b, _ := EncodeJSON(map[string]string{
		"code":    e.code,
		"message": e.message,
	}, false, false)

	return TerminalLogString(string(b))
}

func (e Error) Code() string {
	return e.code
}

func (e Error) Message() string {
	return e.message
}

// TODO remvoe SetMessage, AppendMessage will be used
func (e Error) SetMessage(format string, args ...interface{}) Error {
	return Error{code: e.code, message: fmt.Sprintf(format, args...)}
}

func (e Error) AppendMessage(format string, args ...interface{}) Error {
	return Error{
		code: e.code,
		message: fmt.Sprintf(
			"%s; %s",
			e.message,
			fmt.Sprintf(format, args...),
		),
	}
}

func (e Error) Equal(n error) bool {
	ne, found := n.(Error)
	if !found {
		return false
	}

	return e.Code() == ne.Code()
}

func NewError(name string, number uint, message string) Error {
	return Error{code: fmt.Sprintf("%s-%d", name, number), message: message}
}
