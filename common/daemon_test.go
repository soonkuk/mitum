package common

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testReaderDaemon struct {
	suite.Suite
}

func (t *testReaderDaemon) TestNew() {
	ch := make(chan interface{})
	defer close(ch)

	d := NewReaderDaemon(true)
	d.SetReader(ch)

	count := 10
	var wg sync.WaitGroup
	wg.Add(count)

	d.SetReaderCallback(func(v interface{}) error {
		defer wg.Done()

		return nil
	})

	err := d.Start()
	t.NoError(err)

	for i := 0; i < count; i++ {
		ch <- 1
	}

	wg.Wait()

	err = d.Stop()
	t.NoError(err)

	after := time.After(time.Second * 2)
end:
	for {
		select {
		case <-after:
			t.NoError(xerrors.Errorf("not stopped"))
			break end
		default:
			if d.IsStopped() {
				break end
			}
		}
	}
}

func (t *testReaderDaemon) TestCount() {
	ch := make(chan interface{})
	defer close(ch)

	d := NewReaderDaemon(true)
	d.SetReader(ch)

	limit := 10

	var wg sync.WaitGroup
	wg.Add(limit)

	var sum uint64
	d.SetReaderCallback(func(v interface{}) error {
		defer wg.Done()

		atomic.AddUint64(&sum, uint64(v.(int)))

		return nil
	})

	err := d.Start()
	t.NoError(err)

	for i := 0; i < limit; i++ {
		ch <- i
	}

	wg.Wait()

	sumed := atomic.LoadUint64(&sum)
	t.Equal(45, int(sumed))

	err = d.Stop()
	t.NoError(err)
}

func (t *testReaderDaemon) TestAsynchronous() {
	ch := make(chan interface{})
	defer close(ch)

	d := NewReaderDaemon(false)
	d.SetReader(ch)

	limit := 10

	var wg sync.WaitGroup
	wg.Add(limit)

	var sum uint64
	d.SetReaderCallback(func(v interface{}) error {
		atomic.AddUint64(&sum, uint64(v.(int)))
		defer wg.Done()

		return nil
	})

	err := d.Start()
	t.NoError(err)

	for i := 0; i < limit; i++ {
		ch <- i
	}

	wg.Wait()

	sumed := atomic.LoadUint64(&sum)

	t.Equal(45, int(sumed))

	err = d.Stop()
	t.NoError(err)
}

func (t *testReaderDaemon) TestErrCallback() {
	ch := make(chan interface{})
	defer close(ch)

	d := NewReaderDaemon(false)
	d.SetReader(ch)

	limit := 10

	var wg sync.WaitGroup
	wg.Add(4)

	var sum uint64
	d.SetReaderCallback(func(v interface{}) error {
		if v.(int)%3 == 0 {
			return xerrors.Errorf("%d", v)
		}

		return nil
	})

	d.SetErrCallback(func(err error) {
		defer wg.Done()

		v, _ := strconv.ParseUint(err.Error(), 10, 64)
		atomic.AddUint64(&sum, uint64(v))
	})

	err := d.Start()
	t.NoError(err)

	for i := 0; i < limit; i++ {
		ch <- i
	}

	wg.Wait()

	sumed := atomic.LoadUint64(&sum)

	t.Equal(18, int(sumed))

	err = d.Stop()
	t.NoError(err)
}

func TestReaderDaemon(t *testing.T) {
	suite.Run(t, new(testReaderDaemon))
}
