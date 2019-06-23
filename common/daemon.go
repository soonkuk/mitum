package common

import (
	"sync"
)

const (
	DaemonAleadyStartedErrorCode ErrorCode = iota + 1
)

var (
	DaemonAleadyStartedError = NewError("daemon", DaemonAleadyStartedErrorCode, "daemon already started")
)

type ReaderDaemon struct {
	sync.RWMutex
	*Logger

	stopOnce       *sync.Once
	synchronous    bool
	stop           chan struct{}
	reader         chan interface{}
	readerCallback func(interface{}) error
	errCallback    func(error)
}

func NewReaderDaemon(synchronous bool) *ReaderDaemon {
	return &ReaderDaemon{
		Logger:      NewLogger(log, "module", "reader-daemon"),
		synchronous: synchronous,
	}
}

func (d *ReaderDaemon) SetReader(reader chan interface{}) *ReaderDaemon {
	d.Lock()
	defer d.Unlock()

	d.reader = reader

	return d
}

func (d *ReaderDaemon) SetReaderCallback(readerCallback func(interface{}) error) *ReaderDaemon {
	d.Lock()
	defer d.Unlock()

	d.readerCallback = readerCallback

	return d
}

func (d *ReaderDaemon) SetErrCallback(errCallback func(error)) *ReaderDaemon {
	d.Lock()
	defer d.Unlock()

	d.errCallback = errCallback

	return d
}

func (d *ReaderDaemon) Start() error {
	d.Lock()
	defer d.Unlock()

	if d.stop != nil {
		return DaemonAleadyStartedError
	}

	d.stop = make(chan struct{}, 10)
	d.stopOnce = new(sync.Once)

	go d.loop()

	return nil
}

func (d *ReaderDaemon) Stop() error {
	d.stopOnce.Do(func() {
		d.Lock()
		defer d.Unlock()

		if d.stop == nil {
			return
		}

		d.stop <- struct{}{}
		close(d.stop)
	})

	return nil
}

func (d *ReaderDaemon) IsStopped() bool {
	d.RLock()
	defer d.RUnlock()

	return d.stop == nil
}

func (d *ReaderDaemon) loop() {
end:
	for {
		select {
		case <-d.stop:
			break end
		case v, notClosed := <-d.reader:
			if !notClosed {
				break end
			}

			if d.synchronous {
				d.runCallback(v)
			} else {
				go d.runCallback(v)
			}
		}
	}

	d.Lock()
	defer d.Unlock()

	d.stop = nil
	d.stopOnce = nil
}

func (d *ReaderDaemon) runCallback(v interface{}) {
	d.RLock()
	defer d.RUnlock()

	if d.readerCallback == nil {
		return
	}

	err := d.readerCallback(v)
	if err != nil {
		d.Log().Error("error occurred", "error", err)
	}
	if err != nil && d.errCallback != nil {
		go d.errCallback(err)
	}
}
