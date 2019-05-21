package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/storage/leveldbstorage"
)

type testProposalValidation struct {
	suite.Suite
	home  common.HomeNode
	state *ConsensusState
	st    *leveldbstorage.Storage
}

func (t *testProposalValidation) SetupTest() {
	t.home = common.NewRandomHome()
	t.state = NewConsensusState(t.home)
	t.st = leveldbstorage.NewMemStorage()
}

func TestProposalValidation(t *testing.T) {
	suite.Run(t, new(testProposalValidation))
}
