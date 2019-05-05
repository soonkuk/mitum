package common

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
)

type testSealV1 struct {
	suite.Suite
}

func (t *testSealV1) newRawSeal() RawSeal {
	return RawSeal{
		version:   CurrentSealVersion,
		sealType:  SealType("test-seal"),
		hash:      NewRandomHash("ts"),
		source:    RandomSeed().Address(),
		signature: Signature([]byte("test-sig")),
		signedAt:  Now(),
	}
}

func (t *testSealV1) TestWellFormed() {
	cases := []struct {
		name    string
		getSeal func() RawSeal
		err     string
	}{
		{
			name:    "normal RawSeal",
			getSeal: t.newRawSeal,
		},
		{
			name: "empty version",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
				r.version = Version{}
				return r
			},
			err: "zero version",
		},
		{
			name: "empty sealType",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
				r.sealType = ""
				return r
			},
			err: "SealType",
		},
		{
			name: "empty hash",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
				r.hash = Hash{}
				return r
			},
			err: "hash",
		},
		{
			name: "empty source",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
				r.source = ""
				return r
			},
			err: "Address",
		},
		{
			name: "invalid source",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
				r.source = "invalid source"
				return r
			},
			err: "illegal base32",
		},
		{
			name: "nil signature",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
				r.signature = nil
				return r
			},
			err: "Signature",
		},
		{
			name: "zero signedAt",
			getSeal: func() RawSeal {
				r := t.newRawSeal()
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
				err := c.getSeal().Wellformed()
				if len(c.err) < 1 {
					t.NoError(err, "%d: %v", i, c.name)
				} else {
					t.Contains(err.Error(), c.err, "%d: %v", i, c.name)
				}
			},
		)
	}

}

func (t *testSealV1) TestGenerateHash() {
	r := t.newRawSeal()
	_, err := r.GenerateHash()
	t.Contains(err.Error(), "parent is missing")
}

type CustomSeal struct {
	RawSeal
	fieldA string
	fieldB string
	fieldC []byte
}

func (r CustomSeal) Hint() string {
	return "cs"
}

func (r CustomSeal) SerializeRLP() ([]interface{}, error) {
	s, err := r.RawSeal.SerializeRLPInside()
	if err != nil {
		return nil, err
	}

	return append(s, r.fieldA, r.fieldB, r.fieldC), nil
}

func (r *CustomSeal) UnserializeRLP(m []rlp.RawValue) error {
	if err := r.RawSeal.UnserializeRLPInside(m); err != nil {
		return err
	}

	*r = CustomSeal{RawSeal: r.RawSeal}

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

func (r *CustomSeal) UnmarshalBinary(b []byte) error {
	r.RawSeal.parent = r

	return r.RawSeal.UnmarshalBinaryInside(b)
}

func (r CustomSeal) SerializeMap() (map[string]interface{}, error) {
	m, err := r.RawSeal.SerializeMap()
	if err != nil {
		return nil, err
	}

	m["field_a"] = r.fieldA
	m["field_b"] = r.fieldB
	m["field_c"] = r.fieldC

	return m, nil
}

func (t *testSealV1) newCustomSeal() CustomSeal {
	r := CustomSeal{
		RawSeal: t.newRawSeal(),
		fieldA:  RandomUUID(),
		fieldB:  RandomUUID(),
		fieldC:  []byte(RandomUUID()),
	}
	r.RawSeal.parent = r

	return r
}

func (t *testSealV1) TestCustomSealNew() {
	r := t.newCustomSeal()

	{ // check Seal interface{}
		_, ok := interface{}(r).(SealV1)
		t.True(ok)
	}
}

func (t *testSealV1) TestCustomSealMarshal() {
	defer DebugPanic()

	r := t.newCustomSeal()

	var b []byte
	{
		a, err := r.MarshalBinary()
		t.NoError(err)
		t.NotEmpty(a)

		b = a
	}

	var unmarshaled CustomSeal
	{
		err := unmarshaled.UnmarshalBinary(b)
		t.NoError(err)
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

func (t *testSealV1) TestCustomSealJSONMarshal() {
	r := t.newCustomSeal()

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

func (t *testSealV1) TestCustomSealGenerateHash() {
	r := t.newCustomSeal()

	h, err := r.GenerateHash()
	t.NoError(err)
	t.False(h.Empty())

	t.Equal(r.Hint(), h.Hint())
}

func (t *testSealV1) TestCustomSealSign() {
	r := t.newCustomSeal()

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

func TestSealV1(t *testing.T) {
	suite.Run(t, new(testSealV1))
}
