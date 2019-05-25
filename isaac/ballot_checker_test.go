package isaac

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
)

type testBallotChecker struct {
	suite.Suite
	state            *ConsensusState
	proposal         Proposal
	sealPool         *DefaultSealPool
	proposerSelector *FixedProposerSelector
}

func (t *testBallotChecker) SetupTest() {
	t.state = NewConsensusState(common.NewRandomHome())
	t.state.SetHeight(common.NewBig(33))
	t.state.SetBlock(common.NewRandomHash("bk"))
	t.state.SetState([]byte("state hash"))

	t.proposerSelector = NewFixedProposerSelector()
	t.proposerSelector.SetProposer(t.state.Home())

	proposal := NewTestProposal(t.state.Home().Address(), nil)
	proposal.Block.Height = t.state.Height()
	proposal.Block.Current = t.state.Block()
	proposal.State.Current = t.state.State()
	proposal.State.Next = []byte("next state")
	proposal.Round = Round(0)
	_ = proposal.Sign(common.TestNetworkID, t.state.Home().Seed())

	t.proposal = proposal

	t.sealPool = NewDefaultSealPool()
	_ = t.sealPool.Add(proposal)
}

func (t *testBallotChecker) TestCases() {
	cases := []struct {
		name           string
		ballotProposal common.Hash
		ballotStage    VoteStage
		ballotHeight   common.Big
		ballotRound    Round
		ballotBlock    common.Hash
		ballotProposer common.Address
		err            error
	}{
		{
			name:           "ballot has higher height",
			ballotProposal: t.proposal.Hash(),
			ballotStage:    VoteStageINIT,
			ballotHeight:   t.proposal.Block.Height,
			ballotRound:    t.proposal.Round,
			ballotBlock:    t.proposal.Block.Current,
			ballotProposer: t.proposal.Source(),
			err:            nil,
		},
		{
			name:           "ballot has lower height",
			ballotProposal: t.proposal.Hash(),
			ballotStage:    VoteStageINIT,
			ballotHeight:   t.proposal.Block.Height.Dec(),
			ballotRound:    t.proposal.Round,
			ballotBlock:    t.proposal.Block.Current,
			ballotProposer: t.proposal.Source(),
			err:            common.SealIgnoredError,
		},
		{
			name:           "ballot has unknown proposal",
			ballotProposal: common.NewRandomHash("pp"),
			ballotStage:    VoteStageSIGN,
			ballotHeight:   t.proposal.Block.Height,
			ballotRound:    t.proposal.Round,
			ballotBlock:    t.proposal.Block.Current,
			ballotProposer: t.proposal.Source(),
			err:            SealNotFoundError,
		},
		{
			name:           "ballot has different height with proposal",
			ballotProposal: t.proposal.Hash(),
			ballotStage:    VoteStageSIGN,
			ballotHeight:   t.proposal.Block.Height.Inc(),
			ballotRound:    t.proposal.Round,
			ballotBlock:    t.proposal.Block.Current,
			ballotProposer: t.proposal.Source(),
			err:            BallotNotWellformedError,
		},
		{
			name:           "ballot has different round with proposal",
			ballotProposal: t.proposal.Hash(),
			ballotStage:    VoteStageSIGN,
			ballotHeight:   t.proposal.Block.Height,
			ballotRound:    t.proposal.Round + 1,
			ballotBlock:    t.proposal.Block.Current,
			ballotProposer: t.proposal.Source(),
			err:            BallotNotWellformedError,
		},
		{
			name:           "ballot has different proposer with proposal",
			ballotProposal: t.proposal.Hash(),
			ballotStage:    VoteStageSIGN,
			ballotHeight:   t.proposal.Block.Height,
			ballotRound:    t.proposal.Round,
			ballotBlock:    t.proposal.Block.Current,
			ballotProposer: common.RandomSeed().Address(),
			err:            BallotNotWellformedError,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.T().Run(
			c.name,
			func(*testing.T) {
				var signer common.Signer
				switch c.ballotStage {
				case VoteStageINIT:
					ballot := NewINITBallot(
						c.ballotHeight,
						c.ballotRound,
						c.ballotProposer,
						nil,
					)
					b := &ballot
					signer = interface{}(b).(common.Signer)
				case VoteStageSIGN:
					ballot := NewSIGNBallot(
						c.ballotHeight,
						c.ballotRound,
						c.ballotProposer,
						nil,
						c.ballotProposal,
						c.ballotBlock,
						VoteYES,
					)
					b := &ballot
					signer = interface{}(b).(common.Signer)
				case VoteStageACCEPT:
					ballot := NewACCEPTBallot(
						c.ballotHeight,
						c.ballotRound,
						c.ballotProposer,
						nil,
						c.ballotProposal,
						c.ballotBlock,
					)
					b := &ballot
					signer = interface{}(b).(common.Signer)
				}

				_ = signer.Sign(common.TestNetworkID, t.state.Home().Seed())
				ballot := reflect.ValueOf(signer).Elem().Interface().(Ballot)

				ctx := common.ContextWithValues(
					context.Background(),
					"ballot", ballot,
					"state", t.state,
					"sealPool", t.sealPool,
					"proposerSelector", t.proposerSelector,
					"proposal", t.proposal,
				)
				checker := common.NewChainChecker(
					"ballot-checker",
					ctx,
					CheckerBallotHasValidState,
					CheckerBallotProposal,
					CheckerBallotHasValidProposal,
					CheckerBallotHasValidProposer,
				)
				err := checker.Check()

				if c.err == nil {
					t.NoError(err)
				} else {
					cerr, ok := c.err.(common.Error)
					if ok {
						t.True(cerr.Equal(err), "%d: %v, error=%v", i, c.name, err)
					} else {
						t.Equal(c.err, err)
					}
				}
			},
		)
	}
}

func (t *testBallotChecker) TestBallotProposalHasInvalidProposer() {
	ballot := NewSIGNBallot(
		t.proposal.Block.Height,
		t.proposal.Round,
		t.proposal.Source(),
		nil,
		t.proposal.Hash(),
		t.proposal.Block.Current,
		VoteYES,
	)
	_ = ballot.Sign(common.TestNetworkID, t.state.Home().Seed())

	proposerSelector := NewFixedProposerSelector()
	proposerSelector.SetProposer(common.NewRandomHome())

	ctx := common.ContextWithValues(
		context.Background(),
		"ballot", ballot,
		"state", t.state,
		"sealPool", t.sealPool,
		"proposerSelector", proposerSelector,
		"proposal", t.proposal,
	)
	checker := common.NewChainChecker(
		"ballot-checker",
		ctx,
		CheckerBallotHasValidState,
		CheckerBallotProposal,
		CheckerBallotHasValidProposal,
		CheckerBallotHasValidProposer,
	)
	err := checker.Check()
	t.True(BallotHasInvalidProposerError.Equal(err))
}

func TestBallotChecker(t *testing.T) {
	suite.Run(t, new(testBallotChecker))
}
