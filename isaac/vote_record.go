package isaac

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

type VoteRecords struct {
	sync.RWMutex
	*common.Logger
	hash     hash.Hash
	height   Height
	round    Round
	stage    Stage
	proposal hash.Hash
	voted    map[node.Address]VoteRecord
	closed   bool
}

func NewVoteRecords(
	hash hash.Hash,
	height Height,
	round Round,
	stage Stage,
	proposal hash.Hash,
) *VoteRecords {
	return &VoteRecords{
		Logger: common.NewLogger(log, "module", "node-vote-records").SetLogContext(
			nil,
			"hash", hash,
			"height", height,
			"round", round,
			"stage", stage,
			"proposal", proposal,
		),
		hash:     hash,
		height:   height,
		round:    round,
		stage:    stage,
		proposal: proposal,
		voted:    map[node.Address]VoteRecord{},
	}
}

func (vrs VoteRecords) Hash() hash.Hash {
	return vrs.hash
}

func (vrs VoteRecords) Height() Height {
	return vrs.height
}

func (vrs VoteRecords) Round() Round {
	return vrs.round
}

func (vrs VoteRecords) Stage() Stage {
	return vrs.stage
}

func (vrs VoteRecords) Proposal() hash.Hash {
	return vrs.proposal
}

func (vrs *VoteRecords) Vote(vr VoteRecord) error {
	vrs.Lock()
	defer vrs.Unlock()

	if evr, found := vrs.voted[vr.node]; found {
		// NOTE revoting: same vote will be allowed , but same seal is not allowed
		if evr.Equal(vr) {
			return AlreadyVotedError.Newf("node=%q", vr.node)
		}
	}

	vrs.voted[vr.node] = vr

	return nil
}

func (vrs VoteRecords) Copy() VoteRecords {
	return VoteRecords{
		Logger:   vrs.Logger,
		hash:     vrs.hash,
		height:   vrs.height,
		round:    vrs.round,
		stage:    vrs.stage,
		proposal: vrs.proposal,
		voted:    vrs.voted,
		closed:   vrs.closed,
	}
}

func (vrs VoteRecords) Len() int {
	return len(vrs.voted)
}

func (vrs VoteRecords) IsClosed() bool {
	return vrs.closed
}

func (vrs *VoteRecords) Close() {
	vrs.Lock()
	defer vrs.Unlock()

	if vrs.closed {
		return
	}

	vrs.closed = true
}

func (vrs VoteRecords) Records() map[node.Address]VoteRecord {
	return vrs.voted
}

func (vrs VoteRecords) CheckMajority(total, threshold uint) (VoteResult, error) {
	vr := NewVoteResult(vrs.height, vrs.round, vrs.stage, vrs.proposal, vrs)

	var nextBlocks []hash.Hash
	var counted []uint
	if vrs.stage == StageINIT {
		var c uint
		for _ = range vrs.voted {
			c++
		}
		nextBlocks = append(nextBlocks, hash.Hash{}) // INIT ballot should have empty block
		counted = append(counted, c)
	} else {
		countByNextBlock := map[hash.Hash]uint{}
		for _, vr := range vrs.voted {
			countByNextBlock[vr.nextBlock]++
		}

		for nextBlock, c := range countByNextBlock {
			nextBlocks = append(nextBlocks, nextBlock)
			counted = append(counted, c)
		}
	}

	switch index := common.CheckMajority(total, threshold, counted...); index {
	case -1:
		vr.result = NotYetMajority
		return vr, nil
	case -2:
		vr.result = JustDraw
		return vr, nil
	default:
		if vrs.closed {
			vr.result = FinishedGotMajority
		} else {
			vr.result = GotMajority
		}
		vr.nextBlock = nextBlocks[index]

		return vr, nil
	}
}

func (vrs VoteRecords) MarshalJSON() ([]byte, error) {
	voted := map[string]VoteRecord{}
	for address, vr := range vrs.voted {
		voted[address.String()] = vr
	}

	b, err := json.Marshal(map[string]interface{}{
		"height":   vrs.height,
		"round":    vrs.round,
		"stage":    vrs.stage,
		"proposal": vrs.proposal,
		"voted":    voted,
		"closed":   vrs.closed,
	})

	return b, err
}

func (vnr VoteRecords) String() string {
	b, _ := json.Marshal(vnr)
	return string(b)
}

func (vnr VoteRecords) IsNodeVoted(n node.Address) bool {
	_, found := vnr.voted[n]
	return found
}

func (vnr VoteRecords) NodeVote(n node.Address) (VoteRecord, bool) {
	vr, found := vnr.voted[n]
	return vr, found
}

type VoteRecord struct {
	hash      hash.Hash
	node      node.Address
	nextBlock hash.Hash
	seal      hash.Hash
	votedAt   common.Time
}

func NewVoteRecord(
	node node.Address,
	nextBlock hash.Hash,
	seal hash.Hash,
) (VoteRecord, error) {
	vr := VoteRecord{
		node:      node,
		nextBlock: nextBlock,
		seal:      seal,
		votedAt:   common.Now(),
	}

	b, err := rlp.EncodeToBytes(vr)
	if err != nil {
		return VoteRecord{}, err
	}

	h, err := hash.NewArgon2Hash("vr", b)
	if err != nil {
		return VoteRecord{}, err
	}

	vr.hash = h

	return vr, nil
}

func (vr VoteRecord) Hash() hash.Hash {
	return vr.hash
}

func (vr VoteRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"hash":      vr.hash.String(),
		"node":      vr.node.String(),
		"nextBlock": vr.nextBlock.String(),
		"seal":      vr.seal.String(),
	})
}

func (vr VoteRecord) Equal(nvr VoteRecord) bool {
	if !vr.node.Equal(nvr.node) {
		return false
	}

	if !vr.nextBlock.Equal(nvr.nextBlock) {
		return false
	}

	if !vr.seal.Equal(nvr.seal) {
		return false
	}

	return true
}

func (vr VoteRecord) String() string {
	b, _ := json.Marshal(vr)
	return string(b)
}

func (vr VoteRecord) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		vr.node,
		vr.nextBlock,
		vr.votedAt,
		vr.seal,
	})
}
