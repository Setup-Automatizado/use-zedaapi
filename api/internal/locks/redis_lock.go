package locks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var releaseScript = redis.NewScript(`
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('del', KEYS[1])
else
    return 0
end
`)

// RedisManager implements a basic distributed lock using Redis.
type RedisManager struct {
	client *redis.Client
}

// NewRedisManager instantiates a new manager.
func NewRedisManager(client *redis.Client) *RedisManager {
	return &RedisManager{client: client}
}

// Acquire tries to obtain the lock for key with the provided TTL (seconds).
func (m *RedisManager) Acquire(ctx context.Context, key string, ttlSeconds int) (Lock, bool, error) {
	if m == nil || m.client == nil {
		return nil, false, errors.New("redis manager not configured")
	}
	value := randomToken()
	set, err := m.client.SetNX(ctx, key, value, durationFromSeconds(ttlSeconds)).Result()
	if err != nil {
		return nil, false, err
	}
	if !set {
		return nil, false, nil
	}
	return &redisLock{client: m.client, key: key, value: value}, true, nil
}

// redisLock holds an acquired lock instance.
type redisLock struct {
	client *redis.Client
	key    string
	value  string
}

func (l *redisLock) Refresh(ctx context.Context, ttlSeconds int) error {
	return l.client.Expire(ctx, l.key, durationFromSeconds(ttlSeconds)).Err()
}

func (l *redisLock) Release(ctx context.Context) error {
	_, err := releaseScript.Run(ctx, l.client, []string{l.key}, l.value).Result()
	return err
}

func randomToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func durationFromSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}
