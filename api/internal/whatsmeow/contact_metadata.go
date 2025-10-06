package whatsmeow

import (
	"container/list"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"go.mau.fi/whatsmeow"
	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/types"
)

const (
	defaultCacheCapacity = 10048
	defaultNameTTL       = 12 * time.Hour
	defaultPhotoTTL      = 12 * time.Hour
	defaultErrorTTL      = 5 * time.Minute

	photoTypePreview = "preview"
	photoTypeImage   = "image"
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

type contactMetadataCache struct {
	client *whatsmeow.Client
	log    *slog.Logger

	capacity int
	nameTTL  time.Duration
	photoTTL time.Duration
	errorTTL time.Duration

	mu          sync.Mutex
	nameEntries map[string]*nameCacheEntry
	nameNodes   map[string]*list.Element
	nameOrder   *list.List

	photoEntries map[string]*photoCacheEntry
	photoNodes   map[string]*list.Element
	photoOrder   *list.List

	group singleflight.Group
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

func newContactMetadataCache(client *whatsmeow.Client, log *slog.Logger) *contactMetadataCache {
	return &contactMetadataCache{
		client:       client,
		log:          log,
		capacity:     defaultCacheCapacity,
		nameTTL:      defaultNameTTL,
		photoTTL:     defaultPhotoTTL,
		errorTTL:     defaultErrorTTL,
		nameEntries:  make(map[string]*nameCacheEntry),
		nameNodes:    make(map[string]*list.Element),
		nameOrder:    list.New(),
		photoEntries: make(map[string]*photoCacheEntry),
		photoNodes:   make(map[string]*list.Element),
		photoOrder:   list.New(),
	}
}

func (c *contactMetadataCache) ContactName(ctx context.Context, jid types.JID) string {
	if c == nil {
		return ""
	}

	key := jid.String()
	now := time.Now()
	if value, ok := c.getNameFromCache(key, now); ok {
		return value
	}

	valueAny, _, _ := c.group.Do("name:"+key, func() (interface{}, error) {
		if value, ok := c.getNameFromCache(key, time.Now()); ok {
			return value, nil
		}
		name := c.resolveContactName(ctx, jid)
		c.storeName(key, name, time.Now().Add(c.nameTTL))
		return name, nil
	})

	if valueAny == nil {
		return ""
	}

	return valueAny.(string)
}

func (c *contactMetadataCache) ContactPhoto(ctx context.Context, jid types.JID) string {
	if c == nil {
		return ""
	}

	return c.fetchContactPhoto(ctx, jid, photoTypePreview)
}

func (c *contactMetadataCache) ContactPhotoDetails(ctx context.Context, jid types.JID) eventctx.ContactPhotoDetails {
	var details eventctx.ContactPhotoDetails
	if c == nil {
		return details
	}

	key := jid.String()

	if preview := c.fetchContactPhoto(ctx, jid, photoTypePreview); preview != "" {
		details.PreviewURL = preview
		if entry := c.photoEntrySnapshot(photoCacheKey(key, photoTypePreview)); entry != nil {
			details.PreviewID = entry.id
		}
	}

	if full := c.fetchContactPhoto(ctx, jid, photoTypeImage); full != "" {
		details.FullURL = full
		if entry := c.photoEntrySnapshot(photoCacheKey(key, photoTypeImage)); entry != nil {
			details.FullID = entry.id
		}
	}

	return details
}

func (c *contactMetadataCache) fetchContactPhoto(ctx context.Context, jid types.JID, photoType string) string {
	key := jid.String()
	cacheKey := photoCacheKey(key, photoType)
	now := time.Now()
	if url, ok := c.getPhotoFromCache(cacheKey, now); ok {
		return url
	}

	result, _, _ := c.group.Do("photo:"+cacheKey, func() (interface{}, error) {
		if url, ok := c.getPhotoFromCache(cacheKey, time.Now()); ok {
			return url, nil
		}

		existing := c.photoEntrySnapshot(cacheKey)
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
		c.storePhoto(cacheKey, url, id, photoType, time.Now().Add(ttl))
		return url, nil
	})

	if result == nil {
		return ""
	}

	return result.(string)
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

func (c *contactMetadataCache) resolveContactName(ctx context.Context, jid types.JID) string {
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

	if jid.Server == types.GroupServer {
		if groupInfo, err := c.client.GetGroupInfo(jid); err == nil {
			if groupInfo.GroupName.Name != "" {
				return groupInfo.GroupName.Name
			}
		} else if c.log != nil {
			c.log.Debug("failed to fetch group name",
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

	params := &whatsmeow.GetProfilePictureParams{
		Preview:    requestedType == photoTypePreview,
		ExistingID: existingID,
	}
	if isCommunityJID(jid) {
		params.IsCommunity = true
	}

	info, err := c.client.GetProfilePictureInfo(jid, params)
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
