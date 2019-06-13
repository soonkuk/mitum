package encode

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testEncoders struct {
	suite.Suite
}

func (t *testEncoders) TestRegister() {
	ens := NewEncoders()

	{ // nothing registered
		encoded, err := ens.Encode([]byte("showme"))
		t.True(xerrors.Is(err, EncoderNotRegisteredError))
		t.Empty(encoded)
	}

	{
		err := ens.Register(RLP{})
		t.NoError(err)
		encoded, err := ens.Encode([]byte("showme"))
		t.NoError(err)
		t.NotEmpty(encoded)
		t.Equal(encoded, []byte{4, 0, 0, 0, 1, 0, 0, 0, 7, 0, 0, 0, 134, 115, 104, 111, 119, 109, 101})
	}
}

func (t *testEncoders) TestEncode() {
	ens := NewEncoders()

	err := ens.Register(RLP{})
	t.NoError(err)

	value := []byte("showme")

	encoded, err := ens.Encode(value)
	t.NoError(err)
	t.NotEmpty(encoded)
	t.Equal(encoded, []byte{4, 0, 0, 0, 1, 0, 0, 0, 7, 0, 0, 0, 134, 115, 104, 111, 119, 109, 101})

	var decoded []byte
	err = ens.Decode(encoded, &decoded)
	t.NoError(err)
	t.Equal(value, decoded)
}

func TestEncoders(t *testing.T) {
	suite.Run(t, new(testEncoders))
}
