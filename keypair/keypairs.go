package keypair

import (
	"bytes"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type Keypairs struct {
	sync.RWMutex
	keypairs       map[ /*Type*/ uint]Keypair
	keypairsByName map[ /*MarshalText*/ string]Keypair
	defaultType    Type
}

func NewKeypairs() *Keypairs {
	return &Keypairs{
		keypairs:       map[uint]Keypair{},
		keypairsByName: map[string]Keypair{},
	}
}

func (k *Keypairs) Register(keypair Keypair) error {
	k.Lock()
	defer k.Unlock()

	// check duplication
	for t := range k.keypairs {
		if t == keypair.Type().ID() {
			return KeypairAlreadyRegisteredError.Newf("type=%q", keypair.Type().String())
		}
	}

	for t := range k.keypairsByName {
		if string(t) == keypair.Type().Name() {
			return KeypairAlreadyRegisteredError.Newf("type=%q", keypair.Type().Name())
		}
	}

	k.keypairs[keypair.Type().ID()] = keypair
	k.keypairsByName[keypair.Type().Name()] = keypair

	if k.defaultType.Empty() {
		k.defaultType = keypair.Type()
	}

	return nil
}

func (k *Keypairs) SetDefault(kt Type) error {
	keypair, err := k.Keypair(kt)
	if err != nil {
		return err
	}

	k.Lock()
	defer k.Unlock()

	k.defaultType = keypair.Type()

	return nil
}

func (k *Keypairs) Keypair(kt Type) (Keypair, error) {
	k.RLock()
	defer k.RUnlock()

	var keypair Keypair
	var found bool
	if kt.ID() < 1 {
		keypair, found = k.keypairsByName[kt.Name()]
	} else {
		keypair, found = k.keypairs[kt.ID()]
	}

	if !found {
		return nil, KeypairNotRegisteredError.Newf("type=%q", kt.String())
	}

	return keypair, nil
}

func (k *Keypairs) New() (PrivateKey, error) {
	keypair, err := k.Keypair(k.defaultType)
	if err != nil {
		return nil, err
	}

	return keypair.New()
}

func (k *Keypairs) NewFromSeed(b []byte) (PrivateKey, error) {
	keypair, err := k.Keypair(k.defaultType)
	if err != nil {
		return nil, err
	}

	return keypair.NewFromSeed(b)
}

func (k *Keypairs) NewFromBinary(b []byte) (Key, error) {
	// read keypair type
	et, o := common.ExtractBinary(b)
	if o < 0 {
		return nil, FailedToUnmarshalKeypairError
	}

	var kt Type
	if err := kt.UnmarshalBinary(et); err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	}

	kp, err := k.Keypair(kt)
	if err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	}

	return kp.NewFromBinary(b)
}

func (k *Keypairs) NewFromText(b []byte) (Key, error) {
	n := bytes.SplitN(b, []byte(":"), 3)
	if len(n) < 3 {
		return nil, FailedToUnmarshalKeypairError.Newf("wrong format; length=%d", len(n))
	}

	var kt Type
	if err := kt.UnmarshalText(n[0]); err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	}

	kp, err := k.Keypair(kt)
	if err != nil {
		return nil, FailedToUnmarshalKeypairError.New(err)
	}

	return kp.NewFromText(b)
}
