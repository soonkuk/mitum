package big

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testBig struct {
	suite.Suite
}

func (t *testBig) TestAdd() {
	{ // big.Int: add overflow, but ok
		var a, b, c big.Int
		a.SetUint64(10)
		b.SetUint64(math.MaxUint64)

		t.Equal(1, c.Add(&a, &b).Cmp(&b))
	}

	{
		b := NewBig(math.MaxUint64)

		c, ok := b.AddOK(10)
		t.True(ok)

		t.Equal("10", c.Int.Sub(&c.Int, &b.Int).String())
	}

	{
		a := NewBig(math.MaxUint64)
		b := NewBig(math.MaxUint64)

		c, ok := a.AddOK(b)
		t.True(ok)
		t.Equal("36893488147419103230", c.Int.String())

		t.Equal(a, c.Sub(b))
		t.True(a.Equal(b))
		t.Equal("18446744073709551615", b.Int.String())
	}
}

func (t *testBig) TestSub() {
	{
		a := NewBig(10)
		b := NewBig(math.MaxUint64)

		c, ok := b.SubOK(a)
		t.True(ok)
		t.Equal("18446744073709551605", c.Int.String())
	}

	{
		a := NewBig(10)
		b := NewBig(math.MaxUint64)

		c, ok := a.SubOK(b)
		t.False(ok)
		t.Equal("0", c.Int.String())
	}

	{
		a := NewBig(math.MaxUint64)
		b := NewBig(math.MaxUint64)
		c, _ := a.AddOK(b)

		d, ok := c.SubOK(a)
		t.True(ok)
		t.Equal("18446744073709551615", d.Int.String())
	}
}

func (t *testBig) TestMul() {
	{
		b := NewBig(math.MaxUint64)

		c, ok := b.MulOK(10)
		t.True(ok)
		t.Equal("184467440737095516150", c.Int.String())
	}

	{
		a := NewBig(math.MaxUint64)
		b := NewBig(math.MaxUint64)
		c, _ := a.AddOK(b)

		d, ok := c.MulOK(a)
		t.True(ok)
		t.Equal("680564733841876926852962238568698216450", d.Int.String())
	}
}

func (t *testBig) TestDiv() {
	{
		a := NewBig(10)
		b := NewBig(math.MaxUint64)

		c, ok := b.DivOK(a)
		t.True(ok)
		t.Equal("1844674407370955161", c.Int.String())
	}

	{ // divizion zero
		a := NewBig(10)
		b := NewBig(0)

		c, ok := b.DivOK(a)
		t.True(ok)
		t.Equal("0", c.Int.String())
	}

	{ // zero divizion
		a := NewBig(10)
		b := NewBig(0)

		c, ok := a.DivOK(b)
		t.False(ok)
		t.Equal("0", c.Int.String())
	}
}

func (t *testBig) TestMarshalBinary() {
	a := NewBig(math.MaxUint64).Mul(NewBig(math.MaxUint64))

	var b []byte
	b, err := a.MarshalBinary()
	t.NoError(err)
	t.Equal("340282366920938463426481119284349108225", string(b))

	var n Big
	err = n.UnmarshalBinary(b)
	t.NoError(err)

	t.Equal(0, a.Cmp(n))
}

func (t *testBig) TestJson() {
	a := NewBig(math.MaxUint64).Mul(NewBig(math.MaxUint64))

	var b []byte
	b, err := json.Marshal(a)
	t.NoError(err)
	t.Equal("340282366920938463426481119284349108225", string(b))

	var n Big
	{
		err := json.Unmarshal(b, &n)
		t.NoError(err)
	}

	t.Equal(0, a.Cmp(n))
}

func (t *testBig) TestEncodeDecode() {
	a := NewBig(math.MaxUint64).Mul(NewBig(math.MaxUint64))

	var b []byte
	b, err := common.RLPEncode(a)
	t.NoError(err)

	var n Big
	err = common.RLPDecode(b, &n)
	t.NoError(err)

	t.Equal(0, a.Cmp(n))
}

func (t *testBig) TestFromValue() {
	cases := []struct {
		name     string
		v        interface{}
		expected uint64
		err      string
	}{
		{name: "int", v: int(33), expected: 33},
		{name: "int8", v: int8(33), expected: 33},
		{name: "int16", v: int16(33), expected: 33},
		{name: "int32", v: int32(33), expected: 33},
		{name: "int64", v: int64(33), expected: 33},
		{name: "uint", v: uint(33), expected: 33},
		{name: "uint8", v: uint8(33), expected: 33},
		{name: "uint16", v: uint16(33), expected: 33},
		{name: "uint32", v: uint32(33), expected: 33},
		{name: "uint64", v: uint64(33), expected: 33},
		{name: "negative int", v: int(-33), err: "lower than zero"},
		{name: "negative int8", v: int8(-33), err: "lower than zero"},
		{name: "negative int16", v: int16(-33), err: "lower than zero"},
		{name: "negative int32", v: int32(-33), err: "lower than zero"},
		{name: "negative int64", v: int64(-33), err: "lower than zero"},
		{name: "not acceptable value", v: "showme", err: "invalid value"},
		{name: "from Big", v: NewBig(33), expected: 33},
	}

	for i, c := range cases {
		i := i
		c := c
		t.Run(
			c.name,
			func() {
				result, err := FromValue(c.v)
				if len(c.err) > 0 {
					t.Contains(err.Error(), c.err)
				} else {
					t.NoError(err)
					t.True(NewBig(c.expected).Equal(result), "%d: %v; %v != %v", i, c.name, c.expected, result)
				}
			},
		)
	}
}

func TestBig(t *testing.T) {
	suite.Run(t, new(testBig))
}
