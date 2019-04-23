package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testVersion struct {
	suite.Suite
}

func (t *testVersion) TestEncodeDecode() {
	v, err := NewVersion("0.1.2-proto+findme")
	t.NoError(err)

	// encode
	b, err := v.MarshalBinary()
	t.NoError(err)
	t.NotEmpty(b)

	// decode
	var nv Version
	err = nv.UnmarshalBinary(b)
	t.NoError(err)

	t.True(v.Equal(nv))
}

func TestVersion(t *testing.T) {
	suite.Run(t, new(testVersion))
}
