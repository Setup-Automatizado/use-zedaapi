package zapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/api/internal/events/transform"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/proto/waE2E"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Transformer struct {
	connectedPhone string
}

func NewTransformer(connectedPhone string) *Transformer {
	return &Transformer{
		connectedPhone: connectedPhone,
	}
}

func (t *Transformer) TargetSchema() string {
	return "zapi"
}

func (t *Transformer) SupportsEventType(eventType string) bool {
	switch eventType {
	case "message", "receipt", "chat_presence", "presence", "connected", "disconnected":
		return true
	default:
		return false
	}
}

func (t *Transformer) Transform(ctx context.Context, event *types.InternalEvent) (json.RawMessage, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "zapi_transformer"),
		slog.String("instance_id", event.InstanceID.String()),
		slog.String("event_type", event.EventType),
	)

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

	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize webhook: %w", err)
	}

	logger.InfoContext(ctx, "transformed to Z-API webhook",
		slog.Int("payload_size", len(payload)),
	)

	return payload, nil
}

func (t *Transformer) transformMessage(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ReceivedCallback, error) {
	msgEvent, ok := event.RawPayload.(*events.Message)
	if !ok {
		return nil, fmt.Errorf("invalid message payload type")
	}

	chatJID, chatParseErr := parseJID(event.Metadata["chat"])
	senderJID, senderParseErr := parseJID(event.Metadata["from"])
	provider := eventctx.ContactProvider(ctx)

	chatPhone := normalizeConversationPhone(event.Metadata, chatJID, chatParseErr)
	callback := &ReceivedCallback{
		Type:           "ReceivedCallback",
		InstanceID:     event.InstanceID.String(),
		MessageID:      event.Metadata["message_id"],
		Phone:          chatPhone,
		FromMe:         event.Metadata["from_me"] == "true",
		IsGroup:        event.Metadata["is_group"] == "true",
		Momment:        event.CapturedAt.UnixMilli(),
		Status:         "RECEIVED",
		ConnectedPhone: t.connectedPhone,
		IsEdit:         event.Metadata["is_edit"] == "true",
	}
	callback.ChatLid = deriveChatLID(event.Metadata, chatJID, chatParseErr)

	if provider != nil && chatParseErr == nil {
		if name := provider.ContactName(ctx, chatJID); name != "" {
			callback.ChatName = name
		}
		if photo := provider.ContactPhoto(ctx, chatJID); photo != "" {
			callback.Photo = photo
		}
	}

	if callback.ChatName == "" {
		callback.ChatName = event.Metadata["chat"]
	}

	if pushName, ok := event.Metadata["push_name"]; ok && pushName != "" {
		callback.SenderName = pushName
	}

	if verifiedNameRaw, ok := event.Metadata["verified_name"]; ok && verifiedNameRaw != "" {
		if verifiedName := extractVerifiedBusinessName(verifiedNameRaw); verifiedName != "" {
			callback.ProfileName = verifiedName
		}
	}

	if broadcastOwner, ok := event.Metadata["broadcast_list_owner"]; ok && broadcastOwner != "" {
		callback.Broadcast = true
	}

	if addressingMode, ok := event.Metadata["addressing_mode"]; ok && addressingMode == "lid" {
		if senderAlt, ok := event.Metadata["sender_alt"]; ok {
			callback.SenderLid = normalizeLID(senderAlt)
		}
	}

	if provider != nil && senderParseErr == nil {
		if callback.SenderName == "" {
			if name := provider.ContactName(ctx, senderJID); name != "" {
				callback.SenderName = name
			}
		}
		if photo := provider.ContactPhoto(ctx, senderJID); photo != "" {
			callback.SenderPhoto = photo
		}
	}

	if event.QuotedMessageID != "" {
		callback.ReferenceMessageID = event.QuotedMessageID
	}
	if event.IsForwarded {
		callback.Forwarded = true
	}
	if len(event.MentionedJIDs) > 0 {
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

	if revokedID, ok := event.Metadata["revoked_message_id"]; ok && revokedID != "" {
		callback.RevokedMessageID = revokedID
	}

	if event.Metadata["is_ephemeral"] == "true" {
		callback.MessageExpirationSeconds = 604800
	}

	if event.Metadata["is_view_once"] == "true" {
		callback.ViewOnce = true
	}

	if callback.IsGroup {
		if participant := deriveParticipantPhone(event.Metadata, senderJID, senderParseErr); participant != "" {
			callback.ParticipantPhone = participant
		}
		if participantLID := normalizeLID(event.Metadata["sender_alt"]); participantLID != "" {
			callback.ParticipantLid = participantLID
		}
	}

	if err := t.extractMessageContent(msgEvent.Message, callback, event); err != nil {
		return nil, fmt.Errorf("failed to extract message content: %w", err)
	}

	return callback, nil
}

func (t *Transformer) extractMessageContent(msg *waE2E.Message, callback *ReceivedCallback, event *types.InternalEvent) error {
	if text := msg.GetConversation(); text != "" {
		callback.Text = &TextContent{
			Message: text,
		}
		return nil
	}

	if extText := msg.GetExtendedTextMessage(); extText != nil {
		callback.Text = &TextContent{
			Message: extText.GetText(),
		}
		return nil
	}

	if img := msg.GetImageMessage(); img != nil {
		callback.Image = &ImageContent{
			ImageURL:     event.Metadata["media_url"],
			ThumbnailURL: event.Metadata["thumbnail_url"],
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

	if video := msg.GetVideoMessage(); video != nil {
		callback.Video = &VideoContent{
			VideoURL: event.Metadata["media_url"],
			Caption:  video.GetCaption(),
			MimeType: video.GetMimetype(),
			Seconds:  int(video.GetSeconds()),
			Width:    event.MediaWidth,
			Height:   event.MediaHeight,
			IsGif:    event.MediaIsGIF,
			ViewOnce: video.GetViewOnce(),
		}
		return nil
	}

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

	if contact := msg.GetContactMessage(); contact != nil {
		callback.Contact = &ContactContent{
			DisplayName: contact.GetDisplayName(),
			VCard:       contact.GetVcard(),
		}
		return nil
	}

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

	if pollVote := msg.GetPollUpdateMessage(); pollVote != nil {
		callback.PollVote = &PollVoteContent{
			PollMessageID: pollVote.GetPollCreationMessageKey().GetID(),
			Options:       []PollOption{},
		}
		return nil
	}

	if btnResp := msg.GetButtonsResponseMessage(); btnResp != nil {
		callback.ButtonsResponseMessage = &ButtonsResponseContent{
			ButtonID: btnResp.GetSelectedButtonID(),
			Message:  btnResp.GetSelectedDisplayText(),
		}
		return nil
	}

	if listResp := msg.GetListResponseMessage(); listResp != nil {
		callback.ListResponseMessage = &ListResponseContent{
			Title:         listResp.GetTitle(),
			SelectedRowID: listResp.GetSingleSelectReply().GetSelectedRowID(),
		}
		return nil
	}

	if templateResp := msg.GetTemplateButtonReplyMessage(); templateResp != nil {
		callback.ButtonsResponseMessage = &ButtonsResponseContent{
			ButtonID: templateResp.GetSelectedID(),
			Message:  templateResp.GetSelectedDisplayText(),
		}
		return nil
	}

	return fmt.Errorf("unsupported message type")
}

func (t *Transformer) transformReceipt(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*MessageStatusCallback, error) {
	receiptEvent, ok := event.RawPayload.(*events.Receipt)
	if !ok {
		return nil, fmt.Errorf("invalid receipt payload type")
	}

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

	chatJID, chatErr := parseJID(event.Metadata["chat"])
	callback := &MessageStatusCallback{
		Type:       "MessageStatusCallback",
		InstanceID: event.InstanceID.String(),
		Status:     status,
		IDs:        receiptEvent.MessageIDs,
		Momment:    receiptEvent.Timestamp.UnixMilli(),
		Phone:      normalizeConversationPhone(event.Metadata, chatJID, chatErr),
		IsGroup:    event.Metadata["is_group"] == "true",
	}

	return callback, nil
}

func (t *Transformer) transformChatPresence(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*PresenceChatCallback, error) {
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

	chatJID, chatErr := parseJID(event.Metadata["chat"])
	callback := &PresenceChatCallback{
		Type:       "PresenceChatCallback",
		Phone:      normalizeConversationPhone(event.Metadata, chatJID, chatErr),
		Status:     status,
		InstanceID: event.InstanceID.String(),
	}

	return callback, nil
}

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

	if lastSeenStr, ok := event.Metadata["last_seen"]; ok && lastSeenStr != "" {
		var lastSeen int64
		fmt.Sscanf(lastSeenStr, "%d", &lastSeen)
		callback.LastSeen = &lastSeen
	}

	return callback, nil
}

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

func extractPhoneNumber(jid string) string {
	parsed, err := parseJID(jid)
	if err != nil {
		return sanitizeConversationFallback(jid)
	}
	return userPhoneFromJID(parsed)
}

func parseJID(value string) (watypes.JID, error) {
	if value == "" {
		return watypes.JID{}, fmt.Errorf("empty jid")
	}
	return watypes.ParseJID(value)
}

func normalizeConversationPhone(metadata map[string]string, chat watypes.JID, parseErr error) string {
	if parseErr == nil {
		if chat.Server == watypes.HiddenUserServer {
			if alt := metadata["recipient_alt"]; alt != "" {
				if altJID, err := watypes.ParseJID(alt); err == nil {
					chat = altJID
				} else {
					return sanitizeConversationFallback(alt)
				}
			}
		}
		return conversationIdentifierFromJID(chat)
	}

	return sanitizeConversationFallback(metadata["chat"])
}

func deriveChatLID(metadata map[string]string, chat watypes.JID, parseErr error) *string {
	if parseErr == nil && chat.Server == watypes.HiddenUserServer {
		if normalized := normalizeLID(chat.String()); normalized != "" {
			return stringPtr(normalized)
		}
		return stringPtr(chat.String())
	}

	if raw := metadata["chat"]; raw != "" {
		if strings.HasSuffix(raw, "@"+watypes.HiddenUserServer) {
			if normalized := normalizeLID(raw); normalized != "" {
				return stringPtr(normalized)
			}
			return stringPtr(raw)
		}
	}

	if raw := metadata["chat_lid"]; raw != "" {
		if normalized := normalizeLID(raw); normalized != "" {
			return stringPtr(normalized)
		}
		return stringPtr(raw)
	}

	return nil
}

func extractVerifiedBusinessName(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	var envelope struct {
		Details *struct {
			VerifiedName string `json:"verifiedName"`
		} `json:"Details"`
		VerifiedName string `json:"verified_name"`
	}

	if err := json.Unmarshal([]byte(trimmed), &envelope); err == nil {
		if envelope.Details != nil && envelope.Details.VerifiedName != "" {
			return envelope.Details.VerifiedName
		}
		if envelope.VerifiedName != "" {
			return envelope.VerifiedName
		}
	} else if trimmed != "" && trimmed[0] != '{' {
		return trimmed
	}

	if trimmed != "" && trimmed[0] != '{' {
		return trimmed
	}

	return ""
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func deriveParticipantPhone(metadata map[string]string, sender watypes.JID, parseErr error) string {
	if parseErr == nil {
		if metadata["addressing_mode"] == "lid" {
			if alt := metadata["sender_alt"]; alt != "" {
				if altJID, err := watypes.ParseJID(alt); err == nil {
					sender = altJID
				} else {
					return sanitizeConversationFallback(alt)
				}
			}
		}
		return userPhoneFromJID(sender)
	}

	return sanitizeConversationFallback(metadata["from"])
}

func conversationIdentifierFromJID(jid watypes.JID) string {
	user := sanitizeUserComponent(jid.User)
	switch jid.Server {
	case watypes.GroupServer:
		if user == "" {
			user = jid.User
		}
		return user + "-group"
	case watypes.BroadcastServer:
		if jid.User == watypes.StatusBroadcastJID.User {
			return "status"
		}
		if user == "" {
			user = jid.User
		}
		return user + "-broadcast"
	case watypes.NewsletterServer:
		if user == "" {
			user = jid.User
		}
		return user + "-channel"
	default:
		if user == "" {
			return jid.User
		}
		return user
	}
}

func userPhoneFromJID(jid watypes.JID) string {
	user := sanitizeUserComponent(jid.User)
	if user == "" {
		return jid.User
	}
	return user
}

func normalizeLID(value string) string {
	if value == "" {
		return ""
	}
	jid, err := watypes.ParseJID(value)
	if err != nil {
		return manualNormalizedLID(value)
	}
	if jid.Server != watypes.HiddenUserServer {
		return ""
	}
	user := sanitizeUserComponent(jid.User)
	if user == "" {
		return ""
	}
	return user + "@" + watypes.HiddenUserServer
}

func manualNormalizedLID(value string) string {
	if value == "" {
		return ""
	}
	jidPart := value
	if idx := strings.IndexRune(value, '@'); idx >= 0 {
		jidPart = value[:idx]
	}
	user := sanitizeUserComponent(jidPart)
	if user == "" {
		user = jidPart
	}
	if user == "" {
		return ""
	}
	return user + "@" + watypes.HiddenUserServer
}

func sanitizeConversationFallback(value string) string {
	if value == "" {
		return ""
	}
	if idx := strings.IndexRune(value, '@'); idx >= 0 {
		user := value[:idx]
		server := value[idx+1:]
		sanitized := sanitizeUserComponent(user)
		tempJID := watypes.JID{User: sanitized, Server: server}
		return conversationIdentifierFromJID(tempJID)
	}
	sanitized := sanitizeUserComponent(value)
	if sanitized == "" {
		return value
	}
	return sanitized
}

func sanitizeUserComponent(user string) string {
	if idx := strings.IndexRune(user, ':'); idx >= 0 {
		user = user[:idx]
	}
	if idx := strings.IndexRune(user, '.'); idx >= 0 {
		user = user[:idx]
	}
	return user
}
