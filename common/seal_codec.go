package common

import (
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type SealCodec struct {
	sync.RWMutex
	types map[SealType]reflect.Type
}

func NewSealCodec() *SealCodec {
	return &SealCodec{
		types: map[SealType]reflect.Type{},
	}
}

func (s *SealCodec) Registered() []SealType {
	s.RLock()
	defer s.RUnlock()

	var types []SealType
	for t := range s.types {
		types = append(types, t)
	}

	return types
}

func (s *SealCodec) Register(seal Seal) error {
	s.Lock()
	defer s.Unlock()

	rt := reflect.TypeOf(seal)

	// check RLPUnserializer
	ptrSeal := reflect.New(rt)
	if _, ok := ptrSeal.Interface().(RLPUnserializer); !ok {
		return UnknownSealTypeError.SetMessage("not RLPUnserializer")
	}

	s.types[seal.Type()] = rt

	return nil
}

func (s *SealCodec) Encode(seal Seal) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()

	if err := seal.Wellformed(); err != nil {
		return nil, err
	}

	return EncodeSeal(seal)
}

func (s *SealCodec) Decode(b []byte) (Seal, error) {
	s.RLock()
	defer s.RUnlock()

	// get types
	var rawValues []rlp.RawValue
	if err := Decode(b, &rawValues); err != nil {
		return nil, err
	}

	var sealType SealType
	if err := Decode(rawValues[3], &sealType); err != nil {
		return nil, err
	}

	// check registered
	rt, ok := s.types[sealType]
	if !ok {
		return nil, UnknownSealTypeError
	}

	return decodeSeal(rt, b, rawValues)
}

func EncodeSeal(seal Seal) ([]byte, error) {
	return seal.MarshalBinary()
}

func DecodeSeal(sealStruct interface{}, b []byte) (Seal, error) {
	rt := reflect.TypeOf(sealStruct)

	ptrSeal := reflect.New(rt)
	if _, ok := ptrSeal.Interface().(RLPUnserializer); !ok {
		return nil, UnknownSealTypeError.SetMessage("not RLPUnserializer")
	}

	return decodeSeal(rt, b, nil)
}

func decodeSeal(rt reflect.Type, b []byte, rawValues []rlp.RawValue) (Seal, error) {
	// get types
	if rawValues == nil {
		if err := Decode(b, &rawValues); err != nil {
			return nil, err
		}
	}

	var sealType SealType
	if err := Decode(rawValues[3], &sealType); err != nil {
		return nil, err
	}

	ptrSeal := reflect.New(rt)

	raw := reflect.New(reflect.TypeOf(RawSeal{})).Interface().(*RawSeal)
	raw.parent = ptrSeal.Interface().(Seal)

	if err := raw.UnserializeRLP(rawValues); err != nil {
		return nil, err
	}

	// NOTE remove parent pointer from nested RawSeal
	parent := reflect.ValueOf(raw.parent).Elem()
	nestedRaw := parent.FieldByName("RawSeal").Interface().(RawSeal)
	nestedRaw.parent = nil
	parent.FieldByName("RawSeal").Set(reflect.ValueOf(nestedRaw))

	raw.parent = parent.Interface().(Seal)

	ptrSeal.Elem().FieldByName("RawSeal").Set(reflect.ValueOf(raw).Elem())

	return ptrSeal.Elem().Interface().(Seal), nil
}
