package lib

import (
	"reflect"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
)

func StopProposal(b *WrongSealBroadcaster, seal common.Seal) (common.Seal, bool, error) {
	if seal.Type() != isaac.ProposalSealType {
		return seal, true, nil
	}

	return seal, false, nil
}

func HighHeightACCEPTBallot(b *WrongSealBroadcaster, seal common.Seal) (common.Seal, bool, error) {
	if reflect.ValueOf(seal).Kind() == reflect.Ptr {
		seal = reflect.Indirect(reflect.ValueOf(seal)).Interface().(common.Seal)
	}

	if seal.Type() != isaac.ACCEPTBallotSealType {
		return seal, true, nil
	}

	var ballot isaac.ACCEPTBallot
	if err := common.CheckSeal(seal, &ballot); err != nil {
		panic(err)
		return seal, false, err
	}

	newBallot := isaac.NewACCEPTBallot(
		ballot.Height().Inc(),
		ballot.Round(),
		ballot.Proposer(),
		ballot.Validators(),
		ballot.Proposal(),
		ballot.Block(),
	)

	if err := newBallot.Sign(b.policy.NetworkID, b.state.Home().Seed()); err != nil {
		panic(err)
		return seal, false, err
	}

	b.Log().Info("wrong accept ballot generated", "ballot", newBallot)

	return newBallot, true, nil
}
