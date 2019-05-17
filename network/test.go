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
	chans              map[string]chan<- common.Seal
	validatorsChan     map[common.Address]chan<- common.Seal
	skipCheckValidator bool
}

func NewNodeTestNetwork() *NodeTestNetwork {
	return &NodeTestNetwork{
		chans:          map[string]chan<- common.Seal{},
		validatorsChan: map[common.Address]chan<- common.Seal{},
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

func (n *NodeTestNetwork) AddValidatorChan(validator common.Validator, c chan<- common.Seal) *NodeTestNetwork {
	n.Lock()
	defer n.Unlock()

	n.validatorsChan[validator.Address()] = c

	return n
}

func (n *NodeTestNetwork) Send(node common.Node, seal common.Seal) error {
	n.RLock()
	defer n.RUnlock()

	if reflect.ValueOf(seal).Kind() == reflect.Ptr {
		seal = reflect.Indirect(reflect.ValueOf(seal)).Interface().(common.Seal)
	}

	if err := seal.Wellformed(); err != nil {
		log.Crit("not wellformed seal found", "error", err)
		panic(err)
		return err
	}

	if len(n.chans) > 0 {
		for _, c := range n.chans {
			c <- seal
		}
	}

	sender, found := n.validatorsChan[node.Address()]
	if found {
		sender <- seal
	} else if !n.skipCheckValidator {
		return fmt.Errorf("not registered node for broadcasting: %v", node.Address())
	}

	return nil
}
