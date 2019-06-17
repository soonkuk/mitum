package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

type DataType struct {
	id   uint
	name string
}

func NewDataType(id uint, name string) DataType {
	if id < 1 {
		panic(fmt.Errorf("DataType.id should be greater than 0"))
	}

	return DataType{id: id, name: name}
}

func (i DataType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.name)
}

func (i DataType) ID() uint {
	return i.id
}

func (i DataType) Name() string {
	return i.name
}

func (i DataType) Equal(b DataType) bool {
	return i.id == b.id
}

func (i DataType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(i.id))

	return b, nil
}

func (i *DataType) UnmarshalBinary(b []byte) error {
	i.id = uint(binary.LittleEndian.Uint32(b))

	return nil
}

func (i DataType) MarshalText() ([]byte, error) {
	return json.Marshal(i)
}

func (i *DataType) UnmarshalText(b []byte) error {
	i.name = string(b)
	return nil
}

func (i DataType) Empty() bool {
	return i.id < 1
}

func (i DataType) String() string {
	return i.name
}
