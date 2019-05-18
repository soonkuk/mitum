package network

import "github.com/spikeekips/mitum/common"

type SenderFunc func(common.Node, common.Seal) error
type ReceiverFunc func(common.Seal) error

type NodeNetwork interface {
	Start() error
	Stop() error
	AddReceiver( /* receiver name */ string, ReceiverFunc) error
	RemoveReceiver(string) error
	Send(common.Node, common.Seal) error
}
