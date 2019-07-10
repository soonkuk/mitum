package isaac

import (
	"context"
	"sync"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type BootingStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	homeState *HomeState
	chanState chan<- context.Context
	ctx       context.Context
}

func NewBootingStateHandler(
	homeState *HomeState,
	chanState chan<- context.Context,
) *BootingStateHandler {
	bs := &BootingStateHandler{
		homeState: homeState,
		chanState: chanState,
	}

	bs.ReaderDaemon = common.NewReaderDaemon(false, 0, nil)
	bs.ReaderDaemon.Logger = common.NewLogger(
		Log(),
		"module", "booting-state-handler",
		"state", node.StateBooting,
	)

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

func (bs *BootingStateHandler) StartWithContext(ctx context.Context) error {
	bs.Lock()
	bs.ctx = ctx
	bs.Unlock()

	return bs.Start()
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
		bs.chanState <- common.SetContext(context.TODO(), "state", node.StateSync)
	}()

	return nil
}
