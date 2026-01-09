package statuscache

import (
	"context"
)

// Service defines the interface for status cache operations
type Service interface {
	// CacheStatusEvent caches a status event and returns whether webhook should be suppressed
	CacheStatusEvent(ctx context.Context, event *StatusEvent, rawPayload []byte) (suppress bool, err error)

	// Aggregated queries (summary format)
	GetStatus(ctx context.Context, instanceID, messageID string, includeParticipants bool) (*AggregatedStatus, error)
	QueryByGroup(ctx context.Context, instanceID, groupID string, params QueryParams) (*QueryResult, error)
	QueryByPhone(ctx context.Context, instanceID, phone string, params QueryParams) (*QueryResult, error)

	// Raw queries (FUNNELCHAT webhook format)
	GetRawStatus(ctx context.Context, instanceID, messageID string) (*RawQueryResult, error)
	QueryRawByGroup(ctx context.Context, instanceID, groupID string, params QueryParams) (*RawQueryResult, error)
	QueryRawByPhone(ctx context.Context, instanceID, phone string, params QueryParams) (*RawQueryResult, error)

	// Flush operations (trigger suppressed webhooks)
	FlushMessage(ctx context.Context, instanceID, messageID string) (*FlushResult, error)
	FlushAll(ctx context.Context, instanceID string) (*FlushResult, error)

	// Cache management
	ClearMessage(ctx context.Context, instanceID, messageID string) error
	ClearInstance(ctx context.Context, instanceID string) (int64, error)
	GetStats(ctx context.Context, instanceID string) (*CacheStats, error)

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Health
	IsHealthy(ctx context.Context) bool
}

// WebhookDispatcher is called when flushing pending webhooks
type WebhookDispatcher interface {
	DispatchWebhook(ctx context.Context, instanceID string, payload []byte) error
}
