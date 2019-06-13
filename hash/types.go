package hash

import (
	"encoding/binary"
	"encoding/json"
)

type HashAlgorithmType struct {
	id   uint
	name string
}

func NewHashAlgorithmType(id uint, name string) HashAlgorithmType {
	return HashAlgorithmType{id: id, name: name}
}

func (h HashAlgorithmType) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h HashAlgorithmType) ID() uint {
	return h.id
}

func (h HashAlgorithmType) Equal(b HashAlgorithmType) bool {
	return h.id == b.id
}

func (h HashAlgorithmType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(h.id))

	return b, nil
}

func (h *HashAlgorithmType) UnmarshalBinary(b []byte) error {
	h.id = uint(binary.LittleEndian.Uint32(b))

	return nil
}

func (h HashAlgorithmType) Empty() bool {
	return h.id < 1
}

func (h HashAlgorithmType) String() string {
	return h.name
}
