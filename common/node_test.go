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

	node := NewBaseNode(address, publish)

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

	node := NewBaseNode(address, publish)

	t.Equal(address, node.Address())
	t.Equal(addr, node.Publish().String())
}

func (t testNode) newValidators(count int) []Validator {
	var vs []Validator
	for i := 0; i < count; i++ {
		addr := fmt.Sprintf("http://%s:%d", RandomUUID(), rand.Int())
		publish, _ := NewNetAddr(addr)

		v := NewValidator(RandomSeed().Address(), publish)

		vs = append(vs, v)
	}

	return vs
}

func TestNode(t *testing.T) {
	suite.Run(t, new(testNode))
}
