package seal

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/keypair"
	"github.com/stretchr/testify/suite"
)

type testEncoder struct {
	suite.Suite
}

type tEncoder struct {
	v common.TypeData
}

func (t tEncoder) Type() common.DataType {
	return t.v.Type()
}

func (t tEncoder) Encode(v interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(v)
}

func (t tEncoder) Decode(b []byte) (interface{}, error) {
	var decoded RLPDecodeSeal
	if err := rlp.DecodeBytes(b, &decoded); err != nil {
		return nil, err
	}

	nv := reflect.New(reflect.TypeOf(t.v)).Interface()
	if err := rlp.DecodeBytes(decoded.Body, nv); err != nil {
		return nil, err
	}

	// check type
	seal := BaseSeal{
		t:      decoded.T,
		hash:   decoded.Hash,
		header: decoded.Header,
		body:   nv.(Body),
	}

	return seal, nil
}

func (t *testEncoder) TestEncode() {
	pk, _ := keypair.NewStellarPrivateKey()
	seal, err := newSealBodySigned(pk, "new", 33)
	t.NoError(err)

	err = seal.IsValid()
	t.NoError(err)

	b, err := rlp.EncodeToBytes(seal)
	t.NoError(err)

	encoder := tEncoder{v: tSealBody{T: seal.Body().Type()}}

	decoded, err := encoder.Decode(b)
	t.NoError(err)
	t.NotEmpty(decoded)

	decodedSeal, ok := decoded.(Seal)
	t.True(ok)

	t.True(seal.Equal(decodedSeal))
}

func (t *testEncoder) TestEncoders() {
	pk, _ := keypair.NewStellarPrivateKey()
	seal, err := newSealBodySigned(pk, "new", 33)
	t.NoError(err)

	err = seal.IsValid()
	t.NoError(err)

	b, err := rlp.EncodeToBytes(seal)
	t.NoError(err)

	encoders := NewEncoders()
	err = encoders.Register(tEncoder{v: tSealBody{T: seal.Body().Type()}})
	t.NoError(err)

	decodedSeal, err := encoders.Decode(b)
	t.NoError(err)
	t.NotEmpty(decodedSeal)

	t.True(seal.Equal(decodedSeal))
}

func TestEncoder(t *testing.T) {
	suite.Run(t, new(testEncoder))
}
