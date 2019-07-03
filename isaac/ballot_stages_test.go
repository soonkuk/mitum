package isaac

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type testBallotStages struct {
	suite.Suite
}

func (t *testBallotStages) TestINITBallot() {
	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)
	currentBlock := NewRandomBlockHash()
	proposal := NewRandomProposalHash()
	nextBlock := NewRandomBlockHash()

	ballot, err := NewBallot(n, height, round, StageINIT, proposal, currentBlock, nextBlock)
	t.NoError(err)

	pk, _ := keypair.NewStellarPrivateKey()
	{ // sign
		err = ballot.Sign(pk, []byte{})
		t.NoError(err)
	}

	t.NoError(ballot.IsValid())

	b, err := rlp.EncodeToBytes(ballot)
	t.NoError(err)

	var decoded BaseBallot
	err = rlp.DecodeBytes(b, &decoded)
	t.NoError(err)

	t.Equal(BallotType, decoded.Type())
	t.Equal(BallotType, decoded.Body().Type())
	t.Equal(n, decoded.Node())
	t.Equal(height, decoded.Height())
	t.Equal(round, decoded.Round())
	t.Equal(ballot.Stage(), decoded.Stage())
}

func (t *testBallotStages) TestInvalidINITBallot() {
	defer common.DebugPanic()

	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)

	{ // not empty proposal
		body := BaseBallotBody{
			Node:     n,
			Height:   height,
			Round:    round,
			Stage:    StageINIT,
			Proposal: NewRandomProposalHash(),
		}
		ballot, err := NewBaseBallot(body)
		t.NoError(err)

		pk, _ := keypair.NewStellarPrivateKey()
		{ // sign
			err = ballot.Sign(pk, []byte{})
			t.NoError(err)
		}

		err = ballot.IsValid()
		t.True(xerrors.Is(seal.InvalidSealError, err))

		err = body.IsValid()
		t.True(xerrors.Is(InvalidBallotError, err))
	}

	{ // not empty next block
		body := BaseBallotBody{
			Node:      n,
			Height:    height,
			Round:     round,
			Stage:     StageINIT,
			NextBlock: NewRandomBlockHash(),
		}
		ballot, err := NewBaseBallot(body)
		t.NoError(err)

		pk, _ := keypair.NewStellarPrivateKey()
		{ // sign
			err = ballot.Sign(pk, []byte{})
			t.NoError(err)
		}

		err = ballot.IsValid()
		t.True(xerrors.Is(seal.InvalidSealError, err))

		err = body.IsValid()
		t.True(xerrors.Is(InvalidBallotError, err))
	}
}

func (t *testBallotStages) TestInvalidSIGNBallot() {
	defer common.DebugPanic()

	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)

	{ // empty proposal
		body := BaseBallotBody{
			Node:      n,
			Height:    height,
			Round:     round,
			Stage:     StageSIGN,
			NextBlock: NewRandomBlockHash(),
		}
		ballot, err := NewBaseBallot(body)
		t.NoError(err)

		pk, _ := keypair.NewStellarPrivateKey()
		{ // sign
			err = ballot.Sign(pk, []byte{})
			t.NoError(err)
		}

		err = ballot.IsValid()
		t.True(xerrors.Is(seal.InvalidSealError, err))

		err = body.IsValid()
		t.True(xerrors.Is(InvalidBallotError, err))
	}

	{ // empty next block
		body := BaseBallotBody{
			Node:      n,
			Height:    height,
			Round:     round,
			Stage:     StageINIT,
			NextBlock: NewRandomBlockHash(),
		}
		ballot, err := NewBaseBallot(body)
		t.NoError(err)

		pk, _ := keypair.NewStellarPrivateKey()
		{ // sign
			err = ballot.Sign(pk, []byte{})
			t.NoError(err)
		}

		err = ballot.IsValid()
		t.True(xerrors.Is(seal.InvalidSealError, err))

		err = body.IsValid()
		t.True(xerrors.Is(InvalidBallotError, err))
	}
}

func (t *testBallotStages) TestInvalidACCEPTBallot() {
	defer common.DebugPanic()

	n := node.NewRandomAddress()
	height := NewBlockHeight(33)
	round := Round(0)

	{ // empty proposal
		body := BaseBallotBody{
			Node:      n,
			Height:    height,
			Round:     round,
			Stage:     StageACCEPT,
			NextBlock: NewRandomBlockHash(),
		}
		ballot, err := NewBaseBallot(body)
		t.NoError(err)

		pk, _ := keypair.NewStellarPrivateKey()
		{ // ACCEPT
			err = ballot.Sign(pk, []byte{})
			t.NoError(err)
		}

		err = ballot.IsValid()
		t.True(xerrors.Is(seal.InvalidSealError, err))

		err = body.IsValid()
		t.True(xerrors.Is(InvalidBallotError, err))
	}

	{ // empty next block
		body := BaseBallotBody{
			Node:      n,
			Height:    height,
			Round:     round,
			Stage:     StageACCEPT,
			NextBlock: NewRandomBlockHash(),
		}
		ballot, err := NewBaseBallot(body)
		t.NoError(err)

		pk, _ := keypair.NewStellarPrivateKey()
		{ // sign
			err = ballot.Sign(pk, []byte{})
			t.NoError(err)
		}

		err = ballot.IsValid()
		t.True(xerrors.Is(seal.InvalidSealError, err))

		err = body.IsValid()
		t.True(xerrors.Is(InvalidBallotError, err))
	}
}

func TestBallotStages(t *testing.T) {
	suite.Run(t, new(testBallotStages))
}
