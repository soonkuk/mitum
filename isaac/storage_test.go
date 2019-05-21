package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/storage/leveldbstorage"
)

type testStorage struct {
	suite.Suite
}

func (t *testStorage) TestNewBlock() {
	st := leveldbstorage.NewMemStorage()

	bs, err := NewDefaultBlockStorage(st)
	t.NoError(err)

	_, ok := interface{}(bs).(BlockStorage)
	t.True(ok)

	// store new block
	home := common.NewRandomHome()
	proposal := NewTestProposal(home.Address(), nil)

	{ // correcting proposal
		proposal.Block.Height = common.NewBig(33)
		proposal.Block.Current = common.NewRandomHash("bk")
		proposal.State.Current = []byte("current state")
		proposal.State.Next = []byte("next state")
		proposal.Round = 0
	}

	err = proposal.Sign(common.TestNetworkID, home.Seed())
	t.NoError(err)

	block, batch, err := bs.NewBlock(proposal)
	t.NoError(err)
	t.NotEmpty(block)
	t.NotEmpty(batch)
}

func TestStorage(t *testing.T) {
	suite.Run(t, new(testStorage))
}
