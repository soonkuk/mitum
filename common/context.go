package common

import (
	"context"
	"fmt"
)

func SetContextWithArgs(ctx context.Context, args ...interface{}) context.Context {
	if len(args)%2 != 0 {
		panic(fmt.Errorf("invalid number of args: %v", len(args)))
	}

	if ctx == nil || ctx == context.TODO() {
		ctx = context.Background()
	}

	for i := 0; i < len(args); i += 2 {
		ctx = context.WithValue(ctx, args[i], args[i+1])
	}

	return ctx
}

func SetContext(ctx context.Context, key, value interface{}) context.Context {
	if ctx == nil || ctx == context.TODO() {
		ctx = context.Background()
	}

	return context.WithValue(ctx, key, value)
}
