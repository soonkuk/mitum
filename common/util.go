package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

var (
	basePath string
)

type StartStopper interface {
	Start() error
	Stop() error
}

func init() {
	basePath = filepath.Dir(reflect.TypeOf(OverflowError).PkgPath())
}

func PrettyMap(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for k, v := range m {
		fmt.Fprintf(b, "%s=%v ", k, v)
	}

	return strings.TrimSpace(b.String())
}

func FuncName(f interface{}, full bool) string {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return v.String()
	}

	rf := runtime.FuncForPC(v.Pointer())
	if rf == nil {
		return v.String()
	}

	if full {
		return rf.Name()
	}

	if !strings.HasPrefix(rf.Name(), basePath) {
		return rf.Name()
	}
	return rf.Name()[len(basePath)+1:]
}

func ContextWithValues(ctx context.Context, args ...interface{}) context.Context {
	if len(args)%2 != 0 {
		panic(errors.New(fmt.Sprintf("invalid number of args: %v", len(args))))
	}

	if ctx == nil {
		ctx = context.Background()
	}

	for i := 0; i < len(args); i += 2 {
		ctx = context.WithValue(ctx, args[i], args[i+1])
	}

	return ctx
}
