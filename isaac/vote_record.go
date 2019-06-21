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

type NodeVoteRecords struct {
	sync.RWMutex
	*common.Logger
	voted  map[node.Address]VoteRecord
	vr     VoteRecord
	closed bool
}

func NewNodeVoteRecords(vr VoteRecord) *NodeVoteRecords {
	return &NodeVoteRecords{
		Logger: common.NewLogger(log, "module", "node-vote-records").SetLogContext(
			nil,
			"height", vr.height,
			"round", vr.round,
			"stage", vr.stage,
			"proposal", vr.proposal,
		),
		voted: map[node.Address]VoteRecord{},
		vr:    vr,
	}
}

func (n *NodeVoteRecords) Height() Height {
	return n.vr.height
}

func (n *NodeVoteRecords) Round() Round {
	return n.vr.round
}

func (n *NodeVoteRecords) Stage() Stage {
	return n.vr.stage
}

func (n *NodeVoteRecords) Proposal() hash.Hash {
	return n.vr.proposal
}

func (n *NodeVoteRecords) Vote(vr VoteRecord) ([]uint, error) {
	n.Lock()
	defer n.Unlock()

	if _, found := n.voted[vr.node]; found {
		// NOTE does not allow revoting
		return nil, AlreadyVotedError.Newf("node=%q", vr.node)
	}

	n.voted[vr.node] = vr

	countByNextBlock := map[hash.Hash]uint{}
	for _, vr := range n.voted {
		countByNextBlock[vr.nextBlock]++
	}

	var counted []uint
	for _, c := range countByNextBlock {
		counted = append(counted, c)
	}

	return counted, nil
}

func (n *NodeVoteRecords) Len() int {
	return len(n.voted)
}

func (n *NodeVoteRecords) IsClosed() bool {
	return n.closed
}

func (n *NodeVoteRecords) Close() {
	n.Lock()
	defer n.Unlock()

	n.closed = true
}

func (v NodeVoteRecords) MarshalJSON() ([]byte, error) {
	vrs := map[string]VoteRecord{}
	for address, vr := range v.voted {
		vrs[address.String()] = vr
	}

	return json.Marshal(vrs)
}

type VoteRecord struct {
	node      node.Address
	height    Height
	round     Round
	stage     Stage
	proposal  hash.Hash
	nextBlock hash.Hash
	seal      hash.Hash

	votedAt common.Time

	hash    hash.Hash
	boxHash hash.Hash
}

func NewVoteRecord(
	node node.Address,
	height Height,
	round Round,
	stage Stage,
	proposal hash.Hash,
	nextBlock hash.Hash,
	seal hash.Hash,
) (VoteRecord, error) {
	vr := VoteRecord{
		height:   height,
		round:    round,
		stage:    stage,
		proposal: proposal,
	}

	var boxHash, h hash.Hash

	{ // hash for BallotBox
		b, err := rlp.EncodeToBytes(vr)
		if err != nil {
			return VoteRecord{}, err
		}

		boxHash, err = DefaultHashEncoder.Encode("bb-vr", b)
		if err != nil {
			return VoteRecord{}, err
		}
	}

	{ // hash of VoteRecord
		vr.node = node
		vr.nextBlock = nextBlock
		vr.seal = seal
		vr.votedAt = common.Now()

		b, err := rlp.EncodeToBytes(vr)
		if err != nil {
			return VoteRecord{}, err
		}

		h, err = DefaultHashEncoder.Encode("vr", b)
		if err != nil {
			return VoteRecord{}, err
		}
	}

	vr.boxHash = boxHash
	vr.hash = h

	return vr, nil
}

func (v VoteRecord) Hash() hash.Hash {
	return v.hash
}

func (v VoteRecord) BoxHash() hash.Hash {
	return v.boxHash
}

func (v VoteRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"hash":      v.hash.String(),
		"boxHash":   v.boxHash.String(),
		"node":      v.node.String(),
		"height":    v.height,
		"round":     v.round,
		"stage":     v.stage,
		"proposal":  v.proposal.String(),
		"nextBlock": v.nextBlock.String(),
		"seal":      v.seal.String(),
	})
}

func (v VoteRecord) String() string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (v VoteRecord) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		v.node,
		v.height,
		v.round,
		v.stage,
		v.proposal,
		v.nextBlock,
		v.votedAt,
		v.seal,
	})
}

func (v VoteRecord) generateHash() (hash.Hash, error) {
	b, err := rlp.EncodeToBytes(v)
	if err != nil {
		return hash.Hash{}, err
	}

	h, err := hash.NewArgon2Hash("vr", b)
	if err != nil {
		return hash.Hash{}, err
	}

	return h, nil
}

func (v VoteRecord) generateBoxHash() (hash.Hash, error) {
	b, err := rlp.Encode(w, []interface{}{
		v.height,
		v.round,
		v.stage,
		v.proposal,
	})

	if err != nil {
		return hash.Hash{}, err
	}

	h, err := hash.NewArgon2Hash("bvr", b)
	if err != nil {
		return hash.Hash{}, err
	}

	return h, nil
}
