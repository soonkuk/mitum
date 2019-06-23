package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type testINITBallot struct {
	suite.Suite
}

func (t *testINITBallot) TestEncoders() {
	defer common.DebugPanic()

	encs := seal.NewEncoders()
	err := encs.Register(newBallotEncoder(INITBallotType))
	t.NoError(err)

	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)
	stage := StageSIGN
	proposal := NewRandomProposalHash()
	nextBlock := NewRandomBlockHash()

	var ballot INITBallot
	{
		body := INITBallotBody{
			Node:      n,
			Height:    height,
			Round:     round,
			Stage:     stage,
			Proposal:  proposal,
			NextBlock: nextBlock,
		}
		ballot, err = NewINITBallot(body)
		t.NoError(err)
	}

	pk, _ := keypair.NewStellarPrivateKey()
	{ // sign
		err := ballot.Sign(pk, []byte{})
		t.NoError(err)
	}

	b, err := encs.Encode(ballot)
	t.NoError(err)

	decoded, err := encs.Decode(b)
	t.NoError(err)
	t.True(ballot.Equal(decoded.(seal.Seal)))

	// check values
	decodedBallot, ok := decoded.(INITBallot)
	t.True(ok)

	t.Equal(n, decodedBallot.Node())
	t.Equal(height, decodedBallot.Height())
	t.Equal(round, decodedBallot.Round())
	t.Equal(stage, decodedBallot.Stage())
	t.Equal(proposal, decodedBallot.Proposal())
	t.Equal(nextBlock, decodedBallot.NextBlock())
}

func TestINITBallot(t *testing.T) {
	suite.Run(t, new(testINITBallot))
}
