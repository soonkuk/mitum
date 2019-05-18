package common

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/inconshreveable/log15"
)

type ChainCheckerStop map[string]interface{}

func NewChainCheckerStop(msg string, args ...interface{}) ChainCheckerStop {
	stop, err := ChainCheckerStop{}.SetMessage(msg, args...)
	if err != nil {
		stop, _ = ChainCheckerStop{}.SetMessage(msg)
	}

	return stop
}

func (c ChainCheckerStop) SetMessage(msg string, args ...interface{}) (ChainCheckerStop, error) {
	n := ChainCheckerStop{"msg": msg}

	for i := 0; i < len(args); i += 2 {
		s, ok := args[i].(string)
		if !ok {
			return ChainCheckerStop{}, fmt.Errorf("invalid key found in ChainCheckerStop: %T", args[i])
		}

		n[s] = args[i+1]
	}

	return n, nil
}

func (c ChainCheckerStop) JSONLog() {}

func (c ChainCheckerStop) Error() string {
	b, _ := json.Marshal(c)
	return TerminalLogString(string(b))
}

type ChainCheckerFunc func(*ChainChecker) error

type ChainChecker struct {
	*Logger
	checkers    []ChainCheckerFunc
	originalCtx context.Context
	ctx         context.Context
	deferFunc   func(*ChainChecker, ChainCheckerFunc, error)
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
	return c.success
}

func (c *ChainChecker) Log() log15.Logger {
	return c.log
}

func (c *ChainChecker) Context() context.Context {
	return c.ctx
}

func (c *ChainChecker) SetContext(key, value interface{}) *ChainChecker {
	c.ctx = context.WithValue(c.ctx, key, value)
	return c
}

func (c *ChainChecker) ContextValue(key interface{}, value interface{}) error {
	v := c.Context().Value(key)
	if v == nil {
		return ContextValueNotFoundError.SetMessage("key='%v'", key)
	}

	reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))

	return nil
}

func (c *ChainChecker) Check() error {
	c.success = false
	c.ctx = c.originalCtx // initialize context

	var err error
	var newChecker *ChainChecker

end:
	for _, f := range c.checkers {
		err = f(c)

		if c.deferFunc != nil {
			c.deferFunc(c, f, err)
		}

		if err == nil {
			continue
		}

		switch err.(type) {
		case *ChainChecker:
			newChecker = err.(*ChainChecker)
			break end
		case ChainCheckerStop:
			c.Log().Debug("checker stopped", "stop", err)
			c.success = true
			return nil
		default:
			c.Log().Error("failed to check", "error", err, "func", FuncName(f, false))
			return err
		}
	}

	if newChecker == nil {
		c.success = true
		return nil
	}

	newChecker.SetLogContext(c.LogContext()...)
	err = newChecker.Check()
	c.ctx = newChecker.Context()
	c.success = newChecker.success

	return err
}
