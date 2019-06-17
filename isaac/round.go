package isaac

import "encoding/binary"

type Round uint64

func (r Round) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(r))
	return b, nil
}

func (r *Round) UnmarshalBinary(b []byte) error {
	u := binary.LittleEndian.Uint32(b)

	*r = Round(u)

	return nil
}
