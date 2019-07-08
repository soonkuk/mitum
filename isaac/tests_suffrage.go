// +build test

package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/node"
)

type SuffrageTest struct {
	nodes          []node.Node
	activeSelector func(Height, Round, []node.Node) []node.Node
}

func NewSuffrageTest(
	nodes []node.Node,
	activeSelector func(Height, Round, []node.Node) []node.Node,
) *SuffrageTest {
	return &SuffrageTest{nodes: nodes, activeSelector: activeSelector}
}

func (st *SuffrageTest) Nodes() []node.Node {
	return st.nodes
}

func (st *SuffrageTest) ActiveSuffrage(height Height, round Round) ActiveSuffrage {
	nodes := st.activeSelector(height, round, st.nodes)

	return NewActiveSuffrage(height, round, nodes[0], nodes)
}

func (st SuffrageTest) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"nodes": st.nodes,
	})
}

func (st SuffrageTest) String() string {
	b, _ := json.Marshal(st)
	return string(b)
}
