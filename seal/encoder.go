package seal

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"golang.org/x/xerrors"
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

	raw, err := m.DecodeByType(decoded.T, b)
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
	T      common.DataType
	Hash   rlp.RawValue
	Header rlp.RawValue
	Body   rlp.RawValue
}

type RLPDecodeSeal struct {
	T      common.DataType
	Hash   hash.Hash
	Header Header
	Body   rlp.RawValue
}
