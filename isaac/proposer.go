package isaac

import (
	"crypto/sha256"
	"errors"
	"sort"
	"strconv"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type ProposerSelector interface {
	Select(common.Hash /* block hash */, common.Big /* height */, Round /* round */) (common.Node, error)
}

type DefaultProposerSelector struct {
	candidates map[common.Address]common.Node
	addresses  []common.Address
}

func NewDefaultProposerSelector(candidates []common.Node) (*DefaultProposerSelector, error) {
	if len(candidates) < 1 {
		return nil, FailedToElectProposerError.SetMessage("no candidates TT")
	}

	m := map[common.Address]common.Node{}
	var addresses []common.Address
	for _, c := range candidates {
		addresses = append(addresses, c.Address())
		m[c.Address()] = c
	}

	sort.Sort(common.SortAddress(addresses))

	return &DefaultProposerSelector{
		candidates: m,
		addresses:  addresses,
	}, nil
}

// Select selects the next proposer.
// TODO to make guessing next proposer to be more difficult, we need another
// variables with current block.
func (d *DefaultProposerSelector) Select(block common.Hash, height common.Big, round Round) (common.Node, error) {
	if len(d.addresses) == 1 {
		return d.candidates[d.addresses[0]], nil
	}

	var b []byte = block.Bytes()
	for _, a := range d.addresses {
		b = append(b, []byte(a)...)
	}

	b = append(b, height.Bytes()...)
	b = append(b, []byte(strconv.Itoa(int(round)))...)

	var c [32]byte
	{
		c = sha256.Sum256(b)
		c = sha256.Sum256(c[:])
	}

	var s uint64
	for _, i := range c {
		s += uint64(i)
	}

	return d.candidates[d.addresses[int(s)%len(d.addresses)]], nil
}

type FixedProposerSelector struct {
	sync.RWMutex
	proposer common.Node
}

func NewFixedProposerSelector() *FixedProposerSelector {
	return &FixedProposerSelector{}
}

func (t *FixedProposerSelector) SetProposer(proposer common.Node) {
	t.Lock()
	defer t.Unlock()

	t.proposer = proposer
}

func (t *FixedProposerSelector) Select(block common.Hash, height common.Big, round Round) (common.Node, error) {
	t.RLock()
	defer t.RUnlock()

	if t.proposer == nil {
		return nil, errors.New("empty proposer; `SetProposer()` first")
	}

	return t.proposer, nil
}

type FunctionalProposerSelectorSelectFunc func(block common.Hash, height common.Big, round Round) (common.Node, error)

type FunctionalProposerSelector struct {
	sync.RWMutex
	f FunctionalProposerSelectorSelectFunc
}

func NewFunctionalProposerSelector() *FunctionalProposerSelector {
	return &FunctionalProposerSelector{}
}

func (t *FunctionalProposerSelector) SetSelectFunc(f FunctionalProposerSelectorSelectFunc) {
	t.Lock()
	defer t.Unlock()

	t.f = f
}

func (t *FunctionalProposerSelector) Select(block common.Hash, height common.Big, round Round) (common.Node, error) {
	t.RLock()
	defer t.RUnlock()

	return t.f(block, height, round)
}
