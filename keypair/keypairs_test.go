package keypair

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testKeypairs struct {
	suite.Suite
}

func (t *testKeypairs) TestNew() {
	kp := NewKeypairs()
	err := kp.Register(Stellar{})
	t.NoError(err)

	key, err := kp.New()
	t.NoError(err)
	t.NotEmpty(key)

	t.Equal(StellarType, key.Type())
}

func (t *testKeypairs) TestRegister() {
	kp := NewKeypairs()
	err := kp.Register(Stellar{})
	t.NoError(err)

	// register again
	err = kp.Register(Stellar{})
	t.True(xerrors.Is(err, KeypairAlreadyRegisteredError))
}

func (t *testKeypairs) TestMarshalBinary() {
	kp := NewKeypairs()
	_ = kp.Register(Stellar{})

	pr, _ := kp.New()
	b, err := pr.MarshalBinary()
	t.NoError(err)
	t.NotEmpty(b)

	key, err := kp.NewFromBinary(b)
	t.NoError(err)

	t.True(pr.Equal(key))
}

func (t *testKeypairs) TestMarshalText() {
	kp := NewKeypairs()
	_ = kp.Register(Stellar{})

	pr, _ := kp.New()
	b, err := pr.MarshalText()
	t.NoError(err)
	t.NotEmpty(b)

	key, err := kp.NewFromText(b)
	t.NoError(err)

	t.True(pr.Equal(key))
}

func TestKeypairs(t *testing.T) {
	suite.Run(t, new(testKeypairs))
}
