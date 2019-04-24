// +build test

package common

import (
	"reflect"
	"runtime"

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

func NewRandomHash(hint string) Hash {
	h, _ := NewHash(hint, []byte(RandomUUID()))
	return h
}

func FuncName(f interface{}) string {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return v.String()
	}

	rf := runtime.FuncForPC(v.Pointer())
	if rf == nil {
		return v.String()
	}

	return rf.Name()
}
