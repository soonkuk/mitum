package network

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type ReceiveFunc func(interface{}) error

type Nodes interface {
	common.Daemon
	Home() node.Home
	Receiver() <-chan interface{}
	AddReceiver(node.Address, ReceiveFunc) error
	RemoveReceiver(node.Address) error
	Send(node.Address, interface{}) error
	Broadcast(interface{}) error
}
