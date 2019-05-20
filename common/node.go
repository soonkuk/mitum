package common

import (
	"encoding"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"
)

type Node interface {
	encoding.BinaryMarshaler
	Name() string
	Address() Address
	Publish() NetAddr
	Equal(Node) bool
	String() string
}

type BaseNode struct {
	address Address
	publish NetAddr
}

func NewBaseNode(address Address, publish NetAddr) BaseNode {
	return BaseNode{address: address, publish: publish}
}

func (n BaseNode) Name() string {
	return n.address.Alias()
}

func (n BaseNode) Address() Address {
	return n.address
}

func (n BaseNode) Publish() NetAddr {
	return n.publish
}

func (n BaseNode) Equal(node Node) bool {
	return n.Address() == node.Address()
}

func (n BaseNode) MarshalBinary() ([]byte, error) {
	// NOTE sort valdators by address
	publish, err := n.publish.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return Encode([]interface{}{
		n.address,
		publish,
	})
}

func (n *BaseNode) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := Decode(b, &m); err != nil {
		return err
	}

	var address Address
	if err := Decode(m[0], &address); err != nil {
		return err
	}

	var publish NetAddr
	{
		var vs []byte
		if err := Decode(m[1], &vs); err != nil {
			return err
		}
		if err := publish.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	n.address = address
	n.publish = publish

	return nil
}

func (n BaseNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"address": n.address,
		"publish": n.publish,
	})
}

func (n *BaseNode) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	var address Address
	if err := json.Unmarshal(m["address"], &address); err != nil {
		return err
	}

	var publish NetAddr
	if err := json.Unmarshal(m["publish"], &publish); err != nil {
		return err
	}

	n.address = address
	n.publish = publish

	return nil
}

func (n BaseNode) AsValidator() Validator {
	return Validator{BaseNode: n}
}

func (n BaseNode) String() string {
	b, _ := json.Marshal(n)
	return strings.Replace(string(b), "\"", "'", -1)
}

type HomeNode struct {
	BaseNode
	seed Seed
}

func NewHome(seed Seed, publish NetAddr) HomeNode {
	return HomeNode{
		BaseNode: NewBaseNode(seed.Address(), publish),
		seed:     seed,
	}
}

func (n HomeNode) Seed() Seed {
	return n.seed
}

func (n HomeNode) UnmarshalBinary(b []byte) error {
	var node BaseNode
	if err := Decode(b, &node); err != nil {
		return err
	}

	n.address = node.Address()
	n.publish = node.Publish()

	return nil
}

type Validator struct {
	BaseNode
}

func NewValidator(address Address, publish NetAddr) Validator {
	b := NewBaseNode(address, publish)
	return Validator{BaseNode: b}
}
