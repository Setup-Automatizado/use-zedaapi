package locks

import "context"

// Lock represents an acquired distributed lock.
type Lock interface {
	Refresh(ctx context.Context, ttlSeconds int) error
	Release(ctx context.Context) error
}

// Manager can acquire locks identified by a key.
type Manager interface {
	Acquire(ctx context.Context, key string, ttlSeconds int) (Lock, bool, error)
}
