package isaac

import (
	"context"
	"sync"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type StoppedStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	homeState *HomeState
	ctx       context.Context
}

func NewStoppedStateHandler(homeState *HomeState) *StoppedStateHandler {
	ss := &StoppedStateHandler{
		homeState: homeState,
	}

	ss.ReaderDaemon = common.NewReaderDaemon(false, 0, nil)
	ss.ReaderDaemon.Logger = common.NewLogger(
		Log(),
		"module", "stopped-state-handler",
		"state", node.StateStopped,
	)

	return ss
}

func (ss *StoppedStateHandler) StartWithContext(ctx context.Context) error {
	ss.Lock()
	ss.ctx = ctx
	ss.Unlock()

	return ss.Start()
}

func (ss *StoppedStateHandler) State() node.State {
	return node.StateStopped
}
