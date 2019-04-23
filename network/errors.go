package network

import "github.com/spikeekips/mitum/common"

const (
	_ uint = iota
	ReceiverAlreadyRegisteredCode
	ReceiverNotRegisteredCode
	NoReceiversCode
)

var (
	ReceiverAlreadyRegisteredError common.Error = common.NewError("network", ReceiverAlreadyRegisteredCode, "receiver already registered")
	ReceiverNotRegisteredError     common.Error = common.NewError("network", ReceiverNotRegisteredCode, "receiver not registered")
	NoReceiversError               common.Error = common.NewError("network", NoReceiversCode, "no receivers in network")
)
