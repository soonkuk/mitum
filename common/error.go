package common

import (
	"fmt"
)

type Error struct {
	code    string
	message string // TODO should be slice
	err     error  // TODO should be slice to contain multiple error by called frame
}

func (e Error) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"code":    e.code,
		"message": e.message,
	}
	if e.err != nil {
		m["err"] = e.err
	}

	return EncodeJSON(m, false, false)
}

func (e Error) Error() string {
	b, _ := EncodeJSON(e, false, false)

	return TerminalLogString(string(b))
}

func (e Error) Code() string {
	return e.code
}

func (e Error) Message() string {
	return e.message
}

func (e Error) SetError(err error) Error {
	se, ok := err.(Error)
	if !ok {
		return e.SetMessage(err.Error())
	}

	if e.Equal(se) {
		return e.SetMessage(se.Message())
	}

	return Error{
		code:    e.code,
		message: e.message,
		err:     se,
	}
}

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
