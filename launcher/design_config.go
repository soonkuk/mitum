package launcher

import (
	"time"

	"github.com/spikeekips/mitum/isaac"
)

type NodeConfigDesign struct {
	TimeoutWaitingProposal           time.Duration `yaml:"timeout-waiting-proposal"`
	IntervalBroadcastingINITBallot   time.Duration `yaml:"interval-broadcasting-init-ballot"`
	IntervalBroadcastingProposal     time.Duration `yaml:"interval-broadcasting-proposal"`
	WaitBroadcastingACCEPTBallot     time.Duration `yaml:"wait-broadcasting-accept-ballot"`
	IntervalBroadcastingACCEPTBallot time.Duration `yaml:"interval-broadcasting-accept-ballot"`
	TimespanValidBallot              time.Duration `yaml:"timespan-valid-ballot"`
	TimeoutProcessProposal           time.Duration `yaml:"timeout-process-proposal"`
}

func NewNodeConfigDesign(o *NodeConfigDesign) *NodeConfigDesign {
	if o != nil {
		return &NodeConfigDesign{
			TimeoutWaitingProposal:           o.TimeoutWaitingProposal,
			IntervalBroadcastingINITBallot:   o.IntervalBroadcastingINITBallot,
			IntervalBroadcastingProposal:     o.IntervalBroadcastingProposal,
			WaitBroadcastingACCEPTBallot:     o.WaitBroadcastingACCEPTBallot,
			IntervalBroadcastingACCEPTBallot: o.IntervalBroadcastingACCEPTBallot,
			TimespanValidBallot:              o.TimespanValidBallot,
			TimeoutProcessProposal:           o.TimeoutProcessProposal,
		}
	}

	return &NodeConfigDesign{
		TimeoutWaitingProposal:           isaac.DefaultPolicyTimeoutWaitingProposal,
		IntervalBroadcastingINITBallot:   isaac.DefaultPolicyIntervalBroadcastingINITBallot,
		IntervalBroadcastingProposal:     isaac.DefaultPolicyIntervalBroadcastingProposal,
		WaitBroadcastingACCEPTBallot:     isaac.DefaultPolicyWaitBroadcastingACCEPTBallot,
		IntervalBroadcastingACCEPTBallot: isaac.DefaultPolicyIntervalBroadcastingACCEPTBallot,
		TimespanValidBallot:              isaac.DefaultPolicyTimespanValidBallot,
		TimeoutProcessProposal:           isaac.DefaultPolicyTimeoutProcessProposal,
	}
}

func (nc *NodeConfigDesign) IsValid([]byte) error {
	if nc.TimeoutWaitingProposal < 1 {
		nc.TimeoutWaitingProposal = isaac.DefaultPolicyTimeoutWaitingProposal
	}
	if nc.IntervalBroadcastingINITBallot < 1 {
		nc.IntervalBroadcastingINITBallot = isaac.DefaultPolicyIntervalBroadcastingINITBallot
	}
	if nc.IntervalBroadcastingProposal < 1 {
		nc.IntervalBroadcastingProposal = isaac.DefaultPolicyIntervalBroadcastingProposal
	}
	if nc.WaitBroadcastingACCEPTBallot < 1 {
		nc.WaitBroadcastingACCEPTBallot = isaac.DefaultPolicyWaitBroadcastingACCEPTBallot
	}
	if nc.IntervalBroadcastingACCEPTBallot < 1 {
		nc.IntervalBroadcastingACCEPTBallot = isaac.DefaultPolicyIntervalBroadcastingACCEPTBallot
	}
	if nc.TimespanValidBallot < 1 {
		nc.TimespanValidBallot = isaac.DefaultPolicyTimespanValidBallot
	}
	if nc.TimeoutProcessProposal < 1 {
		nc.TimeoutProcessProposal = isaac.DefaultPolicyTimeoutProcessProposal
	}

	return nil
}

func (nc *NodeConfigDesign) Merge(o *NodeConfigDesign) error {
	if nc.TimeoutWaitingProposal < 1 {
		nc.TimeoutWaitingProposal = o.TimeoutWaitingProposal
	}
	if nc.IntervalBroadcastingINITBallot < 1 {
		nc.IntervalBroadcastingINITBallot = o.IntervalBroadcastingINITBallot
	}
	if nc.IntervalBroadcastingProposal < 1 {
		nc.IntervalBroadcastingProposal = o.IntervalBroadcastingProposal
	}
	if nc.WaitBroadcastingACCEPTBallot < 1 {
		nc.WaitBroadcastingACCEPTBallot = o.WaitBroadcastingACCEPTBallot
	}
	if nc.IntervalBroadcastingACCEPTBallot < 1 {
		nc.IntervalBroadcastingACCEPTBallot = o.IntervalBroadcastingACCEPTBallot
	}
	if nc.TimespanValidBallot < 1 {
		nc.TimespanValidBallot = o.TimespanValidBallot
	}
	if nc.TimeoutProcessProposal < 1 {
		nc.TimeoutProcessProposal = o.TimeoutProcessProposal
	}

	return nil
}
