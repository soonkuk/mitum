package common

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/inconshreveable/log15"
	"golang.org/x/xerrors"
)

const (
	ChainCheckerStopErrorCode ErrorCode = iota + 1
	ContextValueNotFoundErrorCode
)

var (
	ChainCheckerStopError     = NewError("chain", ChainCheckerStopErrorCode, "chain stopped")
	ContextValueNotFoundError = NewError("chain", ContextValueNotFoundErrorCode, "value not found in context")
)

type ChainCheckerFunc func(*ChainChecker) error

type ChainChecker struct {
	sync.RWMutex
	*Logger
	checkers    []ChainCheckerFunc
	originalCtx context.Context
	ctx         context.Context
	success     bool
}

func NewChainChecker(name string, ctx context.Context, checkers ...ChainCheckerFunc) *ChainChecker {
	return &ChainChecker{
		Logger:      NewLogger(log, "module", name),
		checkers:    checkers,
		ctx:         ctx,
		originalCtx: ctx,
	}
}

func (c *ChainChecker) New(ctx context.Context) *ChainChecker {
	c.RLock()
	defer c.RUnlock()

	if ctx == nil {
		ctx = c.originalCtx
	}

	return &ChainChecker{
		Logger:      c.Logger,
		checkers:    c.checkers,
		ctx:         ctx,
		originalCtx: ctx,
	}
}

func (c *ChainChecker) Error() string {
	return "ChainChecker will be also chained"
}

func (c *ChainChecker) Success() bool {
	c.RLock()
	defer c.RUnlock()

	return c.success
}

func (c *ChainChecker) setSuccess(v bool) {
	c.Lock()
	defer c.Unlock()

	c.success = v
}

func (c *ChainChecker) Log() log15.Logger {
	return c.log
}

func (c *ChainChecker) Context() context.Context {
	c.RLock()
	defer c.RUnlock()

	return c.ctx
}

func (c *ChainChecker) SetContext(args ...interface{}) *ChainChecker {
	if len(args)%2 != 0 {
		panic(fmt.Errorf("invalid number of args: %v", len(args)))
	}

	c.Lock()
	defer c.Unlock()

	for i := 0; i < len(args); i += 2 {
		c.ctx = context.WithValue(c.ctx, args[i], args[i+1])
	}

	return c
}

func (c *ChainChecker) ContextValue(key interface{}, value interface{}) error {
	v := c.Context().Value(key)
	if v == nil {
		return ContextValueNotFoundError.Newf("key='%v'", key)
	}

	reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))

	return nil
}

func (c *ChainChecker) Check() error {
	c.setSuccess(false)

	// initialize context
	c.Lock()
	c.originalCtx = c.ctx
	c.ctx = c.originalCtx
	c.Unlock()

	var err error
	var newChecker *ChainChecker

end:
	for _, f := range c.checkers {
		err = f(c)

		if err == nil {
			continue
		}

		switch err.(type) {
		case *ChainChecker:
			newChecker = err.(*ChainChecker)
			break end
		default:
			if xerrors.Is(err, ChainCheckerStopError) {
				c.Log().Debug("checker stopped", "stop", err)
				c.setSuccess(true)
				return nil
			}

			c.Log().Error("failed to check", "error", err, "func", FuncName(f, false))
			return err
		}
	}

	if newChecker == nil {
		c.setSuccess(true)
		return nil
	}

	newChecker.SetLogContext(c.LogContext())
	err = newChecker.Check()
	c.setSuccess(newChecker.success)

	c.Lock()
	c.ctx = newChecker.Context()
	c.Unlock()

	return err
}
