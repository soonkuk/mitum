package consensus

import "github.com/inconshreveable/log15"

var log log15.Logger = log15.New("module", "consensus")

func Log() log15.Logger {
	return log
}
