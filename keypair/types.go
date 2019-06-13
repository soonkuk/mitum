package keypair

import (
	"encoding/binary"
	"encoding/json"
)

type Kind uint

const (
	PublicKeyKind Kind = iota + 1
	PrivateKeyKind
)

func (k Kind) String() string {
	switch k {
	case PublicKeyKind:
		return "public"
	case PrivateKeyKind:
		return "private"
	}

	return ""
}

func (k Kind) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(k))

	return b, nil
}

func (k *Kind) UnmarshalBinary(b []byte) error {
	*k = Kind(uint(binary.LittleEndian.Uint32(b)))

	return nil
}

func (k *Kind) UnmarshalText(b []byte) error {
	switch string(b) {
	case PublicKeyKind.String():
		*k = PublicKeyKind
		return nil
	case PrivateKeyKind.String():
		*k = PrivateKeyKind
		return nil
	default:
		return UnknownKeyKindError
	}
}

func (k Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

type Type struct {
	id   uint
	name string
}

func NewType(id uint, name string) Type {
	return Type{id: id, name: name}
}

func (k Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k Type) ID() uint {
	return k.id
}

func (k Type) Name() string {
	return k.name
}

func (k Type) Equal(b Type) bool {
	return k.id == b.id
}

func (k Type) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(k.id))

	return b, nil
}

func (k *Type) UnmarshalBinary(b []byte) error {
	k.id = uint(binary.LittleEndian.Uint32(b))

	return nil
}

func (k Type) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *Type) UnmarshalText(b []byte) error {
	k.name = string(b)
	return nil
}

func (k Type) Empty() bool {
	return k.id < 1
}

func (k Type) String() string {
	return k.name
}
