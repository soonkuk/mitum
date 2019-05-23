package lib

import (
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
)

type ManipulateBeforeFuncType func(*WrongSealBroadcaster, common.Seal) (common.Seal, bool /* send or not */, error)
type ManipulateAfterFuncType func(*WrongSealBroadcaster, common.Seal) error

type WrongSealBroadcaster struct {
	sync.RWMutex
	*common.Logger
	policy               isaac.ConsensusPolicy
	state                *isaac.ConsensusState
	sender               network.SenderFunc
	manipulateBeforeFunc ManipulateBeforeFuncType
	manipulateAfterFunc  ManipulateAfterFuncType
}

func NewWrongSealBroadcaster(
	policy isaac.ConsensusPolicy,
	state *isaac.ConsensusState,
) (*WrongSealBroadcaster, error) {
	return &WrongSealBroadcaster{
		Logger: common.NewLogger(log, "module", "wrong-broadcaster", "node", state.Home().Name()),
		policy: policy,
		state:  state,
	}, nil
}

func (i *WrongSealBroadcaster) Send(
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

	if i.manipulateBeforeFunc != nil {
		manipulated, keepGoing, err := i.manipulateBeforeFunc(i, message.(common.Seal))
		if err != nil {
			i.Log().Error("failed to before-manipulate seal", "error", err)
			// NOTE just keep going
		} else if !keepGoing {
			i.Log().Warn("seal manipulated, but does not keep going")
			return nil
		} else {
			seal = manipulated
		}
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

	if i.manipulateAfterFunc != nil {
		if err := i.manipulateAfterFunc(i, seal); err != nil {
			i.Log().Error("failed to after-manipulate seal", "error", err)
		}
	}

	return nil
}

func (i *WrongSealBroadcaster) SetSender(sender network.SenderFunc) error {
	i.Lock()
	defer i.Unlock()

	i.sender = sender

	return nil
}

func (i *WrongSealBroadcaster) SetManipulateFuncs(before ManipulateBeforeFuncType, after ManipulateAfterFuncType) {
	i.Lock()
	defer i.Unlock()

	i.manipulateBeforeFunc = before
	i.manipulateAfterFunc = after
}
