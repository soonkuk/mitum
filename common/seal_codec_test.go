package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testSealCodec struct {
	suite.Suite
}

func (t *testSealCodec) newCustomSeal() TestNewSeal {
	r := TestNewSeal{
		fieldA: RandomUUID(),
		fieldB: RandomUUID(),
		fieldC: []byte(RandomUUID()),
	}

	raw := NewRawSeal(r, CurrentSealVersion)
	r.RawSeal = raw

	return r
}

func (t *testSealCodec) TestNew() {
	sc := NewSealCodec()
	t.Equal(0, len(sc.Registered()))
}

func (t *testSealCodec) TestRegister() {
	sc := NewSealCodec()

	err := sc.Register(TestNewSeal{})
	t.NoError(err)
	t.Equal(TestNewSeal{}.Type(), sc.Registered()[0])
}

func (t *testSealCodec) TestEncode() {
	sc := NewSealCodec()

	_ = sc.Register(TestNewSeal{})

	r := t.newCustomSeal()

	seed := RandomSeed()
	err := r.Sign(TestNetworkID, seed)
	t.NoError(err)

	b, err := sc.Encode(r)
	t.NoError(err)
	t.NotEmpty(b)
}

func (t *testSealCodec) TestDecode() {
	sc := NewSealCodec()

	_ = sc.Register(TestNewSeal{})

	r := t.newCustomSeal()
	t.Error(r.Wellformed())

	seed := RandomSeed()
	err := r.Sign(TestNetworkID, seed)
	t.NoError(err)

	// encode
	b, _ := sc.Encode(r)

	// decode
	decoded, err := sc.Decode(b)
	t.NoError(err)
	return
	t.NotNil(decoded)
	t.NoError(decoded.Wellformed())

	// check signature
	t.NoError(decoded.(TestNewSeal).CheckSignature(TestNetworkID))

	t.Equal(r.Version(), decoded.Version())
	t.Equal(r.Type(), decoded.Type())
	t.Equal(r.Hint(), decoded.Hint())
	t.True(r.Hash().Equal(decoded.Hash()))
	t.Equal(r.Source(), decoded.Source())
	t.Equal(r.Signature(), decoded.Signature())
	t.True(r.SignedAt().Equal(decoded.SignedAt()))

	expectedHash, err := r.GenerateHash()
	t.NoError(err)
	decodedHash, err := r.GenerateHash()
	t.NoError(err)
	t.True(expectedHash.Equal(decodedHash))
	t.Equal(r.String(), decoded.String())
}

func (t *testSealCodec) TestDecodeNestedParentNil() {
	sc := NewSealCodec()

	_ = sc.Register(TestNewSeal{})

	r := t.newCustomSeal()

	seed := RandomSeed()
	_ = r.Sign(TestNetworkID, seed)

	b, _ := sc.Encode(r)
	decoded, _ := sc.Decode(b)

	unmarshaledSeal := decoded.(TestNewSeal)

	t.Nil(unmarshaledSeal.RawSeal.parent.(TestNewSeal).RawSeal.parent)
}

func TestSealCodec(t *testing.T) {
	suite.Run(t, new(testSealCodec))
}
