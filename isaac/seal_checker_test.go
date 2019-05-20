package isaac

import (
	"context"
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testSealChecker struct {
	suite.Suite
}

func (t *testSealChecker) TestCheckerSealTypesSealedSeal() {
	ballot := NewTestSealBallot(
		common.NewRandomHash("bl"),
		common.RandomSeed().Address(),
		common.NewBig(33),
		Round(2),
		VoteStageSIGN,
		VoteYES,
		common.NewRandomHash("bk"),
	)

	seal := ballot.(SIGNBallot)
	err := (&seal).Sign(common.TestNetworkID, common.RandomSeed())
	t.NoError(err)

	sealed, err := common.NewSealedSeal(ballot)
	t.NoError(err)

	err = sealed.Sign(common.TestNetworkID, common.RandomSeed())
	t.NoError(err)

	ctx := context.Background()
	checker := common.NewChainChecker(
		"receiveSeal",
		common.ContextWithValues(
			ctx,
			"seal", sealed,
		),
		CheckerSealTypes,
	)
	err = checker.Check()
	t.True(common.UnknownSealTypeError.Equal(err))
}

func TestSealChecker(t *testing.T) {
	suite.Run(t, new(testSealChecker))
}
