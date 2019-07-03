package isaac

import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"golang.org/x/xerrors"
)

func CheckerVoteResultGoToSyncState(ck *common.ChainChecker) error {
	var vr VoteResult
	if err := ck.ContextValue("vr", &vr); err != nil {
		return err
	}

	var homeState *HomeState
	if err := ck.ContextValue("homeState", &homeState); err != nil {
		return err
	}

	sub := vr.Height().Big.Sub(homeState.Height().Big)
	_ = ck.SetContext("heightDiff", 1)

	switch {
	case sub.Equal(big.NewBigFromInt64(0)): // same
		_ = ck.SetContext("heightDiff", 0)
	case sub.Equal(big.NewBigFromInt64(-1)): // 1 lower
		_ = ck.SetContext("heightDiff", -1)
	default: // lower or higher, go to sync
		ck.Log().Debug(
			"VoteResult.Height is different from home",
			"VoteResult.height", vr.Height(),
			"home", homeState.Height(),
			"vr", vr,
		)

		return ChangeNodeStateToSyncError
	}

	return nil
}

func CheckerVoteResultINIT(ck *common.ChainChecker) error {
	var vr VoteResult
	if err := ck.ContextValue("vr", &vr); err != nil {
		return err
	}

	if vr.Stage() != StageINIT {
		return nil
	}

	var homeState *HomeState
	if err := ck.ContextValue("homeState", &homeState); err != nil {
		return err
	}

	var heightDiff int
	if err := ck.ContextValue("heightDiff", &heightDiff); err != nil {
		return err
	}

	log_ := ck.Log().New(log15.Ctx{"vr": vr, "home": homeState})

	switch heightDiff {
	case 0:
		if !vr.CurrentBlock().Equal(homeState.Block().Hash()) {
			log_.Debug(
				"VoteResult.CurrentBlock is different from home's Block",
				"VoteResult.CurrentBlock", vr.CurrentBlock(),
				"homeState.Block().Hash()", homeState.Block().Hash(),
			)
			return ChangeNodeStateToSyncError
		}
	case -1:
		if !vr.NextBlock().Equal(homeState.Block().Hash()) {
			log_.Debug(
				"VoteResult.NextBlock is different from home",
				"VoteResult.NextBlock", vr.NextBlock(),
				"homeState.Block().Hash()", homeState.Block().Hash(),
			)
			return ChangeNodeStateToSyncError
		}
		if !vr.CurrentBlock().Equal(homeState.PreviousBlock().Hash()) {
			log_.Debug(
				"VoteResult.CurrentBlock is different from home's PreviousBlock",
				"VoteResult.CurrentBlock", vr.CurrentBlock(),
				"homeState.PreviousBlock().Hash()", homeState.PreviousBlock().Hash(),
				"vr", vr,
			)
			return ChangeNodeStateToSyncError
		}
		if !vr.Proposal().Equal(homeState.Proposal()) {
			log_.Debug(
				"VoteResult.Proposal is different from home",
				"VoteResult.Proposal", vr.Proposal(),
				"homeState.Proposal()", homeState.Proposal(),
			)
			return ChangeNodeStateToSyncError
		}
	default:
		return ChangeNodeStateToSyncError
	}

	return nil
}

func CheckerVoteResultOtherStage(ck *common.ChainChecker) error {
	var vr VoteResult
	if err := ck.ContextValue("vr", &vr); err != nil {
		return err
	}

	if vr.Stage() == StageINIT {
		return nil
	}

	var homeState *HomeState
	if err := ck.ContextValue("homeState", &homeState); err != nil {
		return err
	}

	var heightDiff int
	if err := ck.ContextValue("heightDiff", &heightDiff); err != nil {
		return err
	}

	log_ := ck.Log().New(log15.Ctx{"vr": vr, "home": homeState})

	switch heightDiff {
	case 0:
		if !vr.CurrentBlock().Equal(homeState.Block().Hash()) {
			log_.Debug(
				"VoteResult.CurrentBlock is different from home's Block",
				"VoteResult.CurrentBlock", vr.CurrentBlock(),
				"homeState.Block().Hash()", homeState.Block().Hash(),
			)
			return xerrors.Errorf("ignore; CurrentBlock is different from home")
		}
	case -1:
		return xerrors.Errorf("ignore; result is behind from home")
	default:
		return ChangeNodeStateToSyncError
	}

	return nil
}
