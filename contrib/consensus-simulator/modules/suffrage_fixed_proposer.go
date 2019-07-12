package modules

import (
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/node"
)

func NewFixedProposerSuffrage(proposer node.Node, nodes []node.Node) *FuncSuffrage {
	return &FuncSuffrage{
		nodes: nodes,
		actingSelector: func(height isaac.Height, round isaac.Round, nodes []node.Node) []node.Node {
			ns := []node.Node{proposer}
			for _, n := range nodes {
				if n.Equal(proposer) {
					continue
				}
				ns = append(ns, n)
			}

			return ns
		},
	}
}
