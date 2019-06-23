package common

import (
	"sync"
)

const (
	EndoderAlreadyRegisteredErrorCode ErrorCode = iota + 1
	EncoderNotRegisteredErrorCode
)

var (
	EndoderAlreadyRegisteredError = NewError("encoders", EndoderAlreadyRegisteredErrorCode, "encoder already registered in Encoders")
	EncoderNotRegisteredError     = NewError("encoders", EncoderNotRegisteredErrorCode, "encoder not registered in Encoders")
)

type TypeData interface {
	Type() DataType
}

type Encoder interface {
	Type() DataType
	Encode(interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
}

type Encoders struct {
	sync.RWMutex
	encoders       map[DataType]Encoder
	defaultEncoder DataType
}

func NewEncoders() *Encoders {
	return &Encoders{
		encoders: map[DataType]Encoder{},
	}
}

func (m *Encoders) Register(encoder Encoder) error {
	m.Lock()
	defer m.Unlock()

	_, found := m.encoders[encoder.Type()]
	if found {
		return EndoderAlreadyRegisteredError.Newf("type=%q", encoder.Type())
	}

	m.encoders[encoder.Type()] = encoder

	if m.defaultEncoder.Empty() {
		m.defaultEncoder = encoder.Type()
	}

	return nil
}

func (m *Encoders) SetDefault(encoderType DataType) error {
	m.Lock()
	defer m.Unlock()

	if _, err := m.Encoder(encoderType); err != nil {
		return err
	}

	m.defaultEncoder = encoderType

	return nil
}

func (m *Encoders) Encoder(encoderType DataType) (Encoder, error) {
	m.RLock()
	defer m.RUnlock()

	h, found := m.encoders[encoderType]
	if !found {
		return nil, EncoderNotRegisteredError
	}

	return h, nil
}

func (m *Encoders) Encode(v interface{}) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	var encoder Encoder
	var err error
	if td, ok := v.(TypeData); ok {
		encoder, err = m.Encoder(td.Type())
	} else {
		encoder, err = m.Encoder(m.defaultEncoder)
	}
	if err != nil {
		return nil, err
	}

	return encoder.Encode(v)
}

func (m *Encoders) EncodeByType(encoderType DataType, v interface{}) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	encoder, err := m.Encoder(encoderType)
	if err != nil {
		return nil, err
	}

	return encoder.Encode(v)
}

func (m *Encoders) Decode(b []byte) (interface{}, error) {
	m.RLock()
	defer m.RUnlock()

	encoder, err := m.Encoder(m.defaultEncoder)
	if err != nil {
		return nil, err
	}

	return encoder.Decode(b)
}

func (m *Encoders) DecodeByType(encoderType DataType, b []byte) (interface{}, error) {
	m.RLock()
	defer m.RUnlock()

	encoder, err := m.Encoder(encoderType)
	if err != nil {
		return nil, err
	}

	return encoder.Decode(b)
}
