package isaac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type testPolicy struct {
	suite.Suite
}

func (t *testPolicy) TestNew() {
	th, err := NewThreshold(10, 66)
	t.NoError(err)

	p := Policy{Threshold: th}

	{
		total, threshold := th.Get(StageINIT)
		pTotal, pThreshold := p.Threshold.Get(StageINIT)
		t.Equal(total, pTotal)
		t.Equal(threshold, pThreshold)
	}

	{
		th.SetBase(6, 4)
		total, threshold := th.Get(StageINIT)
		pTotal, pThreshold := p.Threshold.Get(StageINIT)
		t.Equal(total, pTotal)
		t.Equal(threshold, pThreshold)
	}

	{
		th.Set(StageSIGN, 20, 17)
		total, threshold := th.Get(StageSIGN)
		pTotal, pThreshold := p.Threshold.Get(StageSIGN)
		t.Equal(total, pTotal)
		t.Equal(threshold, pThreshold)
	}
}

func (t *testPolicy) TestCopy() {
	th, err := NewThreshold(10, 66)
	t.NoError(err)

	policy := Policy{Threshold: th}

	check := func(p Policy, stage Stage) {
		total, threshold := th.Get(StageINIT)
		pTotal, pThreshold := p.Threshold.Get(StageINIT)
		t.Equal(total, pTotal)
		t.Equal(threshold, pThreshold)
	}

	{
		th.SetBase(6, 4)
		check(policy, StageINIT)
	}

	{
		th.Set(StageSIGN, 20, 17)
		check(policy, StageSIGN)
	}
}

func TestPolicy(t *testing.T) {
	suite.Run(t, new(testPolicy))
}

func TestThreshold(t *testing.T) {
	cases := []struct {
		name     string
		total    uint
		percent  float64
		expected uint
		err      string
	}{
		{
			name:  "basic",
			total: 10, percent: 66,
			expected: 7,
		},
		{
			name:  "over 100",
			total: 10, percent: 166,
			err: "is over 100",
		},
		{
			name:  "100 percent",
			total: 10, percent: 100,
			expected: 10,
		},
		{
			name:  "100 percent",
			total: 33, percent: 33.45,
			expected: 12,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.Run(
			c.name,
			func(*testing.T) {
				result, err := NewThreshold(c.total, c.percent)
				if len(c.err) > 0 {
					assert.Contains(t, err.Error(), c.err)
				} else {
					assert.Equal(t, c.expected, result.base[1], "%d: %v; %v != %v", i, c.name, c.expected, result.base[1])
				}
			},
		)
	}
}
