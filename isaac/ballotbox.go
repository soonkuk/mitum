package isaac

import (
	"encoding/json"
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
		voted:  map[ /* VoteRecord.VotingHash */ hash.Hash][]VoteRecord{},
	}
}

func (b *BallotBox) Vote(node node.Address, height Height, round Round, stage Stage, proposal hash.Hash, nextBlock hash.Hash) error {
	log_ := b.Log().New(log15.Ctx{
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

	b.Lock()
	b.voted[vr.VotingHash()] = append(b.voted[vr.VotingHash()], vr)
	b.Unlock()

	log_.Debug("voted")

	// check agreement

	return nil
}

type VoteRecord struct {
	hash       hash.Hash
	votingHash hash.Hash
	node       node.Address
	height     Height
	round      Round
	stage      Stage
	proposal   hash.Hash
	nextBlock  hash.Hash
	votedAt    common.Time
}

func NewVoteRecord(node node.Address, height Height, round Round, stage Stage, proposal hash.Hash, nextBlock hash.Hash) (VoteRecord, error) {
	// make hash
	nb, err := node.MarshalBinary()
	if err != nil {
		return VoteRecord{}, err
	}

	hb, err := height.MarshalBinary()
	if err != nil {
		return VoteRecord{}, err
	}

	rb, err := round.MarshalBinary()
	if err != nil {
		return VoteRecord{}, err
	}

	sb, err := stage.MarshalBinary()
	if err != nil {
		return VoteRecord{}, err
	}

	pb, err := proposal.MarshalBinary()
	if err != nil {
		return VoteRecord{}, err
	}

	nbb, err := nextBlock.MarshalBinary()
	if err != nil {
		return VoteRecord{}, err
	}

	var eh []byte
	eh = append(eh, nb...)
	eh = append(eh, hb...)
	eh = append(eh, rb...)
	eh = append(eh, sb...)
	eh = append(eh, pb...)
	eh = append(eh, nbb...)

	hs, err := hash.DefaultHashes.NewHash("vr", eh)
	if err != nil {
		return VoteRecord{}, err
	}

	var ev []byte
	ev = append(ev, hb...)
	ev = append(ev, rb...)
	ev = append(ev, sb...)
	ev = append(ev, pb...)
	ev = append(ev, nbb...)

	votingHash, err := hash.DefaultHashes.NewHash("vrv", ev)
	if err != nil {
		return VoteRecord{}, err
	}

	return VoteRecord{
		hash:       hs,
		votingHash: votingHash,
		node:       node,
		height:     height,
		round:      round,
		stage:      stage,
		proposal:   proposal,
		nextBlock:  nextBlock,
		votedAt:    common.Now(),
	}, nil
}

func (v VoteRecord) Hash() hash.Hash {
	return v.hash
}

func (v VoteRecord) VotingHash() hash.Hash {
	return v.hash
}

func (v VoteRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"hash":       v.hash,
		"votingHash": v.votingHash,
		"node":       v.node,
		"height":     v.height,
		"round":      v.round,
		"stage":      v.stage,
		"proposal":   v.proposal,
		"nextBlock":  v.nextBlock,
	})
}
