// +build test

package isaac

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/node"
)

func init() {
	common.SetTestLogger(Log())
}

func NewRandomProposalHash() hash.Hash {
	b := make([]byte, 4)
	_, _ = rand.Read(b)

	h, _ := NewProposalHash(b)
	return h
}

func NewRandomBlockHash() hash.Hash {
	b := make([]byte, 4)
	_, _ = rand.Read(b)

	h, _ := NewBlockHash(b)
	return h
}

func NewTestPolicy() Policy {
	return Policy{
		TimeoutINITBallot:        time.Second * 10,
		IntervalINITBallotOfJoin: time.Second * 3,
	}
}

func NewRandomBlock() Block {
	b, _ := rand.Int(rand.Reader, big.NewInt(27))

	bk, _ := NewBlock(
		NewBlockHeight(uint64(b.Int64())),
		NewRandomProposalHash(),
	)

	return bk
}

func NewRandomNextBlock(bk Block) Block {
	nbk, _ := NewBlock(
		bk.Height().Add(1),
		NewRandomProposalHash(),
	)

	return nbk
}

func NewRandomHomeState() *HomeState {
	hs := NewHomeState(node.NewRandomHome(), NewRandomBlock())
	hs.SetBlock(NewRandomNextBlock(hs.Block()))

	return hs
}
