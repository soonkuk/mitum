package common

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
)

type testCustomSeal struct {
	RawSeal
	fieldA string
	fieldB string
	fieldC []byte
}

func (r testCustomSeal) Type() SealType {
	return SealType("showme-type")
}

func (r testCustomSeal) Hint() string {
	return "cs"
}

func (r testCustomSeal) SerializeRLP() ([]interface{}, error) {
	return []interface{}{r.fieldA, r.fieldB, r.fieldC}, nil
}

func (r *testCustomSeal) UnserializeRLP(m []rlp.RawValue) error {
	var fieldA string
	if err := Decode(m[6], &fieldA); err != nil {
		return err
	}
	var fieldB string
	if err := Decode(m[7], &fieldB); err != nil {
		return err
	}
	var fieldC []byte
	if err := Decode(m[8], &fieldC); err != nil {
		return err
	}

	r.fieldA = fieldA
	r.fieldB = fieldB
	r.fieldC = fieldC

	return nil
}

func (r testCustomSeal) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"field_a": r.fieldA,
		"field_b": r.fieldB,
		"field_c": r.fieldC,
	}, nil
}

func (r testCustomSeal) Wellformed() error {
	if err := r.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if len(r.fieldA) < 1 || len(r.fieldB) < 1 || len(r.fieldC) < 1 {
		return SealNotWellformedError
	}

	return nil
}

type testSealCodec struct {
	suite.Suite
}

func (t *testSealCodec) newCustomSeal() testCustomSeal {
	r := testCustomSeal{
		fieldA: RandomUUID(),
		fieldB: RandomUUID(),
		fieldC: []byte(RandomUUID()),
	}

	raw := NewRawSeal(
		r,
		CurrentSealVersion,
		r.Type(),
		r.Hint(),
	)
	r.RawSeal = raw

	return r
}

func (t *testSealCodec) TestNew() {
	sc := NewSealCodec()
	t.Equal(0, len(sc.Registered()))
}

func (t *testSealCodec) TestRegister() {
	sc := NewSealCodec()

	err := sc.Register(testCustomSeal{})
	t.NoError(err)
	t.Equal(testCustomSeal{}.Type(), sc.Registered()[0])
}

func (t *testSealCodec) TestEncode() {
	sc := NewSealCodec()

	_ = sc.Register(testCustomSeal{})

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

	_ = sc.Register(testCustomSeal{})

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
	t.NoError(decoded.(testCustomSeal).CheckSignature(TestNetworkID))

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

	_ = sc.Register(testCustomSeal{})

	r := t.newCustomSeal()

	seed := RandomSeed()
	_ = r.Sign(TestNetworkID, seed)

	b, _ := sc.Encode(r)
	decoded, _ := sc.Decode(b)

	unmarshaledSeal := decoded.(testCustomSeal)

	t.Nil(unmarshaledSeal.RawSeal.parent.(testCustomSeal).RawSeal.parent)
}

func TestSealCodec(t *testing.T) {
	suite.Run(t, new(testSealCodec))
}
