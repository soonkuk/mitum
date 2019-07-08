// +build test

package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/node"
)

type SuffrageTest struct {
	nodes          []node.Node
	actingSelector func(Height, Round, []node.Node) []node.Node
}

func NewSuffrageTest(
	nodes []node.Node,
	actingSelector func(Height, Round, []node.Node) []node.Node,
) *SuffrageTest {
	return &SuffrageTest{nodes: nodes, actingSelector: actingSelector}
}

func (st *SuffrageTest) Nodes() []node.Node {
	return st.nodes
}

func (st *SuffrageTest) ActingSuffrage(height Height, round Round) ActingSuffrage {
	nodes := st.actingSelector(height, round, st.nodes)

	return NewActingSuffrage(height, round, nodes[0], nodes)
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
