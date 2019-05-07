package isaac

import (
	"encoding/json"
	"time"

	"github.com/spikeekips/mitum/common"
)

type ConsensusPolicy struct {
	NetworkID                  common.NetworkID `json:"network_id"`
	Total                      uint             `json:"total"`     // total number of validators
	Threshold                  uint             `json:"threshold"` // consensus threshold
	BaseFee                    common.Big       `json:"base_fee"`  // minimum fee for operation
	MaxTransactionsInProposal  uint             `json:"max_transactions_in_proposal"`
	MaxOperationsInTransaction uint             `json:"max_operations_in_transaction"`
	TimeoutWaitSeal            time.Duration    `json:"timeout_wait_seal"`
}

func (c ConsensusPolicy) String() string {
	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}
