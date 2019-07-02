package isaac

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
)

var (
	BlockHashHint string = "bk"
)

func NewBlockHash(b []byte) (hash.Hash, error) {
	return hash.NewArgon2Hash(BlockHashHint, b)
}

// TODO create func to check block hash

type Block struct {
	hash      hash.Hash
	height    Height
	proposal  hash.Hash
	createdAt common.Time
}

func NewBlock(height Height, proposal hash.Hash) (Block, error) {
	bk := Block{
		height:    height,
		proposal:  proposal,
		createdAt: common.Now(),
	}

	h, err := bk.makeHash()
	if err != nil {
		return Block{}, err
	}
	bk.hash = h

	return bk, nil
}

func (bk Block) makeHash() (hash.Hash, error) {
	if err := bk.proposal.IsValid(); err != nil {
		return hash.Hash{}, err
	}

	b, err := rlp.EncodeToBytes([]interface{}{
		bk.height,
		bk.proposal,
	})

	if err != nil {
		return hash.Hash{}, err
	}

	return NewBlockHash(b)
}

func (bk Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"hash":      bk.hash,
		"height":    bk.height,
		"proposal":  bk.proposal,
		"createdAt": bk.createdAt,
	})
}

func (bk Block) String() string {
	b, _ := json.Marshal(bk)
	return string(b)
}

func (bk Block) Hash() hash.Hash {
	return bk.hash
}

func (bk Block) Height() Height {
	return bk.height
}

func (bk Block) Proposal() hash.Hash {
	return bk.proposal
}

func (bk Block) Equal(n Block) bool {
	if !bk.Height().Equal(n.Height()) {
		return false
	}

	if !bk.Hash().Equal(n.Hash()) {
		return false
	}

	if !bk.Proposal().Equal(n.Proposal()) {
		return false
	}

	return true
}

func (bk Block) Empty() bool {
	return bk.hash.Empty()
}
