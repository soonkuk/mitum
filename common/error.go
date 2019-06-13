package common

import (
	"encoding/json"
	"fmt"

	"golang.org/x/xerrors"
)

type ErrorCode uint

type Error struct {
	id      string
	message string
	err     error
	frame   xerrors.Frame
}

func NewErrorType(errorType string, errorID ErrorCode, message string) Error {
	return Error{
		id:      fmt.Sprintf("%s-%d", errorType, errorID),
		message: message,
	}
}

func (e Error) New(err error) Error {
	return Error{
		id:      e.id,
		message: e.message,
		err:     err,
		frame:   xerrors.Caller(1),
	}
}

func (e Error) Newf(s string, args ...interface{}) Error {
	return Error{
		id:      e.id,
		message: e.message,
		err:     fmt.Errorf(s, args...),
		frame:   xerrors.Caller(1),
	}
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) FormatError(p xerrors.Printer) error {
	e.frame.Format(p)
	return e.Unwrap()
}

func (e Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e Error) MarshalJSON() ([]byte, error) {
	var em interface{} = e.err
	if _, ok := e.err.(json.Marshaler); !ok {
		em = e.err.Error()
	}

	return json.Marshal(map[string]interface{}{
		"id":      e.id,
		"message": e.message,
		"error":   em,
	})
}

func (e Error) Is(err error) bool {
	var ae Error
	if !xerrors.As(err, &ae) {
		return false
	}

	return e.id == ae.id
}
