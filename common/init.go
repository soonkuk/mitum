package common

import (
	"time"
)

var (
	timeSyncer *TimeSyncer
)

func Now() Time {
	if timeSyncer == nil {
		return Time{Time: time.Now()}
	}

	return Time{Time: time.Now().Add(timeSyncer.Offset())}
}
