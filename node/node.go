package node

import "github.com/spikeekips/mitum/keypair"

type Node interface {
	Address() Address
	PublicKey() keypair.PublicKey
}

type Home struct {
	address    Address
	publicKey  keypair.PublicKey
	privateKey keypair.PrivateKey
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

type Other struct {
	address   Address
	publicKey keypair.PublicKey
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
