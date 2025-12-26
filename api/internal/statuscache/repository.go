package statuscache

import (
	"context"
)

// Repository defines the interface for status cache storage operations
type Repository interface {
	// Upsert operations
	UpsertStatus(ctx context.Context, entry *StatusCacheEntry) error
	AddParticipantStatus(ctx context.Context, instanceID, messageID string, participant *ParticipantStatus) error

	// Query operations
	GetByMessageID(ctx context.Context, instanceID, messageID string) (*StatusCacheEntry, error)
	GetByGroupID(ctx context.Context, instanceID, groupID string, limit, offset int) ([]*StatusCacheEntry, int64, error)
	GetByPhone(ctx context.Context, instanceID, phone string, limit, offset int) ([]*StatusCacheEntry, int64, error)

	// Pending webhooks (for flush)
	StorePendingWebhook(ctx context.Context, webhook *PendingWebhook) error
	GetPendingWebhooks(ctx context.Context, instanceID string, limit int) ([]*PendingWebhook, error)
	GetPendingWebhooksByMessageID(ctx context.Context, instanceID, messageID string) ([]*PendingWebhook, error)
	DeletePendingWebhooks(ctx context.Context, instanceID string, messageIDs []string) error
	CountPendingWebhooks(ctx context.Context, instanceID string) (int64, error)

	// Cleanup operations
	DeleteByMessageID(ctx context.Context, instanceID, messageID string) error
	DeleteByInstanceID(ctx context.Context, instanceID string) (int64, error)
	DeleteExpired(ctx context.Context, instanceID string) (int64, error)

	// Stats
	Count(ctx context.Context, instanceID string) (int64, error)
	GetStats(ctx context.Context, instanceID string) (*CacheStats, error)

	// Health
	Ping(ctx context.Context) error
}
