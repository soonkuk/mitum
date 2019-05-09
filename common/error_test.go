package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testError struct {
	suite.Suite
}

func (t *testError) TestEscapeMessage() {
	err := NewError("t", 0, "1 < 2")
	t.Equal("{'code':'t-0','message':'1 < 2'}", strings.TrimSpace(err.Error()))
}

func TestError(t *testing.T) {
	suite.Run(t, new(testError))
}
