package common

import (
	"context"
	"reflect"

	"github.com/inconshreveable/log15"
)

type ChainCheckerStop struct {
}

func (c ChainCheckerStop) Error() string {
	return "chain-checker-stop"
}

type ChainCheckerFunc func(*ChainChecker) error

type ChainChecker struct {
	name        string
	checkers    []ChainCheckerFunc
	originalCtx context.Context
	ctx         context.Context
	log         log15.Logger
	deferFunc   func(*ChainChecker, ChainCheckerFunc, error)
	success     bool
}

func NewChainChecker(name string, ctx context.Context, checkers ...ChainCheckerFunc) *ChainChecker {
	return &ChainChecker{
		name:        name,
		checkers:    checkers,
		ctx:         ctx,
		originalCtx: ctx,
		log:         log.New(log15.Ctx{"module": "ChainChecker", "name": name}),
	}
}

func (c *ChainChecker) New(ctx context.Context) *ChainChecker {
	if ctx == nil {
		ctx = c.originalCtx
	}

	return &ChainChecker{
		name:        c.name,
		checkers:    c.checkers,
		ctx:         ctx,
		originalCtx: ctx,
		log:         c.log,
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
			err = nil
			break end
		case ChainCheckerStop:
			c.log.Debug("checker stopped")
			c.success = true
			return nil
		default:
			c.log.Error("checking", "error", err, "func", FuncName(f, false))
			return err
		}
	}

	if newChecker == nil {
		c.success = true
		return nil
	}

	err = newChecker.Check()
	c.ctx = newChecker.Context()
	c.success = newChecker.success

	return err
}
