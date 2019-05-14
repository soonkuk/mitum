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
	AvgBlockRoundInterval      time.Duration    `json:"avg_block_round_interval"` // average interval for each round
	TimeoutWaitSeal            time.Duration    `json:"timeout_wait_seal"`        // wait time for incoming seal
}

func DefaultConsensusPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		BaseFee:                    common.ZeroBig,
		MaxTransactionsInProposal:  100,
		MaxOperationsInTransaction: 100,
		AvgBlockRoundInterval:      time.Second * 3,
		TimeoutWaitSeal:            time.Second * 3,
	}
}

func (c ConsensusPolicy) String() string {
	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}

func (c ConsensusPolicy) JSONLog() {}
