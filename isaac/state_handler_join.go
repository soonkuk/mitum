package isaac

import (
	"context"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type JoinStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	homeState     *HomeState
	policy        Policy
	networkClient NetworkClient
	chanState     chan<- context.Context
	ctx           context.Context
	timer         common.Timer
}

func NewJoinStateHandler(
	homeState *HomeState,
	policy Policy,
	networkClient NetworkClient,
	chanState chan<- context.Context,
) *JoinStateHandler {
	js := &JoinStateHandler{
		homeState:     homeState,
		policy:        policy,
		networkClient: networkClient,
		chanState:     chanState,
	}

	js.ReaderDaemon = common.NewReaderDaemon(false, 1000, js.receive)
	js.ReaderDaemon.Logger = common.NewLogger(
		Log(),
		"module", "join-state-handler",
		"state", node.StateJoin,
	)

	return js
}

func (js *JoinStateHandler) StartWithContext(ctx context.Context) error {
	js.Lock()
	js.ctx = ctx
	js.Unlock()

	return js.Start()
}

func (js *JoinStateHandler) Start() error {
	if err := js.ReaderDaemon.Start(); err != nil {
		return err
	}

	if err := js.start(); err != nil {
		return err
	}

	js.Log().Debug("JoinStateHandler is started")

	return nil
}

func (js *JoinStateHandler) Stop() error {
	if err := js.ReaderDaemon.Stop(); err != nil {
		return err
	}

	js.Lock()
	defer js.Unlock()

	if js.timer != nil {
		if err := js.timer.Stop(); err != nil {
			return err
		}
		js.timer = nil
	}

	js.Log().Debug("JoinStateHandler is stopped")
	return nil
}

func (js *JoinStateHandler) start() error {
	// TODO
	// - basically after sync, join will start
	// - wait INIT VoteResult for current height
	// - store next block with VoteResult.Proposal()
	// - process Proposal of next block
	// - follow next VoteResults
	// - after ACCEPT VoteResult, change to ConsensusStateHandler

	if js.timer != nil {
		if err := js.timer.Stop(); err != nil {
			return err
		}
	}

	// start timer for INITBallot
	js.timer = common.NewCallbackTimer(
		"broadcast-init-ballot-in-join",
		js.policy.IntervalINITBallotOfJoin,
		js.broadcastINITBallot,
	)
	js.timer.(*common.CallbackTimer).SetIntervalFunc(func(count uint, elapsed time.Duration) time.Duration {
		if count == 0 {
			return time.Second * 0
		}

		return js.policy.IntervalINITBallotOfJoin
	})
	js.timer.(*common.CallbackTimer).SetLogContext(js.LogContext())

	return js.timer.Start()
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
		js.Log().Debug("got majority", "vr", vr)
		if err := js.gotMajority(vr); err != nil {
			return err
		}
	}

	return nil
}

func (js *JoinStateHandler) receiveProposal(proposal Proposal) error {
	// TODO process proposal

	return nil
}

func (js *JoinStateHandler) gotMajority(vr VoteResult) error {
	switch stage := vr.Stage(); stage {
	case StageINIT:
		return js.stageINIT(vr)
	case StageACCEPT:
		return js.stageACCEPT(vr)
	default:
		return nil
	}
}

func (js *JoinStateHandler) stageINIT(vr VoteResult) error {
	// TODO checks,
	// - VoteResult.Height() is same with homeState.Height()
	// - VoteResult.Block() is same with homeState.Block().Hash()
	// - VoteResult.Round() is not important :)

	checker := common.NewChainChecker(
		"showme-checker",
		context.Background(),
		CheckerVoteResult,
		CheckerVoteResultINIT,
	)
	_ = checker.SetLogContext(js.LogContext())
	_ = checker.SetContext(
		"homeState", js.homeState,
		"vr", vr,
	)

	err := checker.Check()
	if err != nil {
		if xerrors.Is(err, ChangeNodeStateToSyncError) {
			js.chanState <- common.SetContext(context.TODO(), "state", node.StateSync)
			return nil
		}
		return err
	}

	// store new block
	var heightDiff int
	if err := checker.ContextValue("heightDiff", &heightDiff); err != nil {
		return err
	}

	if heightDiff == -1 { // next height did not processed, go to consensus
		js.chanState <- common.SetContext(
			context.TODO(),
			"state", node.StateConsensus,
			"vr", vr,
		)
		return nil
	} else if heightDiff == 0 {
		// TODO process vr.Proposal()
		// TODO store next block
		nextHeight := vr.Height().Add(1)
		nextBlock, err := NewBlock(nextHeight, vr.Round(), vr.Proposal())
		if err != nil {
			return err
		}

		_ = js.homeState.SetBlock(nextBlock)

		js.Log().Debug(
			"new block created",
			"previous_height", vr.Height(),
			"previous_block", vr.CurrentBlock(),
			"previous_round", vr.Round(),
			"next_height", nextHeight,
			"next_block", vr.NextBlock(),
			"next_round", vr.Round()+1,
			"new_block", nextBlock,
		)
	} else {
		js.Log().Debug("already known block; just ignore it", "diff", heightDiff)
	}

	return nil
}

func (js *JoinStateHandler) stageACCEPT(vr VoteResult) error {
	js.Log().Debug("got accept VoteResult", "vr", vr)

	checker := common.NewChainChecker(
		"showme-checker",
		context.Background(),
		CheckerVoteResult,
		CheckerVoteResultOtherStage,
	)
	_ = checker.SetContext(
		"homeState", js.homeState,
		"vr", vr,
	)

	if err := checker.Check(); err != nil {
		return err
	}

	js.chanState <- common.SetContext(
		context.TODO(),
		"state", node.StateConsensus,
		"vr", vr,
	)

	return nil
}

func (js *JoinStateHandler) broadcastINITBallot(timer common.Timer) error {
	if js.networkClient == nil {
		return xerrors.Errorf("network client is missing")
	}

	t := timer.(*common.CallbackTimer)
	js.Log().Debug(
		"broadcast INITBallot for current block",
		"interval", js.policy.IntervalINITBallotOfJoin,
		"run_count", t.RunCount(),
		"round", js.homeState.Block().Round()+1,
	)

	ballot, err := NewBallot(
		js.homeState.Home().Address(),
		js.homeState.PreviousBlock().Height(),
		js.homeState.Block().Round()+1,
		StageINIT,
		js.homeState.Proposal(),
		js.homeState.PreviousBlock().Hash(),
		js.homeState.Block().Hash(),
	)
	if err != nil {
		return err
	}

	if err := js.networkClient.Vote(&ballot); err != nil {
		return err
	}

	return nil
}
