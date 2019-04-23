package element

import (
	"encoding"
	"encoding/json"

	"github.com/spikeekips/mitum/common"
)

type OperationType interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	json.Marshaler
	json.Unmarshaler
}

type OperationValue interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	json.Marshaler
	json.Unmarshaler
}

type OperationOptions interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	json.Marshaler
	json.Unmarshaler

	Get(string) interface{}
	Set(string) interface{}
}

type Operation interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	json.Marshaler
	json.Unmarshaler

	Type() OperationType
	Value() OperationValue
	Options() OperationOptions
	Target() common.Address
}
