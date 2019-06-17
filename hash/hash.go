package hash

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"

	"github.com/spikeekips/mitum/common"
)

var (
	zeroBody [100]byte = [100]byte{}
)

type Hash struct {
	algorithm common.DataType
	hint      string
	body      [100]byte // NOTE the fixed length array can be possible to make Hash to be comparable
	length    int
}

func NewHash(algorithmType common.DataType, hint string, body []byte) (Hash, error) {
	if len(hint) < 1 {
		return Hash{}, HashFailedError.Newf("zero hint length")
	}

	var b [100]byte
	copy(b[:], body)

	return Hash{
		algorithm: algorithmType,
		hint:      hint,
		body:      b,
		length:    len(body),
	}, nil
}

func (h Hash) MarshalBinary() ([]byte, error) {
	if h.Empty() {
		return nil, nil
	}

	var b []byte

	b = append(b, common.AppendBinary([]byte(h.hint))...)
	b = append(b, common.AppendBinary(h.Body())...)

	{
		algo, err := h.algorithm.MarshalBinary()
		if err != nil {
			return nil, err
		}
		b = append(b, common.AppendBinary([]byte(algo))...)
	}

	return b, nil
}

func (h *Hash) UnmarshalBinary(b []byte) error {
	var offset int

	var hint string
	{
		e, o := common.ExtractBinary(b[offset:])
		if o < 0 {
			return InvalidHashInputError.Newf("not enough to read hint length; length=%d", len(b))
		}
		hint = string(e)

		offset += o
	}

	var length int
	var body [100]byte
	{
		e, o := common.ExtractBinary(b[offset:])
		if o < 0 {
			return InvalidHashInputError.Newf("not enough to read body length; length=%d", len(b))
		}
		offset += o
		length = len(e)

		copy(body[:], e)
	}

	var algorithmType common.DataType
	{
		e, o := common.ExtractBinary(b[offset:])
		if o < 0 {
			return InvalidHashInputError.Newf("not enough to read algorithm length; length=%d", len(b))
		}
		if err := algorithmType.UnmarshalBinary(e); err != nil {
			return err
		}

		offset += o
	}

	h.algorithm = algorithmType
	h.hint = hint
	h.body = body
	h.length = length

	return nil
}

func (h Hash) Empty() bool {
	if len(h.hint) > 0 || h.body != zeroBody || !h.algorithm.Empty() {
		return false
	}

	return true
}

func (h Hash) IsValid() error {
	if h.length < 1 {
		return EmptyHashError.Newf("empty body")
	}

	if len(h.hint) < 1 {
		return EmptyHashError.Newf("empty hint")
	}

	if h.algorithm.Empty() {
		return EmptyHashError.Newf("empty algorithm")
	}

	return nil
}

func (h Hash) Equal(a Hash) bool {
	if h.hint != a.hint {
		return false
	}
	if h.body != a.body {
		return false
	}
	if !h.algorithm.Equal(a.Algorithm()) {
		return false
	}

	for i, b := range h.Body() {
		if b != a.body[i] {
			return false
		}
	}

	return true
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"hint":      h.hint,
		"body":      base58.Encode(h.Body()),
		"algorithm": h.algorithm.String(),
	})
}

func (h Hash) Algorithm() common.DataType {
	return h.algorithm
}

func (h Hash) Hint() string {
	return h.hint
}

func (h Hash) Body() []byte {
	return h.body[:h.length]
}

func (h Hash) Bytes() []byte {
	b, _ := h.MarshalBinary()
	return b
}

func (h Hash) String() string {
	return fmt.Sprintf("%s:%s:%s", h.hint, base58.Encode(h.Body()), h.algorithm.String())
}
