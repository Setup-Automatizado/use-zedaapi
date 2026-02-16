package statuscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// RedisRepository implements Repository using Redis
type RedisRepository struct {
	client    *redis.Client
	ttl       time.Duration
	keyPrefix string
}

// NewRedisRepository creates a new RedisRepository
func NewRedisRepository(client *redis.Client, ttl time.Duration, keyPrefix string) *RedisRepository {
	if keyPrefix == "" {
		keyPrefix = "zedaapi"
	}
	return &RedisRepository{
		client:    client,
		ttl:       ttl,
		keyPrefix: keyPrefix,
	}
}

// Key patterns:
// - Status entry: {prefix}:status:{instanceID}:msg:{messageID}
// - Group index: {prefix}:status:{instanceID}:idx:group:{groupID}
// - Phone index: {prefix}:status:{instanceID}:idx:phone:{phone}
// - Pending webhooks: {prefix}:status:{instanceID}:pending:{messageID}
// - Instance entry list: {prefix}:status:{instanceID}:entries

func (r *RedisRepository) msgKey(instanceID, messageID string) string {
	return fmt.Sprintf("%s:status:%s:msg:%s", r.keyPrefix, instanceID, messageID)
}

func (r *RedisRepository) groupIdxKey(instanceID, groupID string) string {
	return fmt.Sprintf("%s:status:%s:idx:group:%s", r.keyPrefix, instanceID, groupID)
}

func (r *RedisRepository) phoneIdxKey(instanceID, phone string) string {
	return fmt.Sprintf("%s:status:%s:idx:phone:%s", r.keyPrefix, instanceID, phone)
}

func (r *RedisRepository) pendingKey(instanceID, messageID string) string {
	return fmt.Sprintf("%s:status:%s:pending:%s", r.keyPrefix, instanceID, messageID)
}

func (r *RedisRepository) entriesKey(instanceID string) string {
	return fmt.Sprintf("%s:status:%s:entries", r.keyPrefix, instanceID)
}

func (r *RedisRepository) pendingListKey(instanceID string) string {
	return fmt.Sprintf("%s:status:%s:pending:list", r.keyPrefix, instanceID)
}

// UpsertStatus creates or updates a status cache entry
func (r *RedisRepository) UpsertStatus(ctx context.Context, entry *StatusCacheEntry) error {
	if entry == nil {
		return errors.New("entry cannot be nil")
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	pipe := r.client.TxPipeline()

	// Store the main entry
	msgKey := r.msgKey(entry.InstanceID, entry.MessageID)
	pipe.Set(ctx, msgKey, data, r.ttl)

	// Add to instance entries set
	pipe.SAdd(ctx, r.entriesKey(entry.InstanceID), entry.MessageID)
	pipe.Expire(ctx, r.entriesKey(entry.InstanceID), r.ttl)

	// Add to group index if applicable
	if entry.IsGroup && entry.GroupID != "" {
		groupKey := r.groupIdxKey(entry.InstanceID, entry.GroupID)
		pipe.SAdd(ctx, groupKey, entry.MessageID)
		pipe.Expire(ctx, groupKey, r.ttl)
	}

	// Add to phone index
	if entry.Phone != "" {
		phoneKey := r.phoneIdxKey(entry.InstanceID, entry.Phone)
		pipe.SAdd(ctx, phoneKey, entry.MessageID)
		pipe.Expire(ctx, phoneKey, r.ttl)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	return nil
}

// Lua script for atomic participant status update
// This prevents race conditions by doing read-modify-write atomically in Redis
var addParticipantScript = redis.NewScript(`
local key = KEYS[1]
local participantJSON = ARGV[1]
local updatedAt = ARGV[2]
local defaultTTL = tonumber(ARGV[3])

-- Get existing entry
local data = redis.call('GET', key)
if not data then
    return redis.error_reply("entry not found")
end

-- Parse entry
local entry = cjson.decode(data)
if not entry then
    return redis.error_reply("failed to decode entry")
end

-- Parse participant
local participant = cjson.decode(participantJSON)
if not participant then
    return redis.error_reply("failed to decode participant")
end

-- Initialize participants map if nil
if not entry.participants then
    entry.participants = {}
end

-- Update participant
entry.participants[participant.phone] = participant
entry.updatedAt = tonumber(updatedAt)

-- Serialize back
local newData = cjson.encode(entry)

-- Get remaining TTL
local ttl = redis.call('TTL', key)
if ttl <= 0 then
    ttl = defaultTTL
end

-- Save back
redis.call('SETEX', key, ttl, newData)

return "OK"
`)

// AddParticipantStatus adds or updates a participant status in an existing entry
// Uses a Lua script to perform atomic read-modify-write operation
func (r *RedisRepository) AddParticipantStatus(ctx context.Context, instanceID, messageID string, participant *ParticipantStatus) error {
	if participant == nil {
		return errors.New("participant cannot be nil")
	}

	msgKey := r.msgKey(instanceID, messageID)

	participantJSON, err := json.Marshal(participant)
	if err != nil {
		return fmt.Errorf("failed to marshal participant: %w", err)
	}

	updatedAt := time.Now().UnixMilli()
	ttlSeconds := int64(r.ttl.Seconds())

	result, err := addParticipantScript.Run(ctx, r.client, []string{msgKey},
		string(participantJSON),
		updatedAt,
		ttlSeconds,
	).Result()

	if err != nil {
		if err.Error() == "entry not found" {
			return fmt.Errorf("entry not found: %s", messageID)
		}
		return fmt.Errorf("failed to update participant: %w", err)
	}

	if result != "OK" {
		return fmt.Errorf("unexpected result from script: %v", result)
	}

	return nil
}

// GetByMessageID retrieves a status entry by message ID
func (r *RedisRepository) GetByMessageID(ctx context.Context, instanceID, messageID string) (*StatusCacheEntry, error) {
	data, err := r.client.Get(ctx, r.msgKey(instanceID, messageID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	var entry StatusCacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	return &entry, nil
}

// GetByGroupID retrieves status entries by group ID
func (r *RedisRepository) GetByGroupID(ctx context.Context, instanceID, groupID string, limit, offset int) ([]*StatusCacheEntry, int64, error) {
	groupKey := r.groupIdxKey(instanceID, groupID)
	return r.getByIndex(ctx, instanceID, groupKey, limit, offset)
}

// GetByPhone retrieves status entries by phone
func (r *RedisRepository) GetByPhone(ctx context.Context, instanceID, phone string, limit, offset int) ([]*StatusCacheEntry, int64, error) {
	phoneKey := r.phoneIdxKey(instanceID, phone)
	return r.getByIndex(ctx, instanceID, phoneKey, limit, offset)
}

// GetAll retrieves all status entries for an instance with pagination
func (r *RedisRepository) GetAll(ctx context.Context, instanceID string, limit, offset int) ([]*StatusCacheEntry, int64, error) {
	entriesKey := r.entriesKey(instanceID)
	return r.getByIndex(ctx, instanceID, entriesKey, limit, offset)
}

// getByIndex retrieves entries from an index set
func (r *RedisRepository) getByIndex(ctx context.Context, instanceID, indexKey string, limit, offset int) ([]*StatusCacheEntry, int64, error) {
	// Get total count
	total, err := r.client.SCard(ctx, indexKey).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get index count: %w", err)
	}

	if total == 0 {
		return []*StatusCacheEntry{}, 0, nil
	}

	// Get all message IDs from the index
	messageIDs, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get index members: %w", err)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(messageIDs) {
		return []*StatusCacheEntry{}, total, nil
	}
	if end > len(messageIDs) {
		end = len(messageIDs)
	}
	paginatedIDs := messageIDs[start:end]

	// Fetch entries
	entries := make([]*StatusCacheEntry, 0, len(paginatedIDs))
	for _, msgID := range paginatedIDs {
		entry, err := r.GetByMessageID(ctx, instanceID, msgID)
		if err != nil {
			continue // Skip entries that fail to load
		}
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	return entries, total, nil
}

// StorePendingWebhook stores a pending webhook for later flush
func (r *RedisRepository) StorePendingWebhook(ctx context.Context, webhook *PendingWebhook) error {
	if webhook == nil {
		return errors.New("webhook cannot be nil")
	}

	data, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	pipe := r.client.TxPipeline()

	// Store the webhook in a list keyed by messageID
	pendingKey := r.pendingKey(webhook.InstanceID, webhook.MessageID)
	pipe.RPush(ctx, pendingKey, data)
	pipe.Expire(ctx, pendingKey, r.ttl)

	// Add messageID to the pending list for this instance
	listKey := r.pendingListKey(webhook.InstanceID)
	pipe.SAdd(ctx, listKey, webhook.MessageID)
	pipe.Expire(ctx, listKey, r.ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// GetPendingWebhooks retrieves pending webhooks for an instance
func (r *RedisRepository) GetPendingWebhooks(ctx context.Context, instanceID string, limit int) ([]*PendingWebhook, error) {
	listKey := r.pendingListKey(instanceID)

	// Get message IDs with pending webhooks
	messageIDs, err := r.client.SMembers(ctx, listKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending list: %w", err)
	}

	webhooks := make([]*PendingWebhook, 0)
	count := 0
	for _, msgID := range messageIDs {
		if limit > 0 && count >= limit {
			break
		}

		msgWebhooks, err := r.GetPendingWebhooksByMessageID(ctx, instanceID, msgID)
		if err != nil {
			continue
		}
		for _, wh := range msgWebhooks {
			if limit > 0 && count >= limit {
				break
			}
			webhooks = append(webhooks, wh)
			count++
		}
	}

	return webhooks, nil
}

// GetPendingWebhooksByMessageID retrieves pending webhooks for a specific message
func (r *RedisRepository) GetPendingWebhooksByMessageID(ctx context.Context, instanceID, messageID string) ([]*PendingWebhook, error) {
	pendingKey := r.pendingKey(instanceID, messageID)

	data, err := r.client.LRange(ctx, pendingKey, 0, -1).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pending webhooks: %w", err)
	}

	webhooks := make([]*PendingWebhook, 0, len(data))
	for _, d := range data {
		var webhook PendingWebhook
		if err := json.Unmarshal([]byte(d), &webhook); err != nil {
			continue
		}
		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

// DeletePendingWebhooks removes pending webhooks for specified message IDs
func (r *RedisRepository) DeletePendingWebhooks(ctx context.Context, instanceID string, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	pipe := r.client.TxPipeline()

	listKey := r.pendingListKey(instanceID)
	for _, msgID := range messageIDs {
		pipe.Del(ctx, r.pendingKey(instanceID, msgID))
		pipe.SRem(ctx, listKey, msgID)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// CountPendingWebhooks counts pending webhooks for an instance
func (r *RedisRepository) CountPendingWebhooks(ctx context.Context, instanceID string) (int64, error) {
	listKey := r.pendingListKey(instanceID)

	messageIDs, err := r.client.SMembers(ctx, listKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get pending list: %w", err)
	}

	var total int64
	for _, msgID := range messageIDs {
		count, err := r.client.LLen(ctx, r.pendingKey(instanceID, msgID)).Result()
		if err != nil {
			continue
		}
		total += count
	}

	return total, nil
}

// DeleteByMessageID removes a status entry and its indices
func (r *RedisRepository) DeleteByMessageID(ctx context.Context, instanceID, messageID string) error {
	// First get the entry to know which indices to clean
	entry, err := r.GetByMessageID(ctx, instanceID, messageID)
	if err != nil {
		return err
	}
	if entry == nil {
		return nil
	}

	pipe := r.client.TxPipeline()

	// Delete main entry
	pipe.Del(ctx, r.msgKey(instanceID, messageID))

	// Remove from entries set
	pipe.SRem(ctx, r.entriesKey(instanceID), messageID)

	// Remove from group index
	if entry.IsGroup && entry.GroupID != "" {
		pipe.SRem(ctx, r.groupIdxKey(instanceID, entry.GroupID), messageID)
	}

	// Remove from phone index
	if entry.Phone != "" {
		pipe.SRem(ctx, r.phoneIdxKey(instanceID, entry.Phone), messageID)
	}

	// Delete pending webhooks
	pipe.Del(ctx, r.pendingKey(instanceID, messageID))
	pipe.SRem(ctx, r.pendingListKey(instanceID), messageID)

	_, err = pipe.Exec(ctx)
	return err
}

// DeleteByInstanceID removes all entries for an instance
func (r *RedisRepository) DeleteByInstanceID(ctx context.Context, instanceID string) (int64, error) {
	entriesKey := r.entriesKey(instanceID)

	// Get all message IDs
	messageIDs, err := r.client.SMembers(ctx, entriesKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get entries: %w", err)
	}

	if len(messageIDs) == 0 {
		return 0, nil
	}

	// Delete each entry
	var deleted int64
	for _, msgID := range messageIDs {
		if err := r.DeleteByMessageID(ctx, instanceID, msgID); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

// DeleteExpired removes expired entries (handled by Redis TTL, but this cleans up orphaned indices)
func (r *RedisRepository) DeleteExpired(ctx context.Context, instanceID string) (int64, error) {
	entriesKey := r.entriesKey(instanceID)

	// Get all message IDs
	messageIDs, err := r.client.SMembers(ctx, entriesKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get entries: %w", err)
	}

	var cleaned int64
	for _, msgID := range messageIDs {
		// Check if the main key exists
		exists, err := r.client.Exists(ctx, r.msgKey(instanceID, msgID)).Result()
		if err != nil {
			continue
		}
		if exists == 0 {
			// Entry expired, clean up orphaned index reference
			r.client.SRem(ctx, entriesKey, msgID)
			cleaned++
		}
	}

	return cleaned, nil
}

// Count returns the number of cached entries for an instance
func (r *RedisRepository) Count(ctx context.Context, instanceID string) (int64, error) {
	return r.client.SCard(ctx, r.entriesKey(instanceID)).Result()
}

// GetStats returns statistics about the status cache
func (r *RedisRepository) GetStats(ctx context.Context, instanceID string) (*CacheStats, error) {
	stats := &CacheStats{
		EntriesByInstance: make(map[string]int64),
	}

	// Count entries
	count, err := r.Count(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	stats.TotalEntries = count
	stats.EntriesByInstance[instanceID] = count

	// Count pending webhooks
	pending, err := r.CountPendingWebhooks(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	stats.PendingWebhooks = pending

	return stats, nil
}

// Ping checks Redis connectivity
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
