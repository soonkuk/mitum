package encode

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

type Encoder interface {
	Type() EncoderType
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

type Encoders struct {
	sync.RWMutex
	encoders    map[ /*EncoderType*/ uint]Encoder
	defaultType EncoderType
}

func NewEncoders() *Encoders {
	return &Encoders{
		encoders: map[uint]Encoder{},
	}
}

func (e *Encoders) Register(encoder Encoder) error {
	e.Lock()
	defer e.Unlock()

	// check duplication
	for t := range e.encoders {
		if t == encoder.Type().ID() {
			return EncoderAlreadyRegisteredError.Newf("type=%q", encoder.Type().String())
		}
	}

	e.encoders[encoder.Type().ID()] = encoder

	if e.defaultType.Empty() {
		e.defaultType = encoder.Type()
	}

	return nil
}

func (e *Encoders) SetDefault(encoderType EncoderType) error {
	encoder, err := e.encoder(encoderType)
	if err != nil {
		return err
	}

	e.Lock()
	defer e.Unlock()

	e.defaultType = encoder.Type()

	return nil
}

func (e *Encoders) encoder(encoderType EncoderType) (Encoder, error) {
	e.RLock()
	defer e.RUnlock()

	encoder, found := e.encoders[encoderType.ID()]
	if !found {
		return nil, EncoderNotRegisteredError.Newf("type=%q", encoderType.String())
	}

	return encoder, nil
}

func (e *Encoders) Encode(i interface{}) ([]byte, error) {
	encoder, err := e.encoder(e.defaultType)
	if err != nil {
		return nil, err
	}

	return encoder.Encode(i)
}

func (e *Encoders) Decode(b []byte, i interface{}) error {
	a, o := common.ExtractBinary(b)
	if o < 0 {
		return DecodeFailedError.Newf("not enough data; length=%d", len(b))
	}

	var t EncoderType
	if err := t.UnmarshalBinary(a); err != nil {
		return DecodeFailedError.New(err)
	}

	decoder, err := e.encoder(t)
	if err != nil {
		return err
	}

	return decoder.Decode(b, i)
}
