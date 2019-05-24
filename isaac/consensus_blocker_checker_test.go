package isaac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
)

type testConsensusBlockerChecker struct {
	suite.Suite
	state *ConsensusState
}

func (t *testConsensusBlockerChecker) SetupTest() {
	t.state = NewConsensusState(common.NewRandomHome())
	t.state.SetHeight(common.NewBig(33))
	t.state.SetBlock(common.NewRandomHash("bk"))
	t.state.SetState([]byte("state hash"))

}

func (t *testConsensusBlockerChecker) newProposal() Proposal {
	proposal := NewTestProposal(t.state.Home().Address(), nil)

	proposal.Block.Height = t.state.Height()
	proposal.Block.Current = t.state.Block()
	proposal.State.Current = t.state.State()
	proposal.State.Next = []byte("next state")
	proposal.Round = Round(0)

	return proposal
}

// TestProposalDifferentHeight checks,
// - proposal, which has different height on curret state is received
// - it will be ignored
func (t *testConsensusBlockerChecker) TestProposalDifferentHeight() {
	proposal := t.newProposal()

	// set different height
	{
		proposal.Block.Height = proposal.Block.Height.Inc()
		err := proposal.Sign(common.TestNetworkID, t.state.Home().Seed())
		t.NoError(err)
	}

	checker := common.NewChainChecker(
		"blocker-vote-proposal-checker",
		common.ContextWithValues(
			context.Background(),
			"proposal", proposal,
			"state", t.state,
		),
		CheckerBlockerProposalBlock,
	)
	checker.SetLogContext(
		"node", t.state.Home().Name(),
		"seal", proposal.Hash(),
		"seal_type", proposal.Type(),
	)
	err := checker.Check()
	t.True(common.SealIgnoredError.Equal(err))
}

// TestProposalDifferentBlock checks,
// - proposal, which has different block on curret state is received
// - it will be ignored
func (t *testConsensusBlockerChecker) TestProposalDifferentBlock() {
	proposal := t.newProposal()

	// set different block
	{
		proposal.Block.Current = common.NewRandomHash("bk")
		err := proposal.Sign(common.TestNetworkID, t.state.Home().Seed())
		t.NoError(err)
	}

	checker := common.NewChainChecker(
		"blocker-vote-proposal-checker",
		common.ContextWithValues(
			context.Background(),
			"proposal", proposal,
			"state", t.state,
		),
		CheckerBlockerProposalBlock,
	)
	checker.SetLogContext(
		"node", t.state.Home().Name(),
		"seal", proposal.Hash(),
		"seal_type", proposal.Type(),
	)
	err := checker.Check()
	t.True(common.SealIgnoredError.Equal(err))
}

func (t *testConsensusBlockerChecker) TestVotingResult() {
	cases := []struct {
		name   string
		result VoteResultInfo
		last   VoteResultInfo
		err    error
	}{
		{
			name: "proposal, but lower height",
			result: VoteResultInfo{ // proposal opened, but different height
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Height:   t.state.Height().Dec(),
				Round:    Round(1),
				Stage:    VoteStageINIT,
				Proposed: true, // after open proposal
			},
			err: common.SealIgnoredError,
		},
		{
			name: "proposal, but higher height",
			result: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Height:   t.state.Height().Inc(),
				Round:    Round(1),
				Stage:    VoteStageINIT,
				Proposed: true,
			},
			err: DifferentHeightConsensusError,
		},
		{
			name: "init ballot, but lower height",
			result: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Block:    t.state.Block(),
				Height:   t.state.Height().Dec(),
				Round:    Round(2),
				Stage:    VoteStageINIT,
			},
			err: common.SealIgnoredError,
		},
		{
			name: "init ballot, but higher height",
			result: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Block:    t.state.Block(),
				Height:   t.state.Height().Inc(),
				Round:    Round(2),
				Stage:    VoteStageINIT,
			},
			err: DifferentHeightConsensusError,
		},
		{
			name: "sign ballot, but last result is accept",
			result: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Height:   t.state.Height(),
				Round:    Round(1),
				Stage:    VoteStageSIGN,
			},
			last: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Height:   t.state.Height(),
				Round:    Round(1),
				Stage:    VoteStageACCEPT,
			},
			err: common.SealIgnoredError,
		},
		{
			name: "result.Block does not match with last result's in the next stage",
			result: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Height:   t.state.Height(),
				Round:    Round(1),
				Stage:    VoteStageACCEPT,
				Block:    common.NewRandomHash("bk"),
			},
			last: VoteResultInfo{
				Result:   VoteResultYES,
				Proposal: common.NewRandomHash("pp"),
				Proposer: t.state.Home().Address(),
				Height:   t.state.Height(),
				Round:    Round(1),
				Stage:    VoteStageSIGN,
				Block:    common.NewRandomHash("bk"),
			},
			err: ConsensusButBlockDoesNotMatchError,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.T().Run(
			c.name,
			func(*testing.T) {
				resultChecker := common.NewChainChecker(
					"blocker-vote-ballot-result-checker",
					common.ContextWithValues(
						context.Background(),
						"votingResult", c.result,
						"lastVotingResult", c.last,
						"state", t.state,
					),
					CheckerBlockerBallotVotingResult,
					CheckerBlockerVotingBallotResult,
				)
				err := resultChecker.Check()

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

func TestConsensusBlockerChecker(t *testing.T) {
	suite.Run(t, new(testConsensusBlockerChecker))
}
