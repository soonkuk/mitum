package network

import "github.com/spikeekips/mitum/common"

type SenderFunc func(common.Node, common.Seal) error

type NodeNetwork interface {
	Start() error
	Stop() error
	AddReceiver(chan<- common.Seal) error
	RemoveReceiver(chan common.Seal) error
	Send(common.Node, common.Seal) error
}
