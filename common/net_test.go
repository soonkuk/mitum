package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testNetAddr struct {
	suite.Suite
}

func (t testNetAddr) TestBaseNode() {
	cases := []struct {
		name     string
		input    string
		network  string
		expected string
		err      error
	}{
		{
			name:     "normal",
			input:    "dummy://1.2.3.4",
			network:  "dummy",
			expected: "dummy://1.2.3.4",
		},
		{
			name:     "normal",
			input:    "dummy://1.2.3.4",
			network:  "dummy",
			expected: "dummy://1.2.3.4",
		},
		{
			name:  "empty scheme",
			input: "://showme",
			err:   InvalidNetAddrError,
		},
		{
			name:     "with query",
			input:    "killme://showme?b=1&a=2",
			network:  "killme",
			expected: "killme://showme?b=1&a=2",
		},
		{
			name:    "with bad port",
			input:   "killme://showme:abcd",
			network: "killme",
			err:     InvalidNetAddrError,
		},
	}

	for i, c := range cases {
		i := i // NOTE for govet
		c := c // NOTE for govet
		t.T().Run(
			c.name,
			func(*testing.T) {
				addr, err := NewNetAddr(c.input)
				if c.err != nil {
					if e, ok := c.err.(Error); ok {
						t.True(e.Equal(err))
					} else {
						t.Equal(c.err, err)
					}
					return
				}

				t.Equal(c.expected, addr.String(), "%d: %v", i, c.name)
				t.Equal(c.network, addr.Network(), "%d: %v", i, c.name)

				b, err := addr.MarshalBinary()
				t.NoError(err)

				var unmarshaled NetAddr
				err = unmarshaled.UnmarshalBinary(b)
				t.NoError(err)
				t.Equal(c.expected, unmarshaled.String(), "%d: %v", i, c.name)
				t.Equal(c.network, unmarshaled.Network(), "%d: %v", i, c.name)
			},
		)
	}

}

func TestNetAddr(t *testing.T) {
	suite.Run(t, new(testNetAddr))
}
