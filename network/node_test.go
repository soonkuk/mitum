package network

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func (t testHash) Hash() common.Hash {
	encoded, err := t.MarshalBinary()
	if err != nil {
		return common.Hash{}
	}

	hash, _ := common.NewHash("th", encoded)
	return hash
}

type testNodeNetwork struct {
	suite.Suite
}

func (t *testNodeNetwork) newSeal() common.TestNewSeal {
	seal := common.NewTestNewSeal()

	_ = seal.Sign(common.TestNetworkID, common.RandomSeed())

	return seal
}

func (t *testNodeNetwork) TestMultipleReceiver() {
	network := NewNodeTestNetwork()
	network.skipCheckValidator = true

	node := common.NewRandomHome()

	// 2 receiver channel
	receiver0 := make(chan common.Seal)

	err := network.AddReceiver(receiver0)
	t.NoError(err)

	receiver1 := make(chan common.Seal)

	err = network.AddReceiver(receiver1)
	t.NoError(err)

	count := 10
	var wg sync.WaitGroup
	wg.Add(count * 2)

	var counted uint64

	var seals []common.TestNewSeal
	for i := 0; i < count; i++ {
		seals = append(seals, t.newSeal())
	}

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

	var send func(int)
	send = func(c int) {
		seal := seals[c]

		if err := network.Send(node, seal); err != nil {
			log.Error(err.Error())
			t.NoError(err)
			return
		}

		if c == count-1 {
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

func (t *testNodeNetwork) TestRemoveReceiver() {
	network := NewNodeTestNetwork()
	network.skipCheckValidator = true

	node := common.NewRandomHome()

	// 1 receiver channel
	receiver0 := make(chan common.Seal)
	defer close(receiver0)

	err := network.AddReceiver(receiver0)
	t.NoError(err)

	count := 10
	var seals []common.TestNewSeal
	for i := 0; i < count; i++ {
		seals = append(seals, t.newSeal())
	}

	var stop common.Hash = seals[2].Hash()

	var wg sync.WaitGroup
	wg.Add(count)

	var counted uint64
	go func() {
		receive := func(seal common.Seal, notClosed bool) (common.Hash, bool) {
			if !notClosed {
				return common.Hash{}, false
			}
			defer wg.Done()

			return seal.Hash(), true
		}

	end:
		for {
			select {
			case seal, notClosed := <-receiver0:
				hash, received := receive(seal, notClosed)
				if !received {
					break end
				}
				atomic.AddUint64(&counted, 1)

				if hash.Equal(stop) {
					network.RemoveReceiver(receiver0)
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
		seal := seals[c]

		if err := network.Send(node, seal); err != nil {
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
	t.Equal(3, int(countedFinal))
}

func TestNodeNetwork(t *testing.T) {
	suite.Run(t, new(testNodeNetwork))
}
