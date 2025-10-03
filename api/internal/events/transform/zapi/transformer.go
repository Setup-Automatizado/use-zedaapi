package zapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"go.mau.fi/whatsmeow/api/internal/events/transform"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

// Transformer converts InternalEvent to Z-API webhook format.
// This is the target transformer that produces webhook payloads ready for delivery.
type Transformer struct {
	connectedPhone string // The WhatsApp number connected to this instance
}

// NewTransformer creates a new Z-API transformer.
// connectedPhone is the WhatsApp number (e.g., "5544999999999").
func NewTransformer(connectedPhone string) *Transformer {
	return &Transformer{
		connectedPhone: connectedPhone,
	}
}

// TargetSchema returns the target schema identifier.
func (t *Transformer) TargetSchema() string {
	return "zapi"
}

// SupportsEventType returns true if this transformer can handle the given event type.
func (t *Transformer) SupportsEventType(eventType string) bool {
	switch eventType {
	case "message", "receipt", "chat_presence", "presence", "connected", "disconnected":
		return true
	default:
		return false
	}
}

// Transform converts an InternalEvent to Z-API webhook format.
func (t *Transformer) Transform(ctx context.Context, event *types.InternalEvent) (json.RawMessage, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "zapi_transformer"),
		slog.String("instance_id", event.InstanceID.String()),
		slog.String("event_type", event.EventType),
	)

	// Route to specific transformation based on event type
	var result interface{}
	var err error

	switch event.EventType {
	case "message":
		result, err = t.transformMessage(ctx, logger, event)
	case "receipt":
		result, err = t.transformReceipt(ctx, logger, event)
	case "chat_presence":
		result, err = t.transformChatPresence(ctx, logger, event)
	case "presence":
		result, err = t.transformPresence(ctx, logger, event)
	case "connected":
		result, err = t.transformConnected(ctx, logger, event)
	case "disconnected":
		result, err = t.transformDisconnected(ctx, logger, event)
	default:
		logger.Debug("unsupported event type for Z-API transformation")
		return nil, transform.ErrUnsupportedEvent
	}

	if err != nil {
		return nil, err
	}

	// Serialize to JSON
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize webhook: %w", err)
	}

	logger.InfoContext(ctx, "transformed to Z-API webhook",
		slog.Int("payload_size", len(payload)),
	)

	return payload, nil
}

// transformMessage converts a message event to ReceivedCallback.
func (t *Transformer) transformMessage(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ReceivedCallback, error) {
	// Extract whatsmeow Message from RawPayload
	msgEvent, ok := event.RawPayload.(*events.Message)
	if !ok {
		return nil, fmt.Errorf("invalid message payload type")
	}

	// Build base ReceivedCallback
	callback := &ReceivedCallback{
		Type:           "ReceivedCallback",
		InstanceID:     event.InstanceID.String(),
		MessageID:      event.Metadata["message_id"],
		Phone:          extractPhoneFromMetadata(event.Metadata["from"], event.Metadata),
		FromMe:         event.Metadata["from_me"] == "true",
		IsGroup:        event.Metadata["is_group"] == "true",
		Momment:        event.CapturedAt.UnixMilli(),
		Status:         "RECEIVED",
		ConnectedPhone: t.connectedPhone,
		ChatName:       event.Metadata["chat"],
		IsEdit:         event.Metadata["is_edit"] == "true",
	}

	// Add push name if available
	if pushName, ok := event.Metadata["push_name"]; ok && pushName != "" {
		callback.SenderName = pushName
	}

	// Add verified name if available (business accounts)
	if verifiedName, ok := event.Metadata["verified_name"]; ok && verifiedName != "" {
		callback.ProfileName = verifiedName
	}

	// Add broadcast flag
	if broadcastOwner, ok := event.Metadata["broadcast_list_owner"]; ok && broadcastOwner != "" {
		callback.Broadcast = true
	}

	// Add LID if using LID addressing
	if addressingMode, ok := event.Metadata["addressing_mode"]; ok && addressingMode == "lid" {
		if senderAlt, ok := event.Metadata["sender_alt"]; ok {
			callback.SenderLid = extractPhoneNumber(senderAlt)
		}
	}

	// Add ContextInfo fields (quotes, mentions, forwards)
	if event.QuotedMessageID != "" {
		callback.ReferenceMessageID = event.QuotedMessageID
	}
	if event.IsForwarded {
		callback.Forwarded = true
	}
	if len(event.MentionedJIDs) > 0 {
		// Extract phone numbers from mentioned JIDs
		mentioned := make([]string, 0, len(event.MentionedJIDs))
		for _, jid := range event.MentionedJIDs {
			phone := extractPhoneNumber(jid)
			if phone != "" {
				mentioned = append(mentioned, phone)
			}
		}
		if len(mentioned) > 0 {
			callback.Mentioned = mentioned
		}
	}

	// Handle message revocation
	if revokedID, ok := event.Metadata["revoked_message_id"]; ok && revokedID != "" {
		callback.RevokedMessageID = revokedID
	}

	// Handle ephemeral messages
	if event.Metadata["is_ephemeral"] == "true" {
		callback.MessageExpirationSeconds = 604800 // 7 days default
	}

	// Handle view once
	if event.Metadata["is_view_once"] == "true" {
		callback.ViewOnce = true
	}

	// Extract message content based on type
	if err := t.extractMessageContent(msgEvent.Message, callback, event); err != nil {
		return nil, fmt.Errorf("failed to extract message content: %w", err)
	}

	return callback, nil
}

// extractMessageContent extracts the actual message content into the callback.
func (t *Transformer) extractMessageContent(msg *waE2E.Message, callback *ReceivedCallback, event *types.InternalEvent) error {
	// Text message
	if text := msg.GetConversation(); text != "" {
		callback.Text = &TextContent{
			Message: text,
		}
		return nil
	}

	// Extended text message
	if extText := msg.GetExtendedTextMessage(); extText != nil {
		callback.Text = &TextContent{
			Message: extText.GetText(),
		}
		return nil
	}

	// Image message
	if img := msg.GetImageMessage(); img != nil {
		callback.Image = &ImageContent{
			ImageURL:     event.Metadata["media_url"],     // Injected after S3 upload
			ThumbnailURL: event.Metadata["thumbnail_url"], // Injected after S3 upload
			Caption:      img.GetCaption(),
			MimeType:     img.GetMimetype(),
			Width:        event.MediaWidth,
			Height:       event.MediaHeight,
			IsGif:        event.MediaIsGIF,
			IsAnimated:   event.MediaIsAnimated,
			ViewOnce:     img.GetViewOnce(),
		}
		return nil
	}

	// Video message
	if video := msg.GetVideoMessage(); video != nil {
		callback.Video = &VideoContent{
			VideoURL: event.Metadata["media_url"],
			Caption:  video.GetCaption(),
			MimeType: video.GetMimetype(),
			Seconds:  int(video.GetSeconds()),
			Width:    event.MediaWidth,
			Height:   event.MediaHeight,
			IsGif:    event.MediaIsGIF, // GIF playback mode
			ViewOnce: video.GetViewOnce(),
		}
		return nil
	}

	// Audio message
	if audio := msg.GetAudioMessage(); audio != nil {
		callback.Audio = &AudioContent{
			AudioURL: event.Metadata["media_url"],
			MimeType: audio.GetMimetype(),
			PTT:      audio.GetPTT(),
			Seconds:  int(audio.GetSeconds()),
			Waveform: event.MediaWaveform,
		}
		return nil
	}

	// Document message
	if doc := msg.GetDocumentMessage(); doc != nil {
		callback.Document = &DocumentContent{
			DocumentURL:  event.Metadata["media_url"],
			MimeType:     doc.GetMimetype(),
			Title:        doc.GetTitle(),
			FileName:     doc.GetFileName(),
			PageCount:    int(doc.GetPageCount()),
			ThumbnailURL: event.Metadata["thumbnail_url"],
			Caption:      doc.GetCaption(),
		}
		return nil
	}

	// Sticker message
	if sticker := msg.GetStickerMessage(); sticker != nil {
		callback.Sticker = &StickerContent{
			StickerURL: event.Metadata["media_url"],
			MimeType:   sticker.GetMimetype(),
			IsAnimated: event.MediaIsAnimated,
			Width:      event.MediaWidth,
			Height:     event.MediaHeight,
		}
		return nil
	}

	// Location message
	if loc := msg.GetLocationMessage(); loc != nil {
		callback.Location = &LocationContent{
			Latitude:  loc.GetDegreesLatitude(),
			Longitude: loc.GetDegreesLongitude(),
			Name:      loc.GetName(),
			Address:   loc.GetAddress(),
			URL:       loc.GetURL(),
		}
		return nil
	}

	// Contact message
	if contact := msg.GetContactMessage(); contact != nil {
		callback.Contact = &ContactContent{
			DisplayName: contact.GetDisplayName(),
			VCard:       contact.GetVcard(),
		}
		return nil
	}

	// Reaction message
	if reaction := msg.GetReactionMessage(); reaction != nil {
		callback.Reaction = &ReactionContent{
			Value:      reaction.GetText(),
			Time:       reaction.GetSenderTimestampMS(),
			ReactionBy: extractPhoneNumber(event.Metadata["from"]),
			ReferencedMessage: &MessageRef{
				MessageID: reaction.GetKey().GetID(),
				FromMe:    reaction.GetKey().GetFromMe(),
			},
		}
		return nil
	}

	// Poll message
	if poll := msg.GetPollCreationMessage(); poll != nil {
		options := make([]PollOption, 0, len(poll.GetOptions()))
		for _, opt := range poll.GetOptions() {
			options = append(options, PollOption{
				Name: opt.GetOptionName(),
			})
		}
		callback.Poll = &PollContent{
			Question:       poll.GetName(),
			PollMaxOptions: int(poll.GetSelectableOptionsCount()),
			Options:        options,
		}
		return nil
	}

	// Poll vote message
	if pollVote := msg.GetPollUpdateMessage(); pollVote != nil {
		// TODO: Extract selected options from encrypted poll vote
		// The vote is encrypted and needs to be decrypted first
		callback.PollVote = &PollVoteContent{
			PollMessageID: pollVote.GetPollCreationMessageKey().GetID(),
			Options:       []PollOption{}, // Will be populated after decryption
		}
		return nil
	}

	// Buttons response message
	if btnResp := msg.GetButtonsResponseMessage(); btnResp != nil {
		callback.ButtonsResponseMessage = &ButtonsResponseContent{
			ButtonID: btnResp.GetSelectedButtonID(),
			Message:  btnResp.GetSelectedDisplayText(),
		}
		return nil
	}

	// List response message
	if listResp := msg.GetListResponseMessage(); listResp != nil {
		callback.ListResponseMessage = &ListResponseContent{
			Title:         listResp.GetTitle(),
			SelectedRowID: listResp.GetSingleSelectReply().GetSelectedRowID(),
		}
		return nil
	}

	// Template button reply message
	if templateResp := msg.GetTemplateButtonReplyMessage(); templateResp != nil {
		callback.ButtonsResponseMessage = &ButtonsResponseContent{
			ButtonID: templateResp.GetSelectedID(),
			Message:  templateResp.GetSelectedDisplayText(),
		}
		return nil
	}

	// If no specific message type matched, return error
	return fmt.Errorf("unsupported message type")
}

// transformReceipt converts a receipt event to MessageStatusCallback.
func (t *Transformer) transformReceipt(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*MessageStatusCallback, error) {
	// Extract whatsmeow Receipt from RawPayload
	receiptEvent, ok := event.RawPayload.(*events.Receipt)
	if !ok {
		return nil, fmt.Errorf("invalid receipt payload type")
	}

	// Map receipt type to Z-API status
	var status string
	switch receiptEvent.Type {
	case "delivered":
		status = "RECEIVED"
	case "read":
		status = "READ"
	case "played":
		status = "PLAYED"
	case "sender":
		status = "SENT"
	default:
		status = "SENT"
	}

	callback := &MessageStatusCallback{
		Type:       "MessageStatusCallback",
		InstanceID: event.InstanceID.String(),
		Status:     status,
		IDs:        receiptEvent.MessageIDs,
		Momment:    receiptEvent.Timestamp.UnixMilli(),
		Phone:      extractPhoneNumber(event.Metadata["chat"]),
		IsGroup:    event.Metadata["is_group"] == "true",
	}

	return callback, nil
}

// transformChatPresence converts a chat presence event to PresenceChatCallback.
func (t *Transformer) transformChatPresence(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*PresenceChatCallback, error) {
	// Map whatsmeow presence state to Z-API status
	var status string
	switch event.Metadata["state"] {
	case "composing":
		status = "COMPOSING"
	case "paused":
		status = "PAUSED"
	case "recording":
		status = "RECORDING"
	default:
		status = "AVAILABLE"
	}

	callback := &PresenceChatCallback{
		Type:       "PresenceChatCallback",
		Phone:      extractPhoneNumber(event.Metadata["chat"]),
		Status:     status,
		InstanceID: event.InstanceID.String(),
	}

	return callback, nil
}

// transformPresence converts a presence event to PresenceChatCallback.
func (t *Transformer) transformPresence(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*PresenceChatCallback, error) {
	var status string
	if event.Metadata["unavailable"] == "true" {
		status = "UNAVAILABLE"
	} else {
		status = "AVAILABLE"
	}

	callback := &PresenceChatCallback{
		Type:       "PresenceChatCallback",
		Phone:      extractPhoneNumber(event.Metadata["from"]),
		Status:     status,
		InstanceID: event.InstanceID.String(),
	}

	// Add last seen if available
	if lastSeenStr, ok := event.Metadata["last_seen"]; ok && lastSeenStr != "" {
		var lastSeen int64
		fmt.Sscanf(lastSeenStr, "%d", &lastSeen)
		callback.LastSeen = &lastSeen
	}

	return callback, nil
}

// transformConnected converts a connected event to ConnectedCallback.
func (t *Transformer) transformConnected(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ConnectedCallback, error) {
	callback := &ConnectedCallback{
		Type:       "ConnectedCallback",
		Connected:  true,
		Momment:    event.CapturedAt.UnixMilli(),
		InstanceID: event.InstanceID.String(),
		Phone:      t.connectedPhone,
	}

	return callback, nil
}

// transformDisconnected converts a disconnected event to DisconnectedCallback.
func (t *Transformer) transformDisconnected(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*DisconnectedCallback, error) {
	callback := &DisconnectedCallback{
		Type:         "DisconnectedCallback",
		Disconnected: true,
		Momment:      event.CapturedAt.UnixMilli(),
		InstanceID:   event.InstanceID.String(),
		Error:        "Device has been disconnected",
	}

	return callback, nil
}

// extractPhoneNumber extracts just the phone number from a JID string.
// Handles LID addressing mode and alternate JIDs.
// Example: "5544999999999@s.whatsapp.net" → "5544999999999"
func extractPhoneNumber(jid string) string {
	// Find the @ symbol
	for i, c := range jid {
		if c == '@' {
			user := jid[:i]
			server := jid[i+1:]

			// Handle special servers
			switch server {
			case "s.whatsapp.net":
				return user // Regular phone number
			case "lid":
				return user // LID identifier (should use alternate JID if available)
			case "g.us":
				return user // Group ID
			case "broadcast":
				if user == "status" {
					return "status_broadcast"
				}
				return user
			case "newsletter":
				return user // Newsletter ID
			default:
				return user
			}
		}
	}
	return jid
}

// extractPhoneFromMetadata extracts phone number considering addressing mode and alternate JIDs.
// This is used to handle LID addressing properly.
func extractPhoneFromMetadata(jid string, metadata map[string]string) string {
	// Check if LID addressing mode - use alternate JID
	if addressingMode, ok := metadata["addressing_mode"]; ok && addressingMode == "lid" {
		if senderAlt, ok := metadata["sender_alt"]; ok && senderAlt != "" {
			return extractPhoneNumber(senderAlt)
		}
	}

	// Default to regular JID extraction
	return extractPhoneNumber(jid)
}
