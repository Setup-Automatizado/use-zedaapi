package whatsmeow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/api/internal/events/transform"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

// Transformer converts whatsmeow events to InternalEvent format.
// This is the source transformer for WhatsApp events received via whatsmeow library.
type Transformer struct {
	instanceID uuid.UUID
}

// NewTransformer creates a new WhatsmeowTransformer for the given instance.
func NewTransformer(instanceID uuid.UUID) *Transformer {
	return &Transformer{
		instanceID: instanceID,
	}
}

// SourceLib returns the source library identifier.
func (t *Transformer) SourceLib() types.SourceLib {
	return types.SourceLibWhatsmeow
}

// SupportsEvent returns true if this transformer can handle the given event type.
func (t *Transformer) SupportsEvent(eventType reflect.Type) bool {
	switch eventType {
	case reflect.TypeOf(&events.Message{}),
		reflect.TypeOf(&events.Receipt{}),
		reflect.TypeOf(&events.ChatPresence{}),
		reflect.TypeOf(&events.Presence{}),
		reflect.TypeOf(&events.Connected{}),
		reflect.TypeOf(&events.Disconnected{}),
		reflect.TypeOf(&events.JoinedGroup{}),
		reflect.TypeOf(&events.GroupInfo{}),
		reflect.TypeOf(&events.Picture{}):
		return true
	default:
		return false
	}
}

// Transform converts a whatsmeow event to InternalEvent format.
func (t *Transformer) Transform(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "whatsmeow_transformer"),
		slog.String("instance_id", t.instanceID.String()),
	)

	// Route to specific transformation based on event type
	switch evt := rawEvent.(type) {
	case *events.Message:
		return t.transformMessage(ctx, logger, evt)
	case *events.Receipt:
		return t.transformReceipt(ctx, logger, evt)
	case *events.ChatPresence:
		return t.transformChatPresence(ctx, logger, evt)
	case *events.Presence:
		return t.transformPresence(ctx, logger, evt)
	case *events.Connected:
		return t.transformConnected(ctx, logger, evt)
	case *events.Disconnected:
		return t.transformDisconnected(ctx, logger, evt)
	case *events.JoinedGroup:
		return t.transformJoinedGroup(ctx, logger, evt)
	case *events.GroupInfo:
		return t.transformGroupInfo(ctx, logger, evt)
	case *events.Picture:
		return t.transformPicture(ctx, logger, evt)
	default:
		logger.Debug("unsupported event type",
			slog.String("event_type", fmt.Sprintf("%T", rawEvent)),
		)
		return nil, transform.ErrUnsupportedEvent
	}
}

// transformMessage converts a whatsmeow Message event to InternalEvent.
func (t *Transformer) transformMessage(ctx context.Context, logger *slog.Logger, msg *events.Message) (*types.InternalEvent, error) {
	eventID := uuid.New()

	// Unwrap the raw message to get the actual content
	msg.UnwrapRaw()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "message",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: msg,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add metadata
	event.Metadata["message_id"] = msg.Info.ID
	event.Metadata["from"] = msg.Info.Sender.String()
	event.Metadata["chat"] = msg.Info.Chat.String()
	event.Metadata["from_me"] = fmt.Sprintf("%t", msg.Info.IsFromMe)
	event.Metadata["is_group"] = fmt.Sprintf("%t", msg.Info.IsGroup)
	event.Metadata["timestamp"] = fmt.Sprintf("%d", msg.Info.Timestamp.Unix())

	// Add MessageInfo fields
	if msg.Info.PushName != "" {
		event.Metadata["push_name"] = msg.Info.PushName
	}
	if msg.Info.VerifiedName != nil {
		verifiedNameJSON, _ := json.Marshal(msg.Info.VerifiedName)
		event.Metadata["verified_name"] = string(verifiedNameJSON)
	}
	if msg.Info.Category != "" {
		event.Metadata["category"] = msg.Info.Category
	}
	if msg.Info.ServerID != 0 {
		event.Metadata["server_id"] = fmt.Sprintf("%d", msg.Info.ServerID)
	}
	event.Metadata["multicast"] = fmt.Sprintf("%t", msg.Info.Multicast)
	if msg.Info.MediaType != "" {
		event.Metadata["media_type_info"] = msg.Info.MediaType
	}

	// Add edit attributes
	if msg.Info.Edit != "" {
		switch msg.Info.Edit {
		case "1":
			event.Metadata["edit_attribute"] = "message_edit"
		case "2":
			event.Metadata["edit_attribute"] = "pin_in_chat"
		case "3":
			event.Metadata["edit_attribute"] = "admin_edit"
		case "7":
			event.Metadata["edit_attribute"] = "sender_revoke"
		case "8":
			event.Metadata["edit_attribute"] = "admin_revoke"
		default:
			event.Metadata["edit_attribute"] = string(msg.Info.Edit)
		}
	}

	// Add thread/reply metadata
	if msg.Info.MsgMetaInfo.TargetID != "" {
		event.Metadata["reply_to_message_id"] = msg.Info.MsgMetaInfo.TargetID
	}
	if !msg.Info.MsgMetaInfo.TargetSender.IsEmpty() {
		event.Metadata["reply_to_sender"] = msg.Info.MsgMetaInfo.TargetSender.String()
	}
	if !msg.Info.MsgMetaInfo.TargetChat.IsEmpty() {
		event.Metadata["reply_to_chat"] = msg.Info.MsgMetaInfo.TargetChat.String()
	}
	if msg.Info.MsgMetaInfo.ThreadMessageID != "" {
		event.Metadata["thread_message_id"] = msg.Info.MsgMetaInfo.ThreadMessageID
	}
	if !msg.Info.MsgMetaInfo.ThreadMessageSenderJID.IsEmpty() {
		event.Metadata["thread_message_sender"] = msg.Info.MsgMetaInfo.ThreadMessageSenderJID.String()
	}

	// Add DeviceSentMeta
	if msg.Info.DeviceSentMeta != nil {
		if msg.Info.DeviceSentMeta.Phash != "" {
			event.Metadata["device_sent_phash"] = msg.Info.DeviceSentMeta.Phash
		}
	}

	// Add MessageSource fields
	if msg.Info.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(msg.Info.MessageSource.AddressingMode)
	}
	if !msg.Info.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = msg.Info.MessageSource.SenderAlt.String()
	}
	if !msg.Info.MessageSource.RecipientAlt.IsEmpty() {
		event.Metadata["recipient_alt"] = msg.Info.MessageSource.RecipientAlt.String()
	}
	if !msg.Info.MessageSource.BroadcastListOwner.IsEmpty() {
		event.Metadata["broadcast_list_owner"] = msg.Info.MessageSource.BroadcastListOwner.String()
	}

	// Add message type flags
	if msg.IsEphemeral {
		event.Metadata["is_ephemeral"] = "true"
	}
	if msg.IsViewOnce {
		event.Metadata["is_view_once"] = "true"
	}
	if msg.IsEdit {
		event.Metadata["is_edit"] = "true"
	}

	// Handle ProtocolMessage (revoke, edits, etc)
	if protocolMsg := msg.Message.GetProtocolMessage(); protocolMsg != nil {
		msgType := protocolMsg.GetType()
		switch msgType {
		case waE2E.ProtocolMessage_REVOKE:
			// Message revocation/deletion
			event.Metadata["is_revoke"] = "true"
			if key := protocolMsg.GetKey(); key != nil {
				if revokedID := key.GetID(); revokedID != "" {
					event.Metadata["revoked_message_id"] = revokedID
				}
			}
		case waE2E.ProtocolMessage_MESSAGE_EDIT:
			// Message edit (already handled by msg.IsEdit, but keep for completeness)
			event.Metadata["is_edit"] = "true"
		}
	}

	// Extract media information if present
	hasMedia, mediaInfo := t.extractMediaInfo(msg.Message)
	if hasMedia {
		event.HasMedia = true
		event.MediaKey = mediaInfo.MediaKey
		event.DirectPath = mediaInfo.DirectPath
		event.FileSHA256 = mediaInfo.FileSHA256
		event.FileEncSHA256 = mediaInfo.FileEncSHA256
		event.MediaType = mediaInfo.MediaType
		event.MimeType = mediaInfo.MimeType
		event.FileLength = mediaInfo.FileLength
		// Media metadata
		event.MediaIsGIF = mediaInfo.IsGIF
		event.MediaIsAnimated = mediaInfo.IsAnimated
		event.MediaWidth = mediaInfo.Width
		event.MediaHeight = mediaInfo.Height
		event.MediaWaveform = mediaInfo.Waveform
	}

	// Extract ContextInfo (quotes, mentions, forwards, ephemeral)
	if ctxInfo := msg.Message.GetExtendedTextMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetImageMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetVideoMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetAudioMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetDocumentMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetStickerMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetContactMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	} else if ctxInfo := msg.Message.GetLocationMessage().GetContextInfo(); ctxInfo != nil {
		t.extractContextInfo(ctxInfo, event)
	}

	logger.InfoContext(ctx, "transformed message event",
		slog.String("event_id", eventID.String()),
		slog.String("message_id", msg.Info.ID),
		slog.Bool("has_media", hasMedia),
		slog.String("media_type", event.MediaType),
	)

	return event, nil
}

// transformReceipt converts a whatsmeow Receipt event to InternalEvent.
func (t *Transformer) transformReceipt(ctx context.Context, logger *slog.Logger, receipt *events.Receipt) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "receipt",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: receipt,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add metadata
	event.Metadata["chat"] = receipt.Chat.String()
	event.Metadata["sender"] = receipt.Sender.String()
	event.Metadata["receipt_type"] = string(receipt.Type)
	event.Metadata["timestamp"] = fmt.Sprintf("%d", receipt.Timestamp.Unix())

	// Add MessageSource fields
	if receipt.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(receipt.MessageSource.AddressingMode)
	}
	if !receipt.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = receipt.MessageSource.SenderAlt.String()
	}
	if !receipt.MessageSource.RecipientAlt.IsEmpty() {
		event.Metadata["recipient_alt"] = receipt.MessageSource.RecipientAlt.String()
	}

	// Add message IDs
	if len(receipt.MessageIDs) > 0 {
		messageIDsJSON, _ := json.Marshal(receipt.MessageIDs)
		event.Metadata["message_ids"] = string(messageIDsJSON)
	}

	logger.InfoContext(ctx, "transformed receipt event",
		slog.String("event_id", eventID.String()),
		slog.String("receipt_type", string(receipt.Type)),
		slog.Int("message_count", len(receipt.MessageIDs)),
	)

	return event, nil
}

// transformChatPresence converts a whatsmeow ChatPresence event to InternalEvent.
func (t *Transformer) transformChatPresence(ctx context.Context, logger *slog.Logger, presence *events.ChatPresence) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "chat_presence",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: presence,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add metadata
	event.Metadata["chat"] = presence.Chat.String()
	event.Metadata["sender"] = presence.Sender.String()
	event.Metadata["state"] = string(presence.State)
	event.Metadata["media"] = string(presence.Media)

	// Add MessageSource fields
	if presence.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(presence.MessageSource.AddressingMode)
	}
	if !presence.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = presence.MessageSource.SenderAlt.String()
	}
	if !presence.MessageSource.RecipientAlt.IsEmpty() {
		event.Metadata["recipient_alt"] = presence.MessageSource.RecipientAlt.String()
	}

	logger.InfoContext(ctx, "transformed chat presence event",
		slog.String("event_id", eventID.String()),
		slog.String("state", string(presence.State)),
		slog.String("media", string(presence.Media)),
	)

	return event, nil
}

// transformPresence converts a whatsmeow Presence event to InternalEvent.
func (t *Transformer) transformPresence(ctx context.Context, logger *slog.Logger, presence *events.Presence) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "presence",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: presence,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add metadata
	event.Metadata["from"] = presence.From.String()
	event.Metadata["unavailable"] = fmt.Sprintf("%t", presence.Unavailable)
	if !presence.LastSeen.IsZero() {
		event.Metadata["last_seen"] = fmt.Sprintf("%d", presence.LastSeen.Unix())
	}

	logger.InfoContext(ctx, "transformed presence event",
		slog.String("event_id", eventID.String()),
		slog.Bool("unavailable", presence.Unavailable),
	)

	return event, nil
}

// transformConnected converts a whatsmeow Connected event to InternalEvent.
func (t *Transformer) transformConnected(ctx context.Context, logger *slog.Logger, connected *events.Connected) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "connected",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: connected,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	logger.InfoContext(ctx, "transformed connected event",
		slog.String("event_id", eventID.String()),
	)

	return event, nil
}

// transformDisconnected converts a whatsmeow Disconnected event to InternalEvent.
func (t *Transformer) transformDisconnected(ctx context.Context, logger *slog.Logger, disconnected *events.Disconnected) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "disconnected",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: disconnected,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	logger.InfoContext(ctx, "transformed disconnected event",
		slog.String("event_id", eventID.String()),
	)

	return event, nil
}

// MediaInfo holds extracted media information from a message.
type MediaInfo struct {
	MediaKey      string
	DirectPath    string
	FileSHA256    *string
	FileEncSHA256 *string
	MediaType     string
	MimeType      *string
	FileLength    *int64
	// Media metadata
	IsGIF      bool
	IsAnimated bool
	Width      int
	Height     int
	Waveform   []byte
}

// extractMediaInfo extracts media information from a WhatsApp message.
// Returns (hasMedia bool, mediaInfo MediaInfo).
func (t *Transformer) extractMediaInfo(msg *waE2E.Message) (bool, MediaInfo) {
	var info MediaInfo

	// Check for image
	if img := msg.GetImageMessage(); img != nil {
		info.MediaType = "image"
		info.MediaKey = base64.StdEncoding.EncodeToString(img.GetMediaKey())
		info.DirectPath = img.GetDirectPath()
		if sha256 := img.GetFileSHA256(); len(sha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(sha256)
			info.FileSHA256 = &encoded
		}
		if encSha256 := img.GetFileEncSHA256(); len(encSha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(encSha256)
			info.FileEncSHA256 = &encoded
		}
		if mime := img.GetMimetype(); mime != "" {
			info.MimeType = &mime
		}
		if length := img.GetFileLength(); length > 0 {
			lengthInt64 := int64(length)
			info.FileLength = &lengthInt64
		}
		// Extract media metadata
		// For images, detect GIF from mimetype since proto doesn't have isGif field
		if mime := img.GetMimetype(); mime == "image/gif" {
			info.IsGIF = true
			info.IsAnimated = true // GIFs are inherently animated
		}
		if width := img.GetWidth(); width > 0 {
			info.Width = int(width)
		}
		if height := img.GetHeight(); height > 0 {
			info.Height = int(height)
		}
		return true, info
	}

	// Check for video
	if video := msg.GetVideoMessage(); video != nil {
		info.MediaType = "video"
		info.MediaKey = base64.StdEncoding.EncodeToString(video.GetMediaKey())
		info.DirectPath = video.GetDirectPath()
		if sha256 := video.GetFileSHA256(); len(sha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(sha256)
			info.FileSHA256 = &encoded
		}
		if encSha256 := video.GetFileEncSHA256(); len(encSha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(encSha256)
			info.FileEncSHA256 = &encoded
		}
		if mime := video.GetMimetype(); mime != "" {
			info.MimeType = &mime
		}
		if length := video.GetFileLength(); length > 0 {
			lengthInt64 := int64(length)
			info.FileLength = &lengthInt64
		}
		// Extract media metadata
		info.IsGIF = video.GetGifPlayback() // Video plays as GIF
		if width := video.GetWidth(); width > 0 {
			info.Width = int(width)
		}
		if height := video.GetHeight(); height > 0 {
			info.Height = int(height)
		}
		return true, info
	}

	// Check for audio
	if audio := msg.GetAudioMessage(); audio != nil {
		info.MediaType = "audio"
		info.MediaKey = base64.StdEncoding.EncodeToString(audio.GetMediaKey())
		info.DirectPath = audio.GetDirectPath()
		if sha256 := audio.GetFileSHA256(); len(sha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(sha256)
			info.FileSHA256 = &encoded
		}
		if encSha256 := audio.GetFileEncSHA256(); len(encSha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(encSha256)
			info.FileEncSHA256 = &encoded
		}
		if mime := audio.GetMimetype(); mime != "" {
			info.MimeType = &mime
		}
		if length := audio.GetFileLength(); length > 0 {
			lengthInt64 := int64(length)
			info.FileLength = &lengthInt64
		}
		// Extract waveform
		if waveform := audio.GetWaveform(); len(waveform) > 0 {
			info.Waveform = waveform
		}
		return true, info
	}

	// Check for document
	if doc := msg.GetDocumentMessage(); doc != nil {
		info.MediaType = "document"
		info.MediaKey = base64.StdEncoding.EncodeToString(doc.GetMediaKey())
		info.DirectPath = doc.GetDirectPath()
		if sha256 := doc.GetFileSHA256(); len(sha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(sha256)
			info.FileSHA256 = &encoded
		}
		if encSha256 := doc.GetFileEncSHA256(); len(encSha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(encSha256)
			info.FileEncSHA256 = &encoded
		}
		if mime := doc.GetMimetype(); mime != "" {
			info.MimeType = &mime
		}
		if length := doc.GetFileLength(); length > 0 {
			lengthInt64 := int64(length)
			info.FileLength = &lengthInt64
		}
		return true, info
	}

	// Check for sticker
	if sticker := msg.GetStickerMessage(); sticker != nil {
		info.MediaType = "sticker"
		info.MediaKey = base64.StdEncoding.EncodeToString(sticker.GetMediaKey())
		info.DirectPath = sticker.GetDirectPath()
		if sha256 := sticker.GetFileSHA256(); len(sha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(sha256)
			info.FileSHA256 = &encoded
		}
		if encSha256 := sticker.GetFileEncSHA256(); len(encSha256) > 0 {
			encoded := base64.StdEncoding.EncodeToString(encSha256)
			info.FileEncSHA256 = &encoded
		}
		if mime := sticker.GetMimetype(); mime != "" {
			info.MimeType = &mime
		}
		if length := sticker.GetFileLength(); length > 0 {
			lengthInt64 := int64(length)
			info.FileLength = &lengthInt64
		}
		// Extract media metadata
		info.IsAnimated = sticker.GetIsAnimated()
		if width := sticker.GetWidth(); width > 0 {
			info.Width = int(width)
		}
		if height := sticker.GetHeight(); height > 0 {
			info.Height = int(height)
		}
		return true, info
	}

	// No media found
	return false, info
}

// extractContextInfo extracts ContextInfo fields (quotes, mentions, forwards, ephemeral)
func (t *Transformer) extractContextInfo(ctxInfo *waE2E.ContextInfo, event *types.InternalEvent) {
	// Extract quoted message info
	if stanzaID := ctxInfo.GetStanzaID(); stanzaID != "" {
		event.QuotedMessageID = stanzaID
		event.Metadata["quoted_message_id"] = stanzaID
	}
	if participant := ctxInfo.GetParticipant(); participant != "" {
		event.QuotedSender = participant
		event.Metadata["quoted_sender"] = participant
	}
	if remoteJID := ctxInfo.GetRemoteJID(); remoteJID != "" {
		event.QuotedRemoteJID = remoteJID
		event.Metadata["quoted_remote_jid"] = remoteJID
	}

	// Extract quoted message content if present
	if quotedMsg := ctxInfo.GetQuotedMessage(); quotedMsg != nil {
		// Serialize quoted message to JSON for metadata
		if quotedJSON, err := json.Marshal(quotedMsg); err == nil {
			event.Metadata["quoted_message"] = string(quotedJSON)
		}
	}

	// Extract mentioned JIDs
	if mentionedJIDs := ctxInfo.GetMentionedJID(); len(mentionedJIDs) > 0 {
		event.MentionedJIDs = mentionedJIDs
		// Store as comma-separated string for metadata
		event.Metadata["mentioned_jids"] = strings.Join(mentionedJIDs, ",")
	}

	// Extract forward flag
	if isForwarded := ctxInfo.GetIsForwarded(); isForwarded {
		event.IsForwarded = true
		event.Metadata["is_forwarded"] = "true"
	}

	// Extract ephemeral expiry
	if ephemeralExpiry := ctxInfo.GetEphemeralSettingTimestamp(); ephemeralExpiry > 0 {
		event.EphemeralExpiry = ephemeralExpiry
		event.Metadata["ephemeral_expiry"] = fmt.Sprintf("%d", ephemeralExpiry)
	}
}

// transformJoinedGroup transforms a JoinedGroup event to InternalEvent format.
func (t *Transformer) transformJoinedGroup(ctx context.Context, logger *slog.Logger, joined *events.JoinedGroup) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "group_joined",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: joined,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add group info
	event.Metadata["group_id"] = joined.JID.String()
	if joined.Name != "" {
		event.Metadata["group_name"] = joined.Name
	}
	event.Metadata["reason"] = joined.Reason
	event.Metadata["type"] = joined.Type

	if joined.Sender != nil {
		event.Metadata["sender"] = joined.Sender.String()
	}
	if joined.SenderPN != nil {
		event.Metadata["sender_pn"] = joined.SenderPN.String()
	}
	if joined.Notify != "" {
		event.Metadata["notify"] = joined.Notify
	}
	if joined.CreateKey != "" {
		event.Metadata["create_key"] = joined.CreateKey
	}

	logger.InfoContext(ctx, "transformed joined group event",
		slog.String("event_id", eventID.String()),
		slog.String("group_id", joined.JID.String()),
	)

	return event, nil
}

// transformGroupInfo transforms a GroupInfo event to InternalEvent format.
func (t *Transformer) transformGroupInfo(ctx context.Context, logger *slog.Logger, info *events.GroupInfo) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "group_info",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: info,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add group info
	event.Metadata["group_id"] = info.JID.String()
	event.Metadata["timestamp"] = fmt.Sprintf("%d", info.Timestamp.Unix())

	if info.Sender != nil {
		event.Metadata["sender"] = info.Sender.String()
	}
	if info.SenderPN != nil {
		event.Metadata["sender_pn"] = info.SenderPN.String()
	}
	if info.Notify != "" {
		event.Metadata["notify"] = info.Notify
	}

	// Add specific change info
	if info.Name != nil {
		event.Metadata["name_change"] = info.Name.Name
		event.Metadata["name_set_at"] = fmt.Sprintf("%d", info.Name.NameSetAt.Unix())
		if !info.Name.NameSetBy.IsEmpty() {
			event.Metadata["name_set_by"] = info.Name.NameSetBy.String()
		}
	}
	if info.Topic != nil {
		event.Metadata["topic_change"] = info.Topic.Topic
		event.Metadata["topic_set_at"] = fmt.Sprintf("%d", info.Topic.TopicSetAt.Unix())
		if !info.Topic.TopicSetBy.IsEmpty() {
			event.Metadata["topic_set_by"] = info.Topic.TopicSetBy.String()
		}
	}
	if info.Locked != nil {
		event.Metadata["locked"] = fmt.Sprintf("%t", info.Locked.IsLocked)
	}
	if info.Announce != nil {
		event.Metadata["announce"] = fmt.Sprintf("%t", info.Announce.IsAnnounce)
	}

	logger.InfoContext(ctx, "transformed group info event",
		slog.String("event_id", eventID.String()),
		slog.String("group_id", info.JID.String()),
	)

	return event, nil
}

// transformPicture transforms a Picture event to InternalEvent format.
func (t *Transformer) transformPicture(ctx context.Context, logger *slog.Logger, picture *events.Picture) (*types.InternalEvent, error) {
	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "picture",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: picture,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	// Add picture info
	event.Metadata["jid"] = picture.JID.String()
	event.Metadata["author"] = picture.Author.String()
	event.Metadata["timestamp"] = fmt.Sprintf("%d", picture.Timestamp.Unix())
	event.Metadata["remove"] = fmt.Sprintf("%t", picture.Remove)

	if picture.PictureID != "" {
		event.Metadata["picture_id"] = picture.PictureID
	}

	logger.InfoContext(ctx, "transformed picture event",
		slog.String("event_id", eventID.String()),
		slog.String("jid", picture.JID.String()),
		slog.Bool("remove", picture.Remove),
	)

	return event, nil
}
