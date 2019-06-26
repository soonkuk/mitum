package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

// Voting manages voting process; it will block the vote one by one.
type Voting struct {
	sync.RWMutex
	*common.Logger
	ballotBox *BallotBox
	daemon    *common.ReaderDaemon
	receiver  chan interface{}
	threshold *Threshold
	homeState *HomeState
}

func NewVoting(homeState *HomeState, threshold *Threshold) *Voting {
	vt := &Voting{
		Logger:    common.NewLogger(Log(), "module", "voting"),
		ballotBox: NewBallotBox(),
		threshold: threshold,
		homeState: homeState,
	}

	vt.daemon = common.NewReaderDaemon(true, vt.voteCallback)

	return vt
}

func (vt *Voting) Start() error {
	vt.RLock()
	defer vt.RUnlock()

	if !vt.daemon.IsStopped() {
		return common.DaemonAleadyStartedError.Newf("Voting is already running; daemon is still running")
	}

	if err := vt.daemon.Start(); err != nil {
		return err
	}

	return nil
}

func (vt *Voting) Stop() error {
	vt.Lock()
	defer vt.Unlock()

	if err := vt.daemon.Stop(); err != nil {
		return err
	}

	return nil
}

func (vt *Voting) Vote(ballot Ballot) {
	go func() {
		vt.daemon.Write(ballot)
	}()
}

func (vt *Voting) voteCallback(v interface{}) error {
	if v == nil {
		return xerrors.Errorf("Voting.receiver is already closed")
	}

	ballot, ok := v.(Ballot)
	if !ok {
		return xerrors.Errorf("invalid input for Voting.vote; it should be Ballot")
	}

	log_ := vt.Log().New(log15.Ctx{"ballot": ballot.Hash()})
	log_.Debug("trying to vote", "ballot", ballot)

	vrs, err := vt.ballotBox.Vote(
		ballot.Node(),
		ballot.Height(),
		ballot.Round(),
		ballot.Stage(),
		ballot.Proposal(),
		ballot.NextBlock(),
		ballot.Hash(),
	)
	if err != nil {
		return err
	}
	if vrs.IsClosed() {
		log_.Debug("closed", "records", vrs)
		return nil
	}

	var total, threshold uint = vt.threshold.Get(vrs.stage)
	vr, err := vrs.CheckMajority(total, threshold)
	if err != nil {
		return err
	}

	vt.Log().Debug("got vote result", "result", vr)

	// NOTE set node state to StateConsensus
	if vr.Result() != NotYetMajority && vt.homeState.State() != node.StateConsensus {
		if vt.homeState.State() != node.StateSync {
			vt.homeState.SetState(node.StateConsensus)
		}
	}

	switch vr.Result() {
	case NotYetMajority, FinishedGotMajority:
		return nil
	case JustDraw: // NOTE drawed, move to next round
		if err := vt.ballotBox.CloseVoteRecords(vrs.Hash()); err != nil {
			vt.Log().Error("failed to close VoteRecords", "error", err)
		}

		go vt.moveToNextRound(vr)
	case GotMajority: // NOTE move to next stage
		// NOTE if nextBlock is different from current node, go to sync
		homeVote, isVoted := vr.records.NodeVote(vt.homeState.Home().Address())
		if isVoted && !homeVote.nextBlock.Equal(vr.NextBlock()) {
			vt.Log().Debug(
				"nextblock of result is different from home vote",
				"home", homeVote,
				"result", vr,
			)
			go vt.moveToSync(vr)
			return nil
		}

		go vt.moveToNextStage(vr)
	}

	return nil
}

func (vt *Voting) moveToNextRound(vr VoteResult) error {
	if vt.homeState.State() == node.StateSync {
		vt.Log().Debug("move to next round, but home state is StateSync")
		return nil
	}

	// TODO create new ballot; vr.Height(), vr.Round() + 1
	return nil
}

func (vt *Voting) moveToNextStage(vr VoteResult) error {
	if vt.homeState.State() == node.StateSync {
		vt.Log().Debug("move to next stage, but home state is StateSync")
		return nil
	}

	// TODO create new ballot; vr.Stage().Next()
	return nil
}

func (vt *Voting) moveToSync(vr VoteResult) error {
	// TODO sync
	vt.homeState.SetState(node.StateSync)

	return nil
}
