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
	*common.Logger
	chans              map[string]ReceiverFunc
	validatorsChan     map[common.Address]ReceiverFunc
	SkipCheckValidator bool
}

func NewNodeTestNetwork() *NodeTestNetwork {
	return &NodeTestNetwork{
		Logger:         common.NewLogger(log, "module", "node-test-network"),
		chans:          map[string]ReceiverFunc{},
		validatorsChan: map[common.Address]ReceiverFunc{},
	}
}

func (n *NodeTestNetwork) Start() error {
	n.Log().Debug("trying to start node network")
	n.Log().Debug("node network started")
	return nil
}

func (n *NodeTestNetwork) Stop() error {
	n.Log().Debug("trying to stop node network")
	n.Log().Debug("node network stopped")
	return nil
}

func (n *NodeTestNetwork) AddReceiver(name string, f ReceiverFunc) error {
	n.Lock()
	defer n.Unlock()

	if _, found := n.chans[name]; found {
		return ReceiverAlreadyRegisteredError
	}

	n.chans[name] = f
	return nil
}

func (n *NodeTestNetwork) RemoveReceiver(name string) error {
	n.Lock()
	defer n.Unlock()

	if _, found := n.chans[name]; !found {
		return ReceiverNotRegisteredError
	}

	delete(n.chans, name)
	return nil
}

func (n *NodeTestNetwork) AddValidatorChan(validator common.Validator, f ReceiverFunc) *NodeTestNetwork {
	n.Lock()
	defer n.Unlock()

	n.validatorsChan[validator.Address()] = f

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
	}

	if len(n.chans) > 0 {
		for _, f := range n.chans {
			if err := f(seal); err != nil {
				log.Error("failed to receive", "error", err)
				return err
			}
		}
	}

	sender, found := n.validatorsChan[node.Address()]
	if !found {
		if n.SkipCheckValidator {
			return nil
		} else {
			return fmt.Errorf("not registered node for broadcasting: %v", node.Address())
		}
	}

	return sender(seal)
}
