package big

import (
	"encoding/json"
	"math/big"
)

var (
	ZeroBigInt *big.Int = new(big.Int).SetInt64(0)
	ZeroBig    Big      = NewBig(0)
)

type Big struct {
	big.Int
}

func NewBig(i uint64) Big {
	var a big.Int
	a.SetUint64(i)

	return Big{Int: a}
}

func ParseBig(s string) (Big, error) {
	var a big.Int
	err := a.UnmarshalText([]byte(s))
	if err != nil {
		return Big{}, err
	}

	return Big{Int: a}, nil
}

func (a Big) MarshalBinary() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Big) UnmarshalBinary(b []byte) error {
	p, err := ParseBig(string(b))
	if err != nil {
		return err
	}

	*a = p

	return nil
}

func (a Big) MarshalJSON() ([]byte, error) {
	return json.Marshal(&a.Int)
}

func (a Big) String() string {
	return (&a.Int).String()
}

func (a Big) Inc() Big {
	b, _ := a.AddOK(NewBig(1))
	return b
}

func (a Big) Add(n Big) Big {
	b, _ := a.AddOK(n)
	return b
}

func (a Big) AddOK(n Big) (Big, bool) {
	var b big.Int
	b.Add(&a.Int, &n.Int)
	return Big{Int: b}, true
}

func (a Big) Sub(n Big) Big {
	b, _ := a.SubOK(n)
	return b
}

func (a Big) SubOK(n Big) (Big, bool) {
	switch a.Int.Cmp(&n.Int) {
	case -1:
		return Big{}, false
	case 0:
		return Big{}, true
	}

	var b big.Int
	b.Sub(&a.Int, &n.Int)
	return Big{Int: b}, true
}

func (a Big) Dec() Big {
	b, _ := a.SubOK(NewBig(1))
	return b
}

func (a Big) MulOK(n Big) (Big, bool) {
	var b big.Int
	b.Mul(&a.Int, &n.Int)
	return Big{Int: b}, true
}

func (a Big) Div(n Big) Big {
	b, _ := a.DivOK(n)
	return b
}

func (a Big) DivOK(n Big) (Big, bool) {
	if n.Int.Cmp(ZeroBigInt) == 0 {
		return Big{}, false
	}

	var b big.Int
	b.Div(&a.Int, &n.Int)
	return Big{Int: b}, true
}

func (a Big) Rem(n Big) Big {
	b, _ := a.RemOK(n)
	return b
}

func (a Big) RemOK(n Big) (Big, bool) {
	if n.Int.Cmp(ZeroBigInt) == 0 {
		return ZeroBig, true
	}

	var b big.Int
	b.Rem(&a.Int, &n.Int)
	return Big{Int: b}, true
}

func (a Big) Mul(n Big) Big {
	b, _ := a.MulOK(n)
	return b
}

func (a Big) IsZero() bool {
	return a.Int.Cmp(ZeroBigInt) == 0
}

func (a Big) Cmp(b Big) int {
	return a.Int.Cmp(&b.Int)
}

func (a Big) Equal(b Big) bool {
	return a.Int.Cmp(&b.Int) == 0
}

func (a Big) Uint64() uint64 {
	b, _ := a.Uint64Ok()
	return b
}

func (a Big) Uint64Ok() (uint64, bool) {
	return (&(a.Int)).Uint64(), (&(a.Int)).IsUint64()
}
