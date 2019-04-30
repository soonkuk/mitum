package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testCountVoting struct {
	suite.Suite
}

func (t *testCountVoting) TestCanCountVoting() {
	cases := []struct {
		name      string
		total     uint
		threshold uint
		yes       int
		nop       int
		exp       int
		expected  bool
	}{
		{
			name:  "threshold > total",
			total: 10, threshold: 20,
			yes: 1, nop: 1, exp: 1,
			expected: false,
		},
		{
			name:  "not yet",
			total: 10, threshold: 7,
			yes: 1, nop: 1, exp: 1,
			expected: false,
		},
		{
			name:  "yes",
			total: 10, threshold: 7,
			yes: 7, nop: 1, exp: 1,
			expected: true,
		},
		{
			name:  "nop",
			total: 10, threshold: 7,
			yes: 1, nop: 7, exp: 1,
			expected: true,
		},
		{
			name:  "exp",
			total: 10, threshold: 7,
			yes: 1, nop: 1, exp: 7,
			expected: true,
		},
		{
			name:  "not draw",
			total: 10, threshold: 7,
			yes: 3, nop: 3, exp: 0,
			expected: false,
		},
		{
			name:  "draw",
			total: 10, threshold: 7,
			yes: 3, nop: 3, exp: 1,
			expected: true,
		},
		{
			name:  "yes over margin",
			total: 10, threshold: 7,
			yes: 4, nop: 0, exp: 0,
			expected: true,
		},
		{
			name:  "nop over margin",
			total: 10, threshold: 7,
			yes: 0, nop: 4, exp: 0,
			expected: true,
		},
		{
			name:  "exp over margin",
			total: 10, threshold: 7,
			yes: 0, nop: 0, exp: 4,
			expected: true,
		},
		{
			name:  "over total",
			total: 10, threshold: 7,
			yes: 4, nop: 4, exp: 4,
			expected: true,
		},
		{
			name:  "1 total 1 threshold",
			total: 1, threshold: 1,
			yes: 1, nop: 0, exp: 0,
			expected: true,
		},
	}

	for i, c := range cases {
		t.T().Run(
			c.name,
			func(*testing.T) {
				result := canCountVoting(c.total, c.threshold, c.yes, c.nop, c.exp)
				t.Equal(c.expected, result, "%d: %v", i, c.name)
			},
		)
	}
}

func (t *testCountVoting) TestMajority() {
	cases := []struct {
		name      string
		total     uint
		threshold uint
		yes       int
		nop       int
		exp       int
		expected  VoteResult
	}{
		{
			name:  "threshold > total",
			total: 10, threshold: 20,
			yes: 1, nop: 1, exp: 1,
			expected: VoteResultNotYet,
		},
		{
			name:  "not yet",
			total: 10, threshold: 7,
			yes: 1, nop: 1, exp: 1,
			expected: VoteResultNotYet,
		},
		{
			name:  "yes",
			total: 10, threshold: 7,
			yes: 7, nop: 1, exp: 1,
			expected: VoteResultYES,
		},
		{
			name:  "nop",
			total: 10, threshold: 7,
			yes: 1, nop: 7, exp: 1,
			expected: VoteResultNOP,
		},
		{
			name:  "exp",
			total: 10, threshold: 7,
			yes: 1, nop: 1, exp: 7,
			expected: VoteResultEXPIRE,
		},
		{
			name:  "not draw",
			total: 10, threshold: 7,
			yes: 3, nop: 3, exp: 0,
			expected: VoteResultNotYet,
		},
		{
			name:  "draw",
			total: 10, threshold: 7,
			yes: 3, nop: 3, exp: 1,
			expected: VoteResultDRAW,
		},
		{
			name:  "yes over margin",
			total: 10, threshold: 7,
			yes: 4, nop: 0, exp: 0,
			expected: VoteResultDRAW,
		},
		{
			name:  "nop over margin",
			total: 10, threshold: 7,
			yes: 0, nop: 4, exp: 0,
			expected: VoteResultDRAW,
		},
		{
			name:  "exp over margin",
			total: 10, threshold: 7,
			yes: 0, nop: 0, exp: 4,
			expected: VoteResultDRAW,
		},
		{
			name:  "over total",
			total: 10, threshold: 7,
			yes: 4, nop: 4, exp: 4,
			expected: VoteResultDRAW,
		},
		{
			name:  "1 total 1 threshold",
			total: 1, threshold: 1,
			yes: 1, nop: 0, exp: 0,
			expected: VoteResultYES,
		},
	}

	for i, c := range cases {
		t.T().Run(
			c.name,
			func(*testing.T) {
				result := majority(c.total, c.threshold, c.yes, c.nop, c.exp)
				t.Equal(c.expected, result, "%d: %v", i, c.name)
			},
		)
	}
}

func TestCountVoting(t *testing.T) {
	suite.Run(t, new(testCountVoting))
}
