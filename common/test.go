// +build test

package common

import (
	"runtime/debug"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inconshreveable/log15"
)

var (
	TestNetworkID []byte = []byte("this-is-test-network")
)

func init() {
	InTest = true
	SetTestLogger(Log())
}

func SetTestLogger(logger log15.Logger) {
	handler, _ := LogHandler(LogFormatter("terminal"), "")
	handler = log15.CallerFileHandler(handler)
	logger.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, handler))
	//logger.SetHandler(log15.LvlFilterHandler(log15.LvlCrit, handler))
}

func NewRandomHash(hint string) Hash {
	h, _ := NewHash(hint, []byte(RandomUUID()))
	return h
}

func NewRandomHome() *HomeNode {
	return NewHome(RandomSeed(), NetAddr{})
}

func DebugPanic() {
	if r := recover(); r != nil {
		debug.PrintStack()
		panic(r)
	}
}

type TestNewSeal struct {
	RawSeal
	fieldA string
	fieldB string
	fieldC []byte
}

func NewTestNewSeal() TestNewSeal {
	seal := TestNewSeal{
		fieldA: RandomUUID(),
		fieldB: RandomUUID(),
		fieldC: []byte(RandomUUID()),
	}

	raw := NewRawSeal(seal, CurrentSealVersion)
	seal.RawSeal = raw

	return seal
}

func (r TestNewSeal) Type() SealType {
	return SealType("showme-type")
}

func (r TestNewSeal) Hint() string {
	return "cs"
}

func (r TestNewSeal) SerializeRLP() ([]interface{}, error) {
	return []interface{}{r.fieldA, r.fieldB, r.fieldC}, nil
}

func (r *TestNewSeal) UnserializeRLP(m []rlp.RawValue) error {
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

func (r TestNewSeal) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"field_a": r.fieldA,
		"field_b": r.fieldB,
		"field_c": r.fieldC,
	}, nil
}

func (r TestNewSeal) Wellformed() error {
	if err := r.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if len(r.fieldA) < 1 || len(r.fieldB) < 1 || len(r.fieldC) < 1 {
		return SealNotWellformedError
	}

	return nil
}
