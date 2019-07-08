// +build test

package network

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

func init() {
	common.SetTestLogger(Log())
}

type NodesTest struct {
	*common.Logger
	*common.ReaderDaemon
	sync.RWMutex
	home  node.Home
	nodes map[node.Address]ReceiveFunc
}

func NewNodesTest(home node.Home) *NodesTest {
	nt := &NodesTest{
		Logger: common.NewLogger(Log(), "module", "test-nodes-network", "home", home),
		home:   home,
		nodes:  map[node.Address]ReceiveFunc{},
	}

	nt.ReaderDaemon = common.NewReaderDaemon(true, nil)
	nt.AddReceiver(nt.home.Address(), nt.ReceiveFunc)

	return nt
}

func (nt *NodesTest) Home() node.Home {
	return nt.home
}

func (nt *NodesTest) ReceiveFunc(v interface{}) error {
	go nt.ReaderDaemon.Write(v)

	return nil
}

func (nt *NodesTest) AddReceiver(address node.Address, rf ReceiveFunc) error {
	nt.Lock()
	defer nt.Unlock()

	nt.nodes[address] = rf

	return nil
}

func (nt *NodesTest) RemoveReceiver(address node.Address) error {
	nt.Lock()
	defer nt.Unlock()

	if _, found := nt.nodes[address]; !found {
		return xerrors.Errorf("address not found; address=%q", address)
	}

	delete(nt.nodes, address)

	return nil
}

func (nt *NodesTest) Send(v interface{}, addresses ...node.Address) error {
	nt.RLock()

	nodes := map[node.Address]ReceiveFunc{}

	if len(addresses) < 1 {
		nodes = nt.nodes
	} else {
		for _, address := range addresses {
			rf, found := nt.nodes[address]
			if !found {
				nt.Log().Error("node not added", "node", address)
				continue
			}

			nodes[address] = rf
		}
	}
	nt.RUnlock()

	nt.Log().Debug("trying to send message", "addresses", addresses)

	var errs []error
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for address, rf := range nodes {
		go func(address node.Address, rf ReceiveFunc) {
			if err := rf(v); err != nil {
				errs = append(errs, xerrors.Errorf("%s(address=%q)", err.Error(), address))
			}
			wg.Done()
		}(address, rf)
	}
	wg.Wait()

	if len(errs) > 0 {
		return xerrors.Errorf("failed to broadcast; errors=%q")
	}

	return nil
}

func (nt *NodesTest) Broadcast(v interface{}) error {
	return nt.Send(v)
}
