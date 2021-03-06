package isaac

import (
	"context"

	"github.com/rs/zerolog"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
)

type CompilerBallotChecker struct {
	homeState *HomeState
	suffrage  Suffrage
}

func NewCompilerBallotChecker(homeState *HomeState, suffrage Suffrage) *common.ChainChecker {
	cbc := CompilerBallotChecker{
		homeState: homeState,
		suffrage:  suffrage,
	}

	return common.NewChainChecker(
		"ballot-checker",
		context.Background(),
		cbc.initialize,
		cbc.checkInSuffrage,
		cbc.checkHeightAndRound,
		cbc.checkINIT,
		cbc.checkNotINIT,
	)
}

func (cbc CompilerBallotChecker) initialize(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	log_ := c.Log().With().Object("ballot", ballot.Hash()).Logger()
	_ = c.SetContext("log", log_)

	log_.Debug().Msg("will check ballot")

	return nil
}

func (cbc CompilerBallotChecker) checkInSuffrage(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	if ballot.Stage() != StageINIT {
		acting := cbc.suffrage.Acting(ballot.Height(), ballot.Round())
		if !acting.Exists(ballot.Node()) {
			return xerrors.Errorf(
				"%s ballot node does not in acting suffrage; ballot=%v node=%v",
				ballot.Stage(),
				ballot.Hash(),
				ballot.Node(),
			)
		}
	} else if !cbc.suffrage.Exists(ballot.Height().Sub(1), ballot.Node()) {
		return xerrors.Errorf(
			"%s ballot node does not in suffrage; ballot=%v node=%v",
			ballot.Stage(),
			ballot.Hash(),
			ballot.Node(),
		)
	}

	return nil
}

func (cbc CompilerBallotChecker) checkHeightAndRound(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	var log_ zerolog.Logger
	if err := c.ContextValue("log", &log_); err != nil {
		return err
	}

	// NOTE ballot.Height() and ballot.Round() should be same than last init ballot
	var lastINITVoteResult VoteResult
	if err := c.ContextValue("lastINITVoteResult", &lastINITVoteResult); err != nil {
		return err
	}

	// NOTE lastINITVoteResult is not empty, ballot.Height() should be same or
	// greater than lastINITVoteResult.Height()
	if lastINITVoteResult.IsFinished() {
		if ballot.Height().Cmp(lastINITVoteResult.Height()) < 0 {
			err := xerrors.Errorf("lower ballot height")
			log_.Debug().
				Uint64("ballot_height", ballot.Height().Uint64()).
				Uint64("height", lastINITVoteResult.Height().Uint64()).
				Msg("ballot height should be greater than last init ballot; ignore this ballot")
			return err
		}
	}

	return nil
}

func (cbc CompilerBallotChecker) checkINIT(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	if ballot.Stage() != StageINIT {
		return nil
	}

	var log_ zerolog.Logger
	if err := c.ContextValue("log", &log_); err != nil {
		return err
	}

	// NOTE ballot.Height() should be greater than homeState.Block().Height()
	if ballot.Height().Cmp(cbc.homeState.Block().Height()) < 1 {
		err := xerrors.Errorf("lower ballot height")
		log_.Error().
			Uint64("ballot_height", ballot.Height().Uint64()).
			Uint64("height", cbc.homeState.Block().Height().Uint64()).
			Msg("ballot.Height() should be greater than homeState.Block().Height(); ignore this ballot")

		return err
	} else {
		sub := ballot.Height().Sub(cbc.homeState.Block().Height())
		switch sub.Int64() {
		case 2:
			if !ballot.LastBlock().Equal(cbc.homeState.Block().Hash()) {
				log_.Debug().
					Object("last_block of ballot", ballot.LastBlock()).
					Object("block", cbc.homeState.Block().Hash()).
					Msg("last block of ballot does not match with home")
				return common.ChainCheckerStopError
			}
		case 1:
			if !ballot.Block().Equal(cbc.homeState.Block().Hash()) {
				log_.Debug().
					Object("block of ballot", ballot.Block()).
					Object("block", cbc.homeState.Block().Hash()).
					Msg("block of ballot does not match with home")
				return common.ChainCheckerStopError
			}
			if ballot.LastRound() != cbc.homeState.Block().Round() {
				log_.Debug().
					Uint64("last round of ballot", ballot.LastRound().Uint64()).
					Uint64("round", cbc.homeState.Block().Round().Uint64()).
					Msg("last round of ballot does not match with home")
				return common.ChainCheckerStopError
			}
		default:
			log_.Debug().
				Uint64("ballot_height", ballot.Height().Uint64()).
				Uint64("height", cbc.homeState.Block().Height().Uint64()).
				Msg("ballot height is higher than expected")
			return common.ChainCheckerStopError
		}
	}

	var lastINITVoteResult VoteResult
	if err := c.ContextValue("lastINITVoteResult", &lastINITVoteResult); err != nil {
		return err
	}

	if !lastINITVoteResult.IsFinished() {
		log_.Debug().Msg("lastINITVoteResult is empty")
		return nil
	}

	lastHeight := lastINITVoteResult.Height()
	lastRound := lastINITVoteResult.Round()

	if ballot.Height().Equal(lastHeight) { // this should be draw; round should be greater
		if ballot.Round() <= lastRound {
			err := xerrors.Errorf("ballot.Round() should be greater than lastINITVoteResult")
			log_.Debug().
				Err(err).
				Uint64("last_height", lastHeight.Uint64()).
				Uint64("last_round", lastRound.Uint64()).
				Uint64("ballot_height", ballot.Height().Uint64()).
				Uint64("ballot_round", ballot.Round().Uint64()).
				Msg("compared with lastINITVoteResult")
			return err
		}
	} else if ballot.Height().Cmp(lastHeight) < 1 {
		err := xerrors.Errorf("ballot.Height() should be greater than lastINITVoteResult")
		log_.Debug().
			Err(err).
			Uint64("last_height", lastHeight.Uint64()).
			Uint64("last_round", lastRound.Uint64()).
			Uint64("ballot_height", ballot.Height().Uint64()).
			Uint64("ballot_round", ballot.Round().Uint64()).
			Msg("compared with lastINITVoteResult")
		return err
	}

	return nil
}

func (cbc CompilerBallotChecker) checkNotINIT(c *common.ChainChecker) error {
	var ballot Ballot
	if err := c.ContextValue("ballot", &ballot); err != nil {
		return err
	}

	switch ballot.Stage() {
	case StageINIT:
		return nil
	}

	var log_ zerolog.Logger
	if err := c.ContextValue("log", &log_); err != nil {
		return err
	}

	// NOTE ballot.Height() should be greater than homeState.Block().Height()
	if sub := ballot.Height().Sub(cbc.homeState.Block().Height()); sub.Int64() < 1 {
		err := xerrors.Errorf("lower ballot height")
		log_.Error().
			Uint64("ballot_height", ballot.Height().Uint64()).
			Uint64("height", cbc.homeState.Block().Height().Uint64()).
			Msg("ballot height should be greater than homeState.Block().Height() + 1; ignore this ballot")

		return err
	}

	// NOTE ballot.Round() should be greater than homeState.Block().Round()
	if ballot.LastBlock().Equal(cbc.homeState.Block().Hash()) {
		if ballot.LastRound() != cbc.homeState.Block().Round() {
			err := xerrors.Errorf("ballot last round does not match with last block")
			log_.Error().
				Uint64("ballot_round", ballot.Round().Uint64()).
				Uint64("round", cbc.homeState.Block().Round().Uint64()).
				Msg("ballot round should be same with block round of homeState; ignore this ballot")

			return err
		}
	}

	var lastINITVoteResult VoteResult
	if err := c.ContextValue("lastINITVoteResult", &lastINITVoteResult); err != nil {
		return err
	}

	// NOTE without previous lastINITVoteResult, the stages except init will be
	// ignored
	if !lastINITVoteResult.IsFinished() {
		err := xerrors.Errorf("lastINITVoteResult is empty")
		log_.Error().Msg("lastINITVoteResult is empty; ignore this ballot")
		return err
	}

	// NOTE the height of stages except init should be same with
	// lastINITVoteResult.Height()
	if !ballot.Height().Equal(lastINITVoteResult.Height()) {
		err := xerrors.Errorf("lower ballot height")
		log_.Debug().
			Uint64("ballot_height", ballot.Height().Uint64()).
			Uint64("height", lastINITVoteResult.Height().Uint64()).
			Msg("ballot height should be same with last init ballot; ignore this ballot")
		return err
	}

	// NOTE the round of stages except init should be same with
	// lastINITVoteResult.Round()
	if ballot.Round() != lastINITVoteResult.Round() {
		err := xerrors.Errorf("lower ballot round")
		log_.Debug().
			Uint64("ballot_round", ballot.Round().Uint64()).
			Uint64("round", lastINITVoteResult.Round().Uint64()).
			Msg("ballot round should be same with last init ballot; ignore this ballot")
		return err
	}

	return nil
}
