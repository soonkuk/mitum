// +build test

package network

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type NodeTestNetwork struct {
	sync.RWMutex
	local chan<- common.Seal
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
		if reflect.ValueOf(seal).Kind() == reflect.Ptr {
			seal = reflect.Indirect(reflect.ValueOf(seal)).Interface().(common.Seal)
		}
		if err := seal.Wellformed(); err != nil {
			log.Crit("not wellformed seal found", "error", err)
			panic(err)
			return err
		}

		c <- seal
	}

	return nil
}
