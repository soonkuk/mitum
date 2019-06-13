package encode

import (
	"encoding/binary"
	"encoding/json"
)

type EncoderType struct {
	id   uint
	name string
}

func NewEncoderType(id uint, name string) EncoderType {
	return EncoderType{id: id, name: name}
}

func (h EncoderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h EncoderType) ID() uint {
	return h.id
}

func (h EncoderType) Equal(b EncoderType) bool {
	return h.id == b.id
}

func (h EncoderType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(h.id))

	return b, nil
}

func (h *EncoderType) UnmarshalBinary(b []byte) error {
	h.id = uint(binary.LittleEndian.Uint32(b))

	return nil
}

func (h EncoderType) Empty() bool {
	return h.id < 1
}

func (h EncoderType) String() string {
	return h.name
}
