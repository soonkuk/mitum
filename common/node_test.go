package common

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testNode struct {
	suite.Suite
}

func (t testNode) TestBaseNode() {
	address := RandomSeed().Address()
	addr := "unix://little-socket.sock"
	publish, _ := NewNetAddr(addr)

	node := NewBaseNode(address, publish, nil)

	t.Equal(address, node.Address())
	t.Equal(addr, node.Publish().String())

	{ // MarshalBinary
		b, err := node.MarshalBinary()
		t.NoError(err)

		var unmarshaled BaseNode
		err = unmarshaled.UnmarshalBinary(b)
		t.NoError(err)

		t.Equal(address, unmarshaled.Address())
		t.Equal(addr, unmarshaled.Publish().String())
	}
}

func (t testNode) TestPublish() {
	address := RandomSeed().Address()
	addr := "http://showme"
	publish, _ := NewNetAddr(addr)

	node := NewBaseNode(address, publish, nil)

	t.Equal(address, node.Address())
	t.Equal(addr, node.Publish().String())
}

func (t testNode) newValidators(count int) []Validator {
	var vs []Validator
	for i := 0; i < count; i++ {
		addr := fmt.Sprintf("http://%s:%d", RandomUUID(), rand.Int())
		publish, _ := NewNetAddr(addr)

		v := NewValidator(RandomSeed().Address(), publish, nil)

		vs = append(vs, v)
	}

	return vs
}

func (t testNode) TestValidators() {
	address := RandomSeed().Address()
	addr := "http://showme"
	publish, _ := NewNetAddr(addr)
	validators := t.newValidators(3)

	node := NewBaseNode(address, publish, validators)

	t.Equal(address, node.Address())
	t.Equal(addr, node.Publish().String())

	for _, v := range validators {
		rv, ok := node.Validators()[v.Address()]
		t.True(ok)
		t.Equal(v.Address(), rv.Address())
		t.True(v.Publish().Equal(rv.Publish()))
	}

	{ // MarshalBinary
		b, err := node.MarshalBinary()
		t.NoError(err)

		var unmarshaled BaseNode
		err = unmarshaled.UnmarshalBinary(b)
		t.NoError(err)

		t.Equal(address, unmarshaled.Address())
		t.Equal(addr, unmarshaled.Publish().String())

		for _, v := range validators {
			rv, ok := unmarshaled.Validators()[v.Address()]
			t.True(ok)
			t.Equal(v.Address(), rv.Address())
			t.True(v.Publish().Equal(rv.Publish()))
		}
	}
}

func TestNode(t *testing.T) {
	suite.Run(t, new(testNode))
}
