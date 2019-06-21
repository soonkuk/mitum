package isaac

import (
	"fmt"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type BallotBox struct {
	sync.RWMutex
	*common.Logger
	voted map[hash.Hash][]VoteRecord
}

func NewBallotBox() *BallotBox {
	return &BallotBox{
		Logger: common.NewLogger(log, "module", "ballotbox"),
		voted:  map[ /* VoteRecord.boxHash */ hash.Hash][]VoteRecord{},
	}
}

func (bb *BallotBox) Vote(
	node node.Address,
	height Height,
	round Round,
	stage Stage,
	proposal hash.Hash,
	nextBlock hash.Hash,
) error {
	log_ := bb.Log().New(log15.Ctx{
		"node":      node,
		"height":    height,
		"round":     round,
		"stage":     stage,
		"proposal":  proposal,
		"nextBlock": nextBlock,
	})

	log_.Debug("trying to vote")

	// TODO checking CanVote should be done before Vote().
	if !stage.CanVote() {
		return FailedToVoteError.Newf("invalid stage; stage=%q", stage)
	}

	vr, err := NewVoteRecord(node, height, round, stage, proposal, nextBlock)
	if err != nil {
		return FailedToVoteError.New(err)
	}

	fmt.Println(vr)
	log_.Debug("VoteRecord created", "vote_record", vr)

	bb.Lock()
	bb.voted[vr.BoxHash()] = append(bb.voted[vr.BoxHash()], vr)
	bb.Unlock()

	log_.Debug("voted", "vr", vr.Hash())

	// check agreement

	return nil
}
