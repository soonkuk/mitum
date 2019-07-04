package isaac

import "time"

type Policy struct {
	TimeoutINITBallot     time.Duration // wait for INITBallot
	TimeoutINITVoteResult time.Duration // wait for INIT VoteResult
}
