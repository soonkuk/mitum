package network

import "github.com/spikeekips/mitum/common"

type NodeNetwork interface {
	Start() error
	Stop() error
	RegisterReceiver(chan<- common.Seal) error
	UnregisterReceiver(chan common.Seal) error
	Send(common.Node, common.Seal) error
}
