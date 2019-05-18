package common

import (
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

func TestNode(t *testing.T) {
	suite.Run(t, new(testNode))
}
