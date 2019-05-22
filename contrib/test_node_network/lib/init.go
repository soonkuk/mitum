package lib

import "github.com/inconshreveable/log15"

var log log15.Logger = log15.New("module", "test-node-network/lib")

func Log() log15.Logger {
	return log
}
