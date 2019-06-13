package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testDoubleSHA256Hash struct {
	suite.Suite
}

func (t *testDoubleSHA256Hash) TestNew() {
	hint := "block"
	value := []byte("show me")

	hash, err := NewHash(DoubleSHA256HashType, hint, value)
	t.NoError(err)
	t.NotEmpty(hash)

	_, ok := interface{}(hash).(Hash)
	t.True(ok)

	t.Equal(`{"algorithm":"double-sha256","body":"5NfGRdg6ex","hint":"block"}`, hash.String())
}

func (t *testDoubleSHA256Hash) TestEqual() {
	hint := "block"

	hash0, err := NewHash(DoubleSHA256HashType, hint, []byte("show me"))
	t.NoError(err)
	hash1, err := NewHash(DoubleSHA256HashType, hint, []byte("show me"))
	t.NoError(err)
	hash2, err := NewHash(DoubleSHA256HashType, hint, []byte("findme me"))
	t.NoError(err)

	t.True(hash0.Equal(hash1))
	t.True(hash1.Equal(hash0))
	t.False(hash0.Equal(hash2))
}

func (t *testDoubleSHA256Hash) TestIsValid() {
	{
		_, err := NewHash(DoubleSHA256HashType, "", []byte("show me")) // hint should be not empty
		t.Contains(err.Error(), "zero hint length")
	}

	{
		hash, err := NewHash(DoubleSHA256HashType, "hint", []byte("show me"))
		t.NoError(err)
		t.NoError(hash.IsValid())
	}
}

func (t *testDoubleSHA256Hash) TestMarshal() {
	hash, err := NewHash(DoubleSHA256HashType, "hint", []byte("show me"))
	t.NoError(err)

	b, err := hash.MarshalBinary()
	t.NoError(err)

	var uhash Hash
	err = uhash.UnmarshalBinary(b)
	t.NoError(err)

	t.NoError(uhash.IsValid())
	t.True(hash.Equal(uhash))
	t.Equal(DoubleSHA256HashType, hash.Algorithm())
}

func (t *testDoubleSHA256Hash) TestUnmarshal() {
	b := []byte("findme")

	var uhash Hash
	err := uhash.UnmarshalBinary(b)
	t.True(xerrors.Is(InvalidHashInputError, err))
}

func TestDoubleSHA256Hash(t *testing.T) {
	suite.Run(t, new(testDoubleSHA256Hash))
}

type testArgon2Hash struct {
	suite.Suite
}

func (t *testArgon2Hash) TestNew() {
	hint := "block"
	value := []byte("show me")

	hash, err := NewHash(Argon2HashType, hint, value)
	t.NoError(err)
	t.NotEmpty(hash)

	_, ok := interface{}(hash).(Hash)
	t.True(ok)

	t.Equal(`{"algorithm":"argon2","body":"5NfGRdg6ex","hint":"block"}`, hash.String())
}

func (t *testArgon2Hash) TestEqual() {
	hint := "block"

	hash0, err := NewHash(Argon2HashType, hint, []byte("show me"))
	t.NoError(err)
	hash1, err := NewHash(Argon2HashType, hint, []byte("show me"))
	t.NoError(err)
	hash2, err := NewHash(Argon2HashType, hint, []byte("findme me"))
	t.NoError(err)

	t.True(hash0.Equal(hash1))
	t.True(hash1.Equal(hash0))
	t.False(hash0.Equal(hash2))
}

func (t *testArgon2Hash) TestIsValid() {
	{
		_, err := NewHash(Argon2HashType, "", []byte("show me")) // hint should be not empty
		t.Contains(err.Error(), "zero hint length")
	}

	{
		hash, err := NewHash(Argon2HashType, "hint", []byte("show me"))
		t.NoError(err)
		t.NoError(hash.IsValid())
	}
}

func (t *testArgon2Hash) TestMarshal() {
	hash, err := NewHash(Argon2HashType, "hint", []byte("show me"))
	t.NoError(err)

	b, err := hash.MarshalBinary()
	t.NoError(err)

	var uhash Hash
	err = uhash.UnmarshalBinary(b)
	t.NoError(err)

	t.NoError(uhash.IsValid())
	t.True(hash.Equal(uhash))
	t.Equal(Argon2HashType, hash.Algorithm())
}

func (t *testArgon2Hash) TestUnmarshal() {
	b := []byte("findme")

	var uhash Hash
	err := uhash.UnmarshalBinary(b)
	t.True(xerrors.Is(InvalidHashInputError, err))
}

func TestArgon2Hash(t *testing.T) {
	suite.Run(t, new(testArgon2Hash))
}
