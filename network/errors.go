package network

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	ReceiverAlreadyRegisteredErrorCode
	ReceiverNotRegisteredErrorCode
	NoReceiversErrorCode
)

var (
	ReceiverAlreadyRegisteredError common.Error = common.NewError("network", ReceiverAlreadyRegisteredErrorCode, "receiver already registered")
	ReceiverNotRegisteredError     common.Error = common.NewError("network", ReceiverNotRegisteredErrorCode, "receiver not registered")
	NoReceiversError               common.Error = common.NewError("network", NoReceiversErrorCode, "no receivers in network")
)
