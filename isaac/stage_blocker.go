package isaac

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

type StageBlockerDecision uint

const (
	NothingHappended StageBlockerDecision = iota
	ProposalAccepted
	StartNewRound
	GoToNextRound
	GoToNextStage
	FinishRound
)

func (e StageBlockerDecision) String() string {
	switch e {
	case ProposalAccepted:
		return "proposal accepted"
	case StartNewRound:
		return "start new round"
	case GoToNextRound:
		return "go to next round"
	case GoToNextStage:
		return "go to next stage"
	case FinishRound:
		return "finish round"
	}

	return ""
}

type StageBlockerResult struct {
	Decision StageBlockerDecision
	Err      error
}

func NewStageBlockerResult(decision StageBlockerDecision) StageBlockerResult {
	return StageBlockerResult{Decision: decision, Err: nil}
}

func NewStageBlockerResultError(err error) StageBlockerResult {
	return StageBlockerResult{Decision: NothingHappended, Err: err}
}

type stageBlockerChanFunc func() (VoteResultInfo, chan<- StageBlockerResult)

type StageBlocker struct {
	sync.RWMutex
	stopChan chan bool
	voteChan chan stageBlockerChanFunc
}

func NewStageBlocker() *StageBlocker {
	return &StageBlocker{
		voteChan: make(chan stageBlockerChanFunc),
	}
}

func (s *StageBlocker) Start() error {
	s.Lock()
	defer s.Unlock()

	if s.stopChan != nil {
		return common.StartStopperAlreadyStartedError
	}

	s.stopChan = make(chan bool)

	go s.blocking()

	return nil
}

func (s *StageBlocker) Stop() error {
	if s.stopChan == nil {
		return nil
	}

	s.Lock()
	defer s.Unlock()

	s.stopChan <- true
	close(s.stopChan)
	s.stopChan = nil

	return nil
}

func (s *StageBlocker) blocking() {
end:
	for {
		select {
		case <-s.stopChan:
			break end
		case f := <-s.voteChan:
			result, resultChan := f()

			resultChan <- s.check(result)
			close(resultChan)
		}
	}
}

func (s *StageBlocker) Check(result VoteResultInfo) <-chan StageBlockerResult {
	resultChan := make(chan StageBlockerResult, 1)

	if result.NotYet() {
		defer func() {
			resultChan <- NewStageBlockerResultError(
				InvalidVoteResultInfoError.SetMessage("not yet, but should be not"),
			)
			close(resultChan)
		}()

		return resultChan
	}

	go func() {
		s.voteChan <- func() (VoteResultInfo, chan<- StageBlockerResult) {
			return result, resultChan
		}
	}()

	return resultChan
}

func (s *StageBlocker) check(result VoteResultInfo) StageBlockerResult {
	switch result.Stage {
	case VoteStageINIT:
		if result.Proposed {
			return NewStageBlockerResult(ProposalAccepted)
		}

		return NewStageBlockerResult(StartNewRound)
	case VoteStageSIGN:
		if result.Result == VoteResultYES {
			return NewStageBlockerResult(GoToNextStage)
		}

		return NewStageBlockerResult(GoToNextRound)
	case VoteStageACCEPT:
		return NewStageBlockerResult(FinishRound)
	}

	return NewStageBlockerResultError(InvalidVoteStageError)
}
