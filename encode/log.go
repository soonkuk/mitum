package encode

import (
	"github.com/inconshreveable/log15"
)

var log log15.Logger = log15.New("module", "encode")

func Log() log15.Logger {
	return log
}
