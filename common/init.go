package common

import (
	"github.com/inconshreveable/log15"
)

var InTest bool
var DEBUG bool

var log log15.Logger = log15.New("module", "common")

func Log() log15.Logger {
	return log
}
