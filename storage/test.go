// +build test

package storage

type TBatch struct{}

func (t *TBatch) Len() int {
	return 0
}

func (t *TBatch) Put(key []byte, value []byte) {}
func (t *TBatch) Delete(key []byte)            {}
