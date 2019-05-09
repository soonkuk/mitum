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
	log    log15.Logger
	policy ConsensusPolicy
	home   *common.HomeNode
	sender network.SenderFunc
}

func NewDefaultSealBroadcaster(
	policy ConsensusPolicy,
	home *common.HomeNode,
) (*DefaultSealBroadcaster, error) {
	return &DefaultSealBroadcaster{
		log:    log.New(log15.Ctx{"node": home.Name()}),
		policy: policy,
		home:   home,
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

	if err := message.Sign(i.policy.NetworkID, i.home.Seed()); err != nil {
		return err
	}

	log_ := i.log.New(log15.Ctx{"seal": seal.Hash()})
	log_.Debug("seal will be broadcasted")

	var targets = []common.Node{i.home}
	for _, node := range i.home.Validators() {
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
			return err
		}
	}

	return nil
}

func (i *DefaultSealBroadcaster) SetSender(sender network.SenderFunc) error {
	i.Lock()
	defer i.Unlock()

	i.sender = sender

	return nil
}
