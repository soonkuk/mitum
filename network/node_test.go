package network

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testHash struct {
	I uint64
}

func (t testHash) MarshalBinary() ([]byte, error) {
	return []byte(strconv.FormatUint(t.I, 10)), nil
}

func (t *testHash) UnmarshalBinary(b []byte) error {
	i, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}

	t.I = i
	return nil
}

func (t testHash) Hash() (common.Hash, []byte, error) {
	encoded, err := t.MarshalBinary()
	if err != nil {
		return common.Hash{}, nil, err
	}

	hash, _ := common.NewHash("th", encoded)
	return hash, encoded, nil
}

type nodeTestNetwork struct {
	sync.RWMutex
	chans map[int64]chan<- common.Seal
}

func newNodeTestNetwork() *nodeTestNetwork {
	return &nodeTestNetwork{
		chans: map[int64]chan<- common.Seal{},
	}
}

func (n *nodeTestNetwork) addSeal(seal common.Seal) error {
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

func (n *nodeTestNetwork) RegisterReceiver(c chan<- common.Seal) error {
	n.Lock()
	defer n.Unlock()

	p := *(*int64)(unsafe.Pointer(&c))
	if _, found := n.chans[p]; found {
		return ReceiverAlreadyRegisteredError
	}

	n.chans[p] = c
	return nil
}

func (n *nodeTestNetwork) UnregisterReceiver(c chan<- common.Seal) error {
	n.Lock()
	defer n.Unlock()

	p := *(*int64)(unsafe.Pointer(&c))
	if _, found := n.chans[p]; !found {
		return ReceiverNotRegisteredError
	}

	delete(n.chans, p)
	return nil
}

func (n *nodeTestNetwork) Send(node common.Node, seal common.Seal) error {
	return nil
}

type testNodeNetwork struct {
	suite.Suite
}

func (t *testNodeNetwork) newSeal(c uint64) common.Seal {
	seal, err := common.NewSeal(common.NewSealType("test"), testHash{I: c})
	t.NoError(err)

	return seal
}

func (t *testNodeNetwork) TestMultipleReceiver() {
	network := newNodeTestNetwork()

	// 2 receiver channel
	receiver0 := make(chan common.Seal)

	err := network.RegisterReceiver(receiver0)
	t.NoError(err)

	receiver1 := make(chan common.Seal)

	err = network.RegisterReceiver(receiver1)
	t.NoError(err)

	count := 10
	var wg sync.WaitGroup
	wg.Add(count * 2)

	var counted uint64

	go func() {
	end:
		for {
			select {
			case _, notClosed := <-receiver0:
				if !notClosed {
					break end
				}
				wg.Done()
				atomic.AddUint64(&counted, 1)
			case _, notClosed := <-receiver1:
				if !notClosed {
					break end
				}
				wg.Done()
				atomic.AddUint64(&counted, 1)
			case <-time.After(time.Millisecond * time.Duration(200*count)):
				t.NoError(errors.New("failed; timeouted"))
				break end
			}
		}
	}()

	var send func(uint64)
	send = func(c uint64) {
		seal := t.newSeal(c)

		if err := network.addSeal(seal); err != nil {
			return
		}
		if c == uint64(count)-1 {
			return
		}

		go func() {
			time.Sleep(time.Millisecond * 10)
			send(c + 1)
		}()
	}

	send(0)

	wg.Wait()
	close(receiver0)
	close(receiver1)

	countedFinal := atomic.LoadUint64(&counted)
	t.Equal(uint64(count*2), countedFinal)
}

func (t *testNodeNetwork) TestUnregisterReceiver() {
	network := newNodeTestNetwork()

	// 1 receiver channel
	receiver0 := make(chan common.Seal)
	defer close(receiver0)

	err := network.RegisterReceiver(receiver0)
	t.NoError(err)

	count := 10
	var stop uint64 = 3

	var wg sync.WaitGroup
	wg.Add(count)

	var counted uint64
	go func() {
		receive := func(seal common.Seal, notClosed bool) (testHash, bool) {
			if !notClosed {
				return testHash{}, false
			}
			defer wg.Done()

			var th testHash
			err := seal.UnmarshalBody(&th)
			if err != nil {
				return testHash{}, true
			}

			return th, true
		}

	end:
		for {
			select {
			case seal, notClosed := <-receiver0:
				th, received := receive(seal, notClosed)
				if !received {
					break end
				}
				atomic.AddUint64(&counted, 1)

				if th.I == stop-1 {
					network.UnregisterReceiver(receiver0)
				}
			case <-time.After(time.Millisecond * 100):
				break end
			}
		}

	_:
		countedFinal := atomic.LoadUint64(&counted)
		for i := 0; i < count-int(countedFinal); i++ {
			wg.Done()
		}
	}()

	var send func(uint64)
	send = func(c uint64) {
		seal := t.newSeal(c)

		if err := network.addSeal(seal); err != nil {
			return
		}
		if c == uint64(count)-1 {
			return
		}

		go func() {
			time.Sleep(time.Millisecond * 10)
			send(c + 1)
		}()
	}

	send(0)

	wg.Wait()

	countedFinal := atomic.LoadUint64(&counted)
	t.Equal(stop, countedFinal)
}

func TestNodeNetwork(t *testing.T) {
	suite.Run(t, new(testNodeNetwork))
}
