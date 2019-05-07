package common

import (
	"encoding"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type Node interface {
	encoding.BinaryMarshaler
	Address() Address
	Publish() NetAddr
	Validators() map[Address]Validator
	Equal(Node) bool
	String() string
}

type BaseNode struct {
	address    Address
	publish    NetAddr
	validators map[Address]Validator
}

func NewBaseNode(address Address, publish NetAddr, validators []Validator) BaseNode {
	vs := map[Address]Validator{}
	for _, v := range validators {
		vs[v.Address()] = v
	}

	return BaseNode{address: address, publish: publish, validators: vs}
}

func (n BaseNode) Address() Address {
	return n.address
}

func (n BaseNode) Publish() NetAddr {
	return n.publish
}

func (n BaseNode) Validators() map[Address]Validator {
	return n.validators
}

func (n BaseNode) Equal(node Node) bool {
	return n.Address() == node.Address()
}

func (n BaseNode) MarshalBinary() ([]byte, error) {
	// NOTE sort valdators by address
	var validators [][]byte
	if len(n.validators) > 0 {
		var addresses []string
		for a := range n.validators {
			addresses = append(addresses, string(a))
		}
		sort.Strings(addresses)

		for _, address := range addresses {
			b, err := n.validators[Address(address)].MarshalBinary()
			if err != nil {
				return nil, err
			}
			validators = append(validators, b)
		}
	}

	publish, err := n.publish.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return Encode([]interface{}{
		n.address,
		publish,
		validators,
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

	validators := map[Address]Validator{}
	{
		var vs [][]byte
		if err := Decode(m[2], &vs); err != nil {
			return err
		}

		for _, b := range vs {
			var validator Validator
			if err := validator.UnmarshalBinary(b); err != nil {
				return err
			}
			validators[validator.Address()] = validator
		}
	}

	n.address = address
	n.publish = publish
	n.validators = validators

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
	sync.RWMutex
	seed Seed
}

func NewHome(seed Seed, publish NetAddr) *HomeNode {
	return &HomeNode{
		BaseNode: NewBaseNode(seed.Address(), publish, nil),
		seed:     seed,
	}
}

func (n *HomeNode) Seed() Seed {
	n.RLock()
	defer n.RUnlock()

	return n.seed
}

func (n *HomeNode) AddValidators(validators ...Validator) {
	n.Lock()
	defer n.Unlock()

	for _, v := range validators {
		if _, found := n.validators[v.Address()]; found {
			continue
		}
		n.validators[v.Address()] = v
	}
}

func (n *HomeNode) RemoveValidators(validators ...Validator) {
	n.Lock()
	defer n.Unlock()

	for _, v := range validators {
		if _, found := n.validators[v.Address()]; !found {
			continue
		}
		delete(n.validators, v.Address())
	}
}

func (n *HomeNode) UnmarshalBinary(b []byte) error {
	var node BaseNode
	if err := Decode(b, &node); err != nil {
		return err
	}

	n.address = node.Address()
	n.publish = node.Publish()
	n.validators = node.Validators()

	return nil
}

type Validator struct {
	BaseNode
}

func NewValidator(address Address, publish NetAddr, validators []Validator) Validator {
	b := NewBaseNode(address, publish, validators)
	return Validator{BaseNode: b}
}
