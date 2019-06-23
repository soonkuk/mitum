package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

// Voting manages voting process; it will block the vote one by one.
type Voting struct {
	sync.RWMutex
	*common.Logger
	ballotBox *BallotBox
	daemon    *common.ReaderDaemon
	receiver  chan interface{}
	threshold *Threshold
}

func NewVoting(threshold *Threshold) *Voting {
	return &Voting{
		Logger:    common.NewLogger(Log(), "module", "voting"),
		ballotBox: NewBallotBox(),
		daemon:    common.NewReaderDaemon(true),
		threshold: threshold,
	}
}

func (vt *Voting) Start() error {
	vt.RLock()
	defer vt.RUnlock()

	if !vt.daemon.IsStopped() {
		return common.DaemonAleadyStartedError.Newf("Voting is already running; daemon is still running")
	} else if vt.receiver != nil {
		return common.DaemonAleadyStartedError.Newf("Voting is already running; receiver is still not closed")
	}

	if err := vt.daemon.Start(); err != nil {
		return err
	}

	vt.receiver = make(chan interface{})
	vt.daemon.SetReader(vt.receiver)
	vt.daemon.SetReaderCallback(vt.voteCallback)

	return nil
}

func (vt *Voting) Stop() error {
	vt.Lock()
	defer vt.Unlock()

	if err := vt.daemon.Stop(); err != nil {
		return err
	}

	close(vt.receiver)
	vt.receiver = nil

	return nil
}

func (vt *Voting) Vote(ballot Ballot) {
	go func() {
		vt.receiver <- ballot
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
	log_.Debug("trying to vote", "ballot-seal", ballot)

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

	switch vr.Result() {
	case NotYetMajority, FinishedGotMajority:
		return nil
	case JustDraw: // NOTE drawed, move to next round
		if err := vt.ballotBox.CloseVoteRecords(vrs.Hash()); err != nil {
			vt.Log().Error("failed to close VoteRecords", "error", err)
		}
		go vt.moveToNextRound(vr)
	case GotMajority: // NOTE move to next stage
		go vt.moveToNextStage(vr)
	}

	return nil
}

func (vt *Voting) moveToNextRound(vr VoteResult) error {
	// TODO create new ballot; vr.Height(), vr.Round() + 1
	return nil
}

func (vt *Voting) moveToNextStage(vr VoteResult) error {
	// TODO create new ballot; vr.Stage().Next()
	return nil
}
