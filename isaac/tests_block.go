// +build test

package isaac

import "github.com/spikeekips/mitum/hash"

func (b Block) SetHash(newHash hash.Hash) Block {
	b.hash = newHash
	return b
}
