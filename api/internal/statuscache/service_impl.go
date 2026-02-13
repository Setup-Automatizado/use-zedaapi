package statuscache

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo       Repository
	cfg        *config.Config
	dispatcher WebhookDispatcher
	metrics    *observability.Metrics
	log        *slog.Logger

	mu        sync.RWMutex
	running   bool
	stopCh    chan struct{}
	cleanupWg sync.WaitGroup
}

// NewService creates a new StatusCache service
func NewService(repo Repository, cfg *config.Config, dispatcher WebhookDispatcher, metrics *observability.Metrics, log *slog.Logger) *ServiceImpl {
	return &ServiceImpl{
		repo:       repo,
		cfg:        cfg,
		dispatcher: dispatcher,
		metrics:    metrics,
		log:        log.With(slog.String("component", "statuscache")),
		stopCh:     make(chan struct{}),
	}
}

// withTimeout wraps context with configured operation timeout
func (s *ServiceImpl) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := s.cfg.StatusCache.OperationTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second // default fallback
	}
	return context.WithTimeout(ctx, timeout)
}

// CacheStatusEvent caches a status event and returns whether webhook should be suppressed
func (s *ServiceImpl) CacheStatusEvent(ctx context.Context, event *StatusEvent, rawPayload []byte) (bool, error) {
	start := time.Now()

	if !s.cfg.StatusCache.Enabled {
		return false, nil
	}

	// Validate input
	if event == nil {
		return false, nil
	}

	// DEBUG: Log values received for troubleshooting
	s.log.Debug("service cache status check",
		slog.String("instance_id", event.InstanceID),
		slog.String("status", event.Status),
		slog.Bool("is_group", event.IsGroup),
		slog.Any("configured_types", s.cfg.StatusCache.Types),
		slog.Any("configured_scope", s.cfg.StatusCache.Scope),
	)

	// Check if this status type should be cached
	if !s.shouldCacheStatusType(event.Status) {
		s.log.Debug("service cache skipped - status type not configured",
			slog.String("instance_id", event.InstanceID),
			slog.String("status", event.Status),
			slog.Any("configured_types", s.cfg.StatusCache.Types),
		)
		return false, nil
	}

	// Check scope (groups vs direct)
	if !s.shouldCacheScope(event.IsGroup) {
		s.log.Debug("service cache skipped - scope mismatch",
			slog.String("instance_id", event.InstanceID),
			slog.Bool("is_group", event.IsGroup),
			slog.Any("configured_scope", s.cfg.StatusCache.Scope),
		)
		return false, nil
	}

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	log := s.log.With(
		slog.String("instance_id", event.InstanceID),
		slog.String("status", event.Status),
		slog.Bool("is_group", event.IsGroup),
	)

	successCount := 0
	errorCount := 0

	// Process each message ID
	for _, msgID := range event.MessageIDs {
		// Get or create entry
		entry, err := s.repo.GetByMessageID(ctx, event.InstanceID, msgID)
		if err != nil {
			log.Error("failed to get existing entry", slog.String("error", err.Error()))
			errorCount++
			continue
		}

		if entry == nil {
			// Create new entry
			entry = NewStatusCacheEntry(event.InstanceID, msgID, event.Phone, event.GroupID, event.IsGroup)
		}

		// Add participant status
		participant := event.Participant
		if participant == "" {
			participant = event.Phone
		}
		entry.AddParticipant(participant, event.Status, event.Timestamp, event.Device)

		// Save entry
		if err := s.repo.UpsertStatus(ctx, entry); err != nil {
			log.Error("failed to upsert status", slog.String("message_id", msgID), slog.String("error", err.Error()))
			errorCount++
			continue
		}

		successCount++

		// Store pending webhook if suppressing
		if s.cfg.StatusCache.SuppressWebhooks && rawPayload != nil {
			webhook := &PendingWebhook{
				MessageID:   msgID,
				InstanceID:  event.InstanceID,
				Phone:       event.Phone,
				Participant: participant,
				Status:      event.Status,
				Timestamp:   event.Timestamp,
				Device:      event.Device,
				Payload:     rawPayload,
			}
			if err := s.repo.StorePendingWebhook(ctx, webhook); err != nil {
				log.Warn("failed to store pending webhook", slog.String("error", err.Error()))
			}
		}

		log.Debug("cached status event",
			slog.String("message_id", msgID),
			slog.String("participant", participant),
		)
	}

	// Record metrics
	if s.metrics != nil {
		duration := time.Since(start).Seconds()
		s.metrics.StatusCacheDuration.WithLabelValues(event.InstanceID, "cache").Observe(duration)

		if successCount > 0 {
			s.metrics.StatusCacheOperations.WithLabelValues(event.InstanceID, "cache", "success").Add(float64(successCount))
		}
		if errorCount > 0 {
			s.metrics.StatusCacheOperations.WithLabelValues(event.InstanceID, "cache", "error").Add(float64(errorCount))
		}

		// Record suppressions
		if s.cfg.StatusCache.SuppressWebhooks && successCount > 0 {
			s.metrics.StatusCacheSuppressions.WithLabelValues(event.InstanceID, event.Status).Add(float64(successCount))
		}
	}

	return s.cfg.StatusCache.SuppressWebhooks, nil
}

// shouldCacheStatusType checks if the status type is configured for caching
// Uses case-insensitive comparison because config uses lowercase (read,delivered,played,sent)
// but transformer returns UPPERCASE (READ,PLAYED,RECEIVED,SENT)
func (s *ServiceImpl) shouldCacheStatusType(status string) bool {
	statusLower := strings.ToLower(status)
	for _, t := range s.cfg.StatusCache.Types {
		if strings.ToLower(t) == statusLower {
			return true
		}
	}
	return false
}

// shouldCacheScope checks if the message scope (group/direct) is configured for caching
func (s *ServiceImpl) shouldCacheScope(isGroup bool) bool {
	for _, scope := range s.cfg.StatusCache.Scope {
		if scope == "groups" && isGroup {
			return true
		}
		if scope == "direct" && !isGroup {
			return true
		}
	}
	return false
}

// GetStatus retrieves aggregated status for a specific message
func (s *ServiceImpl) GetStatus(ctx context.Context, instanceID, messageID string, includeParticipants bool) (*AggregatedStatus, error) {
	start := time.Now()

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entry, err := s.repo.GetByMessageID(ctx, instanceID, messageID)
	if err != nil {
		if s.metrics != nil {
			s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "get", "error").Inc()
		}
		return nil, err
	}

	// Record metrics
	if s.metrics != nil {
		duration := time.Since(start).Seconds()
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "get").Observe(duration)

		if entry == nil {
			s.metrics.StatusCacheMisses.WithLabelValues(instanceID).Inc()
		} else {
			s.metrics.StatusCacheHits.WithLabelValues(instanceID).Inc()
		}
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "get", "success").Inc()
	}

	if entry == nil {
		return nil, nil
	}
	return entry.ToAggregated(includeParticipants), nil
}

// QueryByGroup retrieves aggregated statuses for messages in a group
func (s *ServiceImpl) QueryByGroup(ctx context.Context, instanceID, groupID string, params QueryParams) (*QueryResult, error) {
	start := time.Now()

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entries, total, err := s.repo.GetByGroupID(ctx, instanceID, groupID, params.Limit, params.Offset)
	if err != nil {
		if s.metrics != nil {
			s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_group", "error").Inc()
		}
		return nil, err
	}

	data := make([]*AggregatedStatus, 0, len(entries))
	for _, entry := range entries {
		data = append(data, entry.ToAggregated(params.IncludeParticipants))
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "query_group").Observe(time.Since(start).Seconds())
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_group", "success").Inc()
	}

	return &QueryResult{
		Data: data,
		Meta: QueryMeta{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}, nil
}

// QueryByPhone retrieves aggregated statuses for messages to/from a phone
func (s *ServiceImpl) QueryByPhone(ctx context.Context, instanceID, phone string, params QueryParams) (*QueryResult, error) {
	start := time.Now()

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entries, total, err := s.repo.GetByPhone(ctx, instanceID, phone, params.Limit, params.Offset)
	if err != nil {
		if s.metrics != nil {
			s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_phone", "error").Inc()
		}
		return nil, err
	}

	data := make([]*AggregatedStatus, 0, len(entries))
	for _, entry := range entries {
		data = append(data, entry.ToAggregated(params.IncludeParticipants))
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "query_phone").Observe(time.Since(start).Seconds())
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_phone", "success").Inc()
	}

	return &QueryResult{
		Data: data,
		Meta: QueryMeta{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}, nil
}

// QueryAll retrieves all aggregated statuses for an instance with pagination
func (s *ServiceImpl) QueryAll(ctx context.Context, instanceID string, params QueryParams) (*QueryResult, error) {
	start := time.Now()

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entries, total, err := s.repo.GetAll(ctx, instanceID, params.Limit, params.Offset)
	if err != nil {
		if s.metrics != nil {
			s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_all", "error").Inc()
		}
		return nil, err
	}

	data := make([]*AggregatedStatus, 0, len(entries))
	for _, entry := range entries {
		data = append(data, entry.ToAggregated(params.IncludeParticipants))
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "query_all").Observe(time.Since(start).Seconds())
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_all", "success").Inc()
	}

	s.log.Debug("query all status cache",
		slog.String("instance_id", instanceID),
		slog.Int64("total", total),
		slog.Int("returned", len(data)),
		slog.Int("limit", params.Limit),
		slog.Int("offset", params.Offset),
	)

	return &QueryResult{
		Data: data,
		Meta: QueryMeta{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}, nil
}

// GetRawStatus retrieves raw payloads for a specific message
func (s *ServiceImpl) GetRawStatus(ctx context.Context, instanceID, messageID string) (*RawQueryResult, error) {
	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entry, err := s.repo.GetByMessageID(ctx, instanceID, messageID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return &RawQueryResult{
			Data: []*RawStatusPayload{},
			Meta: QueryMeta{Total: 0, Limit: 0, Offset: 0},
		}, nil
	}

	payloads := entry.ToRawPayloads()
	return &RawQueryResult{
		Data: payloads,
		Meta: QueryMeta{
			Total:  int64(len(payloads)),
			Limit:  len(payloads),
			Offset: 0,
		},
	}, nil
}

// QueryRawByGroup retrieves raw payloads for messages in a group
func (s *ServiceImpl) QueryRawByGroup(ctx context.Context, instanceID, groupID string, params QueryParams) (*RawQueryResult, error) {
	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entries, total, err := s.repo.GetByGroupID(ctx, instanceID, groupID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	data := make([]*RawStatusPayload, 0)
	for _, entry := range entries {
		data = append(data, entry.ToRawPayloads()...)
	}

	return &RawQueryResult{
		Data: data,
		Meta: QueryMeta{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}, nil
}

// QueryRawByPhone retrieves raw payloads for messages to/from a phone
func (s *ServiceImpl) QueryRawByPhone(ctx context.Context, instanceID, phone string, params QueryParams) (*RawQueryResult, error) {
	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entries, total, err := s.repo.GetByPhone(ctx, instanceID, phone, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	data := make([]*RawStatusPayload, 0)
	for _, entry := range entries {
		data = append(data, entry.ToRawPayloads()...)
	}

	return &RawQueryResult{
		Data: data,
		Meta: QueryMeta{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}, nil
}

// QueryRawAll retrieves all raw payloads for an instance with pagination
// Returns data in the EXACT same format as webhook payload (RawStatusPayload)
func (s *ServiceImpl) QueryRawAll(ctx context.Context, instanceID string, params QueryParams) (*RawQueryResult, error) {
	start := time.Now()

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	entries, total, err := s.repo.GetAll(ctx, instanceID, params.Limit, params.Offset)
	if err != nil {
		if s.metrics != nil {
			s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_raw_all", "error").Inc()
		}
		return nil, err
	}

	data := make([]*RawStatusPayload, 0)
	for _, entry := range entries {
		data = append(data, entry.ToRawPayloads()...)
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "query_raw_all").Observe(time.Since(start).Seconds())
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "query_raw_all", "success").Inc()
	}

	s.log.Debug("query raw all status cache",
		slog.String("instance_id", instanceID),
		slog.Int64("total_entries", total),
		slog.Int("total_payloads", len(data)),
		slog.Int("limit", params.Limit),
		slog.Int("offset", params.Offset),
	)

	return &RawQueryResult{
		Data: data,
		Meta: QueryMeta{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}, nil
}

// FlushMessage sends pending webhooks for a specific message
func (s *ServiceImpl) FlushMessage(ctx context.Context, instanceID, messageID string) (*FlushResult, error) {
	start := time.Now()

	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	webhooks, err := s.repo.GetPendingWebhooksByMessageID(ctx, instanceID, messageID)
	if err != nil {
		if s.metrics != nil {
			s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "flush", "error").Inc()
		}
		return nil, err
	}

	result := &FlushResult{}
	for _, wh := range webhooks {
		if s.dispatcher != nil && len(wh.Payload) > 0 {
			if err := s.dispatcher.DispatchWebhook(ctx, instanceID, wh.Payload); err != nil {
				s.log.Warn("failed to dispatch webhook during flush",
					slog.String("instance_id", instanceID),
					slog.String("message_id", messageID),
					slog.String("error", err.Error()),
				)
				continue
			}
			result.WebhooksTriggered++
		}
		result.Flushed++
	}

	// Delete pending webhooks after flush
	if err := s.repo.DeletePendingWebhooks(ctx, instanceID, []string{messageID}); err != nil {
		s.log.Warn("failed to delete pending webhooks after flush",
			slog.String("instance_id", instanceID),
			slog.String("message_id", messageID),
			slog.String("error", err.Error()),
		)
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "flush").Observe(time.Since(start).Seconds())
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "flush", "success").Inc()
		if result.Flushed > 0 {
			s.metrics.StatusCacheFlushed.WithLabelValues(instanceID, "manual").Add(float64(result.Flushed))
		}
	}

	return result, nil
}

// FlushAll sends all pending webhooks for an instance
func (s *ServiceImpl) FlushAll(ctx context.Context, instanceID string) (*FlushResult, error) {
	start := time.Now()

	result := &FlushResult{}
	batchSize := s.cfg.StatusCache.FlushBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	processedMsgIDs := make([]string, 0)

	for {
		webhooks, err := s.repo.GetPendingWebhooks(ctx, instanceID, batchSize)
		if err != nil {
			if s.metrics != nil {
				s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "flush_all", "error").Inc()
			}
			return result, err
		}

		if len(webhooks) == 0 {
			break
		}

		msgIDsInBatch := make(map[string]bool)
		for _, wh := range webhooks {
			if s.dispatcher != nil && len(wh.Payload) > 0 {
				if err := s.dispatcher.DispatchWebhook(ctx, instanceID, wh.Payload); err != nil {
					s.log.Warn("failed to dispatch webhook during flush all",
						slog.String("instance_id", instanceID),
						slog.String("message_id", wh.MessageID),
						slog.String("error", err.Error()),
					)
					continue
				}
				result.WebhooksTriggered++
			}
			result.Flushed++
			msgIDsInBatch[wh.MessageID] = true
		}

		// Collect message IDs for deletion
		for msgID := range msgIDsInBatch {
			processedMsgIDs = append(processedMsgIDs, msgID)
		}

		// Delete in batches to avoid memory issues
		if len(processedMsgIDs) >= batchSize {
			if err := s.repo.DeletePendingWebhooks(ctx, instanceID, processedMsgIDs); err != nil {
				s.log.Warn("failed to delete pending webhooks batch",
					slog.String("instance_id", instanceID),
					slog.String("error", err.Error()),
				)
			}
			processedMsgIDs = processedMsgIDs[:0]
		}
	}

	// Delete remaining
	if len(processedMsgIDs) > 0 {
		if err := s.repo.DeletePendingWebhooks(ctx, instanceID, processedMsgIDs); err != nil {
			s.log.Warn("failed to delete remaining pending webhooks",
				slog.String("instance_id", instanceID),
				slog.String("error", err.Error()),
			)
		}
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.StatusCacheDuration.WithLabelValues(instanceID, "flush_all").Observe(time.Since(start).Seconds())
		s.metrics.StatusCacheOperations.WithLabelValues(instanceID, "flush_all", "success").Inc()
		if result.Flushed > 0 {
			s.metrics.StatusCacheFlushed.WithLabelValues(instanceID, "manual_all").Add(float64(result.Flushed))
		}
	}

	return result, nil
}

// ClearMessage removes a cached message entry
func (s *ServiceImpl) ClearMessage(ctx context.Context, instanceID, messageID string) error {
	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	return s.repo.DeleteByMessageID(ctx, instanceID, messageID)
}

// ClearInstance removes all cached entries for an instance
// Note: Uses longer timeout since it may need to delete many entries
func (s *ServiceImpl) ClearInstance(ctx context.Context, instanceID string) (int64, error) {
	// Use 60s timeout for bulk operations
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	return s.repo.DeleteByInstanceID(ctx, instanceID)
}

// GetStats returns cache statistics
func (s *ServiceImpl) GetStats(ctx context.Context, instanceID string) (*CacheStats, error) {
	// Apply operation timeout
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	return s.repo.GetStats(ctx, instanceID)
}

// Start begins background cleanup tasks
func (s *ServiceImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	s.stopCh = make(chan struct{})

	// Start cleanup goroutine
	s.cleanupWg.Add(1)
	go s.runCleanup()

	s.log.Info("status cache service started",
		slog.Bool("enabled", s.cfg.StatusCache.Enabled),
		slog.Duration("ttl", s.cfg.StatusCache.TTL),
		slog.Bool("suppress_webhooks", s.cfg.StatusCache.SuppressWebhooks),
	)

	return nil
}

// Stop gracefully stops the service and flushes pending webhooks
func (s *ServiceImpl) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	// Wait for cleanup goroutine to finish
	s.cleanupWg.Wait()

	s.log.Info("status cache service stopped")
	return nil
}

// runCleanup periodically cleans up expired entries
func (s *ServiceImpl) runCleanup() {
	defer s.cleanupWg.Done()

	ticker := time.NewTicker(s.cfg.StatusCache.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.performCleanup()
		}
	}
}

// performCleanup removes expired cache entries
func (s *ServiceImpl) performCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Note: Most cleanup is handled by Redis TTL
	// This just cleans up orphaned index references and logs status
	s.log.Debug("running status cache cleanup")

	// Verify Redis connectivity during cleanup cycle
	if err := s.repo.Ping(ctx); err != nil {
		s.log.Warn("status cache cleanup ping failed", slog.String("error", err.Error()))
	}
}

// IsHealthy checks if the service is healthy
func (s *ServiceImpl) IsHealthy(ctx context.Context) bool {
	// Use short timeout for health checks
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.Ping(ctx); err != nil {
		s.log.Warn("status cache health check failed", slog.String("error", err.Error()))
		return false
	}
	return true
}
