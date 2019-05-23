package isaac

import (
	"encoding/json"
	"time"

	"github.com/spikeekips/mitum/common"
)

type ConsensusPolicy struct {
	NetworkID                  common.NetworkID `json:"network_id"`
	Total                      uint             `json:"total"`     // NOTE total number of validators
	Threshold                  uint             `json:"threshold"` // NOTE consensus threshold
	BaseFee                    common.Big       `json:"base_fee"`  // NOTE minimum fee for operation
	MaxTransactionsInProposal  uint             `json:"max_transactions_in_proposal"`
	MaxOperationsInTransaction uint             `json:"max_operations_in_transaction"`
	AvgBlockRoundInterval      time.Duration    `json:"avg_block_round_interval"`      // NOTE average interval for each round
	TimeoutWaitSeal            time.Duration    `json:"timeout_wait_seal"`             // NOTE wait time for incoming seal
	ExpireDurationVote         time.Duration    `json:"expire_duration_vote"`          // NOTE VotingBoxStageNode.votedAt expires after duration
	SealSignedAtAllowDuration  time.Duration    `json:"seal_signed_at_allowd_uration"` // NOTE Seal.SignedAt() should be within duration; too old should be ignored, and too ahead also too means nodes can send and receive seal within the duration
}

func DefaultConsensusPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		BaseFee:                    common.ZeroBig,
		MaxTransactionsInProposal:  100,
		MaxOperationsInTransaction: 100,
		AvgBlockRoundInterval:      time.Second * 3,
		TimeoutWaitSeal:            time.Second * 3,
		ExpireDurationVote:         time.Second * 10,
		SealSignedAtAllowDuration:  time.Second * 5,
	}
}

func (c ConsensusPolicy) String() string {
	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}

func (c ConsensusPolicy) JSONLog() {}
