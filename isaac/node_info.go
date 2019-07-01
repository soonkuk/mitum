package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
)

type NodeInfo struct {
	address   node.Address
	publicKey keypair.PublicKey
	networkID NetworkID
	startedAt common.Time
	block     hash.Hash
	height    Height
	nodeState node.State
	// TODO basic policy
}

func (ni NodeInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"address":   ni.address,
		"publicKey": ni.publicKey,
		"networkID": ni.networkID,
		"startedAt": ni.startedAt,
		"block":     ni.block,
		"height":    ni.height,
		"nodeState": ni.nodeState,
	})
}

func (ni NodeInfo) String() string {
	b, _ := json.Marshal(ni)
	return string(b)
}
