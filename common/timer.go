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
	c.Log().Debug(
		"trying to start callback-timer",
		"timeout", c.timeout,
		"keep-running", c.keepRunning,
	)

	if c.stopChan != nil {
		return StartStopperAlreadyStartedError
	}

	c.Lock()
	defer c.Unlock()

	c.stopChan = make(chan bool)

	go c.waiting()

	c.Log().Debug("callback-timer started")

	return nil
}

func (c *CallbackTimer) Stop() error {
	c.Lock()
	defer c.Unlock()

	c.Log().Debug("trying to stop callback-timer")

	if c.stopChan != nil {
		c.stopChan <- true
		close(c.stopChan)
		c.stopChan = nil
	}

	return nil
}

func (c *CallbackTimer) waiting() {
	if !c.keepRunning {
		select {
		case <-c.stopChan:
			c.Log().Debug("callback-timer stopped")
			break
		case <-time.After(c.timeout):
			c.Log().Debug("timeout expired")
			if err := c.callback(); err != nil {
				c.Log().Error("failed to callback", "error", err)
			}
		}

		c.Lock()
		if c.stopChan != nil {
			close(c.stopChan)
			c.stopChan = nil
		}
		c.Unlock()
		return
	}

end:
	for {
		select {
		case <-c.stopChan:
			c.Log().Debug("callback-timer stopped")
			break end
		case <-time.After(c.timeout):
			c.Log().Debug("timer expired")
			if c.keepRunning {
				if err := c.callback(); err != nil {
					c.Log().Error("failed to callback", "error", err)
				}
				continue
			}
			break end
		}
	}
}
