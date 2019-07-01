package isaac

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
)

type Ballotbox struct {
	*common.Logger
	voted     map[ /* VoteRecord.BoxHash */ hash.Hash]*VoteRecords
	threshold *Threshold
}

func NewBallotbox(threshold *Threshold) *Ballotbox {
	return &Ballotbox{
		Logger:    common.NewLogger(log, "module", "ballotbox"),
		voted:     map[hash.Hash]*VoteRecords{},
		threshold: threshold,
	}
}

func (bb *Ballotbox) boxHash(ballot Ballot) (hash.Hash, error) {
	var l []interface{}
	if ballot.Stage() == StageINIT {
		l = []interface{}{ballot.Height(), ballot.Round(), ballot.Stage()}
	} else {
		l = []interface{}{ballot.Height(), ballot.Round(), ballot.Stage(), ballot.Proposal()}
	}

	b, err := rlp.EncodeToBytes(l)
	if err != nil {
		return hash.Hash{}, err
	}

	return hash.NewArgon2Hash("bbb", b)
}

func (bb *Ballotbox) Vote(ballot Ballot) (VoteResult, error) {
	log_ := bb.Log().New(log15.Ctx{"ballot": ballot.Hash()})

	log_.Debug("trying to vote")

	// TODO checking CanVote should be done before Vote().
	if !ballot.Stage().CanVote() {
		return VoteResult{}, FailedToVoteError.Newf("invalid stage; stage=%q", ballot.Stage())
	}

	boxHash, err := bb.boxHash(ballot)
	if err != nil {
		return VoteResult{}, err
	}

	vrs, found := bb.voted[boxHash]
	if !found {
		vrs = NewVoteRecords(boxHash, ballot.Height(), ballot.Round(), ballot.Stage(), ballot.Proposal())
		bb.voted[boxHash] = vrs
	}

	vr, err := NewVoteRecord(ballot.Node(), ballot.CurrentBlock(), ballot.NextBlock(), ballot.Hash())
	if err != nil {
		return VoteResult{}, FailedToVoteError.New(err)
	}

	log_.Debug("VoteRecord created", "vote_record", vr)

	err = vrs.Vote(vr)
	if err != nil {
		return VoteResult{}, err
	}

	return bb.CheckMajority(*vrs)
}

func (bb *Ballotbox) CheckMajority(vrs VoteRecords) (VoteResult, error) {
	log_ := bb.Log().New(log15.Ctx{"vrs": vrs})
	log_.Debug("trying to check majority", "vrs", vrs)

	var total, threshold uint = bb.threshold.Get(vrs.stage)
	vr, err := vrs.CheckMajority(total, threshold)
	if err != nil {
		bb.Log().Error("failed to get vote result", "error", err)
		return VoteResult{}, err
	}

	bb.Log().Debug("got vote result", "result", vr)

	switch vr.Result() {
	case JustDraw, GotMajority:
		if vrs.IsClosed() {
			log_.Debug("VoteRecords already closed")
		} else {
			vrs.Close()
		}
	default:
		//
	}

	return vr, nil
}
