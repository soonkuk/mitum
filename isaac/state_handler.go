package isaac

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
)

type StateHandler interface {
	common.Daemon
	StartWithContext(context.Context) error
	State() node.State
	Write(interface{}) bool
	SetLogContext(log15.Ctx, ...interface{}) *common.Logger
}
