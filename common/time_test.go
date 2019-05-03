package common

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type testTime struct {
	suite.Suite
}

func (t *testTime) TestNew() {
	nowString := "2019-04-16T16:39:43.665218000+09:00"
	now, _ := ParseISO8601(nowString)

	t.Equal(nowString, now.String())
}

func (t *testTime) TestJSON() {
	nowString := "2019-04-16T16:39:43.665218000+09:00"
	now, err := ParseISO8601(nowString)
	t.NoError(err)

	b, err := json.Marshal(now)
	t.NoError(err)

	var returned Time
	err = json.Unmarshal(b, &returned)
	t.NoError(err)

	t.Equal(now, returned)
}

func (t *testTime) TestBetween() {
	cases := []struct {
		name     string
		now      time.Time
		center   time.Time
		duration time.Duration
		expected bool
	}{
		{
			name:     "center: same with now",
			now:      time.Now(),
			duration: time.Second * 10,
			center:   time.Now(),
			expected: true,
		},
		{
			name:     "center: 9 second after",
			now:      time.Now(),
			duration: time.Second * 10,
			center:   time.Now().Add(time.Second * 9),
			expected: true,
		},
		{
			name:     "center: 10 second after",
			now:      time.Now(),
			duration: time.Second * 10,
			center:   time.Now().Add(time.Second * 10),
			expected: false,
		},
		{
			name:     "center: 9 second before",
			now:      time.Now(),
			duration: time.Second * 10,
			center:   time.Now().Add(time.Second * -9),
			expected: true,
		},
		{
			name:     "negative duration",
			now:      time.Now(),
			duration: time.Second * -10,
			center:   time.Now().Add(time.Second * -9),
			expected: true,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.T().Run(
			c.name,
			func(*testing.T) {
				m := Time{Time: c.now}
				result := m.Between(Time{Time: c.center}, c.duration)

				t.Equal(c.expected, result, "%d: %v", i, c.name)
			},
		)
	}

}

func TestTime(t *testing.T) {
	suite.Run(t, new(testTime))
}
