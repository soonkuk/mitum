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
		Logger:      &Logger{},
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

	c.Log().Debug("timer stopped")

	return nil
}

func (c *CallbackTimer) waiting() {
end:
	for {
		select {
		case <-c.stopChan:
			c.Log().Debug("timer is stopped")
			break end
		case <-time.After(c.timeout):
			c.Log().Debug("timer expired")
			if err := c.callback(); err != nil {
				log.Error("failed to callback", "error", err)
			}
			if !c.keepRunning {
				c.Log().Debug("callback called and stopped")
				defer c.Stop()
			}
		}
	}
}
