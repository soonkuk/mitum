package hash

import (
	"encoding/json"

	"github.com/btcsuite/btcutil/base58"

	"github.com/spikeekips/mitum/common"
)

type Hash struct {
	algorithm HashAlgorithmType
	hint      string
	body      []byte
}

func NewHash(algorithmType HashAlgorithmType, hint string, body []byte) (Hash, error) {
	if len(hint) < 1 {
		return Hash{}, HashFailedError.Newf("zero hint length")
	}

	return Hash{
		algorithm: algorithmType,
		hint:      hint,
		body:      body,
	}, nil
}

func (h Hash) MarshalBinary() ([]byte, error) {
	var b []byte
	{
		algo, err := h.algorithm.MarshalBinary()
		if err != nil {
			return nil, err
		}
		b = append(b, common.AppendBinary([]byte(algo))...)
	}

	b = append(b, common.AppendBinary([]byte(h.hint))...)
	b = append(b, common.AppendBinary(h.body)...)

	return b, nil
}

func (h *Hash) UnmarshalBinary(b []byte) error {
	var algorithmType HashAlgorithmType
	var offset int
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

	var hint string
	{
		e, o := common.ExtractBinary(b[offset:])
		if o < 0 {
			return InvalidHashInputError.Newf("not enough to read hint length; length=%d", len(b))
		}
		hint = string(e)

		offset += o
	}

	var body []byte
	{
		e, o := common.ExtractBinary(b[offset:])
		if o < 0 {
			return InvalidHashInputError.Newf("not enough to read body length; length=%d", len(b))
		}
		body = e
	}

	h.algorithm = algorithmType
	h.hint = hint
	h.body = body

	return nil
}

func (h Hash) IsValid() error {
	if len(h.hint) < 1 {
		return EmptyHashError.Newf("empty hint; length=%d", len(h.hint))
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
	if len(h.body) != len(a.body) {
		return false
	}
	if !h.algorithm.Equal(a.Algorithm()) {
		return false
	}

	for i, b := range h.body {
		if b != a.body[i] {
			return false
		}
	}

	return true
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"algorithm": h.algorithm.String(),
		"hint":      h.hint,
		"body":      base58.Encode(h.body),
	})
}

func (h Hash) Algorithm() HashAlgorithmType {
	return h.algorithm
}

func (h Hash) Hint() string {
	return h.hint
}

func (h Hash) Body() []byte {
	return h.body
}

func (h Hash) String() string {
	b, _ := json.Marshal(h)
	return string(b)
}
