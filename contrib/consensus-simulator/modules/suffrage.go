package modules

import (
	"encoding/json"

	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/node"
)

type FuncSuffrage struct {
	nodes          []node.Node
	actingSelector func(isaac.Height, isaac.Round, []node.Node) []node.Node
}

func NewFuncSuffrage(
	nodes []node.Node,
	actingSelector func(isaac.Height, isaac.Round, []node.Node) []node.Node,
) *FuncSuffrage {
	return &FuncSuffrage{nodes: nodes, actingSelector: actingSelector}
}

func (st *FuncSuffrage) Nodes() []node.Node {
	return st.nodes
}

func (st *FuncSuffrage) ActingSuffrage(height isaac.Height, round isaac.Round) isaac.ActingSuffrage {
	nodes := st.actingSelector(height, round, st.nodes)

	return isaac.NewActingSuffrage(height, round, nodes[0], nodes)
}

func (st FuncSuffrage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"nodes": st.nodes,
	})
}

func (st FuncSuffrage) String() string {
	b, _ := json.Marshal(st)
	return string(b)
}
