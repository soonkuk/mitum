package common

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testSealCheckers struct {
	suite.Suite
}

func (t testSealCheckers) newSealMessage() (Seal, []byte) {
	body := sealTestBody{A: 1, B: "b"}

	st := NewSealType("body")
	seal, _ := NewSeal(st, body)

	seed := RandomSeed()
	err := seal.Sign(TestNetworkID, seed)
	t.NoError(err)

	b, err := seal.MarshalBinary()
	t.NoError(err)
	return seal, b

}

func (t testSealCheckers) TestCheckerUnmarshalSeal() {
	ctx := context.WithValue(context.Background(), "networkID", TestNetworkID)

	_, message := t.newSealMessage()
	ctx = context.WithValue(ctx, "message", message)

	checker := NewChainChecker(
		"seal-checker",
		ctx,
		CheckerUnmarshalSeal,
	)

	err := checker.Check()
	t.NoError(err)

	checkedSeal := checker.Context().Value("seal")
	t.NotNil(checkedSeal)
}

func (t testSealCheckers) TestCheckerUnmarshalSealFailed() {
	ctx := context.WithValue(context.Background(), "networkID", TestNetworkID)

	var checker *ChainChecker

	{ // invalid message
		sealMessage := []byte{}
		ctx = context.WithValue(ctx, "message", sealMessage)

		checker = NewChainChecker(
			"seal-checker",
			ctx,
			CheckerUnmarshalSeal,
		)

		err := checker.Check()
		t.Equal(io.EOF, err)
	}

	{ // bad seal
		seal, _ := t.newSealMessage()
		seal.hash = NewRandomHash("bad")
		message, err := seal.MarshalBinary()
		t.NoError(err)

		ctx = context.WithValue(ctx, "message", message)
		ctx = context.WithValue(ctx, "networkID", NetworkID("bad-network-id"))

		checker = checker.New(ctx)

		err = checker.Check()
		t.True(SignatureVerificationFailedError.Equal(err))
	}

	{ // invalid networkID
		_, message := t.newSealMessage()
		ctx = context.WithValue(ctx, "message", message)
		ctx = context.WithValue(ctx, "networkID", NetworkID("bad-network-id"))

		checker = checker.New(ctx)

		err := checker.Check()
		t.True(SignatureVerificationFailedError.Equal(err))
	}
}

func TestSealCheckers(t *testing.T) {
	suite.Run(t, new(testSealCheckers))
}
