package storage

import (
	"github.com/spikeekips/mitum/common"
)

func init() {
	common.SetTestLogger(Log())
}
