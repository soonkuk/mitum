package keypair

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spikeekips/mitum/common"
	stellarHash "github.com/stellar/go/hash"
	stellarKeypair "github.com/stellar/go/keypair"
)

var (
	StellarType Type = NewType(1, "stellar")
)

type Stellar struct {
}

func (s Stellar) Type() Type {
	return StellarType
}

// New generates the new random keypair
func (s Stellar) New() (PrivateKey, error) {
	seed, err := stellarKeypair.Random()
	if err != nil {
		return nil, err
	}

	return StellarPrivateKey{kp: seed}, nil
}

// NewFromSeed generates the keypair from raw seed
func (s Stellar) NewFromSeed(b []byte) (PrivateKey, error) {
	seed, err := stellarKeypair.FromRawSeed(stellarHash.Hash(b))
	if err != nil {
		return nil, err
	}

	return StellarPrivateKey{kp: seed}, nil
}

func (s Stellar) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s Stellar) Equal(k Keypair) bool {
	return s.Type().Equal(k.Type())
}

func (s Stellar) NewFromBinary(b []byte) (Key, error) {
	e, o := common.ExtractBinary(b)
	if o < 0 {
		return nil, FailedToUnmarshalKeypairError.Newf("stellar key; failed to parse Type")
	}

	var kt Type
	if err := kt.UnmarshalBinary(e); err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	} else if !kt.Equal(StellarType) {
		return nil, FailedToUnmarshalKeypairError.Newf(
			"stellar key; not stellar keypair; type=%q",
			kt.String(),
		)
	}

	offset := o

	e, o = common.ExtractBinary(b[offset:])
	if o < 0 {
		return nil, FailedToUnmarshalKeypairError.Newf("stellar key; failed to parse Kind")
	}

	var kind Kind
	if err := kind.UnmarshalBinary(e); err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	}

	switch kind {
	case PublicKeyKind:
		var pk StellarPublicKey
		if err := pk.UnmarshalBinary(b); err != nil {
			return nil, err
		}

		return pk, nil
	case PrivateKeyKind:
		var pr StellarPrivateKey
		if err := pr.UnmarshalBinary(b); err != nil {
			return nil, err
		}

		return pr, nil
	default:
		return nil, FailedToUnmarshalKeypairError.Newf("stellar key; unknown Kind")
	}
}

func (s Stellar) NewFromText(b []byte) (Key, error) {
	n := bytes.SplitN(b, []byte(":"), 3)
	if len(n) < 3 {
		return nil, FailedToUnmarshalKeypairError.Newf("stellar key; wrong format; length=%d", len(n))
	}

	var kt Type
	if err := kt.UnmarshalText(n[0]); err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	} else if string(n[0]) != StellarType.Name() {
		return nil, FailedToUnmarshalKeypairError.Newf(
			"stellar key; not stellar keypair; type=%q",
			kt.String(),
		)
	}

	var kind Kind
	if err := kind.UnmarshalText(n[1]); err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	}

	switch kind {
	case PublicKeyKind:
		var pk StellarPublicKey
		if err := pk.UnmarshalText(b); err != nil {
			return nil, err
		}

		return pk, nil
	case PrivateKeyKind:
		var pr StellarPrivateKey
		if err := pr.UnmarshalText(b); err != nil {
			return nil, err
		}

		return pr, nil
	default:
		return nil, FailedToUnmarshalKeypairError.Newf("stellar key; unknown Kind")
	}
}

type StellarPublicKey struct {
	kp stellarKeypair.KP
}

func (s StellarPublicKey) Type() Type {
	return StellarType
}

func (s StellarPublicKey) Kind() Kind {
	return PublicKeyKind
}

func (s StellarPublicKey) Verify(input []byte, sig Signature) error {
	if err := s.kp.Verify(input, sig); err != nil {
		return SignatureVerificationFailedError.New(err)
	}

	return nil
}

func (s StellarPublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": s.Type(),
		"kind": PublicKeyKind,
		"key":  s.kp.Address(),
	})
}

func (s StellarPublicKey) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s StellarPublicKey) MarshalBinary() ([]byte, error) {
	kt, err := StellarType.MarshalBinary()
	if err != nil {
		return nil, err
	}

	b := common.AppendBinary(kt)

	pt, err := PublicKeyKind.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b = append(b, common.AppendBinary(pt)...)
	b = append(b, common.AppendBinary([]byte(s.kp.Address()))...)

	return b, nil
}

func (s *StellarPublicKey) UnmarshalBinary(b []byte) error {
	e, o := common.ExtractBinary(b)
	if o < 0 {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; failed to parse Type")
	}

	var kt Type
	if err := kt.UnmarshalBinary(e); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	} else if !kt.Equal(StellarType) {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; is not stellar keypair")
	}

	offset := o

	e, o = common.ExtractBinary(b[offset:])
	if o < 0 {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; failed to parse Kind")
	}

	offset += o

	var kind Kind
	if err := kind.UnmarshalBinary(e); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	}

	if kind != PublicKeyKind {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; is not public key")
	}

	e, o = common.ExtractBinary(b[offset:])
	if o < 0 {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; failed to parse Kind")
	}

	kp, err := stellarKeypair.Parse(string(e))
	if err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	}

	s.kp = kp

	return nil
}

func (s StellarPublicKey) NewFromBinary(b []byte) (Key, error) {
	var pk StellarPublicKey
	if err := pk.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return pk, nil
}

func (s StellarPublicKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s:%s:%s",
		s.Type().Name(),
		s.Kind().String(),
		s.kp.Address(),
	)), nil
}

func (s *StellarPublicKey) UnmarshalText(b []byte) error {
	n := bytes.SplitN(b, []byte(":"), 3)
	if len(n) < 3 {
		return FailedToUnmarshalKeypairError.Newf("stellar key; wrong format; length=%d", len(n))
	}

	var kt Type
	if err := kt.UnmarshalText(n[0]); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	}

	var kind Kind
	if err := kind.UnmarshalText(n[1]); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	} else if kind != PublicKeyKind {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; is not public key")
	}

	kp, err := stellarKeypair.Parse(string(bytes.Join(n[2:], []byte(":"))))
	if err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	}

	s.kp = kp

	return nil
}

func (s StellarPublicKey) NewFromText(b []byte) (Key, error) {
	var pk StellarPublicKey
	if err := pk.UnmarshalText(b); err != nil {
		return nil, err
	}

	return pk, nil
}

func (s StellarPublicKey) Equal(k Key) bool {
	if !s.Type().Equal(k.Type()) {
		return false
	}

	if s.Kind() != k.Kind() {
		return false
	}

	ks, ok := k.(StellarPublicKey)
	if !ok {
		return false
	}

	return s.kp.Address() == ks.kp.Address()

}

type StellarPrivateKey struct {
	kp *stellarKeypair.Full
}

func (s StellarPrivateKey) Type() Type {
	return StellarType
}

func (s StellarPrivateKey) Kind() Kind {
	return PrivateKeyKind
}

func (s StellarPrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": s.Type(),
		"kind": PrivateKeyKind,
		"key":  s.kp.Seed(),
	})
}

func (s StellarPrivateKey) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s StellarPrivateKey) Sign(b []byte) (Signature, error) {
	sig, err := s.kp.Sign(b)
	if err != nil {
		return nil, err
	}

	return Signature(sig), nil
}

func (s StellarPrivateKey) MarshalBinary() ([]byte, error) {
	kt, err := StellarType.MarshalBinary()
	if err != nil {
		return nil, err
	}

	b := common.AppendBinary(kt)

	pt, err := PrivateKeyKind.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b = append(b, common.AppendBinary(pt)...)
	b = append(b, common.AppendBinary([]byte(s.kp.Seed()))...)

	return b, nil
}

func (s *StellarPrivateKey) UnmarshalBinary(b []byte) error {
	e, o := common.ExtractBinary(b)
	if o < 0 {
		return FailedToUnmarshalKeypairError.Newf("stellar private key; failed to parse Type")
	}

	var kt Type
	if err := kt.UnmarshalBinary(e); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	} else if !kt.Equal(StellarType) {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; is not stellar keypair")
	}

	offset := o

	e, o = common.ExtractBinary(b[offset:])
	if o < 0 {
		return FailedToUnmarshalKeypairError.Newf("stellar private key key; failed to parse Kind")
	}

	offset += o

	var kind Kind
	if err := kind.UnmarshalBinary(e); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	}

	if kind != PrivateKeyKind {
		return FailedToUnmarshalKeypairError.Newf("stellar private key; is not private key")
	}

	e, o = common.ExtractBinary(b[offset:])
	if o < 0 {
		return FailedToUnmarshalKeypairError.Newf("stellar private key; failed to parse Kind")
	}

	if kp, err := stellarKeypair.Parse(string(e)); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	} else if full, ok := kp.(*stellarKeypair.Full); !ok {
		return FailedToUnmarshalKeypairError.Newf("stellar private key; is not *keypair.Full")
	} else {
		s.kp = full
	}

	return nil
}

func (s StellarPrivateKey) NewFromBinary(b []byte) (Key, error) {
	var pk StellarPrivateKey
	if err := pk.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return pk, nil
}

func (s StellarPrivateKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s:%s:%s",
		s.Type().Name(),
		s.Kind().String(),
		s.kp.Seed(),
	)), nil
}

func (s *StellarPrivateKey) UnmarshalText(b []byte) error {
	n := bytes.SplitN(b, []byte(":"), 3)
	if len(n) < 3 {
		return FailedToUnmarshalKeypairError.Newf("stellar key; wrong format; length=%d", len(n))
	}

	var kt Type
	if err := kt.UnmarshalText(n[0]); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	}

	var kind Kind
	if err := kind.UnmarshalText(n[1]); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	} else if kind != PrivateKeyKind {
		return FailedToUnmarshalKeypairError.Newf("stellar public key; is not private key")
	}

	if kp, err := stellarKeypair.Parse(string(bytes.Join(n[2:], []byte(":")))); err != nil {
		return FailedToUnmarshalKeypairError.New(err)
	} else if full, ok := kp.(*stellarKeypair.Full); !ok {
		return FailedToUnmarshalKeypairError.Newf("stellar private key; is not *keypair.Full")
	} else {
		s.kp = full
	}

	return nil
}

func (s StellarPrivateKey) NewFromText(b []byte) (Key, error) {
	var pr StellarPrivateKey
	if err := pr.UnmarshalText(b); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s StellarPrivateKey) Equal(k Key) bool {
	if !s.Type().Equal(k.Type()) {
		return false
	}

	if s.Kind() != k.Kind() {
		return false
	}

	ks, ok := k.(StellarPrivateKey)
	if !ok {
		return false
	}

	return s.kp.Seed() == ks.kp.Seed()
}

func (s StellarPrivateKey) PublicKey() PublicKey {
	return StellarPublicKey{kp: s.kp}
}
