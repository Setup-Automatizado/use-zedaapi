package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildSendExtra(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		msgID  string
		wantID string
	}{
		{"with pre-generated ID", "3EB0ABC123DEF456789012", "3EB0ABC123DEF456789012"},
		{"with empty ID", "", ""},
		{"with zero value (not set)", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			args := &SendMessageArgs{WhatsAppMessageID: tt.msgID}
			extra := BuildSendExtra(args)
			if extra.ID != tt.wantID {
				t.Errorf("BuildSendExtra().ID = %q, want %q", extra.ID, tt.wantID)
			}
		})
	}
}

func TestSendMessageArgsJSONRoundTrip(t *testing.T) {
	t.Parallel()
	original := SendMessageArgs{
		ZaapID:            "test-zaap-id",
		InstanceID:        uuid.New(),
		Phone:             "5511999999999@s.whatsapp.net",
		MessageType:       MessageTypeText,
		WhatsAppMessageID: "3EB0ABCDEF123456789012",
		TextContent:       &TextMessage{Message: "hello"},
		EnqueuedAt:        time.Now().Truncate(time.Microsecond),
		ScheduledFor:      time.Now().Add(time.Minute).Truncate(time.Microsecond),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded SendMessageArgs
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.WhatsAppMessageID != original.WhatsAppMessageID {
		t.Errorf("WhatsAppMessageID = %q, want %q", decoded.WhatsAppMessageID, original.WhatsAppMessageID)
	}
	if decoded.ZaapID != original.ZaapID {
		t.Errorf("ZaapID = %q, want %q", decoded.ZaapID, original.ZaapID)
	}
	if decoded.InstanceID != original.InstanceID {
		t.Errorf("InstanceID = %v, want %v", decoded.InstanceID, original.InstanceID)
	}
	if decoded.MessageType != original.MessageType {
		t.Errorf("MessageType = %q, want %q", decoded.MessageType, original.MessageType)
	}
}

func TestSendMessageArgsJSONRoundTripWithoutWhatsAppID(t *testing.T) {
	t.Parallel()
	original := SendMessageArgs{
		ZaapID:       "test-zaap-id",
		InstanceID:   uuid.New(),
		Phone:        "5511999999999@s.whatsapp.net",
		MessageType:  MessageTypeText,
		TextContent:  &TextMessage{Message: "hello"},
		EnqueuedAt:   time.Now().Truncate(time.Microsecond),
		ScheduledFor: time.Now().Add(time.Minute).Truncate(time.Microsecond),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded SendMessageArgs
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.WhatsAppMessageID != "" {
		t.Errorf("WhatsAppMessageID = %q, want empty", decoded.WhatsAppMessageID)
	}
}
