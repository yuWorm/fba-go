package migration

import "context"

type Lock interface {
	Acquire(ctx context.Context) (release func(context.Context) error, err error)
}

type NoopLock struct{}

func (NoopLock) Acquire(context.Context) (func(context.Context) error, error) {
	return func(context.Context) error { return nil }, nil
}
