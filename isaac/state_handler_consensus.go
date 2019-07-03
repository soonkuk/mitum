package isaac

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
	"golang.org/x/xerrors"
)

type ConsensusStateHandler struct {
	sync.RWMutex
	*common.ReaderDaemon
	*common.Logger
	homeState     *HomeState
	suffrage      Suffrage
	policy        Policy
	ballotbox     *Ballotbox
	networkClient NetworkClient
	chanState     chan<- node.State
	lastRound     Round
	timer         common.Timer
}

func NewConsensusStateHandler(
	homeState *HomeState,
	suffrage Suffrage,
	policy Policy,
	ballotbox *Ballotbox,
	networkClient NetworkClient,
	chanState chan<- node.State,
) *ConsensusStateHandler {
	cs := &ConsensusStateHandler{
		Logger:        common.NewLogger(Log(), "module", "consensus-state-handler", "state", node.StateConsensus),
		homeState:     homeState,
		suffrage:      suffrage,
		policy:        policy,
		ballotbox:     ballotbox,
		networkClient: networkClient,
		chanState:     chanState,
		lastRound:     Round(0),
	}

	cs.ReaderDaemon = common.NewReaderDaemon(true, cs.receiveSeal)

	return cs
}

func (cs *ConsensusStateHandler) Start() error {
	if err := cs.ReaderDaemon.Start(); err != nil {
		return err
	}

	cs.Lock()
	cs.homeState.SetState(node.StateConsensus)
	cs.Unlock()

	if err := cs.start(); err != nil {
		return err
	}

	return nil
}

func (cs *ConsensusStateHandler) Stop() error {
	cs.Lock()
	defer cs.Unlock()

	if err := cs.ReaderDaemon.Stop(); err != nil {
		return err
	}

	if cs.timer != nil {
		if err := cs.timer.Stop(); err != nil {
			return err
		}
		cs.timer = nil
	}

	return nil
}

func (cs *ConsensusStateHandler) State() node.State {
	return node.StateConsensus
}

func (cs *ConsensusStateHandler) start() error {
	if err := cs.startTimeoutINITBallot(); err != nil {
		return err
	}

	return nil
}

func (cs *ConsensusStateHandler) LastRound() Round {
	cs.RLock()
	defer cs.RUnlock()

	return cs.lastRound
}

func (cs *ConsensusStateHandler) setLastRound(round Round) *ConsensusStateHandler {
	cs.Lock()
	defer cs.Unlock()

	cs.lastRound = round

	return cs
}

func (cs *ConsensusStateHandler) receiveSeal(v interface{}) error {
	// TODO store seal

	sl, ok := v.(seal.Seal)
	if !ok {
		return xerrors.Errorf("not Seal")
	}

	// TODO remove; checking IsValid() should be already done in previous
	// process
	if err := sl.IsValid(); err != nil {
		return err
	}

	cs.Log().Debug("got seal", "seal", sl)

	switch t := sl.Type(); t {
	case BallotType:
		ballot, ok := sl.(Ballot)
		if !ok {
			return xerrors.Errorf("is not Ballot; seal=%q", sl)
		}

		// TODO check ballot

		if err := cs.receiveBallot(ballot); err != nil {
			return err
		}
	case ProposalType:
		proposal, ok := sl.(Proposal)
		if !ok {
			return xerrors.Errorf("is not Proposal; seal=%q", sl)
		}

		if err := cs.receiveProposal(proposal); err != nil {
			return err
		}
	default:
		return xerrors.Errorf("not available seal type in JOIN state; type=%q", t)
	}

	return nil
}

func (cs *ConsensusStateHandler) receiveBallot(ballot Ballot) error {
	// TODO checker ballot

	if ballot.Stage() != StageINIT {
		if !ballot.Height().Equal(cs.homeState.Height()) { // ignore it
			return nil
		}

		if ballot.Round() != cs.LastRound() { // ignore it
			return nil
		}
	}

	vr, err := cs.ballotbox.Vote(ballot)
	if err != nil {
		return err
	}

	switch vr.Result() {
	case NotYetMajority, FinishedGotMajority:
	case JustDraw:
		if err := cs.moveToNextRound(vr, vr.Round()+1); err != nil {
			return err
		}
	case GotMajority:
		if err := cs.gotMajority(vr); err != nil {
			return err
		}
	}

	return nil
}

func (cs *ConsensusStateHandler) receiveProposal(proposal Proposal) error {
	// TODO check,
	// - Proposal is already processed
	// - Proposal.CurrentBlock is same with home
	// 	- if not, ignore it
	// - Proposal.Proposer is valid
	// 	- if not, ignore it
	// - Proposal.Height is same with home
	// 	- if not, ignore it
	// - Proposal.Round is same with cs.lastRound
	// 	- if not, ignore it
	// - Proposal.Proposer is valid proposer at this round
	// 	- if not, ignore it
	// - transactions in Proposal.Transactions is already in block or not
	// 	- if not, ignore it

	if !proposal.Height().Equal(cs.homeState.Height()) { // ignore it
		return nil
	}

	if proposal.Round() != cs.LastRound() { // ignore it
		return nil
	}

	// check proposer is valid
	activeSuffrage := cs.suffrage.ActiveSuffrage(proposal.Height(), proposal.Round())
	if !activeSuffrage.Proposer().Address().Equal(proposal.Proposer()) {
		cs.Log().Debug(
			"proposer is not proposer at this round",
			"proposer", proposal.Proposer(),
			"expected_proposer", activeSuffrage.Proposer().Address(),
			"height", proposal.Height(),
			"round", proposal.Round(),
		)
		return nil
	}

	// TODO everyting is ok, validate it and broadcast SIGNBallot

	// reset INITBallot timer
	if err := cs.startTimeoutINITBallot(); err != nil {
		return err
	}

	// broadcast SIGNBallot for next round
	ballot, err := NewBallot(
		cs.homeState.Home().Address(),
		proposal.Height(),
		proposal.Round(),
		StageSIGN,
		proposal.Hash(),
		proposal.CurrentBlock(),
		NewRandomBlockHash(), // TODO remove
	)
	if err != nil {
		return err
	}

	cs.Log().Debug("broadcast ballot for proposal", "proposal", proposal.Hash())

	if err := cs.networkClient.Vote(&ballot); err != nil {
		return err
	}

	return nil
}

func (cs *ConsensusStateHandler) gotMajority(vr VoteResult) error {
	cs.Log().Debug("got majority", "vr", vr)

	// TODO check:
	// - vr.Height is same with home
	// - vr.Proposal is valid
	// - vr.CurrentBlock is same with home
	// - vr.NextBlock is same with the expected

	sub := vr.Height().Big.Sub(cs.homeState.Height().Big)
	switch {
	case sub.Equal(big.NewBigFromInt64(0)): // same
	case sub.Equal(big.NewBigFromInt64(-1)): // 1 lower
		if vr.Stage() != StageINIT {
			cs.Log().Debug(
				"VoteResult.Height is different from home",
				"VoteResult.height", vr.Height(),
				"home", cs.homeState.Height(),
				"vr", vr,
			)

			cs.chanState <- node.StateSync
			return nil
		}

		if !vr.NextBlock().Equal(cs.homeState.Block().Hash()) {
			cs.Log().Debug(
				"VoteResult.NextBlock is different from home",
				"VoteResult.NextBlock", vr.NextBlock(),
				"home", cs.homeState.Block().Hash(),
				"vr", vr,
			)

			cs.chanState <- node.StateSync
			return nil
		}

		if !vr.CurrentBlock().Equal(cs.homeState.PreviousBlock().Hash()) {
			cs.Log().Debug(
				"VoteResult.Block is different from home",
				"VoteResult.Block", vr.CurrentBlock(),
				"home", cs.homeState.PreviousBlock().Hash(),
				"vr", vr,
			)

			cs.chanState <- node.StateSync
			return nil
		}
		if !vr.Proposal().Equal(cs.homeState.Proposal()) {
			cs.Log().Debug(
				"VoteResult.Proposal is different from home",
				"VoteResult.Proposal", vr.Proposal(),
				"home", cs.homeState.Proposal(),
				"vr", vr,
			)

			cs.chanState <- node.StateSync
			return nil
		}
	default: // lower or higher, go to sync
		cs.Log().Debug(
			"VoteResult.Height is different from home",
			"VoteResult.height", vr.Height(),
			"home", cs.homeState.Height(),
			"vr", vr,
		)

		cs.chanState <- node.StateSync
		return nil
	}

	if vr.Stage() != StageINIT && !vr.CurrentBlock().Equal(cs.homeState.Block().Hash()) {
		cs.Log().Debug(
			"VoteResult.CurrentBlock is different from home",
			"VoteResult.CurrentBlock", vr.CurrentBlock(),
			"home", cs.homeState.Block().Hash(),
		)

		cs.chanState <- node.StateSync
		return nil
	}

	switch vr.Stage() {
	case StageINIT:
		// TODO NextBlock() is same with the expected, checks
		// - proposal is same?
		// 	- if proposal is same, but different nextBlock, go to sync
		// 	- if proposal is not same, validate proposal and check nextBlock

		// TODO everyting is ok, save next block and move to next block

		// TODO store next block
		return cs.moveToNextBlock(vr)
	case StageSIGN:
		// TODO checks,
		// - proposal is same?
		// 	- if proposal is same, but different nextBlock, go to sync
		// 	- if proposal is not same, validate proposal and check nextBlock

		// TODO everyting is ok, go to StageACCEPT
		return cs.moveToNextStage(vr)
	case StageACCEPT:
		// TODO checks,
		// - proposal is same?
		// 	- if proposal is same, but different nextBlock, go to sync
		// 	- if proposal is not same, validate proposal and check nextBlock

		// TODO everyting is ok, go to StageINIT
		return cs.moveToNextStage(vr)
	default:
		return xerrors.Errorf("invalid VoteResult.Stage()", "stage", vr.Stage())
	}
}

func (cs *ConsensusStateHandler) moveToNextBlock(vr VoteResult) error {
	if err := cs.startTimeoutINITBallot(); err != nil {
		return err
	}

	nextHeight := vr.Height().Add(1)
	nextRound := vr.Round()

	log_ := cs.Log().New(log15.Ctx{"next_height": nextHeight, "next_round": nextRound})

	// TODO store next block
	nextBlock, err := NewBlock(nextHeight, vr.NextBlock())
	if err != nil {
		return err
	}

	cs.homeState.SetBlock(nextBlock)

	log_.Debug(
		"new block created",
		"previous_height", vr.Height(),
		"previous_block", vr.CurrentBlock(),
		"previous_round", vr.Round(),
		"next_height", nextHeight,
		"next_block", vr.NextBlock(),
		"next_round", vr.Round(),
	)

	_ = cs.setLastRound(nextRound)

	// TODO wait or propose Proposal with vr.Height() + 1
	activeSuffrage := cs.suffrage.ActiveSuffrage(nextHeight, nextRound)

	log_.Debug("move to next block", "vr", vr, "active_suffrage", activeSuffrage)

	log_.Debug(
		"proposer selected",
		"proposer", activeSuffrage.Proposer().Address(),
		"home is proposer?", activeSuffrage.Proposer().Equal(cs.homeState.Home()),
	)

	if activeSuffrage.Proposer().Equal(cs.homeState.Home()) {
		log_.Debug("home is proposer", "proposer", activeSuffrage.Proposer().Address())

		proposal, err := NewProposal(
			cs.homeState.Height(),
			nextRound,
			cs.homeState.Block().Hash(),
			cs.homeState.Home().Address(),
			nil, // TODO set Transactions
		)
		if err != nil {
			return err
		}

		log_.Debug("broadcast proposal for next block", "proposal", proposal)
		if err := cs.networkClient.Propose(&proposal); err != nil {
			return err
		}
	}

	return nil
}

func (cs *ConsensusStateHandler) moveToNextRound(vr VoteResult, round Round) error {
	cs.Log().Debug("move to next round", "vr", vr)

	if err := cs.startTimeoutINITBallot(); err != nil {
		return err
	}

	// broadcast INITBallot for next round
	ballot, err := NewBallot(
		cs.homeState.Home().Address(),
		vr.Height(),
		round,
		StageINIT,
		vr.Proposal(),
		vr.CurrentBlock(),
		vr.NextBlock(),
	)
	if err != nil {
		return err
	}

	cs.Log().Debug("broadcast ballot for next round", "next_round", round, "ballot", ballot)

	if err := cs.networkClient.Vote(&ballot); err != nil {
		return err
	}

	return nil
}

func (cs *ConsensusStateHandler) moveToNextStage(vr VoteResult) error {
	cs.Log().Debug("move to next stage", "vr", vr)

	if err := cs.startTimeoutINITBallot(); err != nil {
		return err
	}

	nextStage := vr.Stage().Next()
	if !nextStage.CanVote() {
		cs.Log().Error("invalid stage for move to next stage", "stage", nextStage)
	}

	// NOTE for next INIT, Round should be initialize to zero
	round := vr.Round()
	if nextStage == StageINIT {
		round = Round(0)
	}

	ballot, err := NewBallot(
		cs.homeState.Home().Address(),
		vr.Height(),
		round,
		nextStage,
		vr.Proposal(),
		vr.CurrentBlock(),
		vr.NextBlock(),
	)
	if err != nil {
		return err
	}

	cs.Log().Debug("broadcast ballot for next stage", "next_stage", nextStage, "ballot", ballot)

	if err := cs.networkClient.Vote(&ballot); err != nil {
		return err
	}

	return nil
}

func (cs *ConsensusStateHandler) startTimeoutINITBallot() error {
	cs.Lock()
	defer cs.Unlock()

	if cs.timer != nil {
		if err := cs.timer.Stop(); err != nil {
			return err
		}
	}

	// start timer for INITBallot
	cs.timer = common.NewCallbackTimer(
		"init-ballot",
		cs.policy.TimeoutINITBallot,
		cs.whenTimeoutINITBallot,
	)

	return cs.timer.Start()
}

func (cs *ConsensusStateHandler) whenTimeoutINITBallot(timer common.Timer) error {
	t := timer.(*common.CallbackTimer)
	cs.Log().Debug(
		"timeout for waiting INITBallot",
		"timeout", cs.policy.TimeoutINITBallot,
		"run_count", t.RunCount(),
	)

	ballot, err := NewBallot(
		cs.homeState.Home().Address(),
		cs.homeState.PreviousHeight(),
		Round(0),
		StageINIT,
		cs.homeState.Proposal(),
		cs.homeState.PreviousBlock().Hash(),
		cs.homeState.Block().Hash(),
	)
	if err != nil {
		return err
	}

	cs.Log().Debug("broadcast init ballot for timeout", "ballot", ballot)

	if err := cs.networkClient.Vote(&ballot); err != nil {
		return err
	}

	return nil
}
