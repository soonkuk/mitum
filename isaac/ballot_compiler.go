package isaac

import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"golang.org/x/xerrors"
)

// BallotCompiler manages voting process; it will block the vote one by one.
type BallotCompiler struct {
	*common.Logger
	ballotBox *BallotBox
	threshold *Threshold
}

func NewBallotCompiler(threshold *Threshold) *BallotCompiler {
	return &BallotCompiler{
		Logger:    common.NewLogger(Log(), "module", "ballot_compiler"),
		ballotBox: NewBallotBox(),
		threshold: threshold,
	}
}

func (bc *BallotCompiler) Vote(ballot Ballot) (VoteResult, error) {
	if ballot == nil {
		return VoteResult{}, xerrors.Errorf("nil ballot received")
	}

	log_ := bc.Log().New(log15.Ctx{"ballot": ballot.Hash()})
	log_.Debug("trying to vote", "ballot", ballot)

	vrs, err := bc.ballotBox.Vote(
		ballot.Node(),
		ballot.Height(),
		ballot.Round(),
		ballot.Stage(),
		ballot.Proposal(),
		ballot.CurrentBlock(),
		ballot.NextBlock(),
		ballot.Hash(),
	)
	if err != nil {
		return VoteResult{}, err
	}

	var total, threshold uint = bc.threshold.Get(vrs.stage)
	vr, err := vrs.CheckMajority(total, threshold)
	if err != nil {
		return VoteResult{}, err
	}

	log_.Debug("got vote result", "result", vr)

	switch vr.Result() {
	case JustDraw, GotMajority:
		if err := bc.ballotBox.CloseVoteRecords(vrs.Hash()); err != nil {
			bc.Log().Error("failed to close VoteRecords", "error", err)
		}
	default:
		//
	}

	return vr, nil

	/*
		switch vr.Result() {
		case NotYetMajority, FinishedGotMajority:
			return nil
		case JustDraw: // NOTE drawed, move to next round
			if err := bc.ballotBox.CloseVoteRecords(vrs.Hash()); err != nil {
				bc.Log().Error("failed to close VoteRecords", "error", err)
			}

			go bc.moveToNextRound(vr)
		case GotMajority: // NOTE move to next stage
			// NOTE if nextBlock is different from current node, go to sync
			homeVote, isVoted := vr.records.NodeVote(bc.homeState.Home().Address())
			if isVoted && !homeVote.nextBlock.Equal(vr.NextBlock()) {
				bc.Log().Debug(
					"nextblock of result is different from home vote",
					"home", homeVote,
					"result", vr,
				)
				go bc.moveToSync(vr)
				return nil
			}

			go bc.moveToNextStage(vr)
		}
	*/
}
