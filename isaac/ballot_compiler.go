package isaac

/*
import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"golang.org/x/xerrors"
)

// BallotCompiler manages voting process; it will block the vote one by one.
type BallotCompiler struct {
	*common.Logger
	ballotbox *Ballotbox
	threshold *Threshold
}

func NewBallotCompiler(threshold *Threshold) *BallotCompiler {
	return &BallotCompiler{
		Logger:    common.NewLogger(Log(), "module", "ballot_compiler"),
		ballotbox: NewBallotbox(),
		threshold: threshold,
	}
}

func (bc *BallotCompiler) Vote(ballot Ballot) (VoteResult, error) {
	if ballot == nil {
		return VoteResult{}, xerrors.Errorf("nil ballot received")
	}

	log_ := bc.Log().New(log15.Ctx{"ballot": ballot.Hash()})
	log_.Debug("trying to vote", "ballot", ballot)

	vrs, err := bc.ballotbox.Vote(
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
		if err := bc.ballotbox.CloseVoteRecords(vrs.Hash()); err != nil {
			bc.Log().Error("failed to close VoteRecords", "error", err)
		}
	default:
		//
	}

	return vr, nil
}
*/
