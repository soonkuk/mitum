package common

import (
	"fmt"
	"sync"
	"time"
)

const (
	_ uint = iota
	TimerCallbackAlreadyStartedErrorCode
	TimerCallbackInvalidTimeoutErrorCode
	TimerCallbackChainAlreadyAddedErrorCode
	InvalidTimerCallbackForChainErrorCode
)

var (
	TimerCallbackAlreadyStartedError Error = NewError(
		"timer", TimerCallbackAlreadyStartedErrorCode, "timer is already started",
	)
	TimerCallbackInvalidTimeoutError Error = NewError(
		"timer", TimerCallbackInvalidTimeoutErrorCode, "invalid timeout value",
	)
	TimerCallbackChainAlreadyAddedError Error = NewError(
		"timer", TimerCallbackChainAlreadyAddedErrorCode, "callback timer already added",
	)
	InvalidTimerCallbackForChainError Error = NewError(
		"timer", InvalidTimerCallbackForChainErrorCode, "invalid callback timer for chain",
	)
)

type TimerCallback interface {
	Start() error
	Stop() error
	ErrorStop() bool
	SetErrorStop(bool) error
	KeepRunning() bool
	SetKeepRunning(bool) error
	Synchronous() bool
	SetSynchronous(bool) error
}

type DefaultTimerCallback struct {
	sync.RWMutex
	stopOnce sync.Once
	*Logger
	initialTimeout time.Duration
	timeout        time.Duration
	stopChan       chan bool
	callback       func() error
	keepRunning    bool
	synchronous    bool
	limit          int
	errorStop      bool
	count          int
}

func NewDefaultTimerCallback(timeout time.Duration, callback func() error) *DefaultTimerCallback {
	return &DefaultTimerCallback{
		Logger:         NewLogger(log, "timer-id", RandomUUID()),
		callback:       callback,
		timeout:        timeout,
		initialTimeout: timeout,
	}
}

func (c *DefaultTimerCallback) Start() error {
	c.Log().Debug(
		"trying to start default timer callback",
		"initial-timeout", c.initialTimeout,
		"timeout", c.timeout,
		"keep-running", c.keepRunning,
		"synchronous", c.synchronous,
		"error-stop", c.errorStop,
	)

	if c.stopChan != nil {
		return StartStopperAlreadyStartedError
	}

	c.Lock()
	c.stopChan = make(chan bool, 1)
	c.Unlock()

	c.Log().Debug("timer callback started")

	if c.synchronous {
		defer func() {
			_ = c.Stop()
		}()

		return c.run()
	}

	go func() {
		defer func() {
			_ = c.Stop()
		}()
		_ = c.run()
	}()

	return nil
}

func (c *DefaultTimerCallback) Stop() error {
	c.stopOnce.Do(func() {
		c.Lock()
		defer c.Unlock()

		c.Log().Debug("trying to stop timer callback")

		c.stopChan <- true
		close(c.stopChan)

		c.Log().Debug("timer callback stopped")
	})

	return nil
}

func (c *DefaultTimerCallback) ErrorStop() bool {
	c.RLock()
	defer c.RUnlock()

	return c.errorStop
}

func (c *DefaultTimerCallback) SetErrorStop(stop bool) error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set errorStop")
	}

	c.errorStop = stop

	return nil
}

func (c *DefaultTimerCallback) SetInitialTimeout(timeout time.Duration) error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set initialTimeout")
	}

	if timeout < 0 {
		return TimerCallbackInvalidTimeoutError.AppendMessage("timeout=%v", timeout)
	}

	c.initialTimeout = timeout

	return nil
}

func (c *DefaultTimerCallback) KeepRunning() bool {
	c.RLock()
	defer c.RUnlock()

	return c.keepRunning
}

func (c *DefaultTimerCallback) SetKeepRunning(keepRunning bool) error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set keepRunning")
	}

	c.keepRunning = keepRunning

	return nil
}

func (c *DefaultTimerCallback) Limit() int {
	c.RLock()
	defer c.RUnlock()

	return c.limit
}

func (c *DefaultTimerCallback) SetLimit(limit int) error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set limit")
	}

	c.limit = limit
	c.keepRunning = limit < 1

	return nil
}

func (c *DefaultTimerCallback) Synchronous() bool {
	c.RLock()
	defer c.RUnlock()

	return c.synchronous
}

func (c *DefaultTimerCallback) SetSynchronous(synchronous bool) error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set synchronous")
	}

	c.synchronous = synchronous

	return nil
}

func (c *DefaultTimerCallback) run() error {
	var limit int = c.limit
	if limit < 1 {
		limit = 1
	}

end0:
	for {
		if c.count >= limit {
			break end0
		}

		select {
		case <-c.stopChan:
			break end0
		case <-time.After(c.initialTimeout):
			c.Log().Debug("timeout expired")
			if err := c.runCallback(); err != nil && c.errorStop {
				return err
			}
		}
	}

	if !c.keepRunning {
		return nil
	}

end1:
	for {
		select {
		case <-c.stopChan:
			break end1
		case <-time.After(c.timeout):
			c.Log().Debug("timout expired in keeprunning")
			if err := c.runCallback(); err != nil && c.errorStop {
				return err
			}

			if c.limit > 0 && c.count >= c.limit {
				break end1
			}
		}
	}

	return nil
}

func (c *DefaultTimerCallback) Count() int {
	c.RLock()
	defer c.RUnlock()

	return c.count
}

func (c *DefaultTimerCallback) runCallback() error {
	err := c.callback()

	c.Lock()
	c.count++
	c.Unlock()

	if err != nil {
		c.Log().Error("failed to callback", "error", err)
	}

	return err
}

type TimerCallbackChain struct {
	sync.RWMutex
	stopOnce sync.Once
	*Logger
	stopChan    chan bool
	synchronous bool
	// TODO
	errorStop bool
	timers    map[ /* callback address */ string]TimerCallback
	seq       []string
	current   int // index of seq
}

func NewTimerCallbackChain() *TimerCallbackChain {
	return &TimerCallbackChain{
		Logger: NewLogger(log, "timer-id", RandomUUID()),
		timers: map[string]TimerCallback{},
	}
}

func (c *TimerCallbackChain) Start() error {
	c.Log().Debug(
		"trying to start timer callback chain",
		"initial-timers", len(c.timers),
		"synchronous", c.synchronous,
	)

	c.RLock()
	stopChanIsNil := c.stopChan != nil
	c.RUnlock()

	if stopChanIsNil {
		return StartStopperAlreadyStartedError
	}

	c.Lock()
	c.stopChan = make(chan bool)
	c.Unlock()

	c.Log().Debug("timer callback chain started")

	if c.synchronous {
		defer func() {
			_ = c.Stop()
		}()
		return c.runTimers()
	}

	go func() {
		_ = c.runTimers()
	}()

	return nil
}

func (c *TimerCallbackChain) Stop() error {
	c.stopOnce.Do(func() {
		c.Lock()
		defer c.Unlock()

		if c.stopChan == nil {
			return
		}

		c.Log().Debug("trying to stop timer callback chain")

		c.stopChan <- true
		close(c.stopChan)

		c.Log().Debug("timer callback chain stopped")
	})

	return nil
}

func (c *TimerCallbackChain) KeepRunning() bool {
	return false
}

func (c *TimerCallbackChain) ErrorStop() bool {
	c.RLock()
	defer c.RUnlock()

	return c.errorStop
}

func (c *TimerCallbackChain) SetErrorStop(stop bool) error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set errorStop")
	}

	c.errorStop = stop

	return nil
}

func (c *TimerCallbackChain) SetSynchronous(synchronous bool) error {
	if c.stopChan != nil {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set synchronous")
	}

	c.Lock()
	defer c.Unlock()

	c.synchronous = synchronous

	return nil
}

func (c *TimerCallbackChain) runTimers() error {
	defer func() {
		c.Lock()
		if c.stopChan != nil {
			close(c.stopChan)
			c.stopChan = nil
		}
		c.Unlock()
		_ = c.Stop()
	}()

	for {
		if c.stopChan == nil {
			return nil
		}

		if c.current >= len(c.timers) {
			return nil
		}

		timer := c.timers[c.seq[c.current]]
		err := c.runTimer(timer)
		if err != nil && c.errorStop {
			return err
		}
		c.current++
	}
}

func (c *TimerCallbackChain) runTimer(timer TimerCallback) error {
	c.Lock()
	defer c.Unlock()

	_ = timer.SetSynchronous(true)
	defer func() {
		_ = timer.Stop()
	}()

	err := timer.Start()
	if err != nil && c.errorStop {
		return err
	}

	return nil
}

func (c *TimerCallbackChain) Append(timer TimerCallback) error {
	if timer.KeepRunning() {
		c.Log().Warn("timer callback is KeepRunning; chain maybe never stop")
	}

	if !timer.ErrorStop() {
		c.Log().Warn("timer callback is not ErrorStop; chain maybe never stop")
	}

	p := fmt.Sprintf("%p", timer)

	c.Lock()
	defer c.Unlock()

	if _, found := c.timers[p]; found {
		return TimerCallbackChainAlreadyAddedError.AppendMessage("timer=%v", p)
	}

	if l, ok := timer.(Loggerable); ok {
		var timerID string
		logContext := c.LogContext()
		for i, v := range logContext {
			if v != "timer-id" {
				continue
			}
			timerID = logContext[i+1].(string)
		}

		l.SetLogContext("from", timerID)
	}

	c.timers[p] = timer
	c.seq = append(c.seq, p)

	return nil
}

type MultiCallbackChain struct {
	sync.RWMutex
	stopOnce sync.Once
	*Logger
	stopped     bool
	errorStop   bool
	synchronous bool
	timers      []TimerCallback
}

func NewMultiCallbackChain(timers ...TimerCallback) *MultiCallbackChain {
	var ts []TimerCallback
	timerID := RandomUUID()
	for _, timer := range timers {
		if l, ok := timer.(Loggerable); ok {
			l.SetLogContext("from", timerID)
		}

		ts = append(ts, timer)
	}

	return &MultiCallbackChain{
		Logger: NewLogger(log, "timer-id", timerID),
		timers: ts,
	}
}

func (m *MultiCallbackChain) Start() error {
	m.Log().Debug(
		"trying to start multiple timer callback chain",
		"initial-timers", len(m.timers),
		"synchronous", m.synchronous,
	)

	m.RLock()
	stopped := m.stopped
	m.RUnlock()

	if stopped {
		return StartStopperAlreadyStartedError
	}

	m.Lock()
	m.stopped = false
	m.Unlock()

	m.Log().Debug("multiple timer callback chain started")

	if m.synchronous {
		defer func() {
			_ = m.Stop()
		}()
		return m.run()
	}

	go func() {
		_ = m.run()
	}()

	return nil
}

func (m *MultiCallbackChain) Stop() error {
	m.stopOnce.Do(func() {
		m.Lock()
		defer m.Unlock()

		if m.stopped {
			return
		}

		m.Log().Debug("trying to stop multiple timer callback chain")

		for _, timer := range m.timers {
			if err := timer.Stop(); err != nil {
				var errorLog func(string, ...interface{})
				if l, ok := timer.(Loggerable); ok {
					errorLog = l.Log().Debug
				} else {
					errorLog = m.Log().Debug
				}

				errorLog("failed to stop timer in multiple timer callback", "error", err)
			}
		}

		m.stopped = true

		m.Log().Debug("multiple timer callback chain stopped")
	})

	return nil
}

func (m *MultiCallbackChain) ErrorStop() bool {
	return m.errorStop
}

func (m *MultiCallbackChain) KeepRunning() bool {
	return true
}

func (m *MultiCallbackChain) SetErrorStop(stop bool) error {
	m.Lock()
	defer m.Unlock()

	if m.stopped {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set errorStop")
	}

	m.errorStop = stop

	return nil
}

func (m *MultiCallbackChain) Synchronous() bool {
	m.RLock()
	defer m.RUnlock()

	return m.synchronous
}

func (m *MultiCallbackChain) SetSynchronous(synchronous bool) error {
	m.Lock()
	defer m.Unlock()

	if m.stopped {
		return TimerCallbackAlreadyStartedError.AppendMessage("can't set synchronous")
	}

	m.synchronous = synchronous

	return nil
}

func (m *MultiCallbackChain) run() error {
	m.RLock()
	defer m.RUnlock()

	if m.synchronous {
		for _, timer := range m.timers {
			_ = timer.SetSynchronous(true)
			_ = timer.SetErrorStop(m.errorStop)
			if err := timer.Start(); err != nil {
				return err
			}
		}

		return nil
	}

	for _, timer := range m.timers {
		_ = timer.SetErrorStop(m.errorStop)
		if err := timer.Start(); err != nil {
			return err
		}
	}

	return nil
}
