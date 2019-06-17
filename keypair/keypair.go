package keypair

import (
	"encoding"
)

type Keypair interface {
	Type() Type
	New() (PrivateKey, error)
	NewFromSeed([]byte) (PrivateKey, error)
	Equal(Keypair) bool
	NewFromBinary([]byte) (Key, error)
	NewFromText([]byte) (Key, error)
	String() string
}

type Key interface {
	encoding.BinaryMarshaler
	encoding.TextMarshaler
	Type() Type
	Kind() Kind
	Equal(Key) bool
	NewFromBinary([]byte) (Key, error)
	NewFromText([]byte) (Key, error)
	NativePublicKey() []byte
	String() string
}

type PublicKey interface {
	Key
	Verify([]byte, Signature) error
}

type PrivateKey interface {
	Key
	Sign([]byte) (Signature, error)
	PublicKey() PublicKey
	NativePrivateKey() []byte
}
