package isaac

import "github.com/spikeekips/mitum/common"

type VotingResultChanFunc func() (common.Seal, chan<- VotingResult)

type VotingResult struct {
	Result VoteResultInfo
	Err    error
}

func NewVotingResult(result VoteResultInfo, err error) VotingResult {
	return VotingResult{Result: result, Err: err}
}

type Voting struct {
	sealChan  chan VotingResultChanFunc
	stopChan  chan bool
	votingBox VotingBox
}

func NewVoting(votingBox VotingBox) *Voting {
	return &Voting{
		sealChan:  make(chan VotingResultChanFunc),
		votingBox: votingBox,
	}
}

func (v *Voting) Start() error {
	if v.stopChan != nil {
		return common.StartStopperAlreadyStartedError
	}

	v.stopChan = make(chan bool)

	go v.blocking()

	return nil
}

func (v *Voting) Stop() error {
	if v.stopChan == nil {
		return nil
	}

	v.stopChan <- true
	close(v.stopChan)
	v.stopChan = nil

	return nil
}

func (v *Voting) blocking() {
end:
	for {
		select {
		case <-v.stopChan:
			break end
		case f := <-v.sealChan:
			seal, resultChan := f()

			var err error
			var result VoteResultInfo
			switch seal.Type {
			case ProposeSealType:
				result, err = v.votingBox.Open(seal)
			case BallotSealType:
				result, err = v.votingBox.Vote(seal)
			default:
				err = common.InvalidSealTypeError
			}

			resultChan <- NewVotingResult(result, err)
		}
	}
}

func (v *Voting) Vote(seal common.Seal) <-chan VotingResult {
	resultChan := make(chan VotingResult)

	go func() {
		v.sealChan <- func() (common.Seal, chan<- VotingResult) {
			return seal, resultChan
		}
	}()

	return resultChan
}
