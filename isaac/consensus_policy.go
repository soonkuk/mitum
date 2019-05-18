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
	AvgBlockRoundInterval      time.Duration    `json:"avg_block_round_interval"`      // average interval for each round
	TimeoutWaitSeal            time.Duration    `json:"timeout_wait_seal"`             // wait time for incoming seal
	ExpireDurationVote         time.Duration    `json:"expire_duration_vote"`          // VotingBoxStageNode.votedAt expires after duration
	SealSignedAtAllowDuration  time.Duration    `json:"seal_signed_at_allowd_uration"` // Seal.SignedAt() should be within duration; too old should be ignored, and too ahead also too
}

func DefaultConsensusPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		BaseFee:                    common.ZeroBig,
		MaxTransactionsInProposal:  100,
		MaxOperationsInTransaction: 100,
		AvgBlockRoundInterval:      time.Second * 3,
		TimeoutWaitSeal:            time.Second * 3,
		ExpireDurationVote:         time.Second * 10,
		SealSignedAtAllowDuration:  time.Second * 30, // 30 seconds
	}
}

func (c ConsensusPolicy) String() string {
	b, _ := json.Marshal(c)
	return common.TerminalLogString(string(b))
}

func (c ConsensusPolicy) JSONLog() {}
