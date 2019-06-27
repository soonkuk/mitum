package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
	"golang.org/x/xerrors"
)

// Consensus manages voting process; it will block the vote one by one.
type Consensus struct {
	sync.RWMutex
	*common.Logger
	*common.ReaderDaemon
	threshold    *Threshold
	homeState    *HomeState
	stateHandler StateHandler
	chanState    chan node.State
}

func NewConsensus(homeState *HomeState, threshold *Threshold) *Consensus {
	cs := &Consensus{
		Logger:    common.NewLogger(Log(), "module", "consensus"),
		threshold: threshold,
		homeState: homeState,
		chanState: make(chan node.State),
	}

	cs.ReaderDaemon = common.NewReaderDaemon(true, cs.sealCallback)

	return cs
}

func (cs *Consensus) Start() error {
	cs.RLock()
	defer cs.RUnlock()

	if !cs.IsStopped() {
		return common.DaemonAleadyStartedError.Newf("Consensus is already running; daemon is still running")
	}

	if err := cs.ReaderDaemon.Start(); err != nil {
		return err
	}

	go func() {
	end:
		for {
			select {
			case nextState := <-cs.chanState:
				if err := cs.runState(nextState); err != nil {
					cs.Log().Error("failed state transition", "current", cs.homeState.State(), "next", nextState)
				}
			default:
				if cs.ReaderDaemon.IsStopped() {
					break end
				}
			}
		}
	}()

	if err := cs.run(); err != nil {
		return err
	}

	return nil
}

func (cs *Consensus) sealCallback(message interface{}) error {
	sl, ok := message.(seal.Seal)
	if !ok {
		cs.Log().Debug("is not seal", "message", message)
	}

	if cs.stateHandler == nil {
		return xerrors.Errorf("something wrong; stateHandler is nil")
	}

	var accepted bool
	for _, t := range cs.stateHandler.AcceptSealTypes() {
		if sl.Type().Equal(t) {
			accepted = true
			break
		}
	}
	if !accepted {
		cs.Log().Debug(
			"not accepted seal found in this state",
			"state", cs.stateHandler.State(),
			"seal_type", sl.Type(),
			"accepted", cs.stateHandler.AcceptSealTypes(),
		)
		return nil
	}

	cs.stateHandler.ReceiveSeal(sl)

	return nil
}

func (cs *Consensus) run() error {
	var nextState node.State
	switch cs.homeState.State() {
	case node.StateBooting:
		nextState = node.StateBooting
	case node.StateJoin:
		nextState = node.StateJoin
	case node.StateConsensus:
		nextState = node.StateConsensus
	case node.StateSync:
		nextState = node.StateSync
	default:
		return xerrors.Errorf("nothing to do in this state, %q", cs.homeState.State())
	}

	cs.chanState <- nextState

	return nil
}

func (cs *Consensus) runState(state node.State) error {
	cs.Log().Debug("trying state transition", "current", cs.homeState.State(), "next", state)
	if cs.stateHandler != nil {
		if err := cs.stateHandler.Stop(); err != nil {
			return err
		}
	} else if cs.stateHandler.State() == state {
		return xerrors.Errorf(
			"same stateHandler is already running; handler state=%q next state=%q",
			cs.stateHandler.State(),
			state,
		)
	}

	switch state {
	case node.StateJoin:
		return cs.startStateJoin()
	case node.StateConsensus:
		return cs.startStateConsensus()
	case node.StateSync:
		return cs.startStateSync()
	default:
		return xerrors.Errorf("no registered state handler; state=%q", state)
	}
}

func (cs *Consensus) startStateBooting() error {
	cs.homeState.SetState(node.StateBooting)

	return nil
}

func (cs *Consensus) startStateJoin() error {
	cs.homeState.SetState(node.StateJoin)

	if cs.stateHandler != nil {
		if err := cs.stateHandler.Stop(); err != nil {
			return err
		}
	}

	cs.stateHandler = NewStateJoinHandler(cs.threshold, cs.homeState, cs.chanState)
	return cs.stateHandler.Start()
}

func (cs *Consensus) startStateConsensus() error {
	cs.homeState.SetState(node.StateConsensus)

	return nil
}

func (cs *Consensus) startStateSync() error {
	cs.homeState.SetState(node.StateSync)

	return nil
}
