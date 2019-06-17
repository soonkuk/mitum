package hash

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testHashes struct {
	suite.Suite
}

func (t *testHashes) TestRegister() {
	hs := NewHashes()
	err := hs.Register(NewArgon2Hash())
	t.NoError(err)

	// register again
	err = hs.Register(NewArgon2Hash())
	t.True(xerrors.Is(err, HashAlgorithmAlreadyRegisteredError))
}

func (t *testHashes) TestSetDefault() {
	hs := NewHashes()
	_ = hs.Register(NewArgon2Hash())

	err := hs.SetDefault(Argon2HashType)
	t.NoError(err)

	hash, err := hs.NewHash("test-hash", []byte("show me"))
	t.NoError(err)
	t.NotEmpty(hash)
}

func (t *testHashes) TestNewHash() {
	hs := NewHashes()
	_ = hs.Register(NewArgon2Hash())
	err := hs.SetDefault(Argon2HashType)
	t.NoError(err)

	hash, err := hs.NewHash("test-hash", []byte("show me"))
	t.NoError(err)
	t.NotEmpty(hash)
}

func (t *testHashes) TestNewHashButUnknownType() {
	hs := NewHashes()
	_ = hs.Register(NewArgon2Hash())
	err := hs.SetDefault(DoubleSHA256HashType)
	t.True(xerrors.Is(err, HashAlgorithmNotRegisteredError))
}

func (t *testHashes) TestUnmarshal() {
	hs := NewHashes()
	_ = hs.Register(NewArgon2Hash())
	err := hs.SetDefault(Argon2HashType)
	t.NoError(err)

	var b []byte
	hash, err := hs.NewHash("test-hash", []byte("show me"))
	t.NoError(err)

	b, err = hash.MarshalBinary()
	t.NoError(err)

	uhash, err := hs.UnmarshalHash(b)
	t.NoError(err)

	t.True(hash.Equal(uhash))
}

func (t *testHashes) TestMerge() {
	hs := NewHashes()
	_ = hs.Register(NewArgon2Hash())
	err := hs.SetDefault(Argon2HashType)
	t.NoError(err)
	_ = hs.Register(NewDoubleSHA256Hash())

	base, _ := hs.NewHash("test-hash", []byte("base"))

	var hashes []Hash
	for i := 0; i < 3; i++ {
		hash, _ := hs.NewHashByType(DoubleSHA256HashType, fmt.Sprintf("%d", i), []byte("others"))
		hashes = append(hashes, hash)
	}

	merged, err := hs.Merge(base, hashes...)
	t.NoError(err)
	t.NoError(merged.IsValid())

	t.Equal(base.Hint(), merged.Hint())
	t.True(base.Algorithm().Equal(merged.Algorithm()))
}

func (t *testHashes) TestMergeWithEmptyHash() {
	hs := NewHashes()
	_ = hs.Register(NewArgon2Hash())
	err := hs.SetDefault(Argon2HashType)
	t.NoError(err)
	_ = hs.Register(NewDoubleSHA256Hash())

	base, _ := hs.NewHash("test-hash", []byte("base"))

	var hashes []Hash
	for i := 0; i < 3; i++ {
		hash, _ := hs.NewHashByType(DoubleSHA256HashType, fmt.Sprintf("%d", i), []byte("others"))
		hashes = append(hashes, hash)
	}
	hashes = append(hashes, Hash{})

	merged, err := hs.Merge(base, hashes...)

	t.True(xerrors.Is(err, EmptyHashError))
	t.True(merged.Empty())
}

func TestHashes(t *testing.T) {
	suite.Run(t, new(testHashes))
}
