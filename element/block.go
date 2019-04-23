package element

import (
	"github.com/spikeekips/mitum/common"
)

type Block interface {
	Version() common.Version
	Hash() common.Hash
	Height() common.Big
	PrevHash() common.Hash
	State() []byte
	PrevState() []byte
	Transactions() []common.Hash
}
