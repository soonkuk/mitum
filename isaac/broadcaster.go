package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/network"
)

type SealBroadcaster interface {
	Send(common.SealType, common.Hasher /* body */, ...common.Address /* excludes */) error
}

type DefaultSealBroadcaster struct {
	sync.RWMutex
	policy   ConsensusPolicy
	homeNode common.HomeNode
	sender   network.SenderFunc
}

func NewDefaultSealBroadcaster(
	policy ConsensusPolicy,
	homeNode common.HomeNode,
) (*DefaultSealBroadcaster, error) {
	return &DefaultSealBroadcaster{
		policy:   policy,
		homeNode: homeNode,
	}, nil
}

func (i *DefaultSealBroadcaster) Send(sealType common.SealType, body common.Hasher, excludes ...common.Address) error {
	seal, err := common.NewSeal(sealType, body)
	if err != nil {
		return err
	}

	if err := seal.Sign(i.policy.NetworkID, i.homeNode.Seed()); err != nil {
		return err
	}

	sHash, _, err := seal.Hash()
	if err != nil {
		return err
	}

	log_ := log.New(log15.Ctx{"seal": sHash})
	log_.Debug("seal will be broadcasted")

	var targets = []common.Node{i.homeNode}
	for _, node := range i.homeNode.Validators() {
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
