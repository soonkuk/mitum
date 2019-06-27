package transaction

import (
	"github.com/inconshreveable/log15"
)

var log log15.Logger = log15.New("module", "transaction")

func Log() log15.Logger {
	return log
}
