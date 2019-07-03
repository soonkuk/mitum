package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

// StateTransition manages voting process; it will block the vote one by one.
type StateTransition struct {
	sync.RWMutex
	*common.Logger
	*common.ReaderDaemon
	threshold    *Threshold
	homeState    *HomeState
	ballotbox    *Ballotbox
	stateHandler StateHandler
	chanState    chan node.State
}

func NewStateTransition(homeState *HomeState, threshold *Threshold) *StateTransition {
	cs := &StateTransition{
		Logger:    common.NewLogger(Log(), "module", "state-transition"),
		threshold: threshold,
		homeState: homeState,
		chanState: make(chan node.State),
		ballotbox: NewBallotbox(threshold),
	}

	cs.ReaderDaemon = common.NewReaderDaemon(true, cs.sealCallback)

	return cs
}

func (cs *StateTransition) Start() error {
	cs.RLock()
	defer cs.RUnlock()

	if !cs.IsStopped() {
		return common.DaemonAleadyStartedError.Newf("StateTransition is already running; daemon is still running")
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

	if err := cs.runState(cs.homeState.State()); err != nil {
		return err
	}

	return nil
}

func (cs *StateTransition) sealCallback(message interface{}) error {
	sl, ok := message.(seal.Seal)
	if !ok {
		cs.Log().Debug("is not seal", "message", message)
	}

	if cs.stateHandler == nil {
		return xerrors.Errorf("something wrong; stateHandler is nil")
	}

	if !cs.stateHandler.Write(sl) {
		return xerrors.Errorf("failed to write seal")
	}

	return nil
}

func (cs *StateTransition) runState(state node.State) error {
	if err := state.IsValid(); err != nil {
		return err
	}

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
	case node.StateBooting:
		go cs.startStateBooting()
	case node.StateJoin:
		go cs.startStateJoin()
	case node.StateConsensus:
		go cs.startStateConsensus()
	case node.StateSync:
		go cs.startStateSync()
	}

	return nil
}

func (cs *StateTransition) startStateBooting() error {
	cs.homeState.SetState(node.StateBooting)

	return nil
}

func (cs *StateTransition) startStateJoin() error {
	cs.homeState.SetState(node.StateJoin)

	if cs.stateHandler != nil {
		if err := cs.stateHandler.Stop(); err != nil {
			return err
		}
	}

	cs.stateHandler = NewStateJoinHandler(cs.threshold, cs.homeState, cs.ballotbox, cs.chanState)
	return cs.stateHandler.Start()
}

func (cs *StateTransition) startStateConsensus() error {
	cs.homeState.SetState(node.StateConsensus)

	return nil
}

func (cs *StateTransition) startStateSync() error {
	cs.homeState.SetState(node.StateSync)

	return nil
}
