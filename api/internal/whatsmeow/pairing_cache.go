package whatsmeow

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const pairingCodeTTL = 20 * time.Second

type PairingCodeType string

const (
	PairingCodeTypeQR    PairingCodeType = "qr"
	PairingCodeTypePhone PairingCodeType = "phone"
)

type CachedPairingCode struct {
	Code      string
	Type      PairingCodeType
	Phone     string
	CreatedAt time.Time
	ExpiresAt time.Time
	SessionID string
}

func (c *CachedPairingCode) IsValid() bool {
	return c != nil && time.Now().Before(c.ExpiresAt)
}

func (c *CachedPairingCode) RemainingTTL() time.Duration {
	if c == nil {
		return 0
	}
	remaining := time.Until(c.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

type pairingCache struct {
	mu    sync.RWMutex
	codes map[uuid.UUID]*CachedPairingCode
}

func newPairingCache() *pairingCache {
	return &pairingCache{
		codes: make(map[uuid.UUID]*CachedPairingCode),
	}
}

func (pc *pairingCache) Get(instanceID uuid.UUID, codeType PairingCodeType) (*CachedPairingCode, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	cached, exists := pc.codes[instanceID]
	if !exists || cached == nil {
		return nil, false
	}

	if cached.Type != codeType || !cached.IsValid() {
		return nil, false
	}

	return cached, true
}

func (pc *pairingCache) GetForPhone(instanceID uuid.UUID, phone string) (*CachedPairingCode, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	cached, exists := pc.codes[instanceID]
	if !exists || cached == nil {
		return nil, false
	}

	if cached.Type != PairingCodeTypePhone || cached.Phone != phone || !cached.IsValid() {
		return nil, false
	}

	return cached, true
}

func (pc *pairingCache) Set(instanceID uuid.UUID, code *CachedPairingCode) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.codes[instanceID] = code
}

func (pc *pairingCache) Invalidate(instanceID uuid.UUID) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.codes, instanceID)
}

func (pc *pairingCache) CleanupExpired() int {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now()
	count := 0
	for id, cached := range pc.codes {
		if cached == nil || now.After(cached.ExpiresAt) {
			delete(pc.codes, id)
			count++
		}
	}
	return count
}
