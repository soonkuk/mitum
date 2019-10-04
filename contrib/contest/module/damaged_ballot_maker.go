package contest_module

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/node"
)

type DamangedBallotMaker struct {
	isaac.DefaultBallotMaker
	*common.Logger
	damaged map[ /* Height + Round + Stage*/ string][]string
}

func NewDamangedBallotMaker(home node.Home) DamangedBallotMaker {
	return DamangedBallotMaker{
		DefaultBallotMaker: isaac.NewDefaultBallotMaker(home),
		Logger: common.NewLogger(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "damanged-ballot_maker")
		}),
		damaged: map[string][]string{},
	}
}

func (db DamangedBallotMaker) key(height, round, stage string) string {
	return fmt.Sprintf("%s-%s-%s", height, round, stage)
}

func (db DamangedBallotMaker) AddPoint(height, round, stage string, kinds ...string) DamangedBallotMaker {
	nk := []string{}

	if len(height) < 1 {
		height = "*"
	}
	if len(round) < 1 {
		round = "*"
	}
	if len(stage) < 1 {
		stage = "*"
	}

	if len(kinds) > 0 {
		seen := map[string]struct{}{}
		for _, k := range kinds {
			if _, found := seen[k]; found {
				continue
			}
			nk = append(nk, k)
			seen[k] = struct{}{}
		}
	}

	k := db.key(height, round, stage)
	db.damaged[k] = nk

	return db
}

func (db DamangedBallotMaker) IsDamaged(height isaac.Height, round isaac.Round, stage isaac.Stage) []string {
	keys := []string{
		db.key(height.String(), round.String(), stage.String()), // exact match
		db.key(height.String(), "*", "*"),                       // height match
		db.key("*", "*", stage.String()),                        // stage match
		db.key("*", "*", "*"),                                   // global match
	}

	for _, k := range keys {
		p := db.isDamaged(k)
		if p != nil {
			return p
		}
	}

	return nil
}

func (db DamangedBallotMaker) isDamaged(key string) []string {
	p, found := db.damaged[key]
	if !found {
		return nil
	}

	return p
}

func (db DamangedBallotMaker) INIT(
	lastBlock hash.Hash,
	lastRound isaac.Round,
	nextHeight isaac.Height,
	nextBlock hash.Hash,
	currentRound isaac.Round,
	currentProposal hash.Hash,
) (isaac.Ballot, error) {
	if p := db.IsDamaged(nextHeight, currentRound, isaac.StageINIT); p != nil {
		db.Log().Debug().
			Uint64("height", nextHeight.Uint64()).
			Uint64("round", currentRound.Uint64()).
			Str("stage", isaac.StageINIT.String()).
			Interface("kinds", p).
			Msg("damaged point")

		if len(p) < 1 {
			nextBlock = NewRandomBlockHash()
		} else {
			for _, k := range p {
				switch k {
				case "lastBlock":
					lastBlock = NewRandomBlockHash()
				case "lastRound":
					lastRound = NewRandomRound()
				case "nextHeight":
					nextHeight = NewRandomHeight()
				case "nextBlock":
					nextBlock = NewRandomBlockHash()
				case "currentRound":
					currentRound = NewRandomRound()
				case "currentProposal":
					currentProposal = NewRandomProposalHash()
				}
			}
		}
	}

	return db.DefaultBallotMaker.INIT(
		lastBlock, lastRound, nextHeight, nextBlock, currentRound, currentProposal,
	)
}

func (db DamangedBallotMaker) SIGN(
	lastBlock hash.Hash,
	lastRound isaac.Round,
	nextHeight isaac.Height,
	nextBlock hash.Hash,
	currentRound isaac.Round,
	currentProposal hash.Hash,
) (isaac.Ballot, error) {
	if p := db.IsDamaged(nextHeight, currentRound, isaac.StageSIGN); p != nil {
		db.Log().Debug().
			Uint64("height", nextHeight.Uint64()).
			Uint64("round", currentRound.Uint64()).
			Str("stage", isaac.StageSIGN.String()).
			Interface("kinds", p).
			Msg("damaged point")

		if len(p) < 1 {
			nextBlock = NewRandomBlockHash()
		} else {
			for _, k := range p {
				switch k {
				case "lastBlock":
					lastBlock = NewRandomBlockHash()
				case "lastRound":
					lastRound = NewRandomRound()
				case "nextHeight":
					nextHeight = NewRandomHeight()
				case "nextBlock":
					nextBlock = NewRandomBlockHash()
				case "currentRound":
					currentRound = NewRandomRound()
				case "currentProposal":
					currentProposal = NewRandomProposalHash()
				}
			}
		}
	}

	return db.DefaultBallotMaker.SIGN(
		lastBlock, lastRound, nextHeight, nextBlock, currentRound, currentProposal,
	)
}

func (db DamangedBallotMaker) ACCEPT(
	lastBlock hash.Hash,
	lastRound isaac.Round,
	nextHeight isaac.Height,
	nextBlock hash.Hash,
	currentRound isaac.Round,
	currentProposal hash.Hash,
) (isaac.Ballot, error) {
	if p := db.IsDamaged(nextHeight, currentRound, isaac.StageACCEPT); p != nil {
		db.Log().Debug().
			Uint64("height", nextHeight.Uint64()).
			Uint64("round", currentRound.Uint64()).
			Str("stage", isaac.StageACCEPT.String()).
			Interface("kinds", p).
			Msg("damaged point")

		if len(p) < 1 {
			nextBlock = NewRandomBlockHash()
		} else {
			for _, k := range p {
				switch k {
				case "lastBlock":
					lastBlock = NewRandomBlockHash()
				case "lastRound":
					lastRound = NewRandomRound()
				case "nextHeight":
					nextHeight = NewRandomHeight()
				case "nextBlock":
					nextBlock = NewRandomBlockHash()
				case "currentRound":
					currentRound = NewRandomRound()
				case "currentProposal":
					currentProposal = NewRandomProposalHash()
				}
			}
		}
	}

	return db.DefaultBallotMaker.ACCEPT(
		lastBlock, lastRound, nextHeight, nextBlock, currentRound, currentProposal,
	)
}

func (db DamangedBallotMaker) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":    "DamangedBallotMaker",
		"damaged": db.damaged,
	})
}
