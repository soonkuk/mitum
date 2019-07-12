package modules

import (
	"sync"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
)

type WrongBlockProposalValidator struct {
	sync.RWMutex
	*common.Logger
	heights    []isaac.Height
	duration   time.Duration
	fabricated map[string]isaac.Round
}

func NewWrongBlockProposalValidator(heights []isaac.Height, duration time.Duration) *WrongBlockProposalValidator {
	return &WrongBlockProposalValidator{
		Logger:     common.NewLogger(log, "duration", duration),
		heights:    heights,
		duration:   duration,
		fabricated: map[string]isaac.Round{},
	}
}

func (dp *WrongBlockProposalValidator) isValid(proposal isaac.Proposal) error {
	if err := proposal.IsValid(); err != nil {
		return err
	}

	<-time.After(dp.duration)

	return nil
}

func (dp *WrongBlockProposalValidator) isFabricated(height isaac.Height) bool {
	dp.RLock()
	defer dp.RUnlock()

	_, found := dp.fabricated[height.String()]
	return found
}

func (dp *WrongBlockProposalValidator) NewBlock(proposal isaac.Proposal) (block isaac.Block, err error) {
	dp.Log().Debug("trying to validate proposal", "proposal", proposal, "heights", dp.heights)
	if err = dp.isValid(proposal); err != nil {
		return
	}

	defer func() {
		dp.Log().Debug("proposal validated", "proposal", proposal, "block", block)
	}()

	height := proposal.Height().Add(1)
	block, err = isaac.NewBlock(height, proposal.Round(), proposal.Hash())
	if err != nil {
		dp.Log().Error("failed to create new block from proposal", "error", err)
		return
	}

	if dp.isFabricated(height) {
		return
	}

	dp.Log().Debug("trying to fabricate proposal", "proposal", proposal.Hash(), "height", height, "heights", dp.heights)

	origHash := block.Hash()
	for _, h := range dp.heights {
		if !height.Equal(h) {
			continue
		}
		block = block.SetHash(isaac.NewRandomBlockHash())
		dp.Log().Debug("block hash fabricated", "orig", origHash, "new", block.Hash())
		break
	}

	dp.Lock()
	defer dp.Unlock()

	dp.fabricated[height.String()] = proposal.Round()

	return
}
