package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testContextWithValues struct {
	suite.Suite
}

func (t *testContextWithValues) TestBasic() {
	ctx := context.Background()
	newCtx := ContextWithValues(
		ctx,
		"showme",
		1,
		2,
		3,
	)
	t.Equal(1, newCtx.Value("showme"))
	t.Equal(3, newCtx.Value(2))
}

func (t *testContextWithValues) TestBadNumberArgs() {
	ctx := context.Background()
	t.Panics(func() {
		ContextWithValues(
			ctx,
			"showme",
			1,
			3,
		)
	})
}

func TestContextWithValues(t *testing.T) {
	suite.Run(t, new(testContextWithValues))
}
