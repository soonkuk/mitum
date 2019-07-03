package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/seal"
)

type SealCompiler struct {
	sync.RWMutex
	*common.Logger
	*common.ReaderDaemon
	homeState *HomeState
	suffrage  Suffrage
	ballotbox *Ballotbox
	lastRound Round
	ch        chan interface{}
}

func NewSealCompiler(
	homeState *HomeState,
	suffrage Suffrage,
	ballotbox *Ballotbox,
	ch chan interface{},
) *SealCompiler {
	sc := &SealCompiler{
		Logger:    common.NewLogger(Log(), "module", "ballot-compiler"),
		homeState: homeState,
		suffrage:  suffrage,
		ballotbox: ballotbox,
		lastRound: Round(0),
		ch:        ch,
	}

	sc.ReaderDaemon = common.NewReaderDaemon(true, sc.receiveSeal)

	return sc
}

func (sc *SealCompiler) LastRound() Round {
	sc.RLock()
	defer sc.RUnlock()

	return sc.lastRound
}

func (sc *SealCompiler) setLastRound(round Round) *SealCompiler {
	sc.Lock()
	defer sc.Unlock()

	sc.lastRound = round

	return sc
}

func (sc *SealCompiler) receiveSeal(v interface{}) error {
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

	sc.Log().Debug("got seal", "seal", sl)

	switch t := sl.Type(); t {
	case BallotType:
		ballot, ok := sl.(Ballot)
		if !ok {
			return xerrors.Errorf("is not Ballot; seal=%q", sl)
		}

		// TODO check ballot
		if err := sc.receiveBallot(ballot); err != nil {
			return err
		}
	case ProposalType:
		proposal, ok := sl.(Proposal)
		if !ok {
			return xerrors.Errorf("is not Proposal; seal=%q", sl)
		}

		if err := sc.receiveProposal(proposal); err != nil {
			return err
		}
	default:
		return xerrors.Errorf("not available seal type in JOIN state; type=%q", t)
	}

	return nil
}

func (sc *SealCompiler) receiveBallot(ballot Ballot) error {
	// TODO checker ballot

	if ballot.Stage() == StageINIT {
		sub := ballot.Height().Big.Sub(sc.homeState.Height().Big)
		switch {
		case sub.Equal(big.NewBigFromInt64(0)): // same
			sc.Log().Debug(
				"received INIT ballot with same height",
				"height", ballot.Height(),
				"home", sc.homeState.Height(),
			)
		case sub.Equal(big.NewBigFromInt64(-1)): // 1 lower
			sc.Log().Debug(
				"received INIT ballot with previous height",
				"height", ballot.Height(),
				"home", sc.homeState.Height(),
			)
		default: // if not, ignore it
			sc.Log().Debug(
				"received INIT ballot with weird height",
				"height", ballot.Height(),
				"home", sc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != Round(0) && ballot.Round() != sc.LastRound() {
			sc.Log().Debug(
				"received INIT ballot with weird round",
				"round", ballot.Round(),
				"expected", sc.LastRound(),
			)
			return nil
		}
	} else {
		if !ballot.Height().Equal(sc.homeState.Height()) { // ignore it
			sc.Log().Debug(
				"received ballot with different height",
				"height", ballot.Height(),
				"home", sc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != sc.LastRound() { // ignore it
			sc.Log().Debug(
				"received ballot with different round",
				"round", ballot.Round(),
				"home", sc.LastRound(),
			)
			return nil
		}
	}

	vr, err := sc.ballotbox.Vote(ballot)
	if err != nil {
		sc.Log().Error("failed to vote", "error", err)
		return err
	}

	switch vr.Result() {
	case GotMajority:
		if vr.Stage() == StageINIT {
			_ = sc.setLastRound(vr.Round()) // set lastRound
			sc.Log().Debug("set LastRound", "round", sc.LastRound())
		}
	}

	// NOTE notify to state handler
	sc.ch <- vr

	return nil
}

func (sc *SealCompiler) receiveProposal(proposal Proposal) error {
	// TODO check,
	// - Proposal is already processed
	// - transactions in Proposal.Transactions is already in block or not.  if not, ignore it

	// - Proposal.Height is same with home.  if not, ignore it
	if !proposal.Height().Equal(sc.homeState.Height()) { // ignore it
		sc.Log().Debug(
			"received proposal with different height",
			"height", proposal.Height(),
			"home", sc.homeState.Height(),
		)
		return nil
	}

	// - Proposal.Round is same with sc.lastRound.  if not, ignore it
	if proposal.Round() != sc.LastRound() { // ignore it
		sc.Log().Debug(
			"received proposal with different round",
			"round", proposal.Round(),
			"home", sc.LastRound(),
		)
		return nil
	}

	// - Proposal.CurrentBlock is same with home.  if not, ignore it
	if !proposal.CurrentBlock().Equal(sc.homeState.Block().Hash()) { // ignore it
		sc.Log().Debug(
			"received proposal with different current block",
			"block", proposal.CurrentBlock(),
			"home", sc.homeState.Block().Hash(),
		)
		return nil
	}

	// - Proposal.Proposer is valid proposer at this round.  if not, ignore it
	activeSuffrage := sc.suffrage.ActiveSuffrage(proposal.Height(), proposal.Round())
	if !activeSuffrage.Proposer().Address().Equal(proposal.Proposer()) {
		sc.Log().Debug(
			"proposer is not proposer at this round",
			"proposer", proposal.Proposer(),
			"expected_proposer", activeSuffrage.Proposer().Address(),
			"height", proposal.Height(),
			"round", proposal.Round(),
		)
		return nil
	}

	// TODO everyting is ok, notify to state handler

	sc.ch <- proposal

	return nil
}
