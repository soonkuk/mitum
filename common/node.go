package common

import "net"

type Node struct {
	Address Address
	Publish net.Addr
}

func NewNode(address Address, publish net.Addr) Node {
	return Node{Address: address, Publish: publish}
}

func (n Node) Equal(node Node) bool {
	return n.Address == node.Address
}
