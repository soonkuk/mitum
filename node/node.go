package node

import (
	"encoding/json"

	"github.com/spikeekips/mitum/keypair"
)

type Node interface {
	Address() Address
	PublicKey() keypair.PublicKey
	Equal(Node) bool
	Alias() string
	SetAlias(string) Node
}

type Home struct {
	address    Address
	publicKey  keypair.PublicKey
	privateKey keypair.PrivateKey
	alias      string
}

func NewHome(address Address, privateKey keypair.PrivateKey) Home {
	return Home{address: address, publicKey: privateKey.PublicKey(), privateKey: privateKey}
}

func (hm Home) Address() Address {
	return hm.address
}

func (hm Home) PublicKey() keypair.PublicKey {
	return hm.publicKey
}

func (hm Home) PrivateKey() keypair.PrivateKey {
	return hm.privateKey
}

func (hm Home) Equal(o Node) bool {
	if !hm.address.Equal(o.Address()) {
		return false
	}

	if !hm.publicKey.Equal(o.PublicKey()) {
		return false
	}

	return true
}

func (hm Home) Other() Other {
	return NewOther(hm.address, hm.publicKey)
}

func (hm Home) Alias() string {
	if len(hm.alias) > 0 {
		return hm.alias
	}

	return hm.address.String()[:6]
}

func (hm Home) SetAlias(alias string) Node {
	if len(alias) < 1 {
		alias = hm.address.String()[:6]
	}

	hm.alias = alias
	return hm
}

func (hm Home) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"address":   hm.address,
		"publickey": hm.publicKey,
	})
}

func (hm Home) String() string {
	b, _ := json.Marshal(hm) // nolint
	return string(b)
}

type Other struct {
	address   Address
	publicKey keypair.PublicKey
	alias     string
}

func NewOther(address Address, publicKey keypair.PublicKey) Other {
	return Other{address: address, publicKey: publicKey}
}

func (ot Other) Address() Address {
	return ot.address
}

func (ot Other) PublicKey() keypair.PublicKey {
	return ot.publicKey
}

func (ot Other) PrivateKey() keypair.PrivateKey {
	return nil
}

func (ot Other) Equal(o Node) bool {
	if !ot.address.Equal(o.Address()) {
		return false
	}

	if !ot.publicKey.Equal(o.PublicKey()) {
		return false
	}

	return true
}

func (ot Other) Alias() string {
	if len(ot.alias) > 0 {
		return ot.alias
	}

	return ot.address.String()[:6]
}

func (ot Other) SetAlias(alias string) Node {
	ot.alias = alias
	return ot
}
