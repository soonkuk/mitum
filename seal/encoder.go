package seal

import (
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
)

type Encoders struct {
	*common.Encoders
}

func NewEncoders() *Encoders {
	return &Encoders{Encoders: common.NewEncoders()}
}

func (m *Encoders) Decode(b []byte) (Seal, error) {
	m.RLock()
	defer m.RUnlock()

	var decoded RLPDecodeSealType
	if err := rlp.DecodeBytes(b, &decoded); err != nil {
		return nil, err
	}

	raw, err := m.DecodeByType(decoded.Type, b)
	if err != nil {
		return nil, err
	}

	seal, ok := raw.(Seal)
	if !ok {
		return nil, xerrors.Errorf("is not Seal")
	}

	return seal, nil
}

type RLPDecodeSealType struct {
	Type   common.DataType
	Hash   rlp.RawValue
	Header rlp.RawValue
	Body   rlp.RawValue
}

type RLPEncodeSeal struct {
	Type   common.DataType
	Hash   hash.Hash
	Header Header
	Body   Body
}

type RLPDecodeSeal struct {
	Type   common.DataType
	Hash   hash.Hash
	Header Header
	Body   rlp.RawValue
}
