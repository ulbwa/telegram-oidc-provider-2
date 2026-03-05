package tx

import "context"

type Manager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type NoOpManager struct{}

func (NoOpManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
