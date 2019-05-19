package leveldbstorage

import (
	"github.com/spikeekips/mitum/common"
)

func init() {
	common.SetTestLogger(Log())
}
