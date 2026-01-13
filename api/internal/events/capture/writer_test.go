package capture

import (
	"testing"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/events/types"
)

func TestResolveWebhookURL(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		fromMe       string
		cfg          *ResolvedWebhookConfig
		expectedURL  string
		expectedType string
	}{
		{
			name:      "notifySentByMe=false, fromMe=false - should use ReceivedURL",
			eventType: "message",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      false,
			},
			expectedURL:  "https://example.com/received",
			expectedType: "received",
		},
		{
			name:      "notifySentByMe=false, fromMe=true - should filter (empty URL)",
			eventType: "message",
			fromMe:    "true",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      false,
			},
			expectedURL:  "",
			expectedType: "",
		},
		{
			name:      "notifySentByMe=true, fromMe=false - should use ReceivedDeliveryURL",
			eventType: "message",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      true,
			},
			expectedURL:  "https://example.com/received-delivery",
			expectedType: "received",
		},
		{
			name:      "notifySentByMe=true, fromMe=true - should use ReceivedDeliveryURL",
			eventType: "message",
			fromMe:    "true",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      true,
			},
			expectedURL:  "https://example.com/received-delivery",
			expectedType: "received",
		},
		{
			name:      "notifySentByMe=true, ReceivedDeliveryURL empty - should fallback to ReceivedURL",
			eventType: "message",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "",
				NotifySentByMe:      true,
			},
			expectedURL:  "https://example.com/received",
			expectedType: "received",
		},
		{
			name:      "receipt event - should use ReceivedDeliveryURL",
			eventType: "receipt",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      false,
			},
			expectedURL:  "https://example.com/received-delivery",
			expectedType: "receipt",
		},
		{
			name:      "connected event - should use ConnectedURL",
			eventType: "connected",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ConnectedURL: "https://example.com/connected",
			},
			expectedURL:  "https://example.com/connected",
			expectedType: "connected",
		},
		{
			name:      "disconnected event - should use DisconnectedURL",
			eventType: "disconnected",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				DisconnectedURL: "https://example.com/disconnected",
			},
			expectedURL:  "https://example.com/disconnected",
			expectedType: "disconnected",
		},
		{
			name:      "chat_presence event - should use ChatPresenceURL",
			eventType: "chat_presence",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ChatPresenceURL: "https://example.com/chat-presence",
			},
			expectedURL:  "https://example.com/chat-presence",
			expectedType: "chat_presence",
		},
		{
			name:      "presence event - should use ChatPresenceURL",
			eventType: "presence",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ChatPresenceURL: "https://example.com/chat-presence",
			},
			expectedURL:  "https://example.com/chat-presence",
			expectedType: "presence",
		},
		{
			name:      "group_info event - should use ReceivedURL",
			eventType: "group_info",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL: "https://example.com/received",
			},
			expectedURL:  "https://example.com/received",
			expectedType: "received",
		},
		{
			name:      "group_joined event - should use ReceivedURL",
			eventType: "group_joined",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL: "https://example.com/received",
			},
			expectedURL:  "https://example.com/received",
			expectedType: "received",
		},
		{
			name:      "undecryptable event - should use ReceivedURL",
			eventType: "undecryptable",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL: "https://example.com/received",
			},
			expectedURL:  "https://example.com/received",
			expectedType: "received",
		},
		{
			name:      "picture event - should use ReceivedURL",
			eventType: "picture",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL: "https://example.com/received",
			},
			expectedURL:  "https://example.com/received",
			expectedType: "received",
		},
		{
			name:         "unknown event type - should return empty",
			eventType:    "unknown_event",
			fromMe:       "false",
			cfg:          &ResolvedWebhookConfig{},
			expectedURL:  "",
			expectedType: "",
		},
		{
			name:         "nil config - should return empty",
			eventType:    "message",
			fromMe:       "false",
			cfg:          nil,
			expectedURL:  "",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &types.InternalEvent{
				EventID:    uuid.New(),
				InstanceID: uuid.New(),
				EventType:  tt.eventType,
				Metadata: map[string]string{
					"from_me": tt.fromMe,
				},
			}

			url, eventType := resolveWebhookURL(event, tt.cfg)

			if url != tt.expectedURL {
				t.Errorf("resolveWebhookURL() URL = %q, want %q", url, tt.expectedURL)
			}
			if eventType != tt.expectedType {
				t.Errorf("resolveWebhookURL() type = %q, want %q", eventType, tt.expectedType)
			}
		})
	}
}

func TestResolveWebhookURL_NotifySentByMeRoutingConsistency(t *testing.T) {
	// This test specifically validates Z-API compatibility:
	// When notifySentByMe=true, ALL messages (received + sent by me) should go to
	// receivedAndDeliveryCallbackUrl (ReceivedDeliveryURL)

	cfg := &ResolvedWebhookConfig{
		ReceivedURL:         "https://example.com/received",
		ReceivedDeliveryURL: "https://example.com/received-and-delivery",
		NotifySentByMe:      true,
	}

	testCases := []struct {
		name   string
		fromMe string
	}{
		{"received message (fromMe=false)", "false"},
		{"sent by me message (fromMe=true)", "true"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &types.InternalEvent{
				EventID:    uuid.New(),
				InstanceID: uuid.New(),
				EventType:  "message",
				Metadata: map[string]string{
					"from_me": tc.fromMe,
				},
			}

			url, eventType := resolveWebhookURL(event, cfg)

			// Both should go to ReceivedDeliveryURL when notifySentByMe=true
			if url != cfg.ReceivedDeliveryURL {
				t.Errorf("Expected URL %q for %s, got %q", cfg.ReceivedDeliveryURL, tc.name, url)
			}
			if eventType != "received" {
				t.Errorf("Expected eventType 'received', got %q", eventType)
			}
		})
	}
}
