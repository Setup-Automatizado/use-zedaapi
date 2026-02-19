package handlers

import (
	"testing"
	"time"
)

func TestValidateAndResolveSchedule(t *testing.T) {
	intPtr := func(v int) *int { return &v }
	strPtr := func(v string) *string { return &v }

	tests := []struct {
		name             string
		delayMessage     *int
		scheduledFor     *string
		wantErr          bool
		wantErrContains  string
		checkDelay       func(int64) bool // optional delay range check
		checkScheduledAt bool             // if true, verify scheduledAt is non-nil
	}{
		{
			name:         "valid scheduledFor in the future",
			scheduledFor: strPtr(time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339)),
			checkDelay: func(d int64) bool {
				// Should be approximately 10 minutes in ms
				return d > 590000 && d < 610000
			},
			checkScheduledAt: true,
		},
		{
			name:            "scheduledFor in the past",
			scheduledFor:    strPtr(time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)),
			wantErr:         true,
			wantErrContains: "must be in the future",
		},
		{
			name:            "scheduledFor invalid format",
			scheduledFor:    strPtr("not-a-timestamp"),
			wantErr:         true,
			wantErrContains: "must be a valid ISO 8601",
		},
		{
			name: "neither scheduledFor nor delayMessage - random 1-3s",
			checkDelay: func(d int64) bool {
				return d >= 1000 && d <= 3000
			},
			checkScheduledAt: true,
		},
		{
			name:         "delayMessage=10 seconds",
			delayMessage: intPtr(10),
			checkDelay: func(d int64) bool {
				return d == 10000
			},
			checkScheduledAt: true,
		},
		{
			name:         "both scheduledFor and delayMessage - scheduledFor wins",
			delayMessage: intPtr(5),
			scheduledFor: strPtr(time.Now().Add(30 * time.Second).UTC().Format(time.RFC3339)),
			checkDelay: func(d int64) bool {
				// Should be approximately 30s, not 5s
				return d > 28000 && d < 32000
			},
			checkScheduledAt: true,
		},
		{
			name:         "delayMessage < 1 rounds up to 1",
			delayMessage: intPtr(0),
			checkDelay: func(d int64) bool {
				return d == 1000
			},
			checkScheduledAt: true,
		},
		{
			name: "scheduledFor with timezone offset",
			scheduledFor: func() *string {
				// Create a future time in a fixed-offset timezone
				loc := time.FixedZone("TEST", 3*60*60) // +03:00
				s := time.Now().Add(5 * time.Minute).In(loc).Format(time.RFC3339)
				return &s
			}(),
			checkDelay: func(d int64) bool {
				return d > 0
			},
			checkScheduledAt: true,
		},
		{
			name:         "scheduledFor empty string - treated as nil",
			scheduledFor: strPtr(""),
			checkDelay: func(d int64) bool {
				return d >= 1000 && d <= 3000 // falls back to random 1-3s
			},
			checkScheduledAt: true,
		},
		{
			name:         "scheduledFor very close (2s in future) - works",
			scheduledFor: strPtr(time.Now().Add(2 * time.Second).UTC().Format(time.RFC3339)),
			checkDelay: func(d int64) bool {
				return d > 0 && d < 3000
			},
			checkScheduledAt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delayMs, scheduledAt, err := validateAndResolveSchedule(tt.delayMessage, tt.scheduledFor)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrContains != "" {
					if !contains(err.Error(), tt.wantErrContains) {
						t.Fatalf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkDelay != nil {
				if !tt.checkDelay(delayMs) {
					t.Errorf("delay check failed: got %d ms", delayMs)
				}
			}

			if tt.checkScheduledAt {
				if scheduledAt == nil {
					t.Fatal("expected scheduledAt to be non-nil")
				}
				// Verify it's valid RFC3339
				if _, err := time.Parse(time.RFC3339, *scheduledAt); err != nil {
					t.Errorf("scheduledAt is not valid RFC3339: %q, err: %v", *scheduledAt, err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
