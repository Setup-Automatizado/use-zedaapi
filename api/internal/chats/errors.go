package chats

import "errors"

var (
	// ErrClientNotConnected is returned when the WhatsApp client is not available or connected.
	ErrClientNotConnected = errors.New("whatsapp client not connected")
)
