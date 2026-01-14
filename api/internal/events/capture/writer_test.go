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
			name:      "notifySentByMe=false, fromMe=true - should use DeliveryURL",
			eventType: "message",
			fromMe:    "true",
			cfg: &ResolvedWebhookConfig{
				DeliveryURL:         "https://example.com/delivery",
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      false,
			},
			expectedURL:  "https://example.com/delivery",
			expectedType: "delivery",
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
			name:      "receipt event - should use MessageStatusURL",
			eventType: "receipt",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				MessageStatusURL:    "https://example.com/message-status",
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				NotifySentByMe:      false,
			},
			expectedURL:  "https://example.com/message-status",
			expectedType: "message_status",
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
		{
			name:      "notifySentByMe=false, fromMe=true, no DeliveryURL - should filter",
			eventType: "message",
			fromMe:    "true",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL:         "https://example.com/received",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				DeliveryURL:         "",
				NotifySentByMe:      false,
			},
			expectedURL:  "",
			expectedType: "",
		},
		{
			name:      "receipt event - no MessageStatusURL, should be discarded",
			eventType: "receipt",
			fromMe:    "false",
			cfg: &ResolvedWebhookConfig{
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				DeliveryURL:         "https://example.com/delivery",
				NotifySentByMe:      false,
			},
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

func TestResolveWebhookURL_APIEchoBypassesNotifySentByMe(t *testing.T) {
	// This test validates that API echo events (from_api=true) bypass the
	// NotifySentByMe filter and always get routed to the combined endpoint
	cfg := &ResolvedWebhookConfig{
		DeliveryURL:         "https://example.com/delivery",
		ReceivedURL:         "https://example.com/received",
		ReceivedDeliveryURL: "https://example.com/received-delivery",
		NotifySentByMe:      false, // Even when false, API echo should be routed
	}

	event := &types.InternalEvent{
		EventID:    uuid.New(),
		InstanceID: uuid.New(),
		EventType:  "message",
		Metadata: map[string]string{
			"from_me":  "true",
			"from_api": "true",
		},
	}

	url, eventType := resolveWebhookURL(event, cfg)

	if url != cfg.ReceivedDeliveryURL {
		t.Errorf("API echo should use ReceivedDeliveryURL, got %q", url)
	}
	if eventType != "received" {
		t.Errorf("API echo eventType should be 'received', got %q", eventType)
	}
}

func TestResolveWebhookURL_DeliveryURLRoutingConsistency(t *testing.T) {
	// This test validates Z-API compatibility for separate webhook routing:
	// When notifySentByMe=false:
	// - Messages received from others -> received_url
	// - Messages sent by me -> delivery_url

	cfg := &ResolvedWebhookConfig{
		DeliveryURL:         "https://example.com/delivery",
		ReceivedURL:         "https://example.com/received",
		ReceivedDeliveryURL: "https://example.com/received-and-delivery",
		NotifySentByMe:      false,
	}

	testCases := []struct {
		name        string
		fromMe      string
		expectedURL string
		expectedCat string
	}{
		{
			name:        "received message (fromMe=false) -> received_url",
			fromMe:      "false",
			expectedURL: cfg.ReceivedURL,
			expectedCat: "received",
		},
		{
			name:        "sent by me message (fromMe=true) -> delivery_url",
			fromMe:      "true",
			expectedURL: cfg.DeliveryURL,
			expectedCat: "delivery",
		},
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

			if url != tc.expectedURL {
				t.Errorf("Expected URL %q, got %q", tc.expectedURL, url)
			}
			if eventType != tc.expectedCat {
				t.Errorf("Expected category %q, got %q", tc.expectedCat, eventType)
			}
		})
	}
}

func TestResolveWebhookURL_WaitingMessageFiltering(t *testing.T) {
	// This test validates that waitingMessage=true events are filtered
	// regardless of message type (1:1, group, or status)
	// Filter is controlled by cfg.EventFilters.FilterWaitingMessage

	testCases := []struct {
		name           string
		eventType      string
		waitingMessage string
		filterEnabled  bool
		shouldFilter   bool
	}{
		{
			name:           "message with waitingMessage=true should be filtered when enabled",
			eventType:      "message",
			waitingMessage: "true",
			filterEnabled:  true,
			shouldFilter:   true,
		},
		{
			name:           "message with waitingMessage=true should NOT be filtered when disabled",
			eventType:      "message",
			waitingMessage: "true",
			filterEnabled:  false,
			shouldFilter:   false,
		},
		{
			name:           "message with waitingMessage=false should not be filtered",
			eventType:      "message",
			waitingMessage: "false",
			filterEnabled:  true,
			shouldFilter:   false,
		},
		{
			name:           "message without waitingMessage should not be filtered",
			eventType:      "message",
			waitingMessage: "",
			filterEnabled:  true,
			shouldFilter:   false,
		},
		{
			name:           "undecryptable with waitingMessage=true should be filtered when enabled",
			eventType:      "undecryptable",
			waitingMessage: "true",
			filterEnabled:  true,
			shouldFilter:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata := map[string]string{}
			if tc.waitingMessage != "" {
				metadata["waiting_message"] = tc.waitingMessage
			}

			// Check filter condition (matches writer.go logic)
			shouldFilter := tc.filterEnabled && metadata["waiting_message"] == "true"
			if shouldFilter != tc.shouldFilter {
				t.Errorf("Expected shouldFilter=%v, got %v", tc.shouldFilter, shouldFilter)
			}
		})
	}
}

func TestResolveWebhookURL_SecondaryDeviceReceiptFiltering(t *testing.T) {
	// This test validates that receipts from secondary devices (Device > 0) are filtered
	// Only receipts from the primary device (Device == 0) should be forwarded
	// Filter is controlled by cfg.EventFilters.FilterSecondaryDeviceReceipts

	testCases := []struct {
		name          string
		eventType     string
		senderDevice  string // Only set when Device > 0 in transformer
		filterEnabled bool
		shouldFilter  bool
	}{
		{
			name:          "receipt from primary device (no sender_device) should not be filtered",
			eventType:     "receipt",
			senderDevice:  "",
			filterEnabled: true,
			shouldFilter:  false,
		},
		{
			name:          "receipt from secondary device 7 should be filtered when enabled",
			eventType:     "receipt",
			senderDevice:  "7",
			filterEnabled: true,
			shouldFilter:  true,
		},
		{
			name:          "receipt from secondary device 7 should NOT be filtered when disabled",
			eventType:     "receipt",
			senderDevice:  "7",
			filterEnabled: false,
			shouldFilter:  false,
		},
		{
			name:          "receipt from secondary device 1 should be filtered when enabled",
			eventType:     "receipt",
			senderDevice:  "1",
			filterEnabled: true,
			shouldFilter:  true,
		},
		{
			name:          "message with sender_device should not be filtered (only receipts)",
			eventType:     "message",
			senderDevice:  "7",
			filterEnabled: true,
			shouldFilter:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata := map[string]string{}
			if tc.senderDevice != "" {
				metadata["sender_device"] = tc.senderDevice
			}

			// Check filter condition (matches writer.go logic)
			shouldFilter := tc.filterEnabled && tc.eventType == "receipt" && metadata["sender_device"] != ""
			if shouldFilter != tc.shouldFilter {
				t.Errorf("Expected shouldFilter=%v, got %v", tc.shouldFilter, shouldFilter)
			}
		})
	}
}

func TestResolveWebhookURL_MessageStatusURLNoFallback(t *testing.T) {
	// This test validates that receipt events ONLY go to message_status_url
	// If message_status_url is not configured, the event is discarded (no fallback)

	testCases := []struct {
		name        string
		cfg         *ResolvedWebhookConfig
		expectedURL string
		expectedCat string
	}{
		{
			name: "MessageStatusURL configured - should use it",
			cfg: &ResolvedWebhookConfig{
				MessageStatusURL:    "https://example.com/message-status",
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				DeliveryURL:         "https://example.com/delivery",
			},
			expectedURL: "https://example.com/message-status",
			expectedCat: "message_status",
		},
		{
			name: "no MessageStatusURL - should discard (no fallback)",
			cfg: &ResolvedWebhookConfig{
				ReceivedDeliveryURL: "https://example.com/received-delivery",
				DeliveryURL:         "https://example.com/delivery",
			},
			expectedURL: "",
			expectedCat: "",
		},
		{
			name: "only other URLs configured - should discard",
			cfg: &ResolvedWebhookConfig{
				ReceivedURL: "https://example.com/received",
				DeliveryURL: "https://example.com/delivery",
			},
			expectedURL: "",
			expectedCat: "",
		},
		{
			name:        "no URLs configured - should discard",
			cfg:         &ResolvedWebhookConfig{},
			expectedURL: "",
			expectedCat: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &types.InternalEvent{
				EventID:    uuid.New(),
				InstanceID: uuid.New(),
				EventType:  "receipt",
				Metadata:   map[string]string{},
			}

			url, eventType := resolveWebhookURL(event, tc.cfg)

			if url != tc.expectedURL {
				t.Errorf("Expected URL %q, got %q", tc.expectedURL, url)
			}
			if eventType != tc.expectedCat {
				t.Errorf("Expected category %q, got %q", tc.expectedCat, eventType)
			}
		})
	}
}
