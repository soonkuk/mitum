package isaac

import (
	"crypto/sha256"
	"sort"
	"strconv"
	"sync"

	"github.com/spikeekips/mitum/common"
)

var (
	proposerSelectBase common.Big = common.NewBig(100)
)

type ProposerSelector interface {
	Select(height common.Big, round Round) (common.Node, error)
}

type DefaultProposerSelector struct {
	sync.RWMutex
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

func (d *DefaultProposerSelector) Select(height common.Big, round Round) (common.Node, error) {
	if len(d.addresses) == 1 {
		return d.candidates[d.addresses[0]], nil
	}

	var b []byte
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
