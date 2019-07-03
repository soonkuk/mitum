package isaac

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
)

type testBaseBallot struct {
	suite.Suite
}

func (t *testBaseBallot) TestEncode() {
	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)
	proposal := NewRandomProposalHash()
	currentBlock := NewRandomBlockHash()
	nextBlock := NewRandomBlockHash()

	ballot, err := NewBallot(n, height, round, StageSIGN, proposal, currentBlock, nextBlock)
	t.NoError(err)

	pk, _ := keypair.NewStellarPrivateKey()
	{ // sign
		err = ballot.Sign(pk, []byte{})
		t.NoError(err)
	}

	b, err := rlp.EncodeToBytes(ballot)
	t.NoError(err)

	var decoded BaseBallot
	err = rlp.DecodeBytes(b, &decoded)
	t.NoError(err)

	t.Equal(BallotType, decoded.Type())
	t.Equal(BallotType, decoded.Body().Type())
	t.Equal(n, decoded.Node())
	t.Equal(height, decoded.Height())
	t.Equal(round, decoded.Round())
	t.Equal(ballot.Stage(), decoded.Stage())
	t.Equal(proposal, decoded.Proposal())
	t.Equal(nextBlock, decoded.NextBlock())
}

func TestBaseBallot(t *testing.T) {
	suite.Run(t, new(testBaseBallot))
}
