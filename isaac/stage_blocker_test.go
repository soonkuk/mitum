package isaac

import (
	"math/rand"
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testStageBlocker struct {
	suite.Suite
	sb       *StageBlocker
	sendChan chan common.Seal
}

func (t *testStageBlocker) SetupTest() {
	t.sb = NewStageBlocker()
	err := t.sb.Start()
	t.NoError(err)

	t.sendChan = make(chan common.Seal)
}

func (t *testStageBlocker) TearDownTest() {
	defer t.sb.Stop()
	defer close(t.sendChan)
}

func (t *testStageBlocker) randomVoteResultInfo() VoteResultInfo {
	var result VoteResultInfo

	result.Result = VoteResultYES
	result.Stage = VoteStageINIT
	result.Proposal = common.NewRandomHash("sl")
	result.Height = common.NewBig(uint64(rand.Intn(100)))
	result.Round = Round(rand.Intn(100))
	result.LastVotedAt = common.Now()
	result.Proposed = false

	return result
}

// TestCheckNotYet, not yet result will be just ignored
func (t *testStageBlocker) TestCheckNotYet() {
	defer common.DebugPanic()

	expectedResult := t.randomVoteResultInfo()
	expectedResult.Result = VoteResultNotYet

	resultChan := t.sb.Check(expectedResult)
	result := <-resultChan
	t.Error(result.Err, InvalidVoteResultInfoError)
}

// TestNewProposeAccepted, new proposal will broadcast sign ballot
func (t *testStageBlocker) TestNewProposeAccepted() {
	expectedResult := t.randomVoteResultInfo()
	expectedResult.Stage = VoteStageINIT
	expectedResult.Proposed = true // NOTE important from just propose seal

	resultChan := t.sb.Check(expectedResult)
	result := <-resultChan
	t.Equal(result.Decision, ProposalAccepted)
}

// TestCheckSIGN, sign ballot will broadcast ACCEPT ballot
func (t *testStageBlocker) TestCheckSIGN() {
	defer common.DebugPanic()

	expectedResult := t.randomVoteResultInfo()
	expectedResult.Stage = VoteStageSIGN

	resultChan := t.sb.Check(expectedResult)
	result := <-resultChan
	t.Equal(result.Decision, GoToNextStage)
}

// TestCheckSIGNButNOP, sign, but nop; it will start new round
func (t *testStageBlocker) TestCheckSIGNButNOP() {
	defer common.DebugPanic()

	expectedResult := t.randomVoteResultInfo()
	expectedResult.Stage = VoteStageSIGN
	expectedResult.Result = VoteResultNOP

	resultChan := t.sb.Check(expectedResult)
	result := <-resultChan
	t.Equal(result.Decision, GoToNextRound)
}

// TestCheckACCEPT, sign ballot will broadcast INIT ballot for next block
func (t *testStageBlocker) TestCheckACCEPT() {
	defer common.DebugPanic()

	expectedResult := t.randomVoteResultInfo()
	expectedResult.Stage = VoteStageACCEPT

	resultChan := t.sb.Check(expectedResult)
	result := <-resultChan
	t.Equal(result.Decision, FinishRound)
}

// TestINITRound, new proposal will broadcast sign ballot
func (t *testStageBlocker) TestINITRound() {
	expectedResult := t.randomVoteResultInfo()
	expectedResult.Stage = VoteStageINIT
	expectedResult.Proposed = false

	resultChan := t.sb.Check(expectedResult)
	result := <-resultChan
	t.Equal(result.Decision, StartNewRound)
}

func TestStageBlocker(t *testing.T) {
	suite.Run(t, new(testStageBlocker))
}
