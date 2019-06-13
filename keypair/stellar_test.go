package keypair

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testStellarKeypair struct {
	suite.Suite
}

func (t *testStellarKeypair) TestNew() {
	st0, _ := Stellar{}.New()
	t.Equal(StellarType, st0.Type())

	st1, _ := Stellar{}.New()
	t.False(st0.Equal(st1))
}

func (t *testStellarKeypair) TestPublicKey() {
	st, _ := Stellar{}.New()
	pk := st.PublicKey()
	t.Equal(StellarType, pk.Type())
	t.NotEmpty(pk)
	t.True(pk.Equal(pk))
	t.Regexp(regexp.MustCompile(`"key":[\s]*"G`), pk.String())
}

func (t *testStellarKeypair) TestMarshalBinary() {
	pr, _ := Stellar{}.New()

	{
		b, err := pr.MarshalBinary()
		t.NoError(err)
		t.NotEmpty(b)

		key, err := Stellar{}.NewFromBinary(b)
		t.NoError(err)
		t.NotEmpty(key)

		t.Equal(PrivateKeyKind, key.Kind())
		t.True(pr.Equal(key))
	}

	{
		pk := pr.PublicKey()

		b, err := pk.MarshalBinary()
		t.NoError(err)
		t.NotEmpty(b)

		key, err := Stellar{}.NewFromBinary(b)
		t.NoError(err)
		t.NotEmpty(key)

		t.Equal(PublicKeyKind, key.Kind())
		t.True(pk.Equal(key))
	}
}

func (t *testStellarKeypair) TestMarshalText() {
	pr, _ := Stellar{}.New()

	{
		b, err := pr.MarshalText()
		t.NoError(err)
		t.NotEmpty(b)

		key, err := Stellar{}.NewFromText(b)
		t.NoError(err)
		t.NotEmpty(key)

		t.Equal(PrivateKeyKind, key.Kind())
		t.True(pr.Equal(key))
	}

	{
		pk := pr.PublicKey()

		b, err := pk.MarshalText()
		t.NoError(err)
		t.NotEmpty(b)

		key, err := Stellar{}.NewFromText(b)
		t.NoError(err)
		t.NotEmpty(key)

		t.Equal(PublicKeyKind, key.Kind())
		t.True(pk.Equal(key))
	}
}

func (t *testStellarKeypair) TestMarshalBinaryPublicKey() {
	st, _ := Stellar{}.New()
	pk := st.PublicKey()

	b, err := pk.MarshalBinary()
	t.NoError(err)
	t.NotEmpty(b)

	var upk StellarPublicKey
	err = upk.UnmarshalBinary(b)
	t.NoError(err)

	var upr StellarPrivateKey
	err = upr.UnmarshalBinary(b)
	t.True(xerrors.Is(err, FailedToUnmarshalKeypairError))
	t.Contains(err.Error(), "is not private key")
}

func (t *testStellarKeypair) TestMarshalBinaryPrivateKey() {
	st := Stellar{}
	pr, _ := st.New()

	b, err := pr.MarshalBinary()
	t.NoError(err)
	t.NotEmpty(b)

	var upr StellarPrivateKey
	err = upr.UnmarshalBinary(b)
	t.NoError(err)

	var upk StellarPublicKey
	err = upk.UnmarshalBinary(b)
	t.True(xerrors.Is(err, FailedToUnmarshalKeypairError))
	t.Contains(err.Error(), "is not public key")
}

func (t *testStellarKeypair) TestMarshalTextPublicKey() {
	st := Stellar{}
	pr, _ := st.New()
	pk := pr.PublicKey()

	b, err := pk.MarshalText()
	t.NoError(err)
	t.NotEmpty(b)

	var upk StellarPublicKey
	err = upk.UnmarshalText(b)
	t.NoError(err)

	t.True(pk.Equal(upk))

	{ // with private key
		b, err := pr.MarshalText()
		t.NoError(err)

		var upr StellarPublicKey
		err = upr.UnmarshalText(b)
		t.True(xerrors.Is(err, FailedToUnmarshalKeypairError))
		t.Contains(err.Error(), "is not public key")
	}
}

func (t *testStellarKeypair) TestMarshalTextPrivateKey() {
	st := Stellar{}
	pr, _ := st.New()

	b, err := pr.MarshalText()
	t.NoError(err)
	t.NotEmpty(b)

	var upr StellarPrivateKey
	err = upr.UnmarshalText(b)
	t.NoError(err)

	t.True(pr.Equal(upr))

	{ // with public key
		pk := pr.PublicKey()
		b, err := pk.MarshalText()
		t.NoError(err)

		var upr StellarPrivateKey
		err = upr.UnmarshalText(b)
		t.True(xerrors.Is(err, FailedToUnmarshalKeypairError))
		t.Contains(err.Error(), "is not private key")
	}
}

func (t *testStellarKeypair) TestSigning() {
	st := Stellar{}
	pr, _ := st.New()

	input := []byte("source")
	sig, err := pr.Sign(input)
	t.NoError(err)
	t.NotEmpty(sig)

	{ // valid input
		err = pr.PublicKey().Verify(input, sig)
		t.NoError(err)
	}

	{ // invalid input
		err = pr.PublicKey().Verify([]byte("killme"), sig)
		t.True(xerrors.Is(err, SignatureVerificationFailedError))
	}
}

func (t *testStellarKeypair) TestFromSeed() {
	seed := []byte("find me")

	pr0, _ := Stellar{}.NewFromSeed(seed)
	pr1, _ := Stellar{}.NewFromSeed(seed)

	t.True(pr0.Equal(pr1))
	t.True(pr0.PublicKey().Equal(pr1.PublicKey()))
}

func TestStellarKeypair(t *testing.T) {
	suite.Run(t, new(testStellarKeypair))
}
