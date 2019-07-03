package isaac

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/seal"
)

type BallotCompiler struct {
	sync.RWMutex
	*common.Logger
	*common.ReaderDaemon
	homeState *HomeState
	suffrage  Suffrage
	ballotbox *Ballotbox
	lastRound Round
}

func NewBallotCompiler(
	homeState *HomeState,
	suffrage Suffrage,
	ballotbox *Ballotbox,
) *BallotCompiler {
	bc := &BallotCompiler{
		Logger:    common.NewLogger(Log(), "module", "ballot-compiler"),
		homeState: homeState,
		suffrage:  suffrage,
		ballotbox: ballotbox,
		lastRound: Round(0),
	}

	bc.ReaderDaemon = common.NewReaderDaemon(true, bc.receiveSeal)

	return bc
}

func (bc *BallotCompiler) LastRound() Round {
	bc.RLock()
	defer bc.RUnlock()

	return bc.lastRound
}

func (bc *BallotCompiler) setLastRound(round Round) *BallotCompiler {
	bc.Lock()
	defer bc.Unlock()

	bc.lastRound = round

	return bc
}

func (bc *BallotCompiler) receiveSeal(v interface{}) error {
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

	bc.Log().Debug("got seal", "seal", sl)

	switch t := sl.Type(); t {
	case BallotType:
		ballot, ok := sl.(Ballot)
		if !ok {
			return xerrors.Errorf("is not Ballot; seal=%q", sl)
		}

		// TODO check ballot
		if err := ballot.IsValid(); err != nil {
			return err
		}

		if err := bc.receiveBallot(ballot); err != nil {
			return err
		}
	case ProposalType:
		proposal, ok := sl.(Proposal)
		if !ok {
			return xerrors.Errorf("is not Proposal; seal=%q", sl)
		}

		if err := proposal.IsValid(); err != nil {
			return err
		}

		if err := bc.receiveProposal(proposal); err != nil {
			return err
		}
	default:
		return xerrors.Errorf("not available seal type in JOIN state; type=%q", t)
	}

	return nil
}

func (bc *BallotCompiler) receiveBallot(ballot Ballot) error {
	// TODO checker ballot

	if ballot.Stage() == StageINIT {

		sub := ballot.Height().Big.Sub(bc.homeState.Height().Big)
		switch {
		case sub.Equal(big.NewBigFromInt64(0)): // same
			bc.Log().Debug(
				"received INIT ballot with same height",
				"height", ballot.Height(),
				"home", bc.homeState.Height(),
			)
		case sub.Equal(big.NewBigFromInt64(-1)): // 1 lower
			bc.Log().Debug(
				"received INIT ballot with previous height",
				"height", ballot.Height(),
				"home", bc.homeState.Height(),
			)
		default: // if not, ignore it
			bc.Log().Debug(
				"received INIT ballot with weird height",
				"height", ballot.Height(),
				"home", bc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != Round(0) && ballot.Round() != bc.LastRound() {
			bc.Log().Debug(
				"received INIT ballot with weird round",
				"round", ballot.Round(),
				"expected", bc.LastRound(),
			)
			return nil
		}
	} else {
		if !ballot.Height().Equal(bc.homeState.Height()) { // ignore it
			bc.Log().Debug(
				"received ballot with different height",
				"height", ballot.Height(),
				"home", bc.homeState.Height(),
			)
			return nil
		}

		if ballot.Round() != bc.LastRound() { // ignore it
			bc.Log().Debug(
				"received ballot with different round",
				"round", ballot.Round(),
				"home", bc.LastRound(),
			)
			return nil
		}
	}

	vr, err := bc.ballotbox.Vote(ballot)
	if err != nil {
		bc.Log().Error("failed to vote", "error", err)
		return err
	}

	switch vr.Result() {
	case GotMajority:
		if vr.Stage() == StageINIT {
			_ = bc.setLastRound(vr.Round()) // set lastRound
			bc.Log().Debug("set LastRound", "round", bc.LastRound())
		}
	}

	// NOTE notify to state handler

	return nil
}

func (bc *BallotCompiler) receiveProposal(proposal Proposal) error {
	// TODO check,
	// - Proposal is already processed
	// - transactions in Proposal.Transactions is already in block or not.  if not, ignore it

	// - Proposal.Height is same with home.  if not, ignore it
	if !proposal.Height().Equal(bc.homeState.Height()) { // ignore it
		bc.Log().Debug(
			"received proposal with different height",
			"height", proposal.Height(),
			"home", bc.homeState.Height(),
		)
		return nil
	}

	// - Proposal.Round is same with bc.lastRound.  if not, ignore it
	if proposal.Round() != bc.LastRound() { // ignore it
		bc.Log().Debug(
			"received proposal with different round",
			"round", proposal.Round(),
			"home", bc.LastRound(),
		)
		return nil
	}

	// - Proposal.CurrentBlock is same with home.  if not, ignore it
	if !proposal.CurrentBlock().Equal(bc.homeState.Block().Hash()) { // ignore it
		bc.Log().Debug(
			"received proposal with different current block",
			"block", proposal.CurrentBlock(),
			"home", bc.homeState.Block().Hash(),
		)
		return nil
	}

	// - Proposal.Proposer is valid proposer at this round.  if not, ignore it
	activeSuffrage := bc.suffrage.ActiveSuffrage(proposal.Height(), proposal.Round())
	if !activeSuffrage.Proposer().Address().Equal(proposal.Proposer()) {
		bc.Log().Debug(
			"proposer is not proposer at this round",
			"proposer", proposal.Proposer(),
			"expected_proposer", activeSuffrage.Proposer().Address(),
			"height", proposal.Height(),
			"round", proposal.Round(),
		)
		return nil
	}

	// TODO everyting is ok, notify to state handler

	return nil
}
