package modules

import (
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/node"
)

func NewRandomSuffrage(nodes []node.Node) *FuncSuffrage {
	return &FuncSuffrage{
		nodes: nodes,
		actingSelector: func(height isaac.Height, round isaac.Round, nodes []node.Node) []node.Node {
			r := height.Big.Add(int(round))
			s := int(r.Rem(len(nodes)).Uint64())

			ns := []node.Node{nodes[s]}
			for i, n := range nodes {
				if i == s {
					continue
				}

				ns = append(ns, n)
			}

			return ns
		},
	}
}
