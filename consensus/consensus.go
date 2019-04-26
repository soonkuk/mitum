package consensus

import "github.com/spikeekips/mitum/common"

type Consensus interface {
	Name() string
	Start() error
	Stop() error
	Receiver() <-chan common.Seal
	SetSender(func(common.HomeNode, common.Seal) error) error
}
