package isaac

import (
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testINITBallot struct {
	suite.Suite
	home common.HomeNode
}

func (t *testINITBallot) SetupTest() {
	t.home = common.NewRandomHome()
}

func (t *testINITBallot) TestNew() {
	height := common.NewBig(33)
	round := Round(0)
	proposer := common.RandomSeed().Address()
	var validators []common.Address
	for i := 0; i < 5; i++ {
		validators = append(validators, common.RandomSeed().Address())
	}

	ballot := NewINITBallot(height, round, proposer, validators)

	err := ballot.Sign(common.TestNetworkID, t.home.Seed())
	t.NoError(err)

	err = ballot.Wellformed()
	t.NoError(err)

	_, ok := interface{}(ballot).(common.Seal)
	t.True(ok)

	_, ok = interface{}(ballot).(Ballot)
	t.True(ok)

	t.Equal(INITBallotSealType, ballot.Type())
	t.Equal(VoteStageINIT, ballot.Stage())
	t.Equal(height, ballot.Height())
	t.Equal(round, ballot.Round())
	t.Equal(proposer, ballot.Proposer())
	t.Equal(validators, ballot.Validators())
	t.Equal(VoteYES, ballot.Vote())
}

func (t *testINITBallot) TestMarshalBinary() {
	height := common.NewBig(33)
	round := Round(0)
	proposer := common.RandomSeed().Address()
	var validators []common.Address
	for i := 0; i < 5; i++ {
		validators = append(validators, common.RandomSeed().Address())
	}

	ballot := NewINITBallot(height, round, proposer, validators)

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

	var initBallot INITBallot
	err = common.CheckSeal(decoded, &initBallot)
	t.NoError(err)

	err = initBallot.Wellformed()
	t.NoError(err)

	t.Equal(INITBallotSealType, initBallot.Type())
	t.Equal(VoteStageINIT, initBallot.Stage())
	t.Equal(height, initBallot.Height())
	t.Equal(round, initBallot.Round())
	t.Equal(proposer, initBallot.Proposer())
	t.Equal(validators, initBallot.Validators())
	t.Equal(VoteYES, initBallot.Vote())
}

func TestINITBallot(t *testing.T) {
	suite.Run(t, new(testINITBallot))
}
