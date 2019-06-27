package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
	"github.com/spikeekips/mitum/transaction"
)

type StateHandler interface {
	common.Daemon
	State() node.State
	ReceiveSeal(seal.Seal)
	AcceptSealTypes() []common.DataType
}

type StateJoinHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	*common.Logger
	threshold      *Threshold
	homeState      *HomeState
	chanState      chan<- node.State
	ballotCompiler *BallotCompiler
}

func NewStateJoinHandler(threshold *Threshold, homeState *HomeState, chanState chan<- node.State) *StateJoinHandler {
	return &StateJoinHandler{
		Logger:         common.NewLogger(Log(), "module", "join-state-handler", "state", "join"),
		threshold:      threshold,
		homeState:      homeState,
		chanState:      chanState,
		ballotCompiler: NewBallotCompiler(threshold),
	}
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

func (sh *StateJoinHandler) AcceptSealTypes() []common.DataType {
	return []common.DataType{
		BallotType,
		transaction.TransactionType,
	}
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

	// TODO round from BlockProof
	round := Round(0)

	// TODO if block of BlockProof is same, keeps to broadcast INIT ballot(round 0)
	initBallot, err := NewINITBallot(
		sh.homeState.Home().Address(),
		sh.homeState.Height(),
		round,
	)
	if err != nil {
		sh.Log().Error(
			"failed to make INIT ballot",
			"node", sh.homeState.Home().Address(),
			"height", sh.homeState.Height(),
			"round", round,
			"error", err,
		)

		return err
	}

	sh.Log().Debug("trying to broadcast INIT ballot", "ballot", initBallot)

	// TODO wait INIT ballots from active suffrage group
	sh.Log().Debug("wait INIT ballot from active suffrage group")

	// TODO if waiting ACCEPT is timeouted, check again
	sh.Log().Debug("waiting INIT ballot timeout")

	return nil
}

func (sh *StateJoinHandler) ReceiveSeal(sl seal.Seal) {
	switch t := sl.Type(); t {
	case BallotType:
		ballot, ok := sl.(Ballot)
		if !ok {
			sh.Log().Error("received, but not ballot")
			return
		}

		if err := sh.receiveBallot(ballot); err != nil {
			sh.Log().Error("failed to receive ballot", "ballot", ballot, "error", err)
			return
		}
	default:
		sh.Log().Error("unknown seal type found", "type", t)
		return
	}
}

func (sh *StateJoinHandler) receiveBallot(ballot Ballot) error {
	log_ := sh.Log().New(log15.Ctx{"ballot": ballot.Hash()})

	vr, err := sh.ballotCompiler.Vote(ballot)
	if err != nil {
		return err
	}

	log_.Debug("got vote result", "result", vr)

	switch vr.Result() {
	case NotYetMajority, FinishedGotMajority, JustDraw:
		return nil
		//case GotMajority:
	}

	// GotMajority
	if vr.Stage() != StageINIT {
		log_.Debug("joining state will only wait INIT ballot", "stage", vr.Stage())
		return nil
	}

	switch n := vr.Height().Cmp(sh.homeState.Height().Add(1)); n {
	case -1: // result is lower than home
		log_.Debug(
			"VoteResult.Height() is lower than home; go to sync",
			"result", vr.Height(),
			"home", sh.homeState.Height(),
			"expected", sh.homeState.Height().Add(1),
		)
		sh.chanState <- node.StateSync
		return nil
	case 1: // higher
		log_.Debug(
			"VoteResult.Height() is heigher than home; go to sync",
			"result", vr.Height(),
			"home", sh.homeState.Height(),
			"expected", sh.homeState.Height().Add(1),
		)
		sh.chanState <- node.StateSync
		return nil
	case 0:
		if !vr.CurrentBlock().Equal(sh.homeState.Block()) {
			log_.Debug(
				"VoteResult.CurrentBlock() is different from home; go to sync",
				"result", vr.CurrentBlock(),
				"home", sh.homeState.Block(),
				"expected", sh.homeState.Block(),
			)
			sh.chanState <- node.StateSync
			return nil
		}
	}

	// TODO stop broadcasting previous INIT ballot

	go sh.processProposal(vr)

	return nil
}

func (sh *StateJoinHandler) processProposal(vr VoteResult) {
	log_ := sh.Log().New(log15.Ctx{"result": vr})
	log_.Debug("trying to process proposal", "proposal", vr.Proposal())

	// get proposal

	// validate proposal

	// store proposal

	return
}
