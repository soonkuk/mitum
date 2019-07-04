package isaac

import (
	"sync"

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
	chanState     chan<- node.State
}

func NewSyncStateHandler(
	homeState *HomeState,
	suffrage Suffrage,
	policy Policy,
	networkClient NetworkClient,
	chanState chan<- node.State,
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

func (ss *SyncStateHandler) Start() error {
	if err := ss.ReaderDaemon.Start(); err != nil {
		return err
	}

	if err := ss.start(); err != nil {
		return err
	}

	return nil
}

func (ss *SyncStateHandler) start() error {
	// TODO
	// request BlockProof to the last

	return nil
}

func (ss *SyncStateHandler) State() node.State {
	return node.StateSync
}

func (ss *SyncStateHandler) receive(v interface{}) error {
	ss.Log().Debug("received", "v", v)

	return nil
}
