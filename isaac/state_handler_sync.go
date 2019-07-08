package isaac

import (
	"context"
	"sync"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type SyncStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	*common.Logger
	homeState     *HomeState
	suffrage      Suffrage
	policy        Policy
	networkClient NetworkClient
	chanState     chan<- context.Context
	ctx           context.Context
}

func NewSyncStateHandler(
	homeState *HomeState,
	suffrage Suffrage,
	policy Policy,
	networkClient NetworkClient,
	chanState chan<- context.Context,
) *SyncStateHandler {
	ss := &SyncStateHandler{
		Logger: common.NewLogger(
			Log(),
			"module", "sync-state-handler",
			"state", node.StateSync,
		),
		homeState:     homeState,
		suffrage:      suffrage,
		policy:        policy,
		networkClient: networkClient,
		chanState:     chanState,
	}

	ss.ReaderDaemon = common.NewReaderDaemon(true, ss.receive)

	return ss
}

func (ss *SyncStateHandler) StartWithContext(ctx context.Context) error {
	ss.Lock()
	ss.ctx = ctx
	ss.Unlock()

	return ss.Start()
}

func (ss *SyncStateHandler) Start() error {
	if err := ss.ReaderDaemon.Start(); err != nil {
		return err
	}

	if err := ss.start(); err != nil {
		return err
	}

	ss.Log().Debug("SyncStateHandler is started")

	return nil
}

func (ss *SyncStateHandler) Stop() error {
	if err := ss.ReaderDaemon.Stop(); err != nil {
		return err
	}

	ss.Log().Debug("SyncStateHandler is stopped")

	return nil
}

func (ss *SyncStateHandler) start() error {
	// TODO
	// - request BlockProof to the last

	// TODO remove
	go func() {
		<-time.After(time.Millisecond * 100)
		ss.chanState <- common.SetContext(nil, "state", node.StateJoin)
	}()

	return nil
}

func (ss *SyncStateHandler) State() node.State {
	return node.StateSync
}

func (ss *SyncStateHandler) receive(v interface{}) error {
	ss.Log().Debug("received", "v", v)

	return nil
}
