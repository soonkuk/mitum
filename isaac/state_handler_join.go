package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type JoinStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	*common.Logger
	homeState     *HomeState
	suffrage      Suffrage
	policy        Policy
	networkClient NetworkClient
	chanState     chan<- node.State
	timer         common.Timer
}

func NewJoinStateHandler(
	homeState *HomeState,
	suffrage Suffrage,
	policy Policy,
	networkClient NetworkClient,
	chanState chan<- node.State,
) *JoinStateHandler {
	js := &JoinStateHandler{
		Logger: common.NewLogger(
			Log(),
			"module", "join-state-handler",
			"state", node.StateConsensus,
		),
		homeState:     homeState,
		suffrage:      suffrage,
		policy:        policy,
		networkClient: networkClient,
		chanState:     chanState,
	}

	js.ReaderDaemon = common.NewReaderDaemon(true, js.receive)

	return js
}

func (js *JoinStateHandler) Start() error {
	if err := js.ReaderDaemon.Start(); err != nil {
		return err
	}

	if err := js.start(); err != nil {
		return err
	}

	return nil
}

func (js *JoinStateHandler) start() error {
	// TODO
	// - basically after sync, join will start
	// - wait INIT VoteResult for current height
	// - store next block with VoteResult.Proposal()
	// - process Proposal
	// - follow next VoteResults
	// - after ACCEPT VoteResult, change to ConsensusStateHandler

	return nil
}

func (js *JoinStateHandler) State() node.State {
	return node.StateJoin
}

func (js *JoinStateHandler) receive(v interface{}) error {
	js.Log().Debug("received", "v", v)
	switch v.(type) {
	case Proposal:
		if err := js.receiveProposal(v.(Proposal)); err != nil {
			return err
		}
	case VoteResult:
		if err := js.receiveVoteResult(v.(VoteResult)); err != nil {
			return err
		}
	default:
		return xerrors.Errorf("invalid seal received", "seal", v)
	}

	return nil
}

func (js *JoinStateHandler) receiveVoteResult(vr VoteResult) error {
	switch vr.Result() {
	case NotYetMajority, FinishedGotMajority:
	case JustDraw:
		js.Log().Debug("just draw, wait another INIT VoteResult", "vr", vr)
	case GotMajority:
		if err := js.gotMajority(vr); err != nil {
			return err
		}
	}

	return nil
}

func (js *JoinStateHandler) receiveProposal(proposal Proposal) error {
	return nil
}

func (js *JoinStateHandler) gotMajority(vr VoteResult) error {
	js.Log().Debug("got majority", "vr", vr)

	switch stage := vr.Stage(); stage {
	case StageINIT:
		return js.stageINIT(vr)
	default:
		return js.stageDefault(vr)
	}

	return nil
}

func (js *JoinStateHandler) stageINIT(vr VoteResult) error {
	// TODO checks,
	// - VoteResult.Height() is same with homeState.Height()
	// - VoteResult.Block() is same with homeState.Block().Hash()
	// - VoteResult.Round() is not important :)

	// store new block

	return nil
}

func (js *JoinStateHandler) stageDefault(vr VoteResult) error {
	return nil
}
