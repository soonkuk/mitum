package isaac

import (
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type StateHandler interface {
	common.Daemon
	State() node.State
	Write(interface{}) bool
}
