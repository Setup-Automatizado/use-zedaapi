package pollstore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
)

type Store interface {
	SaveOptions(ctx context.Context, instanceID uuid.UUID, pollID string, mapping map[string]string) error
	ResolveOptions(ctx context.Context, instanceID uuid.UUID, pollID string, hashes [][]byte) ([]string, error)
}

const (
	MaxTTL     = time.Second * time.Duration(math.MaxInt32-1)
	DefaultTTL = 365 * 24 * time.Hour
)

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStore(client *redis.Client, ttl time.Duration) *RedisStore {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if ttl > MaxTTL {
		ttl = MaxTTL
	}
	return &RedisStore{client: client, ttl: ttl}
}

func (s *RedisStore) SaveOptions(ctx context.Context, instanceID uuid.UUID, pollID string, mapping map[string]string) error {
	if s == nil || s.client == nil {
		return nil
	}
	if pollID == "" || len(mapping) == 0 {
		return nil
	}
	payload, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.key(instanceID, pollID), payload, s.ttl).Err()
}

func (s *RedisStore) ResolveOptions(ctx context.Context, instanceID uuid.UUID, pollID string, hashes [][]byte) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	if pollID == "" || len(hashes) == 0 {
		return nil, nil
	}
	raw, err := s.client.Get(ctx, s.key(instanceID, pollID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var mapping map[string]string
	if err := json.Unmarshal([]byte(raw), &mapping); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		encoded := strings.ToLower(hex.EncodeToString(hash))
		names = append(names, mapping[encoded])
	}
	return names, nil
}

func (s *RedisStore) key(instanceID uuid.UUID, pollID string) string {
	return fmt.Sprintf("funnelchat:instance:%s:poll:%s", instanceID.String(), pollID)
}
