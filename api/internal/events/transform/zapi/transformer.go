package zapi

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/api/internal/events/transform"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/proto/waE2E"
	waWeb "go.mau.fi/whatsmeow/proto/waWeb"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Transformer struct {
	connectedPhone string
	debugRaw       bool
	dumpDir        string
}

func NewTransformer(connectedPhone string, debug bool, dumpDir string) *Transformer {
	return &Transformer{
		connectedPhone: connectedPhone,
		debugRaw:       debug,
		dumpDir:        strings.TrimSpace(dumpDir),
	}
}

func (t *Transformer) TargetSchema() string {
	return "zapi"
}

// TODO: Adicionar novos eventos
func (t *Transformer) SupportsEventType(eventType string) bool {
	switch eventType {
	case "message", "receipt", "chat_presence", "presence", "connected", "disconnected", "undecryptable":
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
	case "undecryptable":
		result, err = t.transformUndecryptable(ctx, logger, event)
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

	if t.debugRaw {
		logger.DebugContext(ctx, "zapi payload",
			slog.Int("payload_size", len(payload)),
			slog.String("payload", string(payload)))
	}

	t.dumpPayload(logger, event, payload)

	logger.InfoContext(ctx, "transformed to Z-API webhook",
		slog.Int("payload_size", len(payload)),
	)

	return payload, nil
}

func (t *Transformer) dumpPayload(logger *slog.Logger, event *types.InternalEvent, payload []byte) {
	if !t.debugRaw {
		return
	}

	if t.dumpDir == "" {
		return
	}

	if err := os.MkdirAll(t.dumpDir, 0o755); err != nil {
		logger.Warn("failed to create debug dump directory",
			slog.String("dir", t.dumpDir),
			slog.String("error", err.Error()))
		return
	}

	record := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339Nano),
		"event_id":    event.EventID.String(),
		"event_type":  event.EventType,
		"instance_id": event.InstanceID.String(),
		"payload_len": len(payload),
	}

	if json.Valid(payload) {
		record["payload"] = json.RawMessage(append([]byte(nil), payload...))
	} else {
		record["payload_base64"] = base64.StdEncoding.EncodeToString(payload)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("%+v", record))
	}

	fileName := fmt.Sprintf("zapi_payload_%s_%s_%d.json", event.EventType, event.EventID.String(), time.Now().UnixNano())
	filePath := filepath.Join(t.dumpDir, fileName)

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		logger.Warn("failed to write zapi payload dump",
			slog.String("path", filePath),
			slog.String("error", err.Error()))
	}
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
	isStatusChat := false
	if chatParseErr == nil && chatJID == watypes.StatusBroadcastJID {
		isStatusChat = true
	}
	if chatPhone == "status" {
		isStatusChat = true
	}
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

	if chatName := strings.TrimSpace(event.Metadata["chat_name"]); chatName != "" {
		callback.ChatName = chatName
	}
	if chatPhoto := strings.TrimSpace(event.Metadata["chat_photo"]); chatPhoto != "" {
		callback.Photo = chatPhoto
	}
	if senderName := strings.TrimSpace(event.Metadata["sender_name"]); senderName != "" {
		callback.SenderName = senderName
	}
	if senderPhoto := strings.TrimSpace(event.Metadata["sender_photo"]); senderPhoto != "" {
		callback.SenderPhoto = senderPhoto
	}

	if chatParseErr == nil && chatJID.Server == watypes.NewsletterServer {
		callback.IsNewsletter = true
	} else if strings.HasSuffix(callback.Phone, "-channel") {
		callback.IsNewsletter = true
	}

	if isStatusChat {
		callback.Broadcast = true
		callback.IsBroadcast = true
	}

	if chatParseErr == nil {
		if chatJID == watypes.StatusBroadcastJID {
			callback.Broadcast = true
		}
		if chatJID.IsBroadcastList() {
			callback.Broadcast = true
		}
	}
	if strings.HasSuffix(callback.Phone, "-broadcast") || callback.Phone == "status" {
		callback.Broadcast = true
	}

	if provider != nil && chatParseErr == nil {
		if callback.ChatName == "" {
			if name := provider.ContactName(ctx, chatJID); name != "" {
				callback.ChatName = name
			}
		}
		if callback.Photo == "" {
			if photo := provider.ContactPhoto(ctx, chatJID); photo != "" {
				callback.Photo = photo
			}
		}
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

	if callback.Broadcast {
		callback.IsGroup = false
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
		if callback.SenderPhoto == "" {
			if photo := provider.ContactPhoto(ctx, senderJID); photo != "" {
				callback.SenderPhoto = photo
			}
		}
	}

	if callback.ChatName == "" {
		if chatParseErr == nil {
			callback.ChatName = conversationIdentifierFromJID(chatJID)
		} else {
			callback.ChatName = sanitizeConversationFallback(event.Metadata["chat"])
		}
	}

	if callback.SenderName == "" {
		if senderParseErr == nil {
			callback.SenderName = userPhoneFromJID(senderJID)
		} else {
			callback.SenderName = sanitizeConversationFallback(event.Metadata["from"])
		}
	}

	callback.FromAPI = event.Metadata["from_api"] == "true"

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

	if event.Metadata["waiting_message"] == "true" {
		callback.WaitingMessage = true
	}

	if quotedRemote := event.Metadata["quoted_remote_jid"]; quotedRemote != "" {
		if isStatusBroadcastReference(quotedRemote) {
			callback.IsStatusReply = true
		}
	}
	if !callback.IsStatusReply {
		if replyChat := event.Metadata["reply_to_chat"]; replyChat != "" {
			if isStatusBroadcastReference(replyChat) {
				callback.IsStatusReply = true
			}
		}
	}
	if !callback.IsStatusReply {
		if quotedSender := event.Metadata["quoted_sender"]; quotedSender != "" {
			if isStatusBroadcastReference(quotedSender) {
				callback.IsStatusReply = true
			}
		}
	}

	if raw := event.Metadata["external_ad_reply"]; raw != "" {
		var external struct {
			Title                 string `json:"title"`
			Body                  string `json:"body"`
			MediaType             int    `json:"mediaType"`
			ThumbnailURL          string `json:"thumbnailUrl"`
			SourceType            string `json:"sourceType"`
			SourceID              string `json:"sourceId"`
			SourceURL             string `json:"sourceUrl"`
			ContainsAutoReply     bool   `json:"containsAutoReply"`
			RenderLargerThumbnail bool   `json:"renderLargerThumbnail"`
			ShowAdAttribution     bool   `json:"showAdAttribution"`
		}
		if err := json.Unmarshal([]byte(raw), &external); err == nil {
			callback.ExternalAdReply = &ExternalAdReplyContent{
				Title:                 external.Title,
				Body:                  external.Body,
				MediaType:             external.MediaType,
				ThumbnailURL:          external.ThumbnailURL,
				SourceType:            external.SourceType,
				SourceID:              external.SourceID,
				SourceURL:             external.SourceURL,
				ContainsAutoReply:     external.ContainsAutoReply,
				RenderLargerThumbnail: external.RenderLargerThumbnail,
				ShowAdAttribution:     external.ShowAdAttribution,
			}
		}
	}

	if msgEvent.SourceWebMsg != nil {
		if stub := msgEvent.SourceWebMsg.GetMessageStubType(); stub != waWeb.WebMessageInfo_UNKNOWN {
			callback.Notification = stub.String()
		}
		if params := msgEvent.SourceWebMsg.GetMessageStubParameters(); len(params) > 0 {
			callback.NotificationParameters = append([]string(nil), params...)
		}
	}

	if call := msgEvent.Message.GetCall(); call != nil {
		if key := call.GetCallKey(); len(key) > 0 {
			callback.CallID = strings.ToUpper(hex.EncodeToString(key))
		}
	}

	if invite := msgEvent.Message.GetGroupInviteMessage(); invite != nil {
		if code := invite.GetInviteCode(); code != "" {
			callback.Code = code
		}
		if callback.ChatName == "" {
			callback.ChatName = invite.GetGroupName()
		}
	}

	if participant := deriveParticipantPhone(event.Metadata, senderJID, senderParseErr); participant != "" {
		callback.ParticipantPhone = participant
	}
	if participantLID := normalizeLID(event.Metadata["sender_alt"]); participantLID != "" {
		callback.ParticipantLid = participantLID
	} else if participantLID := normalizeLID(event.Metadata["sender"]); participantLID != "" {
		callback.ParticipantLid = participantLID
	}

	callback.IsGroup = callback.IsGroup ||
		strings.HasSuffix(callback.Phone, "-group") ||
		strings.HasSuffix(callback.Phone, "@g.us") ||
		strings.HasSuffix(callback.Phone, "@lid")

	if callback.Broadcast {
		callback.IsGroup = false
	}

	if err := t.extractMessageContent(msgEvent.Message, callback, event); err != nil {
		kinds := messagePayloadKinds(msgEvent.Message)
		logger.WarnContext(ctx, "unsupported message content type",
			slog.String("message_id", callback.MessageID),
			slog.String("payload_kinds", strings.Join(kinds, ",")),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to extract message content: %w", err)
	}

	if !callback.WaitingMessage && !hasMessageContent(callback) {
		callback.WaitingMessage = true
	}

	return callback, nil
}

func (t *Transformer) transformUndecryptable(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*ReceivedCallback, error) {
	undecryptable, ok := event.RawPayload.(*events.UndecryptableMessage)
	if !ok {
		return nil, fmt.Errorf("invalid undecryptable payload type")
	}

	chatJID, chatParseErr := parseJID(event.Metadata["chat"])
	senderJID, senderParseErr := parseJID(event.Metadata["from"])
	provider := eventctx.ContactProvider(ctx)

	chatPhone := normalizeConversationPhone(event.Metadata, chatJID, chatParseErr)
	isStatusChat := false
	if chatParseErr == nil && chatJID == watypes.StatusBroadcastJID {
		isStatusChat = true
	}
	if chatPhone == "status" {
		isStatusChat = true
	}
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
		WaitingMessage: true,
	}
	callback.ChatLid = deriveChatLID(event.Metadata, chatJID, chatParseErr)

	if isStatusChat {
		callback.Broadcast = true
		callback.IsBroadcast = true
	}

	if chatParseErr == nil && chatJID.Server == watypes.NewsletterServer {
		callback.IsNewsletter = true
	} else if strings.HasSuffix(callback.Phone, "-channel") {
		callback.IsNewsletter = true
	}

	if chatParseErr == nil {
		if chatJID == watypes.StatusBroadcastJID {
			callback.Broadcast = true
		}
		if chatJID.IsBroadcastList() {
			callback.Broadcast = true
		}
	}
	if strings.HasSuffix(callback.Phone, "-broadcast") || callback.Phone == "status" {
		callback.Broadcast = true
	}

	if provider != nil && chatParseErr == nil {
		if callback.ChatName == "" {
			if name := provider.ContactName(ctx, chatJID); name != "" {
				callback.ChatName = name
			}
		}
		if callback.Photo == "" {
			if photo := provider.ContactPhoto(ctx, chatJID); photo != "" {
				callback.Photo = photo
			}
		}
	}

	if pushName := event.Metadata["push_name"]; pushName != "" {
		callback.SenderName = pushName
	}

	if verifiedNameRaw := event.Metadata["verified_name"]; verifiedNameRaw != "" {
		if verifiedName := extractVerifiedBusinessName(verifiedNameRaw); verifiedName != "" {
			callback.ProfileName = verifiedName
		}
	}

	if callback.Broadcast {
		callback.IsGroup = false
	}

	if quotedRemote := event.Metadata["quoted_remote_jid"]; quotedRemote != "" {
		if isStatusBroadcastReference(quotedRemote) {
			callback.IsStatusReply = true
		}
	}
	if !callback.IsStatusReply {
		if replyChat := event.Metadata["reply_to_chat"]; replyChat != "" {
			if isStatusBroadcastReference(replyChat) {
				callback.IsStatusReply = true
			}
		}
	}
	if !callback.IsStatusReply {
		if quotedSender := event.Metadata["quoted_sender"]; quotedSender != "" {
			if isStatusBroadcastReference(quotedSender) {
				callback.IsStatusReply = true
			}
		}
	}

	if addressingMode := event.Metadata["addressing_mode"]; addressingMode == "lid" {
		if senderAlt := event.Metadata["sender_alt"]; senderAlt != "" {
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

	if callback.ChatName == "" {
		if chatParseErr == nil {
			callback.ChatName = conversationIdentifierFromJID(chatJID)
		} else {
			callback.ChatName = sanitizeConversationFallback(event.Metadata["chat"])
		}
	}

	if callback.SenderName == "" {
		if senderParseErr == nil {
			callback.SenderName = userPhoneFromJID(senderJID)
		} else {
			callback.SenderName = sanitizeConversationFallback(event.Metadata["from"])
		}
	}

	if event.Metadata["is_view_once"] == "true" {
		callback.ViewOnce = true
	}

	if participant := deriveParticipantPhone(event.Metadata, senderJID, senderParseErr); participant != "" {
		callback.ParticipantPhone = participant
	}
	if participantLID := normalizeLID(event.Metadata["sender_alt"]); participantLID != "" {
		callback.ParticipantLid = participantLID
	} else if participantLID := normalizeLID(event.Metadata["sender"]); participantLID != "" {
		callback.ParticipantLid = participantLID
	}

	callback.IsGroup = callback.IsGroup ||
		strings.HasSuffix(callback.Phone, "-group") ||
		strings.HasSuffix(callback.Phone, "@g.us") ||
		strings.HasSuffix(callback.Phone, "@lid")

	if callback.Broadcast {
		callback.IsGroup = false
	}

	if raw := event.Metadata["external_ad_reply"]; raw != "" {
		var external struct {
			Title                 string `json:"title"`
			Body                  string `json:"body"`
			MediaType             int    `json:"mediaType"`
			ThumbnailURL          string `json:"thumbnailUrl"`
			SourceType            string `json:"sourceType"`
			SourceID              string `json:"sourceId"`
			SourceURL             string `json:"sourceUrl"`
			ContainsAutoReply     bool   `json:"containsAutoReply"`
			RenderLargerThumbnail bool   `json:"renderLargerThumbnail"`
			ShowAdAttribution     bool   `json:"showAdAttribution"`
		}
		if err := json.Unmarshal([]byte(raw), &external); err == nil {
			callback.ExternalAdReply = &ExternalAdReplyContent{
				Title:                 external.Title,
				Body:                  external.Body,
				MediaType:             external.MediaType,
				ThumbnailURL:          external.ThumbnailURL,
				SourceType:            external.SourceType,
				SourceID:              external.SourceID,
				SourceURL:             external.SourceURL,
				ContainsAutoReply:     external.ContainsAutoReply,
				RenderLargerThumbnail: external.RenderLargerThumbnail,
				ShowAdAttribution:     external.ShowAdAttribution,
			}
		}
	}

	if undecryptable.UnavailableType == events.UnavailableTypeViewOnce {
		callback.ViewOnce = true
	}

	if provider != nil && callback.SenderPhoto == "" && senderParseErr == nil {
		if photo := provider.ContactPhoto(ctx, senderJID); photo != "" {
			callback.SenderPhoto = photo
		}
	}

	if !hasMessageContent(callback) {
		callback.WaitingMessage = true
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
			ViewOnce: audio.GetViewOnce(),
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
		reactionBy := extractPhoneNumber(event.Metadata["from"])
		if reactionBy == "" {
			reactionBy = sanitizeConversationFallback(event.Metadata["from"])
		}

		key := reaction.GetKey()
		remoteJID, remoteErr := parseJID(key.GetRemoteJID())
		msgRef := &MessageRef{
			MessageID: key.GetID(),
			FromMe:    key.GetFromMe(),
		}

		if remoteErr == nil {
			msgRef.Phone = conversationIdentifierFromJID(remoteJID)
		} else {
			msgRef.Phone = sanitizeConversationFallback(key.GetRemoteJID())
		}

		if key.GetParticipant() != "" {
			participant := extractPhoneNumber(key.GetParticipant())
			if participant == "" {
				participant = sanitizeConversationFallback(key.GetParticipant())
			}
			if participant != "" {
				msgRef.Participant = &participant
			}
		}

		callback.Reaction = &ReactionContent{
			Value:             reaction.GetText(),
			Time:              reaction.GetSenderTimestampMS(),
			ReactionBy:        reactionBy,
			ReferencedMessage: msgRef,
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

	if invite := msg.GetGroupInviteMessage(); invite != nil {
		if callback.Code == "" {
			callback.Code = invite.GetInviteCode()
		}
		return nil
	}

	return nil
}

func messagePayloadKinds(msg *waE2E.Message) []string {
	var kinds []string

	if msg.GetConversation() != "" {
		kinds = append(kinds, "conversation")
	}
	if msg.GetExtendedTextMessage() != nil {
		kinds = append(kinds, "extended_text")
	}
	if msg.GetImageMessage() != nil {
		kinds = append(kinds, "image")
	}
	if msg.GetVideoMessage() != nil {
		kinds = append(kinds, "video")
	}
	if msg.GetAudioMessage() != nil {
		kinds = append(kinds, "audio")
	}
	if msg.GetDocumentMessage() != nil {
		kinds = append(kinds, "document")
	}
	if msg.GetStickerMessage() != nil {
		kinds = append(kinds, "sticker")
	}
	if msg.GetLocationMessage() != nil {
		kinds = append(kinds, "location")
	}
	if msg.GetContactMessage() != nil {
		kinds = append(kinds, "contact")
	}
	if msg.GetReactionMessage() != nil {
		kinds = append(kinds, "reaction")
	}
	if msg.GetPollCreationMessage() != nil {
		kinds = append(kinds, "poll_creation")
	}
	if msg.GetPollUpdateMessage() != nil {
		kinds = append(kinds, "poll_update")
	}
	if msg.GetButtonsResponseMessage() != nil {
		kinds = append(kinds, "buttons_response")
	}
	if msg.GetListResponseMessage() != nil {
		kinds = append(kinds, "list_response")
	}
	if msg.GetTemplateButtonReplyMessage() != nil {
		kinds = append(kinds, "template_button_reply")
	}
	if msg.GetViewOnceMessage() != nil {
		kinds = append(kinds, "view_once")
	}
	if msg.GetProtocolMessage() != nil {
		kinds = append(kinds, "protocol_message")
	}
	if msg.GetKeepInChatMessage() != nil {
		kinds = append(kinds, "keep_in_chat")
	}
	if msg.GetContactsArrayMessage() != nil {
		kinds = append(kinds, "contacts_array")
	}
	if msg.GetLiveLocationMessage() != nil {
		kinds = append(kinds, "live_location")
	}
	if msg.GetProductMessage() != nil {
		kinds = append(kinds, "product")
	}
	if msg.GetOrderMessage() != nil {
		kinds = append(kinds, "order")
	}
	if msg.GetInvoiceMessage() != nil {
		kinds = append(kinds, "invoice")
	}
	if msg.GetGroupInviteMessage() != nil {
		kinds = append(kinds, "group_invite")
	}
	if msg.GetButtonsMessage() != nil {
		kinds = append(kinds, "buttons")
	}
	if msg.GetListMessage() != nil {
		kinds = append(kinds, "list")
	}
	if msg.GetTemplateMessage() != nil {
		kinds = append(kinds, "template")
	}
	if msg.GetSendPaymentMessage() != nil {
		kinds = append(kinds, "send_payment")
	}
	if msg.GetRequestPaymentMessage() != nil {
		kinds = append(kinds, "request_payment")
	}
	if msg.GetCancelPaymentRequestMessage() != nil {
		kinds = append(kinds, "cancel_payment_request")
	}
	if msg.GetDeclinePaymentRequestMessage() != nil {
		kinds = append(kinds, "decline_payment_request")
	}
	if msg.GetCall() != nil {
		kinds = append(kinds, "call")
	}
	if msg.GetChat() != nil {
		kinds = append(kinds, "chat")
	}
	if msg.GetSenderKeyDistributionMessage() != nil {
		kinds = append(kinds, "sender_key_distribution")
	}
	if msg.GetDeviceSentMessage() != nil {
		kinds = append(kinds, "device_sent")
	}
	if msg.GetEditedMessage() != nil {
		kinds = append(kinds, "edited_message")
	}
	if msg.GetPollCreationMessageV2() != nil {
		kinds = append(kinds, "poll_creation_v2")
	}
	if msg.GetPollCreationMessageV3() != nil {
		kinds = append(kinds, "poll_creation_v3")
	}
	if msg.GetScheduledCallCreationMessage() != nil {
		kinds = append(kinds, "scheduled_call_creation")
	}
	if msg.GetPinInChatMessage() != nil {
		kinds = append(kinds, "pin_in_chat")
	}

	return kinds
}

func (t *Transformer) transformReceipt(ctx context.Context, logger *slog.Logger, event *types.InternalEvent) (*MessageStatusCallback, error) {
	receiptEvent, ok := event.RawPayload.(*events.Receipt)
	if !ok {
		return nil, fmt.Errorf("invalid receipt payload type")
	}

	status := mapReceiptStatus(receiptEvent)

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

	// Parse and set PhoneDevice from chat_device metadata
	if chatDevice, ok := event.Metadata["chat_device"]; ok && chatDevice != "" {
		if device, err := strconv.Atoi(chatDevice); err == nil {
			callback.PhoneDevice = device
		}
	}

	// Derive participant from metadata
	var participant string
	if pn := strings.TrimSpace(event.Metadata["sender_pn"]); pn != "" {
		if pnJID, err := parseJID(pn); err == nil {
			switch pnJID.Server {
			case watypes.DefaultUserServer, watypes.LegacyUserServer:
				participant = defaultUserServerJID(pnJID)
			default:
				participant = pnJID.String()
			}
		} else {
			participant = pn
		}
	}
	if participant == "" {
		if senderAlt, ok := event.Metadata["sender_alt"]; ok && senderAlt != "" {
			// Prefer sender_alt - normalize via parseJID
			if senderAltJID, err := parseJID(senderAlt); err == nil {
				switch senderAltJID.Server {
				case watypes.DefaultUserServer, watypes.LegacyUserServer:
					if formatted := defaultUserServerJID(senderAltJID); formatted != "" {
						participant = formatted
					} else {
						participant = sanitizeConversationFallback(senderAlt)
					}
				default:
					participant = sanitizeConversationFallback(senderAltJID.String())
				}
			} else {
				participant = sanitizeConversationFallback(senderAlt)
			}
		}
	}
	if participant == "" {
		if senderJID, ok := event.Metadata["sender_jid"]; ok && senderJID != "" {
			// Fall back to sender_jid
			if parsed, err := parseJID(senderJID); err == nil {
				switch parsed.Server {
				case watypes.DefaultUserServer, watypes.LegacyUserServer:
					if formatted := defaultUserServerJID(parsed); formatted != "" {
						participant = formatted
					} else {
						participant = sanitizeConversationFallback(senderJID)
					}
				default:
					participant = sanitizeConversationFallback(parsed.String())
				}
			} else {
				participant = sanitizeConversationFallback(senderJID)
			}
		}
	}
	if participant == "" {
		if sender, ok := event.Metadata["sender"]; ok && sender != "" {
			// Fall back to sender
			if parsed, err := parseJID(sender); err == nil {
				switch parsed.Server {
				case watypes.DefaultUserServer, watypes.LegacyUserServer:
					if formatted := defaultUserServerJID(parsed); formatted != "" {
						participant = formatted
					} else {
						participant = sanitizeConversationFallback(sender)
					}
				default:
					participant = sanitizeConversationFallback(parsed.String())
				}
			} else {
				participant = sanitizeConversationFallback(sender)
			}
		}
	}
	if participant == "" {
		if storeJID := strings.TrimSpace(event.Metadata["store_jid"]); storeJID != "" {
			if parsed, err := parseJID(storeJID); err == nil {
				switch parsed.Server {
				case watypes.DefaultUserServer, watypes.LegacyUserServer:
					if formatted := defaultUserServerJID(parsed); formatted != "" {
						participant = formatted
					}
				}
			}
		}
	}
	if participant != "" {
		callback.Participant = participant
	}

	// Parse and set ParticipantDevice from sender_device metadata
	if senderDevice, ok := event.Metadata["sender_device"]; ok && senderDevice != "" {
		if device, err := strconv.Atoi(senderDevice); err == nil {
			callback.ParticipantDevice = device
		}
	}

	if callback.IsGroup {
		if participantJID := deriveParticipantJID(event.Metadata); participantJID != "" {
			callback.Participant = participantJID
		}
	}

	// Re-evaluate IsGroup after setting Phone
	// IsGroup should be true if metadata flag is true OR the phone ends with group markers
	callback.IsGroup = event.Metadata["is_group"] == "true" ||
		strings.HasSuffix(callback.Phone, "-group") ||
		strings.HasSuffix(callback.Phone, "@g.us") ||
		strings.HasSuffix(callback.Phone, "@lid")

	return callback, nil
}

func mapReceiptStatus(receiptEvent *events.Receipt) string {
	if receiptEvent == nil {
		return "SENT"
	}

	switch receiptEvent.Type {
	case watypes.ReceiptTypeReadSelf:
		return "READ_BY_ME"
	case watypes.ReceiptTypeRead:
		if receiptEvent.IsFromMe {
			return "READ_BY_ME"
		}
		return "READ"
	case watypes.ReceiptTypePlayedSelf:
		return "PLAYED_BY_ME"
	case watypes.ReceiptTypePlayed:
		if receiptEvent.IsFromMe {
			return "PLAYED_BY_ME"
		}
		return "PLAYED"
	case watypes.ReceiptTypeDelivered:
		if receiptEvent.IsFromMe {
			return "SENT"
		}
		return "RECEIVED"
	case watypes.ReceiptTypeSender:
		return "SENT"
	default:
		return "SENT"
	}
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
	senderJID, senderErr := parseJID(event.Metadata["sender"])
	callback := &PresenceChatCallback{
		Type:       "PresenceChatCallback",
		Phone:      normalizeConversationPhone(event.Metadata, chatJID, chatErr),
		Status:     status,
		InstanceID: event.InstanceID.String(),
	}

	isGroup := event.Metadata["is_group"] == "true" ||
		strings.HasSuffix(callback.Phone, "-group") ||
		strings.HasSuffix(callback.Phone, "@g.us") ||
		(chatErr == nil && chatJID.Server == watypes.GroupServer)
	callback.IsGroup = isGroup

	if isGroup {
		if chatName := strings.TrimSpace(event.Metadata["chat_name"]); chatName != "" {
			callback.ChatName = chatName
		}
		if chatPhoto := strings.TrimSpace(event.Metadata["chat_photo"]); chatPhoto != "" {
			callback.Photo = chatPhoto
		}
		if senderName := strings.TrimSpace(event.Metadata["sender_name"]); senderName != "" {
			callback.SenderName = senderName
		}
		if senderPhoto := strings.TrimSpace(event.Metadata["sender_photo"]); senderPhoto != "" {
			callback.SenderPhoto = senderPhoto
		}

		participantPhone := strings.TrimSpace(event.Metadata["sender_pn"])
		if participantPhone != "" {
			if pnJID, err := parseJID(participantPhone); err == nil {
				participantPhone = userPhoneFromJID(pnJID)
			} else {
				participantPhone = sanitizeConversationFallback(participantPhone)
			}
		}
		if participantPhone == "" {
			if senderErr == nil {
				participantPhone = userPhoneFromJID(senderJID)
			} else if candidate := strings.TrimSpace(event.Metadata["sender"]); candidate != "" {
				participantPhone = sanitizeConversationFallback(candidate)
			}
		}
		if participantPhone != "" {
			if at := strings.IndexRune(participantPhone, '@'); at >= 0 {
				participantPhone = participantPhone[:at]
			}
			participantPhone = sanitizeUserComponent(participantPhone)
			callback.ParticipantPhone = participantPhone
			callback.Participant = participantPhone
		}

		if participantLID := normalizeLID(event.Metadata["sender_alt"]); participantLID != "" {
			callback.ParticipantLid = participantLID
		} else if participantLID := normalizeLID(event.Metadata["sender"]); participantLID != "" {
			callback.ParticipantLid = participantLID
		}
	} else {
		// ensure optional fields remain absent for direct chats
		callback.Participant = ""
		callback.ParticipantPhone = ""
		callback.ParticipantLid = ""
		callback.ChatName = ""
		callback.Photo = ""
		callback.SenderName = ""
		callback.SenderPhoto = ""
		callback.IsGroup = false
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
	if alt := metadata["chat_alt"]; alt != "" {
		if parsed, err := watypes.ParseJID(alt); err == nil {
			return conversationIdentifierFromJID(parsed)
		}
		if normalized := normalizeLID(alt); normalized != "" {
			return normalized
		}
		if fallback := sanitizeConversationFallback(alt); fallback != "" {
			return fallback
		}
	}

	if parseErr == nil {
		if chat.Server == watypes.HiddenUserServer {
			if alt := metadata["recipient_alt"]; alt != "" {
				if altJID, err := watypes.ParseJID(alt); err == nil {
					chat = altJID
				} else {
					return sanitizeConversationFallback(alt)
				}
			} else if normalized := normalizeLID(chat.String()); normalized != "" {
				return normalized
			}
		}
		return conversationIdentifierFromJID(chat)
	}

	if alt := metadata["recipient_alt"]; alt != "" {
		if parsed, err := watypes.ParseJID(alt); err == nil {
			return conversationIdentifierFromJID(parsed)
		}
		if normalized := normalizeLID(alt); normalized != "" {
			return normalized
		}
		if fallback := sanitizeConversationFallback(alt); fallback != "" {
			return fallback
		}
	}

	return sanitizeConversationFallback(metadata["chat"])
}

func deriveChatLID(metadata map[string]string, chat watypes.JID, parseErr error) *string {
	if alt := metadata["chat_alt"]; alt != "" {
		if normalized := normalizeLID(alt); normalized != "" {
			return stringPtr(normalized)
		}
		return stringPtr(alt)
	}

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
	if pn := strings.TrimSpace(metadata["sender_pn"]); pn != "" {
		if parsed, err := watypes.ParseJID(pn); err == nil {
			return userPhoneFromJID(parsed)
		}
		if fallback := sanitizeConversationFallback(pn); fallback != "" {
			return fallback
		}
	}

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

func deriveParticipantJID(metadata map[string]string) string {
	if pn := strings.TrimSpace(metadata["sender_pn"]); pn != "" {
		if jid, err := parseJID(pn); err == nil {
			switch jid.Server {
			case watypes.DefaultUserServer, watypes.LegacyUserServer:
				return defaultUserServerJID(jid)
			default:
				return jid.String()
			}
		}
		if sanitized := sanitizeConversationFallback(pn); sanitized != "" {
			return sanitized
		}
	}

	candidates := []string{
		metadata["sender_alt"],
		metadata["sender"],
		metadata["sender_jid"],
		metadata["message_sender"],
	}

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if jid, err := parseJID(candidate); err == nil {
			switch jid.Server {
			case watypes.HiddenUserServer:
				if normalized := normalizeLID(jid.String()); normalized != "" {
					return normalized
				}
				user := sanitizeUserComponent(jid.User)
				if user == "" {
					user = jid.User
				}
				return user + "@" + watypes.HiddenUserServer
			case watypes.DefaultUserServer, watypes.LegacyUserServer:
				return defaultUserServerJID(jid)
			case watypes.GroupServer:
				user := sanitizeUserComponent(jid.User)
				if user == "" {
					user = jid.User
				}
				return user + "-group"
			default:
				user := sanitizeUserComponent(jid.User)
				if user == "" {
					user = jid.User
				}
				if jid.Server == "" {
					return user
				}
				return user + "@" + jid.Server
			}
		}

		if strings.Contains(candidate, "@"+watypes.HiddenUserServer) {
			if normalized := normalizeLID(candidate); normalized != "" {
				return normalized
			}
		}

		if sanitized := sanitizeConversationFallback(candidate); sanitized != "" {
			return sanitized
		}
	}

	return ""
}

func isStatusBroadcastReference(value string) bool {
	if value == "" {
		return false
	}
	if parsed, err := parseJID(value); err == nil {
		if parsed == watypes.StatusBroadcastJID {
			return true
		}
	}
	if value == watypes.StatusBroadcastJID.String() {
		return true
	}
	if sanitizeConversationFallback(value) == "status" {
		return true
	}
	return false
}

func hasMessageContent(callback *ReceivedCallback) bool {
	if callback == nil {
		return false
	}
	if callback.Text != nil || callback.Image != nil || callback.Audio != nil || callback.Video != nil || callback.Document != nil {
		return true
	}
	if callback.Location != nil || callback.Contact != nil || callback.Sticker != nil || callback.Reaction != nil || callback.Poll != nil || callback.PollVote != nil {
		return true
	}
	if callback.ButtonsResponseMessage != nil || callback.ListResponseMessage != nil || callback.HydratedTemplate != nil || callback.ButtonsMessage != nil || callback.PixKeyMessage != nil {
		return true
	}
	if callback.CarouselMessage != nil || callback.Product != nil || callback.Order != nil || callback.ReviewAndPay != nil || callback.ReviewOrder != nil {
		return true
	}
	if callback.NewsletterAdminInvite != nil || callback.PinMessage != nil || callback.Event != nil || callback.EventResponse != nil {
		return true
	}
	if callback.RequestPayment != nil || callback.SendPayment != nil || callback.ExternalAdReply != nil {
		return true
	}
	if callback.Notification != "" || len(callback.NotificationParameters) > 0 || callback.CallID != "" || callback.Code != "" || callback.RequestMethod != "" {
		return true
	}
	if callback.RevokedMessageID != "" || callback.ReferenceMessageID != "" {
		return true
	}
	if len(callback.Mentioned) > 0 {
		return true
	}
	return false
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
	case watypes.HiddenUserServer:
		if normalized := normalizeLID(jid.String()); normalized != "" {
			return normalized
		}
		if user == "" {
			return jid.User + "@" + watypes.HiddenUserServer
		}
		return user + "@" + watypes.HiddenUserServer
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

func defaultUserServerJID(jid watypes.JID) string {
	user := sanitizeUserComponent(jid.User)
	if user == "" {
		user = jid.User
	}
	if user == "" {
		return ""
	}
	return user + "@" + watypes.DefaultUserServer
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
		if server == watypes.HiddenUserServer {
			if normalized := normalizeLID(value); normalized != "" {
				return normalized
			}
		}
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
