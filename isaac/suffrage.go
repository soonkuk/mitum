package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/node"
)

type Suffrage interface {
	Nodes() []node.Node
	ActiveSuffrage(height Height, round Round) ActiveSuffrage
}

type ActiveSuffrage struct {
	height   Height
	round    Round
	proposer node.Node
	nodes    []node.Node
}

func NewActiveSuffrage(height Height, round Round, proposer node.Node, nodes []node.Node) ActiveSuffrage {
	return ActiveSuffrage{height: height, round: round, proposer: proposer, nodes: nodes}
}

func (af ActiveSuffrage) Proposer() node.Node {
	return af.proposer
}

func (af ActiveSuffrage) Nodes() []node.Node {
	return af.nodes
}

func (af ActiveSuffrage) Height() Height {
	return af.height
}

func (af ActiveSuffrage) Round() Round {
	return af.round
}

func (af ActiveSuffrage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"height":   af.height,
		"round":    af.round,
		"proposer": af.proposer,
		"nodes":    af.nodes,
	})
}
