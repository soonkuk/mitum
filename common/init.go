package common

import (
	logging "github.com/inconshreveable/log15"
)

var InTest bool

var log logging.Logger = logging.New("module", "common")

func Log() logging.Logger {
	return log
}
