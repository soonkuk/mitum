package common

import (
	"sync"
	"time"

	"github.com/inconshreveable/log15"
)

type CallbackTimer struct {
	sync.RWMutex
	name        string
	timeout     time.Duration
	stopChan    chan bool
	callback    func() error
	keepRunning bool
	log         log15.Logger
}

func NewCallbackTimer(name string, timeout time.Duration, callback func() error, keepRunning bool) *CallbackTimer {
	return &CallbackTimer{
		name:        name,
		callback:    callback,
		timeout:     timeout,
		keepRunning: keepRunning,
		log:         log.New(log15.Ctx{"name": name}),
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

	c.log.Debug("timer started")

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

	log.Debug("timer stopped")

	return nil
}

func (c *CallbackTimer) waiting() {
end:
	for {
		select {
		case <-c.stopChan:
			c.log.Debug("timer is stopped")
			break end
		case <-time.After(c.timeout):
			c.log.Debug("timer expired")
			if err := c.callback(); err != nil {
				log.Error("failed to callback", "error", err)
			}
			if !c.keepRunning {
				c.log.Debug("callback called and stopped")
				defer c.Stop()
			}
		}
	}
}
