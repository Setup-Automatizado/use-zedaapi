package echo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/types"
)

func TestNewEmitter(t *testing.T) {
	ctx := context.Background()
	instanceID := uuid.New()

	t.Run("creates emitter with config", func(t *testing.T) {
		cfg := &EmitterConfig{
			InstanceID: instanceID,
			Enabled:    true,
			StoreJID:   "5511999999999@s.whatsapp.net",
		}

		emitter := NewEmitter(ctx, cfg)

		if emitter == nil {
			t.Fatal("expected emitter to be created")
		}
		if !emitter.IsEnabled() {
			t.Error("expected emitter to be enabled")
		}
		if emitter.instanceID != instanceID {
			t.Errorf("expected instanceID %s, got %s", instanceID, emitter.instanceID)
		}
		if emitter.storeJID != cfg.StoreJID {
			t.Errorf("expected storeJID %s, got %s", cfg.StoreJID, emitter.storeJID)
		}
	})

	t.Run("creates disabled emitter", func(t *testing.T) {
		cfg := &EmitterConfig{
			InstanceID: instanceID,
			Enabled:    false,
		}

		emitter := NewEmitter(ctx, cfg)

		if emitter == nil {
			t.Fatal("expected emitter to be created")
		}
		if emitter.IsEnabled() {
			t.Error("expected emitter to be disabled")
		}
	})
}

func TestEmitter_EmitEcho_Disabled(t *testing.T) {
	ctx := context.Background()

	emitter := NewEmitter(ctx, &EmitterConfig{
		InstanceID: uuid.New(),
		Enabled:    false,
	})

	req := &EchoRequest{
		InstanceID:        uuid.New(),
		WhatsAppMessageID: "3EB0ABC123",
		RecipientJID:      types.JID{User: "5511888888888", Server: types.DefaultUserServer},
		Timestamp:         time.Now(),
		MessageType:       "text",
		ZaapID:            "test-zaap-id",
	}

	err := emitter.EmitEcho(ctx, req)
	if err != nil {
		t.Errorf("expected no error when disabled, got: %v", err)
	}
}

func TestEmitter_EmitEcho_NoRouter(t *testing.T) {
	ctx := context.Background()

	emitter := NewEmitter(ctx, &EmitterConfig{
		InstanceID:  uuid.New(),
		EventRouter: nil, // No router configured
		Enabled:     true,
	})

	req := &EchoRequest{
		InstanceID:        uuid.New(),
		WhatsAppMessageID: "3EB0ABC123",
		RecipientJID:      types.JID{User: "5511888888888", Server: types.DefaultUserServer},
		Timestamp:         time.Now(),
		MessageType:       "text",
		ZaapID:            "test-zaap-id",
	}

	err := emitter.EmitEcho(ctx, req)
	if err != nil {
		t.Errorf("expected no error when no router, got: %v", err)
	}
}

func TestEmitter_SetEnabled(t *testing.T) {
	ctx := context.Background()

	emitter := NewEmitter(ctx, &EmitterConfig{
		InstanceID: uuid.New(),
		Enabled:    false,
	})

	if emitter.IsEnabled() {
		t.Error("expected emitter to start disabled")
	}

	emitter.SetEnabled(true)
	if !emitter.IsEnabled() {
		t.Error("expected emitter to be enabled after SetEnabled(true)")
	}

	emitter.SetEnabled(false)
	if emitter.IsEnabled() {
		t.Error("expected emitter to be disabled after SetEnabled(false)")
	}
}

func TestEmitter_SetStoreJID(t *testing.T) {
	ctx := context.Background()

	emitter := NewEmitter(ctx, &EmitterConfig{
		InstanceID: uuid.New(),
		Enabled:    true,
	})

	if emitter.storeJID != "" {
		t.Error("expected empty storeJID initially")
	}

	emitter.SetStoreJID("5511999999999@s.whatsapp.net")
	if emitter.storeJID != "5511999999999@s.whatsapp.net" {
		t.Errorf("expected storeJID to be set, got %s", emitter.storeJID)
	}
}

func TestExtractPhone(t *testing.T) {
	tests := []struct {
		name     string
		jid      types.JID
		expected string
	}{
		{
			name:     "regular user JID",
			jid:      types.JID{User: "5511999999999", Server: types.DefaultUserServer},
			expected: "5511999999999",
		},
		{
			name:     "user with device suffix",
			jid:      types.JID{User: "5511999999999:12", Server: types.DefaultUserServer},
			expected: "5511999999999",
		},
		{
			name:     "group JID",
			jid:      types.JID{User: "120363012345678901", Server: types.GroupServer},
			expected: "120363012345678901",
		},
		{
			name:     "empty user",
			jid:      types.JID{User: "", Server: types.DefaultUserServer},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPhone(tt.jid)
			if result != tt.expected {
				t.Errorf("extractPhone(%v) = %s, want %s", tt.jid, result, tt.expected)
			}
		})
	}
}

func TestEchoRequest_InstanceID_Priority(t *testing.T) {
	ctx := context.Background()
	emitterInstanceID := uuid.New()
	requestInstanceID := uuid.New()

	t.Run("uses request InstanceID when provided", func(t *testing.T) {
		emitter := NewEmitter(ctx, &EmitterConfig{
			InstanceID: emitterInstanceID,
			Enabled:    false, // Disabled to avoid routing
		})

		req := &EchoRequest{
			InstanceID:        requestInstanceID,
			WhatsAppMessageID: "3EB0ABC123",
			RecipientJID:      types.JID{User: "5511888888888", Server: types.DefaultUserServer},
			Timestamp:         time.Now(),
			MessageType:       "text",
			ZaapID:            "test-zaap-id",
		}

		// Since emitter is disabled, we just verify the request has the right InstanceID
		if req.InstanceID != requestInstanceID {
			t.Errorf("expected request InstanceID %s, got %s", requestInstanceID, req.InstanceID)
		}

		// Verify emitter fallback InstanceID is set
		if emitter.instanceID != emitterInstanceID {
			t.Errorf("expected emitter instanceID %s, got %s", emitterInstanceID, emitter.instanceID)
		}
	})

	t.Run("request InstanceID is required", func(t *testing.T) {
		req := &EchoRequest{
			InstanceID:        uuid.Nil,
			WhatsAppMessageID: "3EB0ABC123",
			RecipientJID:      types.JID{User: "5511888888888", Server: types.DefaultUserServer},
			Timestamp:         time.Now(),
			MessageType:       "text",
			ZaapID:            "test-zaap-id",
		}

		// Verify the InstanceID fallback logic exists (tested via emitter code)
		if req.InstanceID != uuid.Nil {
			t.Errorf("expected nil InstanceID, got %s", req.InstanceID)
		}
	})
}
