package statuscache

import (
	"log/slog"
	"os"
	"testing"

	"go.mau.fi/whatsmeow/api/internal/config"
)

// TestShouldCacheStatusType tests the case-insensitive status type matching
func TestShouldCacheStatusType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name       string
		configured []string
		status     string
		expected   bool
	}{
		{
			name:       "lowercase config, UPPERCASE status (READ)",
			configured: []string{"read", "delivered", "played", "sent"},
			status:     "READ",
			expected:   true,
		},
		{
			name:       "lowercase config, UPPERCASE status (PLAYED)",
			configured: []string{"read", "delivered", "played", "sent"},
			status:     "PLAYED",
			expected:   true,
		},
		{
			name:       "lowercase config, UPPERCASE status (SENT)",
			configured: []string{"read", "delivered", "played", "sent"},
			status:     "SENT",
			expected:   true,
		},
		{
			name:       "UPPERCASE config, lowercase status",
			configured: []string{"READ", "DELIVERED", "PLAYED", "SENT"},
			status:     "read",
			expected:   true,
		},
		{
			name:       "mixed case config",
			configured: []string{"Read", "Delivered"},
			status:     "READ",
			expected:   true,
		},
		{
			name:       "exact match lowercase",
			configured: []string{"read", "sent"},
			status:     "read",
			expected:   true,
		},
		{
			name:       "exact match UPPERCASE",
			configured: []string{"READ", "SENT"},
			status:     "READ",
			expected:   true,
		},
		{
			name:       "status not in list",
			configured: []string{"read", "sent"},
			status:     "PLAYED",
			expected:   false,
		},
		{
			name:       "empty config",
			configured: []string{},
			status:     "READ",
			expected:   false,
		},
		{
			name:       "RECEIVED status with delivered config",
			configured: []string{"delivered"},
			status:     "RECEIVED",
			expected:   false, // RECEIVED != delivered (different names)
		},
		{
			name:       "READ_BY_ME status with read config",
			configured: []string{"read"},
			status:     "READ_BY_ME",
			expected:   false, // READ_BY_ME != read (substring, not equal)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.StatusCache.Enabled = true
			cfg.StatusCache.Types = tt.configured

			interceptor := NewInterceptor(nil, cfg, logger)
			result := interceptor.shouldCacheStatusType(tt.status)

			if result != tt.expected {
				t.Errorf("shouldCacheStatusType(%q) with config %v = %v, want %v",
					tt.status, tt.configured, result, tt.expected)
			}
		})
	}
}

// TestShouldCacheScope tests the scope filtering (groups vs direct)
func TestShouldCacheScope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name     string
		scope    []string
		isGroup  bool
		expected bool
	}{
		{
			name:     "groups scope with group message",
			scope:    []string{"groups"},
			isGroup:  true,
			expected: true,
		},
		{
			name:     "groups scope with direct message",
			scope:    []string{"groups"},
			isGroup:  false,
			expected: false,
		},
		{
			name:     "direct scope with direct message",
			scope:    []string{"direct"},
			isGroup:  false,
			expected: true,
		},
		{
			name:     "direct scope with group message",
			scope:    []string{"direct"},
			isGroup:  true,
			expected: false,
		},
		{
			name:     "both scopes with group message",
			scope:    []string{"groups", "direct"},
			isGroup:  true,
			expected: true,
		},
		{
			name:     "both scopes with direct message",
			scope:    []string{"groups", "direct"},
			isGroup:  false,
			expected: true,
		},
		{
			name:     "empty scope",
			scope:    []string{},
			isGroup:  true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.StatusCache.Enabled = true
			cfg.StatusCache.Scope = tt.scope

			interceptor := NewInterceptor(nil, cfg, logger)
			result := interceptor.shouldCacheScope(tt.isGroup)

			if result != tt.expected {
				t.Errorf("shouldCacheScope(%v) with scope %v = %v, want %v",
					tt.isGroup, tt.scope, result, tt.expected)
			}
		})
	}
}

// TestExtractGroupID tests the group ID extraction from phone field
func TestExtractGroupID(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		isGroup  bool
		expected string
	}{
		{
			name:     "standard group JID",
			phone:    "120363182823169824@g.us",
			isGroup:  true,
			expected: "120363182823169824",
		},
		{
			name:     "group JID without suffix",
			phone:    "120363182823169824",
			isGroup:  true,
			expected: "120363182823169824",
		},
		{
			name:     "direct chat (not group)",
			phone:    "5511999999999@s.whatsapp.net",
			isGroup:  false,
			expected: "",
		},
		{
			name:     "empty phone",
			phone:    "",
			isGroup:  true,
			expected: "",
		},
		{
			name:     "group with timestamp suffix",
			phone:    "120363182823169824-1234567890@g.us",
			isGroup:  true,
			expected: "120363182823169824",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGroupID(tt.phone, tt.isGroup)
			if result != tt.expected {
				t.Errorf("extractGroupID(%q, %v) = %q, want %q",
					tt.phone, tt.isGroup, result, tt.expected)
			}
		})
	}
}
