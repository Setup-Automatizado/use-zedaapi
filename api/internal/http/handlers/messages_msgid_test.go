package handlers

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
)

// mockIDGen implements MessageIDGenerator for testing.
type mockIDGen struct {
	client *wameow.Client
	ok     bool
}

func (m *mockIDGen) GetClient(_ uuid.UUID) (*wameow.Client, bool) {
	return m.client, m.ok
}

func TestPreGenerateWhatsAppMessageID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		msgIDGen  MessageIDGenerator
		wantEmpty bool
	}{
		{
			name:      "nil generator returns empty",
			msgIDGen:  nil,
			wantEmpty: true,
		},
		{
			name:      "client not found returns empty",
			msgIDGen:  &mockIDGen{client: nil, ok: false},
			wantEmpty: true,
		},
		{
			name:      "client nil but ok=true returns empty",
			msgIDGen:  &mockIDGen{client: nil, ok: true},
			wantEmpty: true,
		},
		{
			name:      "valid client returns 3EB0 prefixed ID",
			msgIDGen:  &mockIDGen{client: wameow.NewClient(nil, nil), ok: true},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &MessageHandler{
				msgIDGen: tt.msgIDGen,
				log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			got := h.preGenerateWhatsAppMessageID(uuid.New())

			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.NotEmpty(t, got)
				assert.True(t, strings.HasPrefix(got, "3EB0"),
					"expected ID to start with 3EB0, got %q", got)
			}
		})
	}
}

func TestNewSendMessageResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		zaapID        string
		whatsAppMsgID string
		wantMessageID string
		wantID        string
		wantZaapID    string
		wantStatus    string
	}{
		{
			name:          "whatsapp msg ID present uses it for messageId and id",
			zaapID:        "uuid-abc-123",
			whatsAppMsgID: "3EB0ABC123",
			wantMessageID: "3EB0ABC123",
			wantID:        "3EB0ABC123",
			wantZaapID:    "uuid-abc-123",
			wantStatus:    "queued",
		},
		{
			name:          "empty whatsapp msg ID falls back to zaapID",
			zaapID:        "uuid-def-456",
			whatsAppMsgID: "",
			wantMessageID: "uuid-def-456",
			wantID:        "uuid-def-456",
			wantZaapID:    "uuid-def-456",
			wantStatus:    "queued",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &MessageHandler{
				log: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			resp := h.newSendMessageResponse(tt.zaapID, tt.whatsAppMsgID, nil)

			assert.Equal(t, tt.wantZaapID, resp.ZaapID)
			assert.Equal(t, tt.wantMessageID, resp.MessageID)
			assert.Equal(t, tt.wantID, resp.ID)
			assert.Equal(t, tt.wantStatus, resp.Status)
			assert.Nil(t, resp.WhatsAppStatus)
		})
	}
}

func TestEnqueueWithPreGeneratedID(t *testing.T) {
	t.Parallel()

	t.Run("preserves existing WhatsAppMessageID", func(t *testing.T) {
		t.Parallel()

		var captured queue.SendMessageArgs
		h := &MessageHandler{
			msgIDGen: &mockIDGen{client: wameow.NewClient(nil, nil), ok: true},
			log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
			enqueueMessage: func(_ context.Context, _ uuid.UUID, args queue.SendMessageArgs) (string, error) {
				captured = args
				return "mock-zaap-id", nil
			},
		}

		args := &queue.SendMessageArgs{
			WhatsAppMessageID: "3EB0PRE",
		}
		zaapID, waMsgID, err := h.enqueueWithPreGeneratedID(context.Background(), uuid.New(), args)

		require.NoError(t, err)
		assert.Equal(t, "mock-zaap-id", zaapID)
		assert.Equal(t, "3EB0PRE", waMsgID, "should not overwrite existing WhatsAppMessageID")
		assert.Equal(t, "3EB0PRE", captured.WhatsAppMessageID, "enqueued args should keep original ID")
	})

	t.Run("generates ID when empty and client available", func(t *testing.T) {
		t.Parallel()

		var captured queue.SendMessageArgs
		h := &MessageHandler{
			msgIDGen: &mockIDGen{client: wameow.NewClient(nil, nil), ok: true},
			log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
			enqueueMessage: func(_ context.Context, _ uuid.UUID, args queue.SendMessageArgs) (string, error) {
				captured = args
				return "mock-zaap-id", nil
			},
		}

		args := &queue.SendMessageArgs{
			WhatsAppMessageID: "",
		}
		zaapID, waMsgID, err := h.enqueueWithPreGeneratedID(context.Background(), uuid.New(), args)

		require.NoError(t, err)
		assert.Equal(t, "mock-zaap-id", zaapID)
		assert.NotEmpty(t, waMsgID, "should have generated a WhatsApp message ID")
		assert.True(t, strings.HasPrefix(waMsgID, "3EB0"),
			"generated ID should start with 3EB0, got %q", waMsgID)
		assert.Equal(t, waMsgID, captured.WhatsAppMessageID,
			"enqueued args should contain the generated ID")
	})

	t.Run("returns empty ID when no client available", func(t *testing.T) {
		t.Parallel()

		var captured queue.SendMessageArgs
		h := &MessageHandler{
			msgIDGen: &mockIDGen{client: nil, ok: false},
			log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
			enqueueMessage: func(_ context.Context, _ uuid.UUID, args queue.SendMessageArgs) (string, error) {
				captured = args
				return "mock-zaap-id", nil
			},
		}

		args := &queue.SendMessageArgs{
			WhatsAppMessageID: "",
		}
		zaapID, waMsgID, err := h.enqueueWithPreGeneratedID(context.Background(), uuid.New(), args)

		require.NoError(t, err)
		assert.Equal(t, "mock-zaap-id", zaapID)
		assert.Empty(t, waMsgID, "should return empty when client not available")
		assert.Empty(t, captured.WhatsAppMessageID, "enqueued args should have empty ID")
	})

	t.Run("returns empty ID when generator is nil", func(t *testing.T) {
		t.Parallel()

		var captured queue.SendMessageArgs
		h := &MessageHandler{
			msgIDGen: nil,
			log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
			enqueueMessage: func(_ context.Context, _ uuid.UUID, args queue.SendMessageArgs) (string, error) {
				captured = args
				return "mock-zaap-id", nil
			},
		}

		args := &queue.SendMessageArgs{
			WhatsAppMessageID: "",
		}
		zaapID, waMsgID, err := h.enqueueWithPreGeneratedID(context.Background(), uuid.New(), args)

		require.NoError(t, err)
		assert.Equal(t, "mock-zaap-id", zaapID)
		assert.Empty(t, waMsgID)
		assert.Empty(t, captured.WhatsAppMessageID)
	})
}
