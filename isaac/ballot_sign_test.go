package isaac

import (
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testSIGNBallot struct {
	suite.Suite
	home common.HomeNode
}

func (t *testSIGNBallot) SetupTest() {
	t.home = common.NewRandomHome()
}

func (t *testSIGNBallot) TestNew() {
	height := common.NewBig(33)
	round := Round(0)
	proposer := common.RandomSeed().Address()
	var validators []common.Address
	for i := 0; i < 5; i++ {
		validators = append(validators, common.RandomSeed().Address())
	}

	proposal := common.NewRandomHash("pp")
	block := common.NewRandomHash("bk")
	vote := VoteYES

	ballot := NewSIGNBallot(
		height, round, proposer, validators,
		proposal, block, vote,
	)

	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)

	err = ballot.Wellformed()
	t.NoError(err)

	_, ok := interface{}(ballot).(common.Seal)
	t.True(ok)

	_, ok = interface{}(ballot).(Ballot)
	t.True(ok)

	t.Equal(SIGNBallotSealType, ballot.Type())
	t.Equal(VoteStageSIGN, ballot.Stage())
	t.Equal(height, ballot.Height())
	t.Equal(round, ballot.Round())
	t.Equal(proposer, ballot.Proposer())
	t.Equal(validators, ballot.Validators())
	t.Equal(proposal, ballot.Proposal())
	t.Equal(block, ballot.Block())
	t.Equal(vote, ballot.Vote())
}

func (t *testSIGNBallot) TestMarshalBinary() {
	height := common.NewBig(33)
	round := Round(0)
	proposer := common.RandomSeed().Address()
	var validators []common.Address
	for i := 0; i < 5; i++ {
		validators = append(validators, common.RandomSeed().Address())
	}

	proposal := common.NewRandomHash("pp")
	block := common.NewRandomHash("bk")
	vote := VoteYES

	ballot := NewSIGNBallot(
		height, round, proposer, validators,
		proposal, block, vote,
	)

	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)

	err = ballot.Wellformed()
	t.NoError(err)

	var b []byte
	{ // marshal
		b, err = sealCodec.Encode(ballot)
		t.NoError(err)
		t.NotEmpty(b)
	}

	decoded, err := sealCodec.Decode(b)
	t.NoError(err)

	var signBallot SIGNBallot
	err = common.CheckSeal(decoded, &signBallot)
	t.NoError(err)

	err = signBallot.Wellformed()
	t.NoError(err)

	t.Equal(SIGNBallotSealType, signBallot.Type())
	t.Equal(VoteStageSIGN, signBallot.Stage())
	t.Equal(height, signBallot.Height())
	t.Equal(round, signBallot.Round())
	t.Equal(proposer, signBallot.Proposer())
	t.Equal(validators, signBallot.Validators())
	t.Equal(proposal, signBallot.Proposal())
	t.Equal(block, signBallot.Block())
	t.Equal(vote, signBallot.Vote())
}

func TestSIGNBallot(t *testing.T) {
	suite.Run(t, new(testSIGNBallot))
}
