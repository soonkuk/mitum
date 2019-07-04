package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

// StateTransition manages consensus process by node state
type StateTransition struct {
	sync.RWMutex
	*common.Logger
	*common.ReaderDaemon
	homeState     *HomeState
	voteCompiler  *VoteCompiler
	chanState     chan node.State
	stateHandler  StateHandler
	stateHandlers map[node.State]StateHandler
}

func NewStateTransition(homeState *HomeState, voteCompiler *VoteCompiler) *StateTransition {
	cs := &StateTransition{
		Logger:        common.NewLogger(Log(), "module", "state-transition"),
		homeState:     homeState,
		chanState:     make(chan node.State),
		voteCompiler:  voteCompiler,
		stateHandlers: map[node.State]StateHandler{},
	}

	cs.ReaderDaemon = common.NewReaderDaemon(true, cs.receiveFromVoteCompiler)

	return cs
}

func (cs *StateTransition) Start() error {
	if err := cs.ReaderDaemon.Start(); err != nil {
		return err
	}

	cs.voteCompiler.RegisterCallback(
		"state-transition",
		func(v interface{}) error {
			wrote := cs.Write(v)
			cs.Log().Debug("sent VoteCompiler result to state handler", "wrote", wrote)

			return nil
		},
	)

	go func() {
	end:
		for {
			select {
			case nextState := <-cs.chanState:
				go func(nextState node.State) {
					if err := cs.runState(nextState); err != nil {
						cs.Log().Error(
							"failed state transition",
							"current", cs.homeState.State(),
							"next", nextState,
						)
					}
				}(nextState)
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

func (cs *StateTransition) ChanState() chan<- node.State {
	return cs.chanState
}

func (cs *StateTransition) SetStateHandler(stateHandler StateHandler) error {
	cs.Lock()
	defer cs.Unlock()

	if _, found := cs.stateHandlers[stateHandler.State()]; found {
		return xerrors.Errorf("StateHandler already registered; state=%q", stateHandler.State())
	}

	cs.stateHandlers[stateHandler.State()] = stateHandler

	return nil
}

func (cs *StateTransition) receiveFromVoteCompiler(v interface{}) error {
	if cs.stateHandler == nil {
		return xerrors.Errorf("something wrong; stateHandler is nil")
	}

	if !cs.stateHandler.Write(v) {
		return xerrors.Errorf("failed to write seal")
	}

	return nil
}

func (cs *StateTransition) runState(state node.State) error {
	if err := state.IsValid(); err != nil {
		return err
	}

	cs.Log().Debug("trying state transition", "current", cs.homeState.State(), "next", state)

	cs.Lock()
	defer cs.Unlock()

	if cs.stateHandler.State() == state {
		return xerrors.Errorf(
			"same stateHandler is already running; handler state=%q next state=%q",
			cs.stateHandler.State(),
			state,
		)
	}

	if cs.stateHandler != nil {
		if err := cs.stateHandler.Stop(); err != nil {
			return err
		}
	}

	stateHandler, found := cs.stateHandlers[state]
	if !found {
		return xerrors.Errorf("stateHandler not registered yet; state=%q", state)
	}

	cs.stateHandler = stateHandler
	if err := cs.stateHandler.Start(); err != nil {
		cs.Log().Error("failed to start stateHandler", "state", state, "error", err)
		return err
	}
	cs.homeState.SetState(state)

	return nil
}
