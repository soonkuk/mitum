package common

import (
	"encoding"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/rlp"
)

type SealV1 interface {
	encoding.BinaryMarshaler
	RLPSerializer

	Version() Version
	Type() SealType
	Hint() string
	Hash() Hash
	GenerateHash() (Hash, error)
	Source() Address
	Signature() Signature
	SignedAt() Time // signed time
	Wellformed() error
	String() string
}

type RawSeal struct {
	parent    SealV1
	version   Version
	sealType  SealType
	hash      Hash
	source    Address
	signature Signature
	signedAt  Time
}

func (r RawSeal) SerializeRLPInside() ([]interface{}, error) {
	version, err := r.version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	hash, err := r.hash.MarshalBinary()
	if err != nil {
		return nil, err
	}

	signedAt, err := r.signedAt.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return []interface{}{
		hash,
		r.signature,
		version,
		r.sealType,
		r.source,
		signedAt,
	}, nil
}

func (r *RawSeal) UnserializeRLPInside(m []rlp.RawValue) error {
	if len(m) < 6 {
		return SealNotWellformedError
	}

	var hash Hash
	{
		var vs []byte
		if err := Decode(m[0], &vs); err != nil {
			return err
		} else if err := hash.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var signature Signature
	if err := Decode(m[1], &signature); err != nil {
		return err
	}

	var version Version
	{
		var vs []byte
		if err := Decode(m[2], &vs); err != nil {
			return err
		} else if err := version.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var sealType SealType
	if err := Decode(m[3], &sealType); err != nil {
		return err
	}

	var source Address
	if err := Decode(m[4], &source); err != nil {
		return err
	}

	var signedAt Time
	{
		var vs []byte
		if err := Decode(m[5], &vs); err != nil {
			return err
		} else if err := signedAt.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	r.hash = hash
	r.signature = signature
	r.version = version
	r.sealType = sealType
	r.source = source
	r.signedAt = signedAt

	if err := r.Wellformed(); err != nil {
		return err
	}

	return nil
}

func (r RawSeal) MarshalBinary() ([]byte, error) {
	if r.parent == nil {
		return nil, errors.New("parent is missing")
	}

	s, err := r.parent.SerializeRLP()
	if err != nil {
		return nil, err
	}

	return Encode(s)
}

func (r RawSeal) GenerateHash() (Hash, error) {
	if r.parent == nil {
		return Hash{}, errors.New("parent is missing")
	}

	var s []interface{}
	var err error
	if r.parent != nil {
		s, err = r.parent.SerializeRLP()
	} else {
		s, err = r.SerializeRLPInside()
	}

	if err != nil {
		return Hash{}, err
	}

	encoded, err := Encode(s[2:]) // hash and signature will be excluded for hash
	if err != nil {
		return Hash{}, err
	}

	hash, err := NewHash(r.parent.Hint(), encoded)
	if err != nil {
		return Hash{}, err
	}

	return hash, nil
}

func (r *RawSeal) UnmarshalBinaryInside(b []byte) error {
	if r.parent == nil {
		return errors.New("parent is missing")
	}

	u, ok := r.parent.(RLPUnserializer)
	if !ok {
		return errors.New("parent is not RLPUnserializer")
	}

	var m []rlp.RawValue
	if err := Decode(b, &m); err != nil {
		return err
	}

	return u.UnserializeRLP(m)
}

func (r *RawSeal) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"version":   r.version,
		"type":      r.sealType,
		"hash":      r.hash,
		"source":    r.source,
		"signature": r.signature,
		"signedAt":  r.signedAt,
	}, nil
}

func (r RawSeal) MarshalJSON() ([]byte, error) {
	if r.parent == nil {
		return nil, errors.New("parent is missing")
	}

	u, ok := r.parent.(JSONMapSerializer)
	if !ok {
		return nil, errors.New("parent is not JSONMapSerializer")
	}

	m, err := u.SerializeMap()
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func (r RawSeal) String() string {
	return TerminalLogString(PrintJSON(r, true, false))
}

func (r RawSeal) Version() Version {
	return r.version
}

func (r RawSeal) Type() SealType {
	return r.sealType
}

func (r RawSeal) Hash() Hash {
	return r.hash
}

func (r RawSeal) Source() Address {
	return r.source
}

func (r RawSeal) Signature() Signature {
	return r.signature
}

func (r *RawSeal) Sign(networkID NetworkID, seed Seed) error {
	r.source = seed.Address()
	r.signedAt = Now()

	if hash, err := r.GenerateHash(); err != nil {
		return err
	} else {
		r.hash = hash
	}

	signature, err := NewSignature(networkID, seed, r.hash)
	if err != nil {
		return err
	}

	r.signature = signature

	return nil
}

func (r RawSeal) CheckSignature(networkID NetworkID) error {
	return r.source.Verify(
		append(networkID, r.hash.Bytes()...),
		[]byte(r.signature),
	)
}

func (r RawSeal) SignedAt() Time {
	return r.signedAt
}

func (r RawSeal) Wellformed() error {
	if err := r.wellformed(); err != nil {
		return SealNotWellformedError.SetMessage(err.Error())
	}

	return nil
}

func (r RawSeal) wellformed() error {
	if r.version.Equal(ZeroVersion) {
		return errors.New("zero version found")
	}

	if len(r.sealType) < 1 {
		return errors.New("empty SealType")
	}

	if r.hash.Empty() {
		return errors.New("empty hash")
	}

	if _, err := r.source.IsValid(); err != nil {
		return err
	}

	if err := r.signature.IsValid(); err != nil {
		return err
	}

	if r.signedAt.IsZero() {
		return errors.New("zero signedAt")
	}

	return nil
}
