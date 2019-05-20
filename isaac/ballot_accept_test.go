package isaac

import (
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testACCEPTBallot struct {
	suite.Suite
	home common.HomeNode
}

func (t *testACCEPTBallot) SetupTest() {
	t.home = common.NewRandomHome()
}

func (t *testACCEPTBallot) TestNew() {
	height := common.NewBig(33)
	round := Round(0)
	proposer := common.RandomSeed().Address()
	var validators []common.Address
	for i := 0; i < 5; i++ {
		validators = append(validators, common.RandomSeed().Address())
	}

	proposal := common.NewRandomHash("pp")
	block := common.NewRandomHash("bk")

	ballot := NewACCEPTBallot(
		height, round, proposer, validators,
		proposal, block,
	)

	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)

	err = ballot.Wellformed()
	t.NoError(err)

	_, ok := interface{}(ballot).(common.Seal)
	t.True(ok)

	_, ok = interface{}(ballot).(Ballot)
	t.True(ok)

	t.Equal(ACCEPTBallotSealType, ballot.Type())
	t.Equal(VoteStageACCEPT, ballot.Stage())
	t.Equal(height, ballot.Height())
	t.Equal(round, ballot.Round())
	t.Equal(proposer, ballot.Proposer())
	t.Equal(validators, ballot.Validators())
	t.Equal(proposal, ballot.Proposal())
	t.Equal(block, ballot.Block())
	t.Equal(VoteYES, ballot.Vote())
}

func (t *testACCEPTBallot) TestMarshalBinary() {
	height := common.NewBig(33)
	round := Round(0)
	proposer := common.RandomSeed().Address()
	var validators []common.Address
	for i := 0; i < 5; i++ {
		validators = append(validators, common.RandomSeed().Address())
	}

	proposal := common.NewRandomHash("pp")
	block := common.NewRandomHash("bk")

	ballot := NewACCEPTBallot(
		height, round, proposer, validators,
		proposal, block,
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

	var acceptBallot ACCEPTBallot
	err = common.CheckSeal(decoded, &acceptBallot)
	t.NoError(err)

	err = acceptBallot.Wellformed()
	t.NoError(err)

	t.Equal(ACCEPTBallotSealType, acceptBallot.Type())
	t.Equal(VoteStageACCEPT, acceptBallot.Stage())
	t.Equal(height, acceptBallot.Height())
	t.Equal(round, acceptBallot.Round())
	t.Equal(proposer, acceptBallot.Proposer())
	t.Equal(validators, acceptBallot.Validators())
	t.Equal(proposal, acceptBallot.Proposal())
	t.Equal(block, acceptBallot.Block())
	t.Equal(VoteYES, acceptBallot.Vote())
}

func TestACCEPTBallot(t *testing.T) {
	suite.Run(t, new(testACCEPTBallot))
}
