package isaac

import (
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type BallotBox struct {
	sync.RWMutex
	*common.Logger
	voted map[ /* VoteRecord.BoxHash */ hash.Hash]*VoteRecords
}

func NewBallotBox() *BallotBox {
	return &BallotBox{
		Logger: common.NewLogger(log, "module", "ballotbox"),
		voted:  map[hash.Hash]*VoteRecords{},
	}
}

func (bb *BallotBox) boxHash(height Height, round Round, stage Stage, proposal hash.Hash) (hash.Hash, error) {
	var l []interface{}
	if stage == StageINIT {
		l = []interface{}{
			height,
			round,
			stage,
		}
	} else {
		l = []interface{}{
			height,
			round,
			stage,
			proposal,
		}
	}

	b, err := rlp.EncodeToBytes(l)
	if err != nil {
		return hash.Hash{}, err
	}

	return hash.NewArgon2Hash("bbb", b)
}

func (bb *BallotBox) Vote(
	node node.Address,
	height Height,
	round Round,
	stage Stage,
	proposal hash.Hash,
	nextBlock hash.Hash,
	seal hash.Hash,
) (VoteRecords, error) {
	log_ := bb.Log().New(log15.Ctx{
		"node":      node,
		"height":    height,
		"round":     round,
		"stage":     stage,
		"proposal":  proposal,
		"nextBlock": nextBlock,
		"seal":      seal,
	})

	log_.Debug("trying to vote")

	// TODO checking CanVote should be done before Vote().
	if !stage.CanVote() {
		return VoteRecords{}, FailedToVoteError.Newf("invalid stage; stage=%q", stage)
	}

	boxHash, err := bb.boxHash(height, round, stage, proposal)
	if err != nil {
		return VoteRecords{}, err
	}

	nr, found := bb.voted[boxHash]
	if !found {
		nr = NewVoteRecords(boxHash, height, round, stage, proposal)
		bb.voted[boxHash] = nr
	}

	vr, err := NewVoteRecord(node, nextBlock, seal)
	if err != nil {
		return VoteRecords{}, FailedToVoteError.New(err)
	}

	log_.Debug("VoteRecord created", "vote_record", vr)

	bb.Lock()
	defer bb.Unlock()

	err = nr.Vote(vr)

	return nr.Copy(), err
}

func (bb *BallotBox) CloseVoteRecords(boxHash hash.Hash) error {
	nr, found := bb.voted[boxHash]
	if !found {
		return common.NotFoundError.Newf("VoteRecords not found; boxHash=%q", boxHash)
	}
	nr.Close()

	return nil
}
