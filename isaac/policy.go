package isaac

import "time"

type Policy struct {
	TimeoutINITBallot        time.Duration // wait for INITBallot
	IntervalINITBallotOfJoin time.Duration // interval for broadcasting init ballot in join state
}
