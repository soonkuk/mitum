package isaac

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type StoppedStateHandler struct {
	*common.ReaderDaemon
	*common.Logger
	homeState *HomeState
}

func NewStoppedStateHandler(homeState *HomeState) *StoppedStateHandler {
	ss := &StoppedStateHandler{
		Logger: common.NewLogger(
			Log(),
			"module", "stopped-state-handler",
			"state", node.StateStopped,
		),
		homeState: homeState,
	}

	ss.ReaderDaemon = common.NewReaderDaemon(true, func(interface{}) error { return nil })

	return ss
}

func (ss *StoppedStateHandler) State() node.State {
	return node.StateStopped
}
