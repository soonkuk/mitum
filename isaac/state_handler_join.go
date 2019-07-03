package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type StateJoinHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	*common.Logger
	threshold *Threshold
	homeState *HomeState
	chanState chan<- node.State
	ballotbox *Ballotbox
}

func NewStateJoinHandler(threshold *Threshold, homeState *HomeState, ballotbox *Ballotbox, chanState chan<- node.State) *StateJoinHandler {
	sh := &StateJoinHandler{
		Logger:    common.NewLogger(Log(), "module", "join-state-handler", "state", node.StateJoin),
		threshold: threshold,
		homeState: homeState,
		chanState: chanState,
		ballotbox: ballotbox,
	}

	sh.ReaderDaemon = common.NewReaderDaemon(true, sh.receiveSeal)

	return sh
}

func (sh *StateJoinHandler) Start() error {
	if err := sh.ReaderDaemon.Start(); err != nil {
		return err
	}

	go sh.start()

	return nil
}

func (sh *StateJoinHandler) State() node.State {
	return node.StateJoin
}

func (sh *StateJoinHandler) start() {
	// TODO check last block
	sh.Log().Debug("trying to check last block")

	if err := sh.check(); err != nil {
		sh.Log().Error("failed to check", "error", err)
	}
}

func (sh *StateJoinHandler) check() error {
	// TODO request BlockProof to active suffrage members of the last block
	sh.Log().Debug("trying to request BlockProof")

	// TODO if block of BlockProof is higher than homeState, go to sync

	// TODO wait INIT ballots from active suffrage group
	sh.Log().Debug("wait INIT ballot from active suffrage group")

	// TODO if waiting ACCEPT is timeouted, check again
	sh.Log().Debug("waiting INIT ballot timeout")

	return nil
}

func (sh *StateJoinHandler) receiveSeal(v interface{}) error {
	sl, ok := v.(seal.Seal)
	if !ok {
		return xerrors.Errorf("not Seal")
	}

	switch t := sl.Type(); t {
	case BallotType:
	case ProposalType:
	default:
		return xerrors.Errorf("not available seal type in JOIN state; type=%q", t)
	}

	return nil
}

/*
func (sh *StateJoinHandler) ReceiveVoteResult(vr VoteResult) error {
	sh.Log().Debug("got vote result", "result", vr)

	if vr.Stage() != StageINIT {
		sh.Log().Debug("joining state will only wait INIT ballot", "stage", vr.Stage())
		return nil
	}

	// home will wait the agreed INIT ballots of next block
	switch vr.Height().Cmp(sh.homeState.Height()) {
	case -1, 0: // same or lower than home, go to sync
		sh.Log().Debug(
			"agreed height is same or lower than home",
			"result", vr.Height(),
			"home", sh.homeState.Height(),
		)
		sh.chanState <- node.StateSync
		return nil
	}

	if vr.Height().Cmp(sh.homeState.Height().Add(1)) > 0 { // higher than next block
		sh.Log().Debug(
			"agreed height is higher than home",
			"result", vr.Height(),
			"home", sh.homeState.Height(),
		)
		sh.chanState <- node.StateSync
		return nil
	}

	// TODO prepare to store next block of current height
	// 1. request BP to active suffrage members
	// 1. wait Proposal
	// 1. voting

	return nil
}
*/
