package big

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
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
		a := NewBig(10)
		b := NewBig(math.MaxUint64)

		c, ok := a.AddOK(b)
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
		a := NewBig(10)
		b := NewBig(math.MaxUint64)

		c, ok := b.MulOK(a)
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
	b, err := RLPEncode(a)
	t.NoError(err)

	var n Big
	err = RLPDecode(b, &n)
	t.NoError(err)

	t.Equal(0, a.Cmp(n))
}

func TestBig(t *testing.T) {
	suite.Run(t, new(testBig))
}

func RLPEncode(i interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(i)
}

func RLPDecode(b []byte, i interface{}) error {
	return rlp.DecodeBytes(b, i)
}
