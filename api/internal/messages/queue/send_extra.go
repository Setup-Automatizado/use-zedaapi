package queue

import (
	wameow "go.mau.fi/whatsmeow"
)

// BuildSendExtra constructs a SendRequestExtra with the pre-generated WhatsApp
// message ID from args. When WhatsAppMessageID is set, whatsmeow will use it
// instead of generating a new one, enabling end-to-end message ID correlation.
func BuildSendExtra(args *SendMessageArgs) wameow.SendRequestExtra {
	var extra wameow.SendRequestExtra
	if args.WhatsAppMessageID != "" {
		extra.ID = args.WhatsAppMessageID
	}
	return extra
}
