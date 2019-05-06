package common

import (
	"encoding"
	"encoding/json"
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type SealV1 interface {
	encoding.BinaryMarshaler
	RLPSerializer
	JSONMapSerializer
	WellformChecker

	Version() Version
	Type() SealType
	Hint() string
	Hash() Hash
	Source() Address
	Signature() Signature
	SignedAt() Time // signed time
	GenerateHash() (Hash, error)
	Raw() RawSeal
	String() string
}

type RawSeal struct {
	sync.RWMutex
	parent    SealV1
	version   Version
	sealType  SealType
	hash      Hash
	source    Address
	signature Signature
	signedAt  Time
}

func NewRawSeal(
	parent SealV1,
	version Version,
	sealType SealType,
) RawSeal {
	return RawSeal{
		parent:   parent,
		version:  version,
		sealType: sealType,
	}
}

func (r RawSeal) SerializeRLP() ([]interface{}, error) {
	if r.parent == nil {
		return nil, errors.New("parent is missing")
	}

	version, err := r.version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	hash, err := r.hash.MarshalBinary()
	if err != nil {
		hash = []byte{} // skipped
	}

	signedAt, err := r.signedAt.MarshalBinary()
	if err != nil {
		return nil, err
	}

	l := []interface{}{
		hash,
		r.signature,
		version,
		r.sealType,
		r.source,
		signedAt,
	}

	p, err := r.parent.SerializeRLP()
	if err != nil {
		return nil, err
	}

	return append(l, p...), nil
}

func (r *RawSeal) UnserializeRLP(m []rlp.RawValue) error {
	if r.parent == nil {
		return errors.New("parent is missing")
	}

	var parent RLPUnserializer
	if r.parent == nil {
		return errors.New("parent is missing")
	} else if u, ok := r.parent.(RLPUnserializer); !ok {
		return errors.New("parent is not RLPUnserializer")
	} else {
		parent = u
	}

	if err := parent.UnserializeRLP(m); err != nil {
		return err
	}

	if len(m) < 6 {
		return SealNotWellformedError.SetMessage("invalid rlp value count: %d", len(m))
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

	return nil
}

func (r RawSeal) MarshalBinary() ([]byte, error) {
	if r.parent == nil {
		return nil, errors.New("parent is missing")
	}

	s, err := r.SerializeRLP()
	if err != nil {
		return nil, err
	}

	return Encode(s)
}

func (r *RawSeal) UnmarshalBinaryRaw(b []byte) error {
	return nil
}

func (r *RawSeal) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := Decode(b, &m); err != nil {
		return err
	}

	if err := r.UnserializeRLP(m); err != nil {
		return err
	}

	if r.parent == nil {
		return errors.New("parent is missing")
	}

	return nil
}

func (r RawSeal) GenerateHash() (Hash, error) {
	if r.parent == nil {
		return Hash{}, errors.New("parent is missing")
	}

	s, err := r.SerializeRLP()
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

func (r RawSeal) SerializeMap() (map[string]interface{}, error) {
	var parent JSONMapSerializer
	if r.parent == nil {
		return nil, errors.New("parent is missing")
	} else if u, ok := r.parent.(JSONMapSerializer); !ok {
		return nil, errors.New("parent is not JSONMapSerializer")
	} else {
		parent = u
	}

	l := map[string]interface{}{
		"version":   r.version,
		"type":      r.sealType,
		"hash":      r.hash,
		"source":    r.source,
		"signature": r.signature,
		"signedAt":  r.signedAt,
	}

	m, err := parent.SerializeMap()
	if err != nil {
		return nil, err
	}
	for k, v := range m {
		l[k] = v
	}

	return l, nil
}

func (r RawSeal) MarshalJSON() ([]byte, error) {
	if r.parent == nil {
		return nil, errors.New("parent is missing")
	}

	m, err := r.SerializeMap()
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func (r RawSeal) String() string {
	return TerminalLogString(PrintJSON(r, true, false))
}

func (r RawSeal) Raw() RawSeal {
	return r
}

func (r *RawSeal) SetParent(seal SealV1) {
	r.Lock()
	defer r.Unlock()

	r.parent = seal
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

func (r RawSeal) WellformedRaw() error {
	if r.parent == nil {
		return errors.New("parent is missing")
	}

	if err := r.wellformed(); err != nil {
		return SealNotWellformedError.SetMessage(err.Error())
	}

	if hash, err := r.GenerateHash(); err != nil {
		return err
	} else if !r.hash.Equal(hash) {
		return HashDoesNotMatchError
	}

	return nil
}

func (r RawSeal) wellformed() error {
	if r.hash.Empty() {
		return errors.New("empty hash")
	}

	if err := r.signature.IsValid(); err != nil {
		return err
	}

	if r.version.Equal(ZeroVersion) {
		return errors.New("zero version found")
	}

	if len(r.sealType) < 1 {
		return errors.New("empty SealType")
	}

	if _, err := r.source.IsValid(); err != nil {
		return err
	}

	if r.signedAt.IsZero() {
		return errors.New("zero signedAt")
	}

	return nil
}