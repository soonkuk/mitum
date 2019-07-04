package isaac

import (
	"sync"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type BootingStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	*common.Logger
	homeState *HomeState
	chanState chan<- node.State
}

func NewBootingStateHandler(
	homeState *HomeState,
	chanState chan<- node.State,
) *BootingStateHandler {
	bs := &BootingStateHandler{
		Logger: common.NewLogger(
			Log(),
			"module", "booting-state-handler",
			"state", node.StateBooting,
		),
		homeState: homeState,
		chanState: chanState,
	}

	bs.ReaderDaemon = common.NewReaderDaemon(true, func(interface{}) error { return nil })

	return bs
}

func (bs *BootingStateHandler) Start() error {
	if err := bs.ReaderDaemon.Start(); err != nil {
		return err
	}

	if err := bs.start(); err != nil {
		return err
	}

	bs.Log().Debug("BootingStateHandler is started")

	return nil
}

func (bs *BootingStateHandler) Stop() error {
	if err := bs.ReaderDaemon.Stop(); err != nil {
		return err
	}

	bs.Log().Debug("BootingStateHandler is stopped")

	return nil
}

func (bs *BootingStateHandler) State() node.State {
	return node.StateBooting
}

func (bs *BootingStateHandler) start() error {
	// TODO
	// - check homeState
	// - check blocks

	// TODO remove
	go func() {
		<-time.After(time.Millisecond * 100)
		bs.chanState <- node.StateSync
	}()

	return nil
}
