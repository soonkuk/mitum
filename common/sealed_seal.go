package common

import (
	"encoding/base64"

	"github.com/ethereum/go-ethereum/rlp"
)

type SealedSeal struct {
	RawSeal
	sealType SealType
	binary   []byte
}

// NOTE seal should pass wellformed.
func NewSealedSeal(seal SealV1) (SealedSeal, error) {
	encoded, err := EncodeSeal(seal)
	if err != nil {
		return SealedSeal{}, err
	}

	s := SealedSeal{
		sealType: seal.Type(),
		binary:   encoded,
	}

	raw := NewRawSeal(s, CurrentSealVersion, SealedSealType)
	s.RawSeal = raw

	return s, nil
}

func (r SealedSeal) Binary() []byte {
	return r.binary
}

func (r SealedSeal) Hint() string {
	return "ss"
}

func (r SealedSeal) SerializeRLP() ([]interface{}, error) {
	return []interface{}{r.sealType, r.binary}, nil
}

func (r *SealedSeal) UnserializeRLP(m []rlp.RawValue) error {
	if len(m) < 8 {
		return SealNotWellformedError.SetMessage("invalid marshaled data: %d < 8", len(m))
	}

	var sealType SealType
	if err := Decode(m[6], &sealType); err != nil {
		return err
	}

	var binary []byte
	if err := Decode(m[7], &binary); err != nil {
		return err
	}

	r.sealType = sealType
	r.binary = binary

	return nil
}

func (r SealedSeal) SerializeMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"sealType": r.sealType,
		"binary":   base64.StdEncoding.EncodeToString(r.binary),
	}, nil
}

func (r SealedSeal) Wellformed() error {
	if err := r.RawSeal.WellformedRaw(); err != nil {
		return err
	}

	if len(r.sealType) < 1 {
		return SealNotWellformedError.SetMessage("empty sealType")
	}

	if len(r.binary) < 1 {
		return SealNotWellformedError.SetMessage("empty binary")
	}

	return nil
}
