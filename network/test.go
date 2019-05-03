// +build test

package network

import (
	"fmt"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type NodeTestNetwork struct {
	sync.RWMutex
	chans map[string]chan<- common.Seal
}

func NewNodeTestNetwork() *NodeTestNetwork {
	return &NodeTestNetwork{
		chans: map[string]chan<- common.Seal{},
	}
}

func (n *NodeTestNetwork) Start() error {
	return nil
}

func (n *NodeTestNetwork) Stop() error {
	return nil
}

func (n *NodeTestNetwork) AddReceiver(c chan<- common.Seal) error {
	n.Lock()
	defer n.Unlock()

	p := fmt.Sprintf("%p", c)
	if _, found := n.chans[p]; found {
		return ReceiverAlreadyRegisteredError
	}

	n.chans[p] = c
	return nil
}

func (n *NodeTestNetwork) RemoveReceiver(c chan common.Seal) error {
	n.Lock()
	defer n.Unlock()

	p := fmt.Sprintf("%p", c)
	if _, found := n.chans[p]; !found {
		return ReceiverNotRegisteredError
	}

	delete(n.chans, p)
	return nil
}

func (n *NodeTestNetwork) Send(node common.Node, seal common.Seal) error {
	n.RLock()
	defer n.RUnlock()

	if len(n.chans) < 1 {
		return NoReceiversError
	}

	for _, c := range n.chans {
		c <- seal
	}

	return nil
}
