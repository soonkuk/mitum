// +build test

package common

import (
	"github.com/inconshreveable/log15"
)

var (
	TestNetworkID []byte = []byte("this-is-test-network")
)

func SetTestLogger(logger log15.Logger) {
	handler, _ := LogHandler(LogFormatter("json"), "")
	logger.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, handler))
}

func init() {
	InTest = true
}
