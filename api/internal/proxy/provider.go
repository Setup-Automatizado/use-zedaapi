package proxy

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProviderType identifies a proxy provider implementation.
type ProviderType string

const (
	ProviderTypeWebshare ProviderType = "webshare"
	ProviderTypeCustom   ProviderType = "custom"
)

// ErrReplaceNotSupported is returned when a provider does not support proxy replacement.
var ErrReplaceNotSupported = errors.New("provider does not support proxy replacement")

// ProxyProvider defines the interface for external proxy sources.
type ProxyProvider interface {
	// Type returns the provider type identifier.
	Type() ProviderType

	// ListProxies returns proxies matching the filter with pagination.
	ListProxies(ctx context.Context, filter ProxyFilter) ([]ProxyEntry, int, error)

	// ValidateProxy checks if a proxy is still valid via the provider API (not connectivity).
	ValidateProxy(ctx context.Context, externalID string) (bool, error)

	// ReplaceProxy requests the provider to replace a failed proxy.
	// Returns ErrReplaceNotSupported if not supported.
	ReplaceProxy(ctx context.Context, externalID string) (*ReplaceResult, error)

	// GetInfo returns provider metadata and capabilities.
	GetInfo(ctx context.Context) (*ProviderInfo, error)

	// SyncAll fetches ALL proxies from the provider for full pool sync.
	SyncAll(ctx context.Context) ([]ProxyEntry, error)

	// Close releases any resources held by the provider.
	Close() error
}

// ProxyFilter configures proxy listing criteria.
type ProxyFilter struct {
	CountryCodes []string
	Page         int
	PageSize     int
}

// ProxyEntry represents a single proxy from a provider.
type ProxyEntry struct {
	ExternalID   string
	ProxyURL     string
	CountryCode  string
	City         string
	Username     string
	Password     string
	Address      string
	Port         int
	Valid        bool
	LastVerified time.Time
}

// ReplaceResult holds the outcome of a proxy replacement request.
type ReplaceResult struct {
	OldExternalID string
	NewProxy      *ProxyEntry
	Success       bool
	Message       string
}

// ProviderInfo describes a provider's current state and capabilities.
type ProviderInfo struct {
	Type             ProviderType
	Name             string
	TotalProxies     int
	AvailableProxies int
	SupportsReplace  bool
	RateLimitRPM     int
	CountryCodes     []string
}

// RegistrySwapper abstracts whatsmeow.ClientRegistry for proxy operations.
type RegistrySwapper interface {
	// ApplyProxy sets the proxy on an existing client (no reconnect).
	ApplyProxy(ctx context.Context, instanceID uuid.UUID, proxyURL string, noWebsocket, onlyLogin, noMedia bool) error
	// SwapProxy hot-swaps the proxy on an active connection (disconnect + reconnect).
	SwapProxy(ctx context.Context, instanceID uuid.UUID, proxyURL string, noWebsocket, onlyLogin, noMedia bool) error
}

// InstanceProxyUpdater abstracts instance repository operations for proxy config.
type InstanceProxyUpdater interface {
	UpdateProxyURL(ctx context.Context, id uuid.UUID, proxyURL *string, enabled bool) error
	ClearProxyConfig(ctx context.Context, id uuid.UUID) error
}

// AssignOptions configures proxy assignment behavior.
type AssignOptions struct {
	ProviderID   *uuid.UUID
	CountryCodes []string
	GroupID      *uuid.UUID
	NoWebsocket  bool
	OnlyLogin    bool
	NoMedia      bool
}

// InstanceQueuePauser allows the proxy system to pause/resume message queue processing
// for specific instances during proxy failures and hot-swaps.
type InstanceQueuePauser interface {
	PauseInstance(ctx context.Context, instanceID uuid.UUID, reason string) error
	ResumeInstance(ctx context.Context, instanceID uuid.UUID) error
	IsInstancePaused(instanceID uuid.UUID) bool
}
