package network

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type testNodes struct {
	suite.Suite
}

func (t *testNodes) newNetwork(n int) ([]*NodesTest, func()) {
	var networks []*NodesTest
	for i := 0; i < n; i++ {
		home := node.NewRandomHome()
		networks = append(networks, NewNodesTest(home))
	}

	for _, nt := range networks {
		for _, ot := range networks {
			nt.AddReceiver(ot.Home().Address(), ot.ReceiveFunc)
		}
		err := nt.Start()
		t.NoError(err)
	}

	closeFunc := func() {
		for _, nt := range networks {
			err := nt.Stop()
			t.NoError(err)
		}
	}

	return networks, closeFunc
}

func (t *testNodes) TestNew() {
	nodeCount := 3

	var chans []<-chan interface{}
	networks, closeFunc := t.newNetwork(nodeCount)
	defer closeFunc()

	for _, nt := range networks {
		chans = append(chans, nt.Reader())
	}

	var wg sync.WaitGroup
	wg.Add(nodeCount)

	message_id := common.RandomUUID()
	go func() {
		for _, ch := range chans {
			m, ok := <-ch
			if !ok {
				continue
			}
			t.Equal(message_id, m)
			wg.Done()
		}
	}()

	err := networks[0].Broadcast(message_id)
	t.NoError(err)

	wg.Wait()
}

func (t *testNodes) TestSendToOneNode() {
	nodeCount := 3

	var chans []<-chan interface{}
	networks, closeFunc := t.newNetwork(nodeCount)
	defer closeFunc()

	for _, nt := range networks {
		chans = append(chans, nt.Reader())
	}

	var wg sync.WaitGroup
	wg.Add(nodeCount)

	var message_ids []string
	for range networks {
		message_ids = append(message_ids, common.RandomUUID())
	}

	go func() {
		for i, ch := range chans {
			m, ok := <-ch
			if !ok {
				continue
			}
			t.Equal(message_ids[i], m)
			wg.Done()
		}
	}()

	// send the given message to the next node
	for i, nt := range networks {
		n := (i + 1) % len(networks)
		err := nt.Send(message_ids[n], networks[n].Home().Address())
		t.NoError(err)
	}

	wg.Wait()
}

func TestNodes(t *testing.T) {
	suite.Run(t, new(testNodes))
}
