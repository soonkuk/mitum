package common

import (
	"encoding/json"
	"fmt"

	"github.com/stellar/go/keypair"
)

type Address string

func (a Address) IsValid() (keypair.KP, error) {
	return keypair.Parse(string(a))
}

func (a Address) Alias() string {
	return fmt.Sprintf("%s.%s", a[:4], a[len(a)-4:])
}

func (a Address) Verify(input []byte, sig []byte) error {
	kp, err := a.IsValid()
	if err != nil {
		return err
	}

	err = kp.Verify(input, sig)
	switch err {
	case keypair.ErrInvalidSignature:
		err = SignatureVerificationFailedError
	}

	return err
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

func (a *Address) UnmarshalJSON(b []byte) error {
	var n string
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}

	na := Address(n)
	if _, err := na.IsValid(); err != nil {
		return err
	}

	*a = na
	return nil
}

type Seed struct {
	*keypair.Full
}

func RandomSeed() Seed {
	seed, _ := keypair.Random()
	return Seed{Full: seed}
}

func NewSeed(raw []byte) Seed {
	seed, _ := keypair.FromRawSeed([32]byte(RawHash(raw)))
	return Seed{Full: seed}
}

func (s Seed) Address() Address {
	return Address(s.Full.Address())
}
