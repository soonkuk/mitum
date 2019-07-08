package isaac

import (
	"context"
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
	chanState     chan context.Context
	stateHandler  StateHandler
	stateHandlers map[node.State]StateHandler
}

func NewStateTransition(homeState *HomeState, voteCompiler *VoteCompiler) *StateTransition {
	cs := &StateTransition{
		Logger:        common.NewLogger(Log(), "module", "state-transition"),
		homeState:     homeState,
		chanState:     make(chan context.Context),
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

	err := cs.voteCompiler.RegisterCallback(
		"state-transition",
		func(v interface{}) error {
			wrote := cs.Write(v)
			cs.Log().Debug("sent VoteCompiler result to state handler", "wrote", wrote)

			return nil
		},
	)
	if err != nil {
		cs.Log().Error("error in cs.voteCompiler.RegisterCallback", "error", err)
	}

	go func() {
	end:
		for {
			select {
			case ctx := <-cs.chanState:
				go func(ctx context.Context) {
					if err := cs.runState(ctx); err != nil {
						cs.Log().Error(
							"failed state transition",
							"current", cs.homeState.State(),
							"error", err,
						)
					}
				}(ctx)
			default:
				if cs.ReaderDaemon.IsStopped() {
					break end
				}
			}
		}
	}()

	// if err := cs.runState(cs.homeState.State()); err != nil {
	// 	return err
	// }

	return nil
}

func (cs *StateTransition) Stop() error {
	if err := cs.ReaderDaemon.Stop(); err != nil {
		return err
	}

	if cs.stateHandler != nil {
		if err := cs.stateHandler.Stop(); err != nil {
			return err
		}
	}

	if err := cs.voteCompiler.UnregisterCallback("state-transition"); err != nil {
		cs.Log().Error("error in cs.voteCompiler.UnregisterCallback", "error", err)
	}

	return nil
}

func (cs *StateTransition) StateHandler() StateHandler {
	return cs.stateHandler
}

func (cs *StateTransition) ChanState() chan<- context.Context {
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

func (cs *StateTransition) runState(ctx context.Context) error {
	if ctx == nil {
		return xerrors.Errorf("empty context")
	}

	v := ctx.Value("state")
	if v == nil {
		return xerrors.Errorf("context for chanState should have 'state' value")
	}

	nextState, ok := v.(node.State)
	if !ok {
		return xerrors.Errorf("invalid 'state' value found; state for chanState should be node.State")
	}

	if err := nextState.IsValid(); err != nil {
		return err
	}

	cs.Lock()
	defer cs.Unlock()

	if cs.stateHandler != nil && cs.stateHandler.State() == nextState {
		return xerrors.Errorf(
			"same stateHandler is already running; handler state=%q next state=%q",
			cs.stateHandler.State(),
			nextState,
		)
	}

	cs.Log().Debug("trying state transition", "current", cs.homeState.State(), "next", nextState)
	if cs.stateHandler != nil {
		if err := cs.stateHandler.Stop(); err != nil {
			return err
		}
	}

	stateHandler, found := cs.stateHandlers[nextState]
	if !found {
		return xerrors.Errorf("stateHandler not registered yet; state=%q", nextState)
	}

	cs.stateHandler = stateHandler
	if err := cs.stateHandler.StartWithContext(ctx); err != nil {
		cs.Log().Error("failed to start stateHandler", "state", nextState, "error", err)
		return err
	}
	cs.homeState.SetState(nextState)

	return nil
}
