package isaac

import (
	"github.com/spikeekips/mitum/big"
)

type Height struct {
	big.Big
}

func NewBlockHeight(height uint64) Height {
	return Height{Big: big.NewBigFromUint64(height)}
}

func (ht Height) Equal(height Height) bool {
	return ht.Big.Equal(height.Big)
}

func (ht Height) Add(v interface{}) Height {
	return Height{Big: ht.Big.Add(v)}
}

func (ht Height) Sub(v interface{}) Height {
	return Height{Big: ht.Big.Sub(v)}
}

func (ht Height) Cmp(height Height) int {
	return ht.Big.Cmp(height.Big)
}
