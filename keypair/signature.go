package keypair

import "github.com/btcsuite/btcutil/base58"

type Signature []byte

func (s Signature) String() string {
	return base58.Encode(s)
}
