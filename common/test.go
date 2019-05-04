// +build test

package common

import (
	"runtime/debug"

	"github.com/inconshreveable/log15"
)

var (
	TestNetworkID []byte = []byte("this-is-test-network")
)

func init() {
	InTest = true
	SetTestLogger(Log())
}

func SetTestLogger(logger log15.Logger) {
	handler, _ := LogHandler(LogFormatter("terminal"), "")
	logger.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, handler))
	//logger.SetHandler(log15.LvlFilterHandler(log15.LvlCrit, handler))
}

func NewRandomHash(hint string) Hash {
	h, _ := NewHash(hint, []byte(RandomUUID()))
	return h
}

func NewRandomHome() *HomeNode {
	return NewHome(RandomSeed(), NetAddr{})
}

func DebugPanic() {
	if r := recover(); r != nil {
		debug.PrintStack()
		panic(r)
	}
}
