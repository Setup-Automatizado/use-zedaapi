package locks

import "context"

type Lock interface {
	Refresh(ctx context.Context, ttlSeconds int) error
	Release(ctx context.Context) error
	GetValue() string
}

type Manager interface {
	Acquire(ctx context.Context, key string, ttlSeconds int) (Lock, bool, error)
}
