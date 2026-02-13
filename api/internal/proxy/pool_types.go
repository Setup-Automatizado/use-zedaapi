package proxy

import (
	"time"

	"github.com/google/uuid"
)

// ProviderRecord represents a row in proxy_providers table.
type ProviderRecord struct {
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	ProviderType         string     `json:"providerType"`
	Enabled              bool       `json:"enabled"`
	Priority             int        `json:"priority"`
	APIKey               *string    `json:"-"` // Never expose in JSON
	APIEndpoint          *string    `json:"apiEndpoint,omitempty"`
	MaxProxies           int        `json:"maxProxies"`
	MaxInstancesPerProxy int        `json:"maxInstancesPerProxy"`
	CountryCodes         []string   `json:"countryCodes"`
	RateLimitRPM         int        `json:"rateLimitRpm"`
	LastSyncAt           *time.Time `json:"lastSyncAt,omitempty"`
	SyncError            *string    `json:"syncError,omitempty"`
	ProxyCount           int        `json:"proxyCount"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

// PoolProxyRecord represents a row in proxy_pool table.
type PoolProxyRecord struct {
	ID              uuid.UUID  `json:"id"`
	ProviderID      uuid.UUID  `json:"providerId"`
	ExternalID      *string    `json:"externalId,omitempty"`
	ProxyURL        string     `json:"proxyUrl"`
	CountryCode     *string    `json:"countryCode,omitempty"`
	City            *string    `json:"city,omitempty"`
	Status          string     `json:"status"`
	HealthStatus    string     `json:"healthStatus"`
	HealthFailures  int        `json:"healthFailures"`
	LastHealthCheck *time.Time `json:"lastHealthCheck,omitempty"`
	AssignedCount   int        `json:"assignedCount"`
	MaxAssignments  int        `json:"maxAssignments"`
	Valid           bool       `json:"valid"`
	LastVerifiedAt  *time.Time `json:"lastVerifiedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// PoolProxyStatus constants.
const (
	PoolProxyStatusAvailable = "available"
	PoolProxyStatusAssigned  = "assigned"
	PoolProxyStatusUnhealthy = "unhealthy"
	PoolProxyStatusRetired   = "retired"
)

// AssignmentRecord represents a row in proxy_assignments table.
type AssignmentRecord struct {
	ID            uuid.UUID  `json:"id"`
	PoolProxyID   uuid.UUID  `json:"poolProxyId"`
	InstanceID    uuid.UUID  `json:"instanceId"`
	GroupID       *uuid.UUID `json:"groupId,omitempty"`
	Status        string     `json:"status"`
	AssignedAt    time.Time  `json:"assignedAt"`
	ReleasedAt    *time.Time `json:"releasedAt,omitempty"`
	AssignedBy    string     `json:"assignedBy"`
	ReleaseReason *string    `json:"releaseReason,omitempty"`
	ProxyURL      string     `json:"proxyUrl,omitempty"`
}

// AssignmentStatus constants.
const (
	AssignmentStatusActive      = "active"
	AssignmentStatusPendingSwap = "pending_swap"
	AssignmentStatusInactive    = "inactive"
)

// GroupRecord represents a row in proxy_groups table.
type GroupRecord struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	ProviderID   *uuid.UUID `json:"providerId,omitempty"`
	PoolProxyID  *uuid.UUID `json:"poolProxyId,omitempty"`
	MaxInstances int        `json:"maxInstances"`
	CountryCode  *string    `json:"countryCode,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// PoolStats aggregates pool health and assignment statistics.
type PoolStats struct {
	TotalProxies     int            `json:"totalProxies"`
	AvailableProxies int            `json:"availableProxies"`
	AssignedProxies  int            `json:"assignedProxies"`
	UnhealthyProxies int            `json:"unhealthyProxies"`
	RetiredProxies   int            `json:"retiredProxies"`
	TotalAssignments int            `json:"totalAssignments"`
	ByProvider       []ProviderStat `json:"byProvider"`
}

// ProviderStat holds per-provider statistics.
type ProviderStat struct {
	ProviderID   uuid.UUID `json:"providerId"`
	ProviderName string    `json:"providerName"`
	Total        int       `json:"total"`
	Available    int       `json:"available"`
	Assigned     int       `json:"assigned"`
	Unhealthy    int       `json:"unhealthy"`
}

// CreateProviderRequest is the input for creating a new provider.
type CreateProviderRequest struct {
	Name                 string   `json:"name"`
	ProviderType         string   `json:"providerType"`
	Enabled              bool     `json:"enabled"`
	Priority             int      `json:"priority"`
	APIKey               string   `json:"apiKey"`
	APIEndpoint          string   `json:"apiEndpoint"`
	MaxProxies           int      `json:"maxProxies"`
	MaxInstancesPerProxy int      `json:"maxInstancesPerProxy"`
	CountryCodes         []string `json:"countryCodes"`
	RateLimitRPM         int      `json:"rateLimitRpm"`
}

// UpdateProviderRequest is the input for updating a provider.
type UpdateProviderRequest struct {
	Name                 *string  `json:"name,omitempty"`
	Enabled              *bool    `json:"enabled,omitempty"`
	Priority             *int     `json:"priority,omitempty"`
	APIKey               *string  `json:"apiKey,omitempty"`
	APIEndpoint          *string  `json:"apiEndpoint,omitempty"`
	MaxProxies           *int     `json:"maxProxies,omitempty"`
	MaxInstancesPerProxy *int     `json:"maxInstancesPerProxy,omitempty"`
	CountryCodes         []string `json:"countryCodes,omitempty"`
	RateLimitRPM         *int     `json:"rateLimitRpm,omitempty"`
}

// AssignPoolProxyRequest is the input for assigning a pool proxy to an instance.
type AssignPoolProxyRequest struct {
	ProviderID   *uuid.UUID `json:"providerId,omitempty"`
	CountryCodes []string   `json:"countryCodes,omitempty"`
	NoWebsocket  bool       `json:"noWebsocket,omitempty"`
	OnlyLogin    bool       `json:"onlyLogin,omitempty"`
	NoMedia      bool       `json:"noMedia,omitempty"`
}

// AssignGroupRequest is the input for assigning an instance to a proxy group.
type AssignGroupRequest struct {
	GroupID uuid.UUID `json:"groupId"`
}

// BulkAssignRequest is the input for bulk-assigning pool proxies to multiple instances.
type BulkAssignRequest struct {
	InstanceIDs  []uuid.UUID `json:"instanceIds,omitempty"` // If empty, assigns to all unassigned active instances
	ProviderID   *uuid.UUID  `json:"providerId,omitempty"`
	CountryCodes []string    `json:"countryCodes,omitempty"`
}

// BulkAssignResult reports the outcome of a bulk assignment operation.
type BulkAssignResult struct {
	Total    int      `json:"total"`
	Assigned int      `json:"assigned"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}
