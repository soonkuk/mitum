package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/seal"
)

type VoteCompilerCallback func(interface{}) error

type VoteCompiler struct {
	sync.RWMutex
	*common.ReaderDaemon
	homeState *HomeState
	suffrage  Suffrage
	ballotbox *Ballotbox
	lastRound Round
	callbacks map[string]VoteCompilerCallback
}

func NewVoteCompiler(
	homeState *HomeState,
	suffrage Suffrage,
	ballotbox *Ballotbox,
) *VoteCompiler {
	vc := &VoteCompiler{
		homeState: homeState,
		suffrage:  suffrage,
		ballotbox: ballotbox,
		lastRound: homeState.Block().Round(),
		callbacks: map[string]VoteCompilerCallback{},
	}

	vc.ReaderDaemon = common.NewReaderDaemon(true, 1000, vc.receiveSeal)
	vc.ReaderDaemon.Logger = common.NewLogger(Log(), "module", "vote-compiler")

	return vc
}

func (vc *VoteCompiler) sendResult(v interface{}) {
	vc.RLock()
	defer vc.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(vc.callbacks))

	for name, callback := range vc.callbacks {
		go func(name string, callback VoteCompilerCallback) {
			if err := callback(v); err != nil {
				vc.Log().Error("failed to run callback", "callback", name, "error", err)
			}

			wg.Done()
		}(name, callback)
	}

	wg.Wait()
}

func (vc *VoteCompiler) Callbacks() map[string]VoteCompilerCallback {
	vc.RLock()
	defer vc.RUnlock()

	return vc.callbacks
}

func (vc *VoteCompiler) RegisterCallback(name string, callback VoteCompilerCallback) error {
	vc.Lock()
	defer vc.Unlock()

	if _, found := vc.callbacks[name]; found {
		return xerrors.Errorf("VoteCompilerCallback already registered; name=%q", name)
	}

	vc.callbacks[name] = callback

	return nil
}

func (vc *VoteCompiler) UnregisterCallback(name string) error {
	vc.Lock()
	defer vc.Unlock()

	if _, found := vc.callbacks[name]; !found {
		return xerrors.Errorf("VoteCompilerCallback not registered; name=%q", name)
	}

	delete(vc.callbacks, name)

	return nil
}

func (vc *VoteCompiler) LastRound() Round {
	vc.RLock()
	defer vc.RUnlock()

	return vc.lastRound
}

func (vc *VoteCompiler) setLastRound(round Round) *VoteCompiler {
	vc.Lock()
	defer vc.Unlock()

	vc.lastRound = round

	return vc
}

func (vc *VoteCompiler) receiveSeal(v interface{}) error {
	// TODO store seal

	sl, ok := v.(seal.Seal)
	if !ok {
		return xerrors.Errorf("not Seal")
	}

	// TODO remove; checking IsValid() should be already done in previous
	// process
	if err := sl.IsValid(); err != nil {
		return err
	}

	vc.Log().Debug("got seal", "seal", sl)

	switch t := sl.Type(); t {
	case BallotType:
		ballot, ok := sl.(Ballot)
		if !ok {
			return xerrors.Errorf("is not Ballot; seal=%q", sl)
		}

		// TODO check ballot
		if err := vc.receiveBallot(ballot); err != nil {
			return err
		}
	case ProposalType:
		proposal, ok := sl.(Proposal)
		if !ok {
			return xerrors.Errorf("is not Proposal; seal=%q", sl)
		}

		if err := vc.receiveProposal(proposal); err != nil {
			return err
		}
	default:
		return xerrors.Errorf("not available seal type in JOIN state; type=%q", t)
	}

	return nil
}

func (vc *VoteCompiler) receiveBallot(ballot Ballot) error {
	// TODO checker ballot

	log_ := vc.Log().New(log15.Ctx{"ballot": ballot.Hash()})

	if ballot.Stage() == StageINIT {
		sub := ballot.Height().Big.Sub(vc.homeState.Height().Big)
		switch {
		case sub.Equal(big.NewBigFromInt64(0)): // same
			log_.Debug(
				"received INIT ballot with same height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
		case sub.Equal(big.NewBigFromInt64(-1)): // 1 lower
			log_.Warn(
				"received INIT ballot with previous height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
		default: // if not, ignore it
			log_.Warn(
				"received INIT ballot with weird height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() == vc.LastRound() {
			log_.Warn(
				"received INIT ballot with previous round",
				"round", ballot.Round(),
				"expected", vc.LastRound(),
			)
		} else if ballot.Round() == vc.LastRound()+1 {
			log_.Debug(
				"received INIT ballot with running round",
				"round", ballot.Round(),
				"expected", vc.LastRound()+1,
			)
		} else {
			log_.Warn(
				"received INIT ballot with weird round",
				"round", ballot.Round(),
				"expected", []Round{vc.LastRound(), vc.LastRound()},
			)
			return nil
		}
	} else {
		if !ballot.Height().Equal(vc.homeState.Height()) { // ignore it
			log_.Warn(
				"received ballot with different height",
				"height", ballot.Height(),
				"expected", vc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != vc.LastRound()+1 { // ignore it
			log_.Warn(
				"received ballot with different round",
				"round", ballot.Round(),
				"expected", vc.LastRound()+1,
			)
			return nil
		}
	}

	vr, err := vc.ballotbox.Vote(ballot)
	if err != nil {
		log_.Error("failed to vote", "error", err)
		return err
	}

	switch vr.Result() {
	case GotMajority:
		if vr.Stage() == StageINIT {
			_ = vc.setLastRound(vr.Round()) // set lastRound
			log_.Debug("set LastRound", "round", vc.LastRound())
		}
	}

	// NOTE notify to state handler

	vc.sendResult(vr)

	return nil
}

func (vc *VoteCompiler) receiveProposal(proposal Proposal) error {
	// TODO check,
	// - Proposal is already processed
	// - transactions in Proposal.Transactions is already in block or not.  if not, ignore it

	// - Proposal.Height is same with home.  if not, ignore it
	if !proposal.Height().Equal(vc.homeState.Height()) { // ignore it
		vc.Log().Debug(
			"received proposal with different height",
			"height", proposal.Height(),
			"home", vc.homeState.Height(),
		)
		return nil
	}

	// - Proposal.Round is same with vc.lastRound.  if not, ignore it
	if proposal.Round() != vc.LastRound()+1 { // ignore it
		vc.Log().Warn(
			"received proposal with different round",
			"round", proposal.Round(),
			"home", vc.LastRound(),
		)
		return nil
	}

	// - Proposal.CurrentBlock is same with home.  if not, ignore it
	if !proposal.CurrentBlock().Equal(vc.homeState.Block().Hash()) { // ignore it
		vc.Log().Warn(
			"received proposal with different current block",
			"block", proposal.CurrentBlock(),
			"home", vc.homeState.Block().Hash(),
		)
		return nil
	}

	// - Proposal.Proposer is valid proposer at this round.  if not, ignore it
	actingSuffrage := vc.suffrage.ActingSuffrage(proposal.Height(), proposal.Round())
	if !actingSuffrage.Proposer().Address().Equal(proposal.Proposer()) {
		vc.Log().Warn(
			"proposer is not proposer at this round",
			"proposer", proposal.Proposer(),
			"expected_proposer", actingSuffrage.Proposer().Address(),
			"height", proposal.Height(),
			"round", proposal.Round(),
		)
		return nil
	}

	// TODO everyting is ok, notify to state handler

	vc.sendResult(proposal)

	return nil
}
