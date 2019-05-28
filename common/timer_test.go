package common

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type testTimerCallback struct {
	suite.Suite
}

// TestCallbackErrorStop; with SetErrorStop(true), if error occurred in
// callback(), TimerCallback will be stopped.
func (t *testTimerCallback) TestCallbackErrorStop() {
	var count int
	ch := make(chan struct{})

	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		ch <- struct{}{}
		return errors.New("something wrong")
	})
	err := tm.SetKeepRunning(true)
	t.NoError(err)

	err = tm.SetErrorStop(true)
	t.NoError(err)

	err = tm.Start()
	t.NoError(err)
	defer tm.Stop()
	defer close(ch)

end:
	for {
		select {
		case <-time.After(time.Millisecond * 30):
			break end
		case <-ch:
			count++
		}
	}

	t.Equal(1, count) // NOTE called once, but failed :)
}

// TestCallbackErrorStopSynchronous; with SetSynchronous(true), if error
// occurred in callback(), Start() will return the callback error.
func (t *testTimerCallback) TestCallbackErrorStopSynchronous() {
	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		return errors.New("something wrong")
	})
	defer tm.Stop()

	err := tm.SetKeepRunning(true)
	t.NoError(err)

	err = tm.SetErrorStop(true)
	t.NoError(err)

	err = tm.SetSynchronous(true)
	t.NoError(err)

	err = tm.Start()
	t.Contains(err.Error(), "something wrong")
}

func (t *testTimerCallback) TestKeepRunning() {
	var count int
	limit := 3
	ch := make(chan struct{})

	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		ch <- struct{}{}
		return nil
	})
	err := tm.SetKeepRunning(true)
	t.NoError(err)

	err = tm.Start()
	t.NoError(err)
	defer tm.Stop()
	defer close(ch)

end:
	for {
		select {
		case <-ch:
			if count == limit {
				break end
			}

			count++
		}
	}

	t.Equal(limit, count)
}

func (t *testTimerCallback) TestNotKeepRunning() {
	var count int
	ch := make(chan struct{})

	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		ch <- struct{}{}
		return nil
	})

	err := tm.SetKeepRunning(false)
	t.NoError(err)

	err = tm.Start()
	t.NoError(err)

	defer tm.Stop()
	defer close(ch)

end:
	for {
		select {
		case <-time.After(time.Millisecond * 30):
			break end
		case <-ch:
			count++
		}
	}

	t.Equal(1, count)
}

func (t *testTimerCallback) TestUntilCount() {
	var count int
	ch := make(chan struct{})

	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		ch <- struct{}{}
		return nil
	})
	err := tm.SetKeepRunning(false)
	t.NoError(err)

	limit := 3
	err = tm.SetLimit(limit)
	t.NoError(err)

	t.False(tm.KeepRunning())

	err = tm.Start()
	t.NoError(err)

	defer tm.Stop()
	defer close(ch)

	stopChan := make(chan bool)
	defer close(stopChan)

	go func() {
		<-time.After(time.Millisecond * 50)
		stopChan <- true
	}()

end:
	for {
		select {
		case <-stopChan:
			break end
		case <-ch:
			count++
		}
	}

	t.Equal(limit, count)
}

func (t *testTimerCallback) TestSetInitialTimeout() {
	var count int
	ch := make(chan struct{})

	tm := NewDefaultTimerCallback(time.Millisecond*100, func() error {
		ch <- struct{}{}
		return nil
	})
	err := tm.SetKeepRunning(true)
	t.NoError(err)

	initialTimeout := time.Millisecond * 1
	err = tm.SetInitialTimeout(initialTimeout)
	t.NoError(err)

	err = tm.Start()
	t.NoError(err)

	defer tm.Stop()
	defer close(ch)

	stopChan := make(chan bool)
	defer close(stopChan)

	go func() {
		<-time.After(time.Millisecond * 10)
		stopChan <- true
	}()

end:
	for {
		select {
		case <-stopChan:
			break end
		case <-ch:
			count++
		}
	}

	t.Equal(1, count)
}

func (t *testTimerCallback) TestSetSynchronous() {
	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		return nil
	})

	limit := 3
	err := tm.SetLimit(limit)
	t.NoError(err)

	err = tm.SetSynchronous(true)
	t.NoError(err)

	err = tm.Start()
	t.NoError(err)

	defer tm.Stop()

	t.Equal(limit, tm.count)
}

func (t *testTimerCallback) TestStop() {
	defer DebugPanic()

	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		return nil
	})
	_ = tm.SetKeepRunning(true)

	_ = tm.Start()
	<-time.After(time.Millisecond * 30)
	err := tm.Stop()
	t.NoError(err)
}

func (t *testTimerCallback) TestStopInsideCallback() {
	var tm *DefaultTimerCallback
	tm = NewDefaultTimerCallback(time.Millisecond*10, func() error {
		go func() {
			<-time.After(time.Millisecond * 5)
			tm.Stop()
		}()
		return nil
	})
	_ = tm.SetKeepRunning(true)

	_ = tm.Start()
	<-time.After(time.Millisecond * 30)
	t.Equal(1, tm.Count())
}

func (t *testTimerCallback) TestStopSynchronous() {
	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		return nil
	})
	_ = tm.SetKeepRunning(true)
	_ = tm.SetSynchronous(true)

	go func() {
		<-time.After(time.Millisecond * 25)
		err := tm.Stop()
		t.NoError(err)
	}()

	_ = tm.Start()
	t.Equal(2, tm.Count())
}

func TestTimerCallback(t *testing.T) {
	suite.Run(t, new(testTimerCallback))
}

type testTimerCallbackChain struct {
	suite.Suite
}

func (t *testTimerCallbackChain) TestNew() {
	limit := 3

	ch := make(chan bool)
	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		ch <- true
		return nil
	})
	_ = tm.SetKeepRunning(true)

	chain := NewTimerCallbackChain()

	_ = tm.SetLimit(limit)

	err := chain.Append(tm)
	t.NoError(err)

	err = chain.Start()
	t.NoError(err)

	var count int
end:
	for {
		select {
		case <-time.After(time.Millisecond * 100):
			break end
		case <-ch:
			count++
		}
	}

	t.Equal(limit, count)
}

func (t *testTimerCallbackChain) TestErrorStop() {
	tm := NewDefaultTimerCallback(time.Millisecond*10, func() error {
		return errors.New("something wrong")
	})
	_ = tm.SetKeepRunning(false)
	_ = tm.SetErrorStop(true)

	chain := NewTimerCallbackChain()
	err := chain.SetSynchronous(true)
	t.NoError(err)
	err = chain.SetErrorStop(true)
	t.NoError(err)

	err = chain.Append(tm)
	t.NoError(err)

	err = chain.Start()
	t.Contains(err.Error(), "something wrong")
}

func (t *testTimerCallbackChain) TestChaining() {
	limit := 3
	ch := make(chan string)
	chain := NewTimerCallbackChain()

	names := []string{
		RandomUUID(),
		RandomUUID(),
		RandomUUID(),
	}

	for _, name := range names {
		n := name
		tm := NewDefaultTimerCallback(time.Millisecond*2, func() error {
			ch <- n
			return nil
		})
		tm.SetLogContext("name", name)
		_ = tm.SetKeepRunning(false)
		_ = tm.SetLimit(limit)

		err := chain.Append(tm)
		t.NoError(err)
	}

	err := chain.Start()
	t.NoError(err)

	var seq []string
	count := map[string]int{}
end:
	for {
		select {
		case <-time.After(time.Millisecond * 20):
			break end
		case name := <-ch:
			if _, ok := count[name]; !ok {
				seq = append(seq, name)
			}

			count[name]++
		}
	}

	t.Equal(names, seq) // NOTE check run order
	t.Equal(limit, count[names[0]])
	t.Equal(limit, count[names[1]])
	t.Equal(limit, count[names[2]])
}

func TestTimerCallbackChain(t *testing.T) {
	suite.Run(t, new(testTimerCallbackChain))
}

type testMultiTimerCallback struct {
	suite.Suite
}

func (t *testMultiTimerCallback) TestNew() {
	limit := 3
	countTimer := 4
	ch := make(chan int)

	var timers []TimerCallback
	for i := 0; i < countTimer; i++ {
		id := i
		timer := NewDefaultTimerCallback(time.Millisecond*3, func() error {
			ch <- id
			return nil
		})
		timer.SetLimit(limit)

		timers = append(timers, timer)
	}

	m := NewMultiCallbackChain(timers...)
	err := m.Start()
	t.NoError(err)
	defer m.Stop()

	collected := map[int]int{}

end:
	for {
		select {
		case <-time.After(time.Millisecond * 100):
			break end
		case id := <-ch:
			collected[id]++
		}
	}

	t.Equal(countTimer, len(collected))
	for _, v := range collected {
		t.Equal(limit, v)
	}
}

func (t *testMultiTimerCallback) TestSynchronous() {
	limit := 3
	ch := make(chan int)

	var timers []TimerCallback

	var counter0 int = 100
	timer0 := NewDefaultTimerCallback(time.Millisecond*3, func() error {
		ch <- counter0
		counter0++
		return nil
	})
	timer0.SetLimit(limit)

	var counter1 int
	timer1 := NewDefaultTimerCallback(time.Millisecond*3, func() error {
		ch <- counter1
		counter1++
		return nil
	})
	timer1.SetLimit(limit)

	timers = append(timers, timer0, timer1)

	m := NewMultiCallbackChain(timers...)
	m.SetSynchronous(true)

	go func() {
		err := m.Start()
		t.NoError(err)
		defer m.Stop()
	}()

	var collected []int

end:
	for {
		select {
		case <-time.After(time.Millisecond * 100):
			break end
		case id := <-ch:
			collected = append(collected, id)
		}
	}

	t.Equal(limit*2, len(collected))
	t.Equal(
		[]int{100, 101, 102, 0, 1, 2},
		collected,
	)
}

func (t *testMultiTimerCallback) TestStop() {
	ch := make(chan string)

	var timers []TimerCallback

	timer0 := NewDefaultTimerCallback(time.Millisecond*3, func() error {
		ch <- RandomUUID()
		return nil
	})
	timer0.SetKeepRunning(true)

	timer1 := NewDefaultTimerCallback(time.Millisecond*3, func() error {
		ch <- RandomUUID()
		return nil
	})
	timer1.SetKeepRunning(true)

	timers = append(timers, timer0, timer1)

	m := NewMultiCallbackChain(timers...)

	err := m.Start()
	t.NoError(err)

	var collected []string

	after := time.After(time.Millisecond * 20)
end:
	for {
		select {
		case <-after:
			err = m.Stop()
			t.NoError(err)
			break end
		case id := <-ch:
			collected = append(collected, id)
		}
	}

	t.True(len(collected) >= 6)
}

func TestMultiTimerCallback(t *testing.T) {
	suite.Run(t, new(testMultiTimerCallback))
}
