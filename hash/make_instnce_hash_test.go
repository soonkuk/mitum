package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testMakeInstanceHash struct {
	suite.Suite
}

type testHashable0 struct {
	b string
}

func (t testHashable0) Hash() Hash {
	hash, _ := NewHash(DoubleSHA256HashType, "th", []byte(t.b))
	return hash
}

type testHashable1 struct {
	b string
}

func (t testHashable1) MarshalBinary() ([]byte, error) {
	return []byte(t.b), nil
}

func (t *testMakeInstanceHash) TestRegister() {
	hashes := NewHashes()
	err := hashes.Register(NewArgon2Hash())
	t.NoError(err)

	cases := []struct {
		name      string
		hashes    *Hashes
		hint      string
		input     interface{}
		expected  string
		err       error
		errString string
	}{
		{
			name:     "simple",
			hint:     "s",
			input:    []byte("showme"),
			expected: `s:6GYMhLR37x3qMerFZkNcVPnSo8tLqkfB5dMQbeSm2G25:argon2`,
		},
		{
			name:     "Hashable",
			hint:     "th",
			input:    testHashable0{b: "find me"},
			expected: `th:4t6WsYK5h6:double-sha256`,
		},
		{
			name:      "map will be failed",
			hint:      "m",
			input:     map[string]interface{}{},
			err:       HashFailedError,
			errString: "type map[string]interface {} is not RLP-serializable",
		},
		{
			name:     "MarshalBinary",
			hint:     "mb",
			input:    testHashable1{b: "kill me"},
			expected: `mb:H2P71TxSg3ZaRFXTwqUW8ma89ToCQCPz3bGWE8YtxdRR:argon2`,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.T().Run(
			c.name,
			func(*testing.T) {
				var hashes *Hashes = hashes
				if c.hashes != nil {
					hashes = c.hashes
				}

				hash, err := MakeInstanceHash(hashes, c.hint, c.input)
				if c.err != nil {
					t.True(xerrors.Is(err, c.err), "%d: %q", i, c.name)
					t.Contains(err.Error(), c.errString)
				} else if err != nil {
					t.NoError(err)
				}

				if err == nil {
					t.Equal(c.hint, hash.hint)
					t.NotEmpty(hash, "%d: %q", i, c.name)
					t.Equal(c.expected, hash.String(), "%d: %q", i, c.name)
				}
			},
		)
	}
}

func TestMakeInstanceHash(t *testing.T) {
	suite.Run(t, new(testMakeInstanceHash))
}
