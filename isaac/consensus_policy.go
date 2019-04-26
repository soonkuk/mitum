package isaac

import (
	"github.com/spikeekips/mitum/common"
)

type ConsensusPolicy struct {
	NetworkID                  common.NetworkID
	Total                      uint // total number of validators
	Threshold                  uint // consensus threshold
	BaseFee                    common.Big
	MaxTransactionsInBallot    uint
	MaxOperationsInTransaction uint
}
