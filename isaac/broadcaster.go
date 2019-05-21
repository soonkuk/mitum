package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/network"
)

type SealBroadcaster interface {
	Send(common.Signer /* seal */, ...common.Address /* excludes */) error
}

type DefaultSealBroadcaster struct {
	sync.RWMutex
	*common.Logger
	policy ConsensusPolicy
	state  *ConsensusState
	sender network.SenderFunc
}

func NewDefaultSealBroadcaster(
	policy ConsensusPolicy,
	state *ConsensusState,
) (*DefaultSealBroadcaster, error) {
	return &DefaultSealBroadcaster{
		Logger: common.NewLogger(log, "module", "broadcaster", "node", state.Home().Name()),
		policy: policy,
		state:  state,
	}, nil
}

func (i *DefaultSealBroadcaster) Send(
	message common.Signer,
	excludes ...common.Address,
) error {
	var seal common.Seal
	if s, ok := message.(common.Seal); !ok {
		return common.InvalidSealTypeError
	} else {
		seal = s
	}

	if err := message.Sign(i.policy.NetworkID, i.state.Home().Seed()); err != nil {
		return err
	}

	log_ := i.Log().New(log15.Ctx{"seal": seal.Hash()})
	log_.Debug(
		"seal will be broadcasted",
		"excludes", excludes,
		"validators", i.state.Validators(),
	)

	var targets = []common.Node{i.state.Home()}
	for _, node := range i.state.Validators() {
		var exclude bool
		for _, a := range excludes {
			if a == node.Address() {
				exclude = true
				break
			}
		}
		if exclude {
			continue
		}
		targets = append(targets, node)
	}

	for _, node := range targets {
		if err := i.sender(node, seal); err != nil {
			log_.Error("failed to broadcast seal", "target-node", node.Name(), "seal-original", seal, "error", err)
			continue
		}
		log_.Debug("seal broadcasted", "target-node", node.Name(), "seal-original", seal)
	}

	return nil
}

func (i *DefaultSealBroadcaster) SetSender(sender network.SenderFunc) error {
	i.Lock()
	defer i.Unlock()

	i.sender = sender

	return nil
}
