package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/common"
)

type ConsensusPolicy struct {
	NetworkID                  common.NetworkID `json:"network_id"`
	Total                      uint             `json:"total"`     // total number of validators
	Threshold                  uint             `json:"threshold"` // consensus threshold
	BaseFee                    common.Big       `json:"base_fee"`  // minimum fee for operation
	MaxTransactionsInPropose   uint             `json:"max_transactions_in_propose"`
	MaxOperationsInTransaction uint             `json:"max_operations_in_transaction"`
}

func (c ConsensusPolicy) String() string {
	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}
