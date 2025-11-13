package queue

import (
	"context"
	"fmt"
	"time"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// PresenceHelper handles WhatsApp presence simulation (typing, recording, etc.)
type PresenceHelper struct{}

// NewPresenceHelper creates a new presence helper
func NewPresenceHelper() *PresenceHelper {
	return &PresenceHelper{}
}

// SimulateTyping sends typing presence indicator and waits for the specified duration
// Shows "typing..." indicator in WhatsApp
func (h *PresenceHelper) SimulateTyping(client *wameow.Client, recipientJID types.JID, delayMs int64) error {
	if delayMs <= 0 {
		return nil
	}

	// Send composing presence
	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return fmt.Errorf("send typing presence: %w", err)
	}

	// Wait for delay duration
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Send paused presence (stop typing)
	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return fmt.Errorf("send paused presence: %w", err)
	}

	return nil
}

// SimulateRecording sends recording audio presence and waits for the specified duration
// Shows "recording audio..." indicator in WhatsApp
func (h *PresenceHelper) SimulateRecording(client *wameow.Client, recipientJID types.JID, delayMs int64) error {
	if delayMs <= 0 {
		return nil
	}

	// Send recording presence
	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaAudio); err != nil {
		return fmt.Errorf("send recording presence: %w", err)
	}

	// Wait for delay duration
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Send paused presence (stop recording)
	if err := client.SendChatPresence(context.Background(), recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaAudio); err != nil {
		return fmt.Errorf("send paused presence: %w", err)
	}

	return nil
}
