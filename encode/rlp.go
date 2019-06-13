package encode

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
)

var (
	RLPEncoderType EncoderType = NewEncoderType(1, "rlp")
)

type RLP struct {
}

func (r RLP) Type() EncoderType {
	return RLPEncoderType
}

func (r RLP) Encode(i interface{}) ([]byte, error) {
	t, err := RLPEncoderType.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encoded, err := rlp.EncodeToBytes(i)
	if err != nil {
		return nil, err
	}

	b := common.AppendBinary(t)
	b = append(b, common.AppendBinary(encoded)...)

	return b, nil
}

func (r RLP) Decode(b []byte, i interface{}) error {
	e, o := common.ExtractBinary(b)
	if o < 0 {
		return DecodeFailedError.Newf("not enough data; length=%d", len(b))
	}
	var t EncoderType
	if err := t.UnmarshalBinary(e); err != nil {
		return DecodeFailedError.New(err)
	} else if !t.Equal(RLPEncoderType) {
		return DecodeFailedError.Newf("not rlp encoded; type=%q", t.String())
	}

	e, o = common.ExtractBinary(b[o:])
	if o < 0 {
		return DecodeFailedError.Newf("not enough data; length=%d", len(b))
	}
	return rlp.DecodeBytes(e, i)
}
