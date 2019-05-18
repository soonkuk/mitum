package common

import (
	"sync"
	"time"
)

type CallbackTimer struct {
	sync.RWMutex
	*Logger
	timeout     time.Duration
	stopChan    chan bool
	callback    func() error
	keepRunning bool
}

func NewCallbackTimer(timeout time.Duration, callback func() error, keepRunning bool) *CallbackTimer {
	return &CallbackTimer{
		Logger:      NewLogger(log),
		callback:    callback,
		timeout:     timeout,
		keepRunning: keepRunning,
	}
}

func (c *CallbackTimer) Start() error {
	if c.stopChan != nil {
		return StartStopperAlreadyStartedError
	}

	c.Lock()
	defer c.Unlock()

	c.stopChan = make(chan bool)

	go c.waiting()

	c.Log().Debug("timer started", "timeout", c.timeout, "keep-running", c.keepRunning)

	return nil
}

func (c *CallbackTimer) Stop() error {
	c.Lock()
	defer c.Unlock()

	if c.stopChan == nil {
		return nil
	}

	c.stopChan <- true
	close(c.stopChan)
	c.stopChan = nil

	return nil
}

func (c *CallbackTimer) waiting() {
end:
	for {
		select {
		case <-c.stopChan:
			c.Log().Debug("timer stopped")
			break end
		case <-time.After(c.timeout):
			c.Log().Debug("timer expired")
			if err := c.callback(); err != nil {
				c.Log().Error("failed to callback", "error", err)
			}
			if !c.keepRunning {
				c.Log().Debug("callback called and stopped")
				defer func() {
					_ = c.Stop()
				}()
			}
		}
	}
}
