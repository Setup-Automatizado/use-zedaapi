package whatsmeow

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.mau.fi/whatsmeow"
	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/types"
)

const (
	defaultCacheCapacity = 50000
	defaultNameTTL       = 24 * time.Hour
	defaultPhotoTTL      = 24 * time.Hour
	defaultErrorTTL      = 24 * time.Hour
	photoTypePreview     = "preview"
	photoTypeImage       = "image"
)

type nameCacheEntry struct {
	value   string
	expires time.Time
}

type photoCacheEntry struct {
	url     string
	id      string
	typ     string
	expires time.Time
}

type fetchKind int

const (
	fetchKindName fetchKind = iota
	fetchKindPhoto
)

type fetchRequest struct {
	kind      fetchKind
	jid       types.JID
	photoType string
}

type contactMetadataCache struct {
	client      *whatsmeow.Client
	log         *slog.Logger
	instanceID  uuid.UUID
	capacity    int
	nameTTL     time.Duration
	photoTTL    time.Duration
	errorTTL    time.Duration
	mu          sync.Mutex
	nameEntries map[string]*nameCacheEntry
	nameNodes   map[string]*list.Element
	nameOrder   *list.List

	photoEntries map[string]*photoCacheEntry
	photoNodes   map[string]*list.Element
	photoOrder   *list.List

	redis             *redis.Client
	redisPrefix       string
	legacyRedisPrefix string

	fetchQueue   chan fetchRequest
	fetchWorkers int
	stopCh       chan struct{}
	wg           sync.WaitGroup
	inflight     sync.Map
	closeOnce    sync.Once
}

type redisNameEntry struct {
	Value      string `json:"value"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

type redisPhotoEntry struct {
	URL        string `json:"url"`
	ID         string `json:"id"`
	Type       string `json:"type"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

func photoCacheKey(jidKey, photoType string) string {
	if photoType == "" {
		photoType = photoTypePreview
	}
	return jidKey + "|" + photoType
}

var (
	_ eventctx.ContactMetadataProvider    = (*contactMetadataCache)(nil)
	_ eventctx.ContactPhotoDetailProvider = (*contactMetadataCache)(nil)
)

func newContactMetadataCache(
	client *whatsmeow.Client,
	log *slog.Logger,
	instanceID uuid.UUID,
	cfg ContactMetadataConfig,
	redisClient *redis.Client,
) *contactMetadataCache {
	if log == nil {
		log = slog.Default()
	}
	if cfg.CacheCapacity <= 0 {
		cfg.CacheCapacity = defaultCacheCapacity
	}
	if cfg.NameTTL <= 0 {
		cfg.NameTTL = defaultNameTTL
	}
	if cfg.PhotoTTL <= 0 {
		cfg.PhotoTTL = defaultPhotoTTL
	}
	if cfg.ErrorTTL <= 0 {
		cfg.ErrorTTL = defaultErrorTTL
	}
	if cfg.PrefetchWorkers <= 0 {
		cfg.PrefetchWorkers = 1
	}
	if cfg.FetchQueueSize <= 0 {
		cfg.FetchQueueSize = 1024
	}

	cache := &contactMetadataCache{
		client:            client,
		log:               log,
		instanceID:        instanceID,
		capacity:          cfg.CacheCapacity,
		nameTTL:           cfg.NameTTL,
		photoTTL:          cfg.PhotoTTL,
		errorTTL:          cfg.ErrorTTL,
		nameEntries:       make(map[string]*nameCacheEntry),
		nameNodes:         make(map[string]*list.Element),
		nameOrder:         list.New(),
		photoEntries:      make(map[string]*photoCacheEntry),
		photoNodes:        make(map[string]*list.Element),
		photoOrder:        list.New(),
		redis:             redisClient,
		redisPrefix:       fmt.Sprintf("zedaapi:instance:%s:contact", instanceID.String()),
		legacyRedisPrefix: fmt.Sprintf("contactmeta:%s", instanceID.String()),
		fetchQueue:        make(chan fetchRequest, cfg.FetchQueueSize),
		fetchWorkers:      cfg.PrefetchWorkers,
		stopCh:            make(chan struct{}),
	}

	cache.startWorkers()
	return cache
}

func (c *contactMetadataCache) ContactName(ctx context.Context, jid types.JID) string {
	if c == nil {
		return ""
	}

	if ctx == nil {
		ctx = context.Background()
	}

	key := jid.String()
	now := time.Now()
	if value, ok := c.getNameFromCache(key, now); ok && value != "" {
		return value
	}

	if value, ok := c.getNameFromRedis(ctx, jid, now); ok && value != "" {
		return value
	}

	c.requestNameFetch(jid)

	allowRemote := jid.Server == types.GroupServer || jid.Server == types.NewsletterServer
	name := strings.TrimSpace(c.resolveContactName(ctx, jid, allowRemote))
	if name == "" {
		return ""
	}

	fallback := sanitizeUserComponent(jid.User)
	if fallback == "" {
		fallback = jid.String()
	}

	ttl := c.nameTTL
	if allowRemote && name == fallback {
		ttl = c.errorTTL
	}
	if ttl <= 0 {
		ttl = defaultNameTTL
	}

	expires := now.Add(ttl)
	c.storeName(key, name, expires)
	c.storeNameRedis(ctx, jid, name, ttl)

	return name
}

func (c *contactMetadataCache) ContactPhoto(ctx context.Context, jid types.JID) string {
	if c == nil {
		return ""
	}

	return c.fetchContactPhoto(ctx, jid, photoTypeImage)
}

func (c *contactMetadataCache) ContactPhotoDetails(ctx context.Context, jid types.JID) eventctx.ContactPhotoDetails {
	var details eventctx.ContactPhotoDetails
	if c == nil {
		return details
	}

	key := jid.String()

	if full := c.fetchContactPhoto(ctx, jid, photoTypeImage); full != "" {
		details.FullURL = full
		if entry := c.photoEntrySnapshot(photoCacheKey(key, photoTypeImage)); entry != nil {
			details.FullID = entry.id
		}
	} else if entry := c.photoEntrySnapshot(photoCacheKey(key, photoTypePreview)); entry != nil {
		if entry.url != "" {
			details.PreviewURL = entry.url
		}
		if entry.id != "" {
			details.PreviewID = entry.id
		}
	}

	return details
}

func (c *contactMetadataCache) fetchContactPhoto(ctx context.Context, jid types.JID, photoType string) string {
	if ctx == nil {
		ctx = context.Background()
	}

	key := jid.String()
	cacheKey := photoCacheKey(key, photoType)
	now := time.Now()
	if url, ok := c.getPhotoFromCache(cacheKey, now); ok {
		return url
	}

	if entry, ok := c.getPhotoFromRedis(ctx, jid, photoType, now); ok {
		return entry
	}

	existing := c.photoEntrySnapshot(cacheKey)
	if photoType == photoTypePreview {
		if existing != nil && now.Before(existing.expires) {
			return existing.url
		}
		return ""
	}

	url, id, ttl, reuse := c.resolveContactPhoto(ctx, jid, existing, photoType)
	if reuse && existing != nil {
		if url == "" {
			url = existing.url
		}
		if id == "" {
			id = existing.id
		}
	}

	if ttl <= 0 {
		ttl = c.photoTTL
	}
	if ttl <= 0 {
		ttl = defaultPhotoTTL
	}

	expires := now.Add(ttl)
	c.storePhoto(cacheKey, url, id, photoType, expires)
	c.storePhotoRedis(ctx, jid, photoType, url, id, ttl)

	if url != "" {
		return url
	}
	if existing != nil && now.Before(existing.expires) {
		return existing.url
	}

	return ""
}

func (c *contactMetadataCache) startWorkers() {
	if c.fetchWorkers <= 0 {
		c.fetchWorkers = 1
	}
	for i := 0; i < c.fetchWorkers; i++ {
		c.wg.Add(1)
		go c.worker()
	}
}

func (c *contactMetadataCache) worker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopCh:
			return
		case req := <-c.fetchQueue:
			c.processFetchRequest(req)
		}
	}
}

func (c *contactMetadataCache) processFetchRequest(req fetchRequest) {
	switch req.kind {
	case fetchKindName:
		c.refreshName(req.jid)
	case fetchKindPhoto:
		c.refreshPhoto(req.jid, req.photoType)
	}
}

func (c *contactMetadataCache) requestNameFetch(jid types.JID) {
	if c == nil {
		return
	}
	key := c.inflightKey(fetchKindName, jid, "")
	if !c.markInflight(key) {
		return
	}

	select {
	case <-c.stopCh:
		c.inflight.Delete(key)
		return
	case c.fetchQueue <- fetchRequest{kind: fetchKindName, jid: jid}:
	default:
		c.inflight.Delete(key)
		if c.log != nil {
			c.log.Debug("contact metadata name fetch queue full",
				slog.String("jid", jid.String()))
		}
	}
}

func (c *contactMetadataCache) requestPhotoFetch(jid types.JID, photoType string) {
	if c == nil {
		return
	}
	if photoType == "" {
		photoType = photoTypeImage
	}
	if photoType == photoTypePreview {
		return
	}
	key := c.inflightKey(fetchKindPhoto, jid, photoType)
	if !c.markInflight(key) {
		return
	}

	select {
	case <-c.stopCh:
		c.inflight.Delete(key)
		return
	case c.fetchQueue <- fetchRequest{kind: fetchKindPhoto, jid: jid, photoType: photoType}:
	default:
		c.inflight.Delete(key)
		if c.log != nil {
			c.log.Debug("contact metadata photo fetch queue full",
				slog.String("jid", jid.String()),
				slog.String("photo_type", photoType))
		}
	}
}

func (c *contactMetadataCache) refreshName(jid types.JID) {
	key := c.inflightKey(fetchKindName, jid, "")
	defer c.inflight.Delete(key)

	name := c.resolveContactName(context.Background(), jid, true)
	ttl := c.nameTTL
	if name == "" {
		ttl = c.errorTTL
	}
	if ttl <= 0 {
		ttl = defaultNameTTL
	}

	expires := time.Now().Add(ttl)
	c.storeName(jid.String(), name, expires)
	c.storeNameRedis(context.Background(), jid, name, ttl)
}

func (c *contactMetadataCache) refreshPhoto(jid types.JID, photoType string) {
	if photoType == "" {
		photoType = photoTypeImage
	}
	key := c.inflightKey(fetchKindPhoto, jid, photoType)
	defer c.inflight.Delete(key)

	if photoType == photoTypePreview {
		return
	}

	cacheKey := photoCacheKey(jid.String(), photoType)
	existing := c.photoEntrySnapshot(cacheKey)

	url, id, ttl, reuse := c.resolveContactPhoto(context.Background(), jid, existing, photoType)
	if reuse && existing != nil {
		if url == "" {
			url = existing.url
		}
		if id == "" {
			id = existing.id
		}
	}
	if ttl <= 0 {
		ttl = c.photoTTL
	}
	if ttl <= 0 {
		ttl = defaultPhotoTTL
	}

	expires := time.Now().Add(ttl)
	c.storePhoto(cacheKey, url, id, photoType, expires)
	c.storePhotoRedis(context.Background(), jid, photoType, url, id, ttl)
}

func (c *contactMetadataCache) markInflight(key string) bool {
	_, loaded := c.inflight.LoadOrStore(key, struct{}{})
	return !loaded
}

func (c *contactMetadataCache) getNameFromRedis(ctx context.Context, jid types.JID, now time.Time) (string, bool) {
	if c.redis == nil {
		return "", false
	}

	type redisKey struct {
		value  string
		legacy bool
	}

	keys := []redisKey{{value: c.redisNameKey(jid)}}
	if legacy := c.redisLegacyNameKey(jid); legacy != "" {
		keys = append(keys, redisKey{value: legacy, legacy: true})
	}

	for _, key := range keys {
		if key.value == "" {
			continue
		}

		data, err := c.redis.Get(ctx, key.value).Result()
		if errors.Is(err, redis.Nil) || data == "" {
			continue
		}
		if err != nil {
			if c.log != nil {
				c.log.Debug("redis name lookup failed",
					slog.String("jid", jid.String()),
					slog.Bool("legacy_key", key.legacy),
					slog.String("error", err.Error()))
			}
			continue
		}

		var entry redisNameEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			if c.log != nil {
				c.log.Debug("redis name decode failed",
					slog.String("jid", jid.String()),
					slog.Bool("legacy_key", key.legacy),
					slog.String("error", err.Error()))
			}
			continue
		}

		ttl := c.nameTTL
		if entry.TTLSeconds > 0 {
			ttl = time.Duration(entry.TTLSeconds) * time.Second
		}
		if ttl <= 0 {
			ttl = defaultNameTTL
		}

		expires := now.Add(ttl)
		c.storeName(jid.String(), entry.Value, expires)

		if key.legacy {
			c.storeNameRedis(ctx, jid, entry.Value, ttl)
		}

		return entry.Value, true
	}

	return "", false
}

func (c *contactMetadataCache) storeNameRedis(ctx context.Context, jid types.JID, value string, ttl time.Duration) {
	if c.redis == nil {
		return
	}
	if ttl <= 0 {
		ttl = c.nameTTL
	}
	if ttl <= 0 {
		ttl = defaultNameTTL
	}

	entry := redisNameEntry{
		Value:      value,
		TTLSeconds: int64(ttl / time.Second),
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		if c.log != nil {
			c.log.Debug("marshal redis name entry failed",
				slog.String("jid", jid.String()),
				slog.String("error", err.Error()))
		}
		return
	}

	if err := c.redis.Set(ctx, c.redisNameKey(jid), payload, ttl).Err(); err != nil && c.log != nil {
		c.log.Debug("redis name set failed",
			slog.String("jid", jid.String()),
			slog.String("error", err.Error()))
	}
}

func (c *contactMetadataCache) getPhotoFromRedis(ctx context.Context, jid types.JID, photoType string, now time.Time) (string, bool) {
	if c.redis == nil {
		return "", false
	}

	type redisKey struct {
		value  string
		legacy bool
	}

	keys := []redisKey{{value: c.redisPhotoKey(jid, photoType)}}
	if legacy := c.redisLegacyPhotoKey(jid, photoType); legacy != "" {
		keys = append(keys, redisKey{value: legacy, legacy: true})
	}

	for _, key := range keys {
		if key.value == "" {
			continue
		}

		data, err := c.redis.Get(ctx, key.value).Result()
		if errors.Is(err, redis.Nil) || data == "" {
			continue
		}
		if err != nil {
			if c.log != nil {
				c.log.Debug("redis photo lookup failed",
					slog.String("jid", jid.String()),
					slog.String("photo_type", photoType),
					slog.Bool("legacy_key", key.legacy),
					slog.String("error", err.Error()))
			}
			continue
		}

		var entry redisPhotoEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			if c.log != nil {
				c.log.Debug("redis photo decode failed",
					slog.String("jid", jid.String()),
					slog.String("photo_type", photoType),
					slog.Bool("legacy_key", key.legacy),
					slog.String("error", err.Error()))
			}
			continue
		}

		ttl := c.photoTTL
		if entry.TTLSeconds > 0 {
			ttl = time.Duration(entry.TTLSeconds) * time.Second
		}
		if ttl <= 0 {
			ttl = defaultPhotoTTL
		}

		expires := now.Add(ttl)
		cacheKey := photoCacheKey(jid.String(), photoType)
		c.storePhoto(cacheKey, entry.URL, entry.ID, photoType, expires)

		if key.legacy {
			c.storePhotoRedis(ctx, jid, photoType, entry.URL, entry.ID, ttl)
		}

		return entry.URL, true
	}

	return "", false
}

func (c *contactMetadataCache) storePhotoRedis(ctx context.Context, jid types.JID, photoType, url, id string, ttl time.Duration) {
	if c.redis == nil {
		return
	}
	if photoType == photoTypePreview {
		return
	}

	if ttl <= 0 {
		ttl = c.photoTTL
	}
	if ttl <= 0 {
		ttl = defaultPhotoTTL
	}

	entry := redisPhotoEntry{
		URL:        url,
		ID:         id,
		Type:       photoType,
		TTLSeconds: int64(ttl / time.Second),
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		if c.log != nil {
			c.log.Debug("marshal redis photo entry failed",
				slog.String("jid", jid.String()),
				slog.String("photo_type", photoType),
				slog.String("error", err.Error()))
		}
		return
	}

	if err := c.redis.Set(ctx, c.redisPhotoKey(jid, photoType), payload, ttl).Err(); err != nil && c.log != nil {
		c.log.Debug("redis photo set failed",
			slog.String("jid", jid.String()),
			slog.String("photo_type", photoType),
			slog.String("error", err.Error()))
	}
}

func (c *contactMetadataCache) redisNameKey(jid types.JID) string {
	return fmt.Sprintf("%s:name:%s", c.redisPrefix, jid.String())
}

func (c *contactMetadataCache) redisLegacyNameKey(jid types.JID) string {
	if c.legacyRedisPrefix == "" {
		return ""
	}
	return fmt.Sprintf("%s:name:%s", c.legacyRedisPrefix, jid.String())
}

func (c *contactMetadataCache) redisPhotoKey(jid types.JID, photoType string) string {
	if photoType == "" {
		photoType = photoTypePreview
	}
	return fmt.Sprintf("%s:photo:%s:%s", c.redisPrefix, photoType, jid.String())
}

func (c *contactMetadataCache) redisLegacyPhotoKey(jid types.JID, photoType string) string {
	if c.legacyRedisPrefix == "" {
		return ""
	}
	if photoType == "" {
		photoType = photoTypePreview
	}
	return fmt.Sprintf("%s:photo:%s:%s", c.legacyRedisPrefix, photoType, jid.String())
}

func (c *contactMetadataCache) inflightKey(kind fetchKind, jid types.JID, photoType string) string {
	switch kind {
	case fetchKindName:
		return "name:" + jid.String()
	case fetchKindPhoto:
		if photoType == "" {
			photoType = photoTypePreview
		}
		return "photo:" + photoType + ":" + jid.String()
	default:
		return jid.String()
	}
}

func (c *contactMetadataCache) Close() {
	if c == nil {
		return
	}
	c.closeOnce.Do(func() {
		close(c.stopCh)
		c.wg.Wait()
	})
}

func (c *contactMetadataCache) getNameFromCache(key string, now time.Time) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.nameEntries[key]
	if !ok {
		return "", false
	}
	if now.After(entry.expires) {
		c.removeNameLocked(key)
		return "", false
	}
	c.touchNameLocked(key)
	return entry.value, true
}

func (c *contactMetadataCache) storeName(key, value string, expires time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.nameEntries[key]; ok {
		entry.value = value
		entry.expires = expires
		c.touchNameLocked(key)
		return
	}

	if len(c.nameEntries) >= c.capacity {
		c.evictOldestNameLocked()
	}

	c.nameEntries[key] = &nameCacheEntry{value: value, expires: expires}
	elem := c.nameOrder.PushFront(key)
	c.nameNodes[key] = elem
}

func (c *contactMetadataCache) getPhotoFromCache(key string, now time.Time) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.photoEntries[key]
	if !ok {
		return "", false
	}
	if now.After(entry.expires) {
		c.removePhotoLocked(key)
		return "", false
	}
	c.touchPhotoLocked(key)
	return entry.url, true
}

func (c *contactMetadataCache) storePhoto(key, url, id, photoType string, expires time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.photoEntries[key]; ok {
		entry.url = url
		entry.id = id
		entry.typ = photoType
		entry.expires = expires
		c.touchPhotoLocked(key)
		return
	}

	if len(c.photoEntries) >= c.capacity {
		c.evictOldestPhotoLocked()
	}

	c.photoEntries[key] = &photoCacheEntry{url: url, id: id, typ: photoType, expires: expires}
	elem := c.photoOrder.PushFront(key)
	c.photoNodes[key] = elem
}

func (c *contactMetadataCache) photoEntrySnapshot(key string) *photoCacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.photoEntries[key]
	if !ok {
		return nil
	}
	copy := *entry
	return &copy
}

func (c *contactMetadataCache) touchNameLocked(key string) {
	if node, ok := c.nameNodes[key]; ok {
		c.nameOrder.MoveToFront(node)
	}
}

func (c *contactMetadataCache) touchPhotoLocked(key string) {
	if node, ok := c.photoNodes[key]; ok {
		c.photoOrder.MoveToFront(node)
	}
}

func (c *contactMetadataCache) removeNameLocked(key string) {
	if node, ok := c.nameNodes[key]; ok {
		c.nameOrder.Remove(node)
		delete(c.nameNodes, key)
	}
	delete(c.nameEntries, key)
}

func (c *contactMetadataCache) removePhotoLocked(key string) {
	if node, ok := c.photoNodes[key]; ok {
		c.photoOrder.Remove(node)
		delete(c.photoNodes, key)
	}
	delete(c.photoEntries, key)
}

func (c *contactMetadataCache) evictOldestNameLocked() {
	elem := c.nameOrder.Back()
	if elem == nil {
		return
	}
	key := elem.Value.(string)
	c.nameOrder.Remove(elem)
	delete(c.nameNodes, key)
	delete(c.nameEntries, key)
}

func (c *contactMetadataCache) evictOldestPhotoLocked() {
	elem := c.photoOrder.Back()
	if elem == nil {
		return
	}
	key := elem.Value.(string)
	c.photoOrder.Remove(elem)
	delete(c.photoNodes, key)
	delete(c.photoEntries, key)
}

func (c *contactMetadataCache) resolveContactName(ctx context.Context, jid types.JID, allowRemote bool) string {
	if c.client.Store != nil && c.client.Store.Contacts != nil {
		if info, err := c.client.Store.Contacts.GetContact(ctx, jid); err == nil {
			if info.Found {
				if info.FullName != "" {
					return info.FullName
				}
				if info.FirstName != "" {
					return info.FirstName
				}
				if info.PushName != "" {
					return info.PushName
				}
			}
		}
	}

	if allowRemote && jid.Server == types.GroupServer {
		if groupInfo, err := c.client.GetGroupInfo(context.Background(), jid); err == nil {
			if groupInfo.GroupName.Name != "" {
				return groupInfo.GroupName.Name
			}
		} else if c.log != nil {
			c.log.Debug("failed to fetch group name",
				slog.String("jid", jid.String()),
				slog.String("error", err.Error()))
		}
	} else if allowRemote && jid.Server == types.NewsletterServer {
		if info, err := c.client.GetNewsletterInfo(context.Background(), jid); err == nil && info != nil {
			if name := strings.TrimSpace(info.ThreadMeta.Name.Text); name != "" {
				return name
			}
		} else if err != nil && c.log != nil {
			c.log.Debug("failed to fetch newsletter name",
				slog.String("jid", jid.String()),
				slog.String("error", err.Error()))
		}
	}

	sanitized := sanitizeUserComponent(jid.User)
	if sanitized == "" {
		return jid.String()
	}
	return sanitized
}

func (c *contactMetadataCache) resolveContactPhoto(ctx context.Context, jid types.JID, existing *photoCacheEntry, photoType string) (url, id string, ttl time.Duration, reuse bool) {
	existingID := ""
	existingURL := ""
	requestedType := photoType
	if requestedType == "" {
		requestedType = photoTypePreview
	}
	if existing != nil {
		existingID = existing.id
		existingURL = existing.url
	}

	if jid.Server == types.NewsletterServer {
		if metadata, err := c.client.GetNewsletterInfo(context.Background(), jid); err == nil && metadata != nil {
			if metadata.ThreadMeta.Picture != nil {
				if url := normalizeProfilePictureURL(metadata.ThreadMeta.Picture); url != "" {
					return url, metadata.ThreadMeta.Picture.ID, c.photoTTL, false
				}
			}
			if previewURL := normalizeProfilePictureURL(&metadata.ThreadMeta.Preview); previewURL != "" {
				return previewURL, metadata.ThreadMeta.Preview.ID, c.photoTTL, false
			}
		} else if err != nil && c.log != nil {
			c.log.Debug("failed to fetch newsletter photo",
				slog.String("jid", jid.String()),
				slog.String("error", err.Error()))
		}
	}

	params := &whatsmeow.GetProfilePictureParams{
		Preview:    requestedType == photoTypePreview,
		ExistingID: existingID,
	}
	if isCommunityJID(jid) {
		params.IsCommunity = true
	}

	info, err := c.client.GetProfilePictureInfo(context.Background(), jid, params)
	if err == nil {
		if info == nil {
			return existingURL, existingID, c.photoTTL, true
		}
		return info.URL, info.ID, c.photoTTL, false
	}

	switch {
	case errors.Is(err, whatsmeow.ErrProfilePictureNotSet):
		fallthrough
	case errors.Is(err, whatsmeow.ErrProfilePictureUnauthorized):
		return "", "", c.photoTTL, false
	default:
		if c.log != nil {
			c.log.Debug("failed to fetch profile photo",
				slog.String("jid", jid.String()),
				slog.String("error", err.Error()))
		}
		return existingURL, existingID, c.errorTTL, true
	}
}

func isCommunityJID(jid types.JID) bool {
	return jid.Server == types.HiddenUserServer && jid.Device == 0
}

func sanitizeUserComponent(user string) string {
	if idx := strings.IndexRune(user, ':'); idx >= 0 {
		user = user[:idx]
	}
	if idx := strings.IndexRune(user, '.'); idx >= 0 {
		user = user[:idx]
	}
	return user
}

func normalizeProfilePictureURL(info *types.ProfilePictureInfo) string {
	if info == nil {
		return ""
	}
	url := strings.TrimSpace(info.URL)
	if url != "" {
		return url
	}
	direct := strings.TrimSpace(info.DirectPath)
	if direct == "" {
		return ""
	}
	if strings.HasPrefix(direct, "http://") || strings.HasPrefix(direct, "https://") {
		return direct
	}
	if strings.HasPrefix(direct, "//") {
		return "https:" + direct
	}
	if !strings.HasPrefix(direct, "/") {
		direct = "/" + direct
	}
	base := "https://pps.whatsapp.net"
	lower := strings.ToLower(direct)
	if strings.HasPrefix(lower, "/m1/") || strings.HasPrefix(lower, "/m2/") || strings.HasPrefix(lower, "/m3/") || strings.HasPrefix(lower, "/m4/") || strings.HasPrefix(lower, "/mmg/") {
		base = "https://mmg.whatsapp.net"
	}
	return base + direct
}
