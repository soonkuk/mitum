package common

import "encoding/binary"

func AppendBinary(b []byte) []byte {
	var t []byte

	l := make([]byte, 4)
	binary.LittleEndian.PutUint32(l, uint32(len(b)))
	t = append(t, l...)
	t = append(t, b...)

	return t
}

func ExtractBinary(b []byte) ([]byte, int) {
	if len(b) < 4 {
		return nil, -1
	}

	l := int(binary.LittleEndian.Uint32(b[:4]))
	if len(b) < 4+l {
		return nil, -1
	}

	return b[4 : 4+l], 4 + l
}
