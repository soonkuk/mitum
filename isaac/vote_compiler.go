package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/seal"
)

type VoteCompiler struct {
	sync.RWMutex
	*common.Logger
	*common.ReaderDaemon
	homeState *HomeState
	suffrage  Suffrage
	ballotbox *Ballotbox
	lastRound Round
	ch        chan interface{}
}

func NewVoteCompiler(
	homeState *HomeState,
	suffrage Suffrage,
	ballotbox *Ballotbox,
	ch chan interface{},
) *VoteCompiler {
	vc := &VoteCompiler{
		Logger:    common.NewLogger(Log(), "module", "ballot-compiler"),
		homeState: homeState,
		suffrage:  suffrage,
		ballotbox: ballotbox,
		lastRound: Round(0),
		ch:        ch,
	}

	vc.ReaderDaemon = common.NewReaderDaemon(true, vc.receiveSeal)

	return vc
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

	if ballot.Stage() == StageINIT {
		sub := ballot.Height().Big.Sub(vc.homeState.Height().Big)
		switch {
		case sub.Equal(big.NewBigFromInt64(0)): // same
			vc.Log().Debug(
				"received INIT ballot with same height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
		case sub.Equal(big.NewBigFromInt64(-1)): // 1 lower
			vc.Log().Debug(
				"received INIT ballot with previous height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
		default: // if not, ignore it
			vc.Log().Debug(
				"received INIT ballot with weird height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != Round(0) && ballot.Round() != vc.LastRound() {
			vc.Log().Debug(
				"received INIT ballot with weird round",
				"round", ballot.Round(),
				"expected", vc.LastRound(),
			)
			return nil
		}
	} else {
		if !ballot.Height().Equal(vc.homeState.Height()) { // ignore it
			vc.Log().Debug(
				"received ballot with different height",
				"height", ballot.Height(),
				"home", vc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != vc.LastRound() { // ignore it
			vc.Log().Debug(
				"received ballot with different round",
				"round", ballot.Round(),
				"home", vc.LastRound(),
			)
			return nil
		}
	}

	vr, err := vc.ballotbox.Vote(ballot)
	if err != nil {
		vc.Log().Error("failed to vote", "error", err)
		return err
	}

	switch vr.Result() {
	case GotMajority:
		if vr.Stage() == StageINIT {
			_ = vc.setLastRound(vr.Round()) // set lastRound
			vc.Log().Debug("set LastRound", "round", vc.LastRound())
		}
	}

	// NOTE notify to state handler
	vc.ch <- vr

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
	if proposal.Round() != vc.LastRound() { // ignore it
		vc.Log().Debug(
			"received proposal with different round",
			"round", proposal.Round(),
			"home", vc.LastRound(),
		)
		return nil
	}

	// - Proposal.CurrentBlock is same with home.  if not, ignore it
	if !proposal.CurrentBlock().Equal(vc.homeState.Block().Hash()) { // ignore it
		vc.Log().Debug(
			"received proposal with different current block",
			"block", proposal.CurrentBlock(),
			"home", vc.homeState.Block().Hash(),
		)
		return nil
	}

	// - Proposal.Proposer is valid proposer at this round.  if not, ignore it
	activeSuffrage := vc.suffrage.ActiveSuffrage(proposal.Height(), proposal.Round())
	if !activeSuffrage.Proposer().Address().Equal(proposal.Proposer()) {
		vc.Log().Debug(
			"proposer is not proposer at this round",
			"proposer", proposal.Proposer(),
			"expected_proposer", activeSuffrage.Proposer().Address(),
			"height", proposal.Height(),
			"round", proposal.Round(),
		)
		return nil
	}

	// TODO everyting is ok, notify to state handler

	vc.ch <- proposal

	return nil
}
