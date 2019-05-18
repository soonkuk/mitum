package common

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
)

type testSeal struct {
	suite.Suite
}

func newRawSeal() RawSeal {
	return RawSeal{
		version:   CurrentSealVersion,
		sealType:  SealType("test-seal"),
		hash:      NewRandomHash("ts"),
		source:    RandomSeed().Address(),
		signature: Signature([]byte("test-sig")),
		signedAt:  Now(),
	}
}

func NewCustomSeal() CustomSeal {
	r := CustomSeal{
		RawSeal: newRawSeal(),
		fieldA:  RandomUUID(),
		fieldB:  RandomUUID(),
		fieldC:  []byte(RandomUUID()),
	}
	r.RawSeal.parent = r
	r.RawSeal.sealType = r.Type()
	r.RawSeal.hint = r.Hint()

	return r
}

func (t *testSeal) TestWellFormed() {
	cases := []struct {
		name    string
		getSeal func() RawSeal
		err     string
	}{
		{
			name:    "normal RawSeal",
			getSeal: newRawSeal,
		},
		{
			name: "empty version",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.version = Version{}
				return r
			},
			err: "zero version",
		},
		{
			name: "empty sealType",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.sealType = ""
				return r
			},
			err: "SealType",
		},
		{
			name: "empty hash",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.hash = Hash{}
				return r
			},
			err: "hash",
		},
		{
			name: "empty source",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.source = ""
				return r
			},
			err: "Address",
		},
		{
			name: "invalid source",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.source = "invalid source"
				return r
			},
			err: "illegal base32",
		},
		{
			name: "nil signature",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.signature = nil
				return r
			},
			err: "Signature",
		},
		{
			name: "zero signedAt",
			getSeal: func() RawSeal {
				r := newRawSeal()
				r.signedAt = Time{}
				return r
			},
			err: "signedAt",
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.Run(
			c.name,
			func() {
				err := c.getSeal().wellformed()
				if len(c.err) < 1 {
					t.NoError(err, "%d: %v", i, c.name)
				} else {
					t.Contains(err.Error(), c.err, "%d: %v", i, c.name)
				}
			},
		)
	}

}

func (t *testSeal) TestGenerateHash() {
	r := newRawSeal()
	_, err := r.GenerateHash()
	t.Contains(err.Error(), "parent is missing")
}

type CustomSeal struct {
	RawSeal
	fieldA string
	fieldB string
	fieldC []byte
}

func (r CustomSeal) Type() SealType {
	return SealType("custom-seal")
}

func (r CustomSeal) Hint() string {
	return "cs"
}

func (r CustomSeal) SerializeRLP() ([]interface{}, error) {
	return []interface{}{r.fieldA, r.fieldB, r.fieldC}, nil
}

func (r *CustomSeal) UnserializeRLP(m []rlp.RawValue) error {
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

func (r CustomSeal) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"field_a": r.fieldA,
		"field_b": r.fieldB,
		"field_c": r.fieldC,
	}, nil
}

func (r CustomSeal) Wellformed() error {
	if err := r.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if len(r.fieldA) < 1 || len(r.fieldB) < 1 || len(r.fieldC) < 1 {
		return SealNotWellformedError
	}

	return nil
}

func (t *testSeal) TestCustomSealNew() {
	r := NewCustomSeal()

	{ // check Seal interface{}
		_, ok := interface{}(r).(Seal)
		t.True(ok)
	}
}

func (t *testSeal) TestCustomSealMarshal() {
	defer DebugPanic()

	r := NewCustomSeal()
	seed := RandomSeed()
	err := r.Sign(TestNetworkID, seed)
	t.NoError(err)

	var b []byte
	{
		a, err := r.MarshalBinary()
		t.NoError(err)
		t.NotEmpty(a)

		b = a
	}

	var unmarshaled CustomSeal
	{
		decoded, err := DecodeSeal(CustomSeal{}, b)
		t.NoError(err)

		_, ok := decoded.(Seal)
		t.True(ok)

		unmarshaled = decoded.(CustomSeal)
	}

	t.Equal(r.version, unmarshaled.version)
	t.Equal(r.sealType, unmarshaled.sealType)
	t.True(r.hash.Equal(unmarshaled.hash))
	t.Equal(r.source, unmarshaled.source)
	t.Equal(r.signature, unmarshaled.signature)
	t.True(r.signedAt.Equal(unmarshaled.signedAt))
	t.Equal(r.fieldA, unmarshaled.fieldA)
	t.Equal(r.fieldB, unmarshaled.fieldB)
	t.Equal(r.fieldC, unmarshaled.fieldC)
}

func (t *testSeal) TestCustomSealMarshalAfterSign() {
	defer DebugPanic()

	r := NewCustomSeal()

	seed := RandomSeed()
	err := r.Sign(TestNetworkID, seed)
	t.NoError(err)

	signature, err := NewSignature(TestNetworkID, seed, r.hash)
	t.NoError(err)

	var unmarshaled CustomSeal
	{

		b, err := r.MarshalBinary()
		t.NoError(err)
		t.NotEmpty(b)

		decoded, err := DecodeSeal(CustomSeal{}, b)
		t.NoError(err)
		unmarshaled = decoded.(CustomSeal)

		err = unmarshaled.Wellformed()
		t.NoError(err)
	}

	t.Equal(r.version, unmarshaled.version)
	t.Equal(r.sealType, unmarshaled.sealType)
	t.True(r.hash.Equal(unmarshaled.hash))
	t.Equal(r.source, unmarshaled.source)
	t.Equal(r.signature, unmarshaled.signature)
	t.Equal(signature, unmarshaled.signature)
	t.True(r.signedAt.Equal(unmarshaled.signedAt))
	t.Equal(r.fieldA, unmarshaled.fieldA)
	t.Equal(r.fieldB, unmarshaled.fieldB)
	t.Equal(r.fieldC, unmarshaled.fieldC)
}

func (t *testSeal) TestCustomSealJSONMarshal() {
	r := NewCustomSeal()

	b, err := json.Marshal(r)
	t.NoError(err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	t.NoError(err)

	t.Equal(r.fieldA, m["field_a"])
	t.Equal(r.fieldB, m["field_b"])

	c, err := base64.StdEncoding.DecodeString(m["field_c"].(string))
	t.NoError(err)

	t.Equal(r.fieldC, c)
}

func (t *testSeal) TestCustomSealGenerateHash() {
	r := NewCustomSeal()

	h, err := r.GenerateHash()
	t.NoError(err)
	t.True(h.IsValid())

	t.Equal(r.Hint(), h.Hint())
}

func (t *testSeal) TestCustomSealSign() {
	r := NewCustomSeal()

	seed := RandomSeed()
	err := r.Sign(TestNetworkID, seed)
	t.NoError(err)

	signature, err := NewSignature(TestNetworkID, seed, r.hash)
	t.NoError(err)

	t.Equal(r.signature, signature)

	// check signature
	err = r.CheckSignature(TestNetworkID)
	t.NoError(err)
}

func TestSeal(t *testing.T) {
	suite.Run(t, new(testSeal))
}

type testSealedSeal struct {
	suite.Suite
}

func (t *testSealedSeal) TestNew() {
	seal := TestNewSeal{
		fieldA: RandomUUID(),
		fieldB: RandomUUID(),
		fieldC: []byte(RandomUUID()),
	}

	sealSeed := RandomSeed()
	{
		raw := NewRawSeal(seal, CurrentSealVersion)
		seal.RawSeal = raw

		err := seal.Sign(TestNetworkID, sealSeed)
		t.NoError(err)
	}

	sealed, err := NewSealedSeal(seal)
	t.NoError(err)

	sealedSeed := RandomSeed()
	err = sealed.Sign(TestNetworkID, sealedSeed)
	t.NoError(err)
	t.NoError(sealed.Wellformed())
	t.NoError(sealed.CheckSignature(TestNetworkID))
}

func (t *testSealedSeal) TestMarshal() {
	seal := TestNewSeal{
		fieldA: RandomUUID(),
		fieldB: RandomUUID(),
		fieldC: []byte(RandomUUID()),
	}

	sealSeed := RandomSeed()
	{
		raw := NewRawSeal(seal, CurrentSealVersion)
		seal.RawSeal = raw

		_ = seal.Sign(TestNetworkID, sealSeed)
	}

	sealed, _ := NewSealedSeal(seal)

	sealedSeed := RandomSeed()
	_ = sealed.Sign(TestNetworkID, sealedSeed)

	b, err := EncodeSeal(sealed) // encode
	t.NoError(err)

	decoded, err := DecodeSeal(SealedSeal{}, b)
	t.NoError(err)

	t.Equal(sealed.Version(), decoded.Version())
	t.Equal(SealedSealType, decoded.Type())
	t.Equal(sealed.Hint(), decoded.Hint())
	t.True(sealed.Hash().Equal(decoded.Hash()))
	t.Equal(sealed.Source(), decoded.Source())
	t.Equal(sealed.Signature(), decoded.Signature())
	t.True(sealed.SignedAt().Equal(decoded.SignedAt()))
}

func (t *testSealedSeal) TestUnmarshalInsideSeal() {
	seal := TestNewSeal{
		fieldA: RandomUUID(),
		fieldB: RandomUUID(),
		fieldC: []byte(RandomUUID()),
	}

	sealSeed := RandomSeed()
	{
		raw := NewRawSeal(seal, CurrentSealVersion)
		seal.RawSeal = raw

		_ = seal.Sign(TestNetworkID, sealSeed)
	}

	sealed, _ := NewSealedSeal(seal)

	sealedSeed := RandomSeed()
	_ = sealed.Sign(TestNetworkID, sealedSeed)

	b, err := EncodeSeal(sealed) // encode
	t.NoError(err)

	decoded, err := DecodeSeal(SealedSeal{}, b)
	t.NoError(err)

	decodedSeal := decoded.(SealedSeal)

	// to decode inside seal, SealType should be known
	insideSeal, err := DecodeSeal(TestNewSeal{}, decodedSeal.Binary())
	t.NoError(err)

	t.Equal(seal.Version(), insideSeal.Version())
	t.Equal(seal.Type(), insideSeal.Type())
	t.Equal(seal.Hint(), insideSeal.Hint())
	t.True(seal.Hash().Equal(insideSeal.Hash()))
	t.Equal(seal.Source(), insideSeal.Source())
	t.Equal(seal.Signature(), insideSeal.Signature())
	t.True(seal.SignedAt().Equal(insideSeal.SignedAt()))
}

func TestSealedSeal(t *testing.T) {
	suite.Run(t, new(testSealedSeal))
}
