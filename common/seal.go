package common

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"reflect"

	"github.com/ethereum/go-ethereum/rlp"
)

var (
	CurrentSealVersion Version = MustParseVersion("0.1.0-proto")
)

type SealType string

func NewSealType(t string) SealType {
	return SealType(t)
}

type Seal struct {
	Version   Version
	Type      SealType
	hash      Hash
	Source    Address
	Signature Signature
	bodyHash  Hash
	Body      []byte
	CreatedAt Time

	encoded []byte
	body    interface{}
}

func NewSeal(t SealType, body Hasher) (Seal, error) {
	bodyHash, encoded, err := body.Hash()
	if err != nil {
		return Seal{}, err
	}

	s := Seal{
		Type:     t,
		Version:  CurrentSealVersion,
		bodyHash: bodyHash,
		Body:     encoded,
	}

	return s, nil
}

func (s Seal) makeHash() (Hash, []byte, error) {
	encoded, err := s.MarshalBinary()
	if err != nil {
		return Hash{}, nil, err
	}

	hash, err := NewHash("sl", encoded)
	if err != nil {
		return Hash{}, nil, err
	}

	return hash, encoded, nil
}

func (s Seal) Hash() (Hash, []byte, error) {
	if s.hash.Empty() {
		return s.makeHash()
	}

	return s.hash, s.encoded, nil
}

func (s Seal) BodyHash() Hash {
	return s.bodyHash
}

func (s *Seal) Sign(networkID NetworkID, seed Seed) error {
	signature, err := NewSignature(networkID, seed, s.bodyHash)
	if err != nil {
		return err
	}

	s.Source = seed.Address()
	s.Signature = signature
	s.CreatedAt = Now()

	hash, encoded, err := s.makeHash()
	if err != nil {
		return err
	}
	s.hash = hash
	s.encoded = encoded

	return nil
}

func (s Seal) CheckSignature(networkID NetworkID) error {
	err := s.Source.Verify(
		append(networkID, s.bodyHash.Bytes()...),
		[]byte(s.Signature),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s Seal) MarshalBinary() ([]byte, error) {
	version, err := s.Version.MarshalBinary()
	if err != nil {
		return nil, err
	}

	bodyHash, err := s.bodyHash.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return Encode([]interface{}{
		version,
		s.Type,
		s.Source,
		s.Signature,
		bodyHash,
		s.Body,
		s.CreatedAt,
	})
}

func (s *Seal) UnmarshalBinary(b []byte) error {
	var m []rlp.RawValue
	if err := Decode(b, &m); err != nil {
		return err
	}

	var version Version
	{
		var vs []byte
		if err := Decode(m[0], &vs); err != nil {
			return err
		} else if err := version.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var sealType SealType
	if err := Decode(m[1], &sealType); err != nil {
		return err
	}

	var source Address
	if err := Decode(m[2], &source); err != nil {
		return err
	}

	var signature Signature
	if err := Decode(m[3], &signature); err != nil {
		return err
	}

	var bodyHash Hash
	{
		var vs []byte
		if err := Decode(m[4], &vs); err != nil {
			return err
		} else if err := bodyHash.UnmarshalBinary(vs); err != nil {
			return err
		}
	}

	var body []byte
	if err := Decode(m[5], &body); err != nil {
		return err
	}

	var createdAt Time
	if err := Decode(m[6], &createdAt); err != nil {
		return err
	}

	s.Version = version
	s.Type = sealType
	s.Signature = signature
	s.Source = source
	s.Body = body
	s.bodyHash = bodyHash
	s.CreatedAt = createdAt

	hash, encoded, err := s.makeHash()
	if err != nil {
		return err
	}
	s.hash = hash
	s.encoded = encoded

	return nil
}

func (s Seal) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"version":    s.Version,
		"type":       s.Type,
		"hash":       s.hash,
		"source":     s.Source,
		"signature":  s.Signature,
		"body_hash":  s.bodyHash,
		"body":       base64.StdEncoding.EncodeToString(s.Body),
		"created_at": s.CreatedAt,
	})
}

func (s *Seal) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var version Version
	if err := json.Unmarshal(raw["version"], &version); err != nil {
		return err
	}

	var source Address
	if err := json.Unmarshal(raw["source"], &source); err != nil {
		return err
	}

	var sealType SealType
	if err := json.Unmarshal(raw["type"], &sealType); err != nil {
		return err
	}

	var signature Signature
	if err := json.Unmarshal(raw["signature"], &signature); err != nil {
		return err
	}

	var hash Hash
	if err := json.Unmarshal(raw["hash"], &hash); err != nil {
		return err
	}

	var bodyHash Hash
	if err := json.Unmarshal(raw["body_hash"], &bodyHash); err != nil {
		return err
	}

	var body []byte
	{
		var c string
		if err := json.Unmarshal(raw["body"], &c); err != nil {
			return err
		} else if d, err := base64.StdEncoding.DecodeString(c); err != nil {
			return err
		} else {
			body = d
		}
	}

	var createdAt Time
	if err := json.Unmarshal(raw["created_at"], &createdAt); err != nil {
		return err
	}

	s.Version = version
	s.Type = sealType
	s.hash = hash
	s.Source = source
	s.Signature = signature
	s.bodyHash = bodyHash
	s.Body = body
	s.CreatedAt = createdAt

	{
		hash, encoded, err := s.makeHash()
		if err != nil {
			return err
		}
		if !s.hash.Equal(hash) {
			return HashDoesNotMatchError
		}

		s.encoded = encoded
	}

	return nil
}

func (s *Seal) UnmarshalBody(i encoding.BinaryUnmarshaler) error {
	if s.body == nil {
		if err := i.UnmarshalBinary(s.Body); err != nil {
			return err
		}

		s.body = reflect.ValueOf(i).Elem().Interface()
	} else {
		reflect.ValueOf(i).Elem().Set(reflect.ValueOf(s.body))
	}

	return nil
}

func (s Seal) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}
