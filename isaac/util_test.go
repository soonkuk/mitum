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
		expected  bool
	}{
		{
			name:  "threshold > total; yes",
			total: 10, threshold: 20,
			yes: 10, nop: 0,
			expected: true,
		},
		{
			name:  "threshold > total; o",
			total: 10, threshold: 20,
			yes: 0, nop: 10,
			expected: true,
		},
		{
			name:  "not yet",
			total: 10, threshold: 7,
			yes: 1, nop: 1,
			expected: false,
		},
		{
			name:  "yes",
			total: 10, threshold: 7,
			yes: 7, nop: 1,
			expected: true,
		},
		{
			name:  "nop",
			total: 10, threshold: 7,
			yes: 1, nop: 7,
			expected: true,
		},
		{
			name:  "not draw #0",
			total: 10, threshold: 7,
			yes: 3, nop: 3,
			expected: false,
		},
		{
			name:  "not draw #1",
			total: 10, threshold: 7,
			yes: 0, nop: 4,
			expected: false,
		},
		{
			name:  "draw #0",
			total: 10, threshold: 7,
			yes: 4, nop: 4,
			expected: true,
		},
		{
			name:  "draw #1",
			total: 10, threshold: 7,
			yes: 5, nop: 5,
			expected: true,
		},
		{
			name:  "over total",
			total: 10, threshold: 7,
			yes: 4, nop: 4,
			expected: true,
		},
		{
			name:  "1 total 1 threshold",
			total: 1, threshold: 1,
			yes: 1, nop: 0,
			expected: true,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.T().Run(
			c.name,
			func(*testing.T) {
				result := canCountVoting(c.total, c.threshold, c.yes, c.nop)
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
		expected  VoteResult
	}{
		{
			name:  "threshold > total; yes",
			total: 10, threshold: 20,
			yes: 10, nop: 0,
			expected: VoteResultYES,
		},
		{
			name:  "threshold > total; nop",
			total: 10, threshold: 20,
			yes: 0, nop: 10,
			expected: VoteResultNOP,
		},
		{
			name:  "not yet",
			total: 10, threshold: 7,
			yes: 1, nop: 1,
			expected: VoteResultNotYet,
		},
		{
			name:  "yes",
			total: 10, threshold: 7,
			yes: 7, nop: 1,
			expected: VoteResultYES,
		},
		{
			name:  "nop",
			total: 10, threshold: 7,
			yes: 1, nop: 7,
			expected: VoteResultNOP,
		},
		{
			name:  "not draw #1",
			total: 10, threshold: 7,
			yes: 3, nop: 3,
			expected: VoteResultNotYet,
		},
		{
			name:  "not draw #1",
			total: 10, threshold: 7,
			yes: 4, nop: 0,
			expected: VoteResultNotYet,
		},
		{
			name:  "draw #0",
			total: 10, threshold: 7,
			yes: 4, nop: 4,
			expected: VoteResultDRAW,
		},
		{
			name:  "draw #1",
			total: 10, threshold: 7,
			yes: 5, nop: 5,
			expected: VoteResultDRAW,
		},
		{
			name:  "over total",
			total: 10, threshold: 7,
			yes: 4, nop: 10,
			expected: VoteResultNOP,
		},
		{
			name:  "1 total 1 threshold",
			total: 1, threshold: 1,
			yes: 1, nop: 0,
			expected: VoteResultYES,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.T().Run(
			c.name,
			func(*testing.T) {
				result := majority(c.total, c.threshold, c.yes, c.nop)
				t.Equal(c.expected, result, "%d: %v; %v != %v", i, c.name, c.expected, result)
			},
		)
	}
}

func TestCountVoting(t *testing.T) {
	suite.Run(t, new(testCountVoting))
}
