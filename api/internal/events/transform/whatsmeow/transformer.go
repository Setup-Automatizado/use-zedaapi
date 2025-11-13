package whatsmeow

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	whatsmeowcore "go.mau.fi/whatsmeow"
	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	"go.mau.fi/whatsmeow/api/internal/events/transform"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/proto/waE2E"
	watypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var debugProtoMarshalOpts = protojson.MarshalOptions{EmitUnpopulated: true}

type Transformer struct {
	instanceID uuid.UUID
	debug      bool
	pollStore  pollstore.Store
}

func NewTransformer(instanceID uuid.UUID, debug bool, store pollstore.Store) *Transformer {
	return &Transformer{
		instanceID: instanceID,
		debug:      debug,
		pollStore:  store,
	}
}

func (t *Transformer) SourceLib() types.SourceLib {
	return types.SourceLibWhatsmeow
}

// TODO: Adicionar novos eventos
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
		reflect.TypeOf(&events.Picture{}),
		reflect.TypeOf(&events.HistorySync{}),
		reflect.TypeOf(&events.UndecryptableMessage{}),
		reflect.TypeOf(&events.CallOffer{}),
		reflect.TypeOf(&events.CallOfferNotice{}),
		reflect.TypeOf(&events.CallRelayLatency{}),
		reflect.TypeOf(&events.CallTransport{}),
		reflect.TypeOf(&events.CallTerminate{}),
		reflect.TypeOf(&events.CallReject{}),
		reflect.TypeOf(&events.NewsletterJoin{}),
		reflect.TypeOf(&events.NewsletterLeave{}),
		reflect.TypeOf(&events.NewsletterMuteChange{}),
		reflect.TypeOf(&events.NewsletterLiveUpdate{}),
		reflect.TypeOf(&events.PushName{}),
		reflect.TypeOf(&events.BusinessName{}),
		reflect.TypeOf(&events.UserAbout{}):
		return true
	default:
		return false
	}
}

// TODO: Adicionar novos eventos
func (t *Transformer) Transform(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "whatsmeow_transformer"),
		slog.String("instance_id", t.instanceID.String()),
	)

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
	case *events.HistorySync:
		return t.transformHistorySync(ctx, logger, evt)
	case *events.UndecryptableMessage:
		return t.transformUndecryptable(ctx, logger, evt)
	case *events.CallOffer:
		return t.transformCallOffer(ctx, logger, evt)
	case *events.CallOfferNotice:
		return t.transformCallOfferNotice(ctx, logger, evt)
	case *events.CallRelayLatency:
		return t.transformCallRelayLatency(ctx, logger, evt)
	case *events.CallTransport:
		return t.transformCallTransport(ctx, logger, evt)
	case *events.CallTerminate:
		return t.transformCallTerminate(ctx, logger, evt)
	case *events.CallReject:
		return t.transformCallReject(ctx, logger, evt)
	case *events.NewsletterJoin:
		return t.transformNewsletterJoin(ctx, logger, evt)
	case *events.NewsletterLeave:
		return t.transformNewsletterLeave(ctx, logger, evt)
	case *events.NewsletterMuteChange:
		return t.transformNewsletterMuteChange(ctx, logger, evt)
	case *events.NewsletterLiveUpdate:
		return t.transformNewsletterLiveUpdate(ctx, logger, evt)
	case *events.PushName:
		return t.transformPushName(ctx, logger, evt)
	case *events.BusinessName:
		return t.transformBusinessName(ctx, logger, evt)
	case *events.UserAbout:
		return t.transformUserAbout(ctx, logger, evt)
	default:
		logger.Debug("unsupported event type",
			slog.String("event_type", fmt.Sprintf("%T", rawEvent)),
		)
		return nil, transform.ErrUnsupportedEvent
	}
}

func (t *Transformer) transformMessage(ctx context.Context, logger *slog.Logger, msg *events.Message) (*types.InternalEvent, error) {
	eventID := uuid.New()

	msg.UnwrapRaw()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "message",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: msg,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	event.Metadata["message_id"] = msg.Info.ID
	event.Metadata["from"] = msg.Info.Sender.String()
	event.Metadata["sender"] = msg.Info.Sender.String()
	event.Metadata["chat"] = msg.Info.Chat.String()
	event.Metadata["from_me"] = fmt.Sprintf("%t", msg.Info.IsFromMe)
	event.Metadata["is_group"] = fmt.Sprintf("%t", msg.Info.IsGroup)
	event.Metadata["timestamp"] = fmt.Sprintf("%d", msg.Info.Timestamp.Unix())

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

	if _, exists := event.Metadata["sender_pn"]; !exists {
		if pn := resolvedPNString(ctx, lidResolver, msg.Info.Sender); pn != "" {
			event.Metadata["sender_pn"] = pn
		} else if pn := resolvedPNString(ctx, lidResolver, msg.Info.MessageSource.SenderAlt); pn != "" {
			event.Metadata["sender_pn"] = pn
		}
	}

	if _, exists := event.Metadata["chat_pn"]; !exists {
		if pn := deriveChatPNString(ctx, lidResolver, msg.Info.Chat, msg.Info.MessageSource.RecipientAlt); pn != "" {
			event.Metadata["chat_pn"] = pn
		}
	}

	isGroupChat := msg.Info.IsGroup

	if provider := eventctx.ContactProvider(ctx); provider != nil {
		senderLookup := msg.Info.Sender.ToNonAD()
		senderIsHidden := senderLookup.Server == watypes.HiddenUserServer
		if senderIsHidden {
			if alt := msg.Info.MessageSource.SenderAlt.ToNonAD(); !alt.IsEmpty() && alt.Server != watypes.HiddenUserServer {
				senderLookup = alt
				senderIsHidden = false
			} else if pn := strings.TrimSpace(event.Metadata["sender_pn"]); pn != "" {
				if pnJID, err := watypes.ParseJID(pn); err == nil {
					senderLookup = pnJID.ToNonAD()
					senderIsHidden = senderLookup.Server == watypes.HiddenUserServer
				}
			}
		}

		chatLookup := msg.Info.Chat.ToNonAD()
		if chatLookup.Server == watypes.HiddenUserServer {
			if alt := msg.Info.MessageSource.RecipientAlt.ToNonAD(); !alt.IsEmpty() {
				chatLookup = alt
			}
		}

		if _, ok := event.Metadata["chat_name"]; !ok {
			if name := provider.ContactName(ctx, chatLookup); name != "" {
				event.Metadata["chat_name"] = name
			}
		}
		if isGroupChat || chatLookup.Server != watypes.HiddenUserServer {
			t.populatePhotoMetadata(ctx, provider, event.Metadata, "chat", chatLookup)
		}
		if !senderIsHidden {
			if _, ok := event.Metadata["sender_name"]; !ok {
				if name := provider.ContactName(ctx, senderLookup); name != "" {
					event.Metadata["sender_name"] = name
				}
			}
			t.populatePhotoMetadata(ctx, provider, event.Metadata, "sender", senderLookup)
		}
	}
	if _, ok := event.Metadata["sender_name"]; !ok && msg.Info.PushName != "" {
		event.Metadata["sender_name"] = msg.Info.PushName
	}

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

	if msg.Info.DeviceSentMeta != nil {
		if msg.Info.DeviceSentMeta.Phash != "" {
			event.Metadata["device_sent_phash"] = msg.Info.DeviceSentMeta.Phash
		}
	}

	if msg.Info.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(msg.Info.MessageSource.AddressingMode)
	}
	if !msg.Info.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = msg.Info.MessageSource.SenderAlt.String()
	}
	if !msg.Info.MessageSource.RecipientAlt.IsEmpty() {
		recipientAlt := msg.Info.MessageSource.RecipientAlt.String()
		event.Metadata["recipient_alt"] = recipientAlt
		if msg.Info.MessageSource.RecipientAlt.Server == watypes.GroupServer {
			event.Metadata["chat_alt"] = recipientAlt
		}
	}
	if !msg.Info.MessageSource.BroadcastListOwner.IsEmpty() {
		event.Metadata["broadcast_list_owner"] = msg.Info.MessageSource.BroadcastListOwner.String()
	}

	if msg.IsEphemeral {
		event.Metadata["is_ephemeral"] = "true"
	}
	if msg.IsViewOnce {
		event.Metadata["is_view_once"] = "true"
	}
	if msg.IsEdit {
		event.Metadata["is_edit"] = "true"
	}

	if protocolMsg := msg.Message.GetProtocolMessage(); protocolMsg != nil {
		msgType := protocolMsg.GetType()
		switch msgType {
		case waE2E.ProtocolMessage_REVOKE:
			event.Metadata["is_revoke"] = "true"
			if key := protocolMsg.GetKey(); key != nil {
				if revokedID := key.GetID(); revokedID != "" {
					event.Metadata["revoked_message_id"] = revokedID
				}
			}
		case waE2E.ProtocolMessage_MESSAGE_EDIT:
			event.Metadata["is_edit"] = "true"
		}
	}

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
		event.MediaIsGIF = mediaInfo.IsGIF
		event.MediaIsAnimated = mediaInfo.IsAnimated
		event.MediaWidth = mediaInfo.Width
		event.MediaHeight = mediaInfo.Height
		event.MediaWaveform = mediaInfo.Waveform
	}

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

	t.logMessageDebug(ctx, logger, msg)

	t.capturePollMetadata(ctx, event, msg)

	return event, nil
}

func (t *Transformer) transformUndecryptable(ctx context.Context, logger *slog.Logger, msg *events.UndecryptableMessage) (*types.InternalEvent, error) {
	eventID := uuid.New()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "undecryptable",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: msg,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	info := msg.Info

	event.Metadata["message_id"] = info.ID
	event.Metadata["from"] = info.Sender.String()
	event.Metadata["sender"] = info.Sender.String()
	event.Metadata["chat"] = info.Chat.String()
	event.Metadata["from_me"] = fmt.Sprintf("%t", info.IsFromMe)
	event.Metadata["is_group"] = fmt.Sprintf("%t", info.IsGroup)
	event.Metadata["timestamp"] = fmt.Sprintf("%d", info.Timestamp.Unix())
	if info.PushName != "" {
		event.Metadata["push_name"] = info.PushName
	}
	if info.VerifiedName != nil {
		verifiedNameJSON, _ := json.Marshal(info.VerifiedName)
		event.Metadata["verified_name"] = string(verifiedNameJSON)
	}
	if info.Category != "" {
		event.Metadata["category"] = info.Category
	}
	if info.ServerID != 0 {
		event.Metadata["server_id"] = fmt.Sprintf("%d", info.ServerID)
	}
	event.Metadata["multicast"] = fmt.Sprintf("%t", info.Multicast)
	if info.MediaType != "" {
		event.Metadata["media_type_info"] = info.MediaType
	}

	if _, exists := event.Metadata["sender_pn"]; !exists {
		sender := info.Sender.ToNonAD()
		switch sender.Server {
		case watypes.DefaultUserServer:
			event.Metadata["sender_pn"] = sender.String()
		case watypes.HiddenUserServer:
			if lidResolver != nil {
				if pn, ok := lidResolver.PNForLID(ctx, sender); ok {
					event.Metadata["sender_pn"] = pn.String()
				}
			}
		}
		if _, hasPN := event.Metadata["sender_pn"]; !hasPN {
			alt := info.MessageSource.SenderAlt.ToNonAD()
			if alt.Server == watypes.DefaultUserServer && !alt.IsEmpty() {
				event.Metadata["sender_pn"] = alt.String()
			}
		}
	}

	if provider := eventctx.ContactProvider(ctx); provider != nil {
		senderJID := info.Sender.ToNonAD()
		chatJID := info.Chat.ToNonAD()
		if _, ok := event.Metadata["chat_name"]; !ok {
			if name := provider.ContactName(ctx, chatJID); name != "" {
				event.Metadata["chat_name"] = name
			}
		}
		t.populatePhotoMetadata(ctx, provider, event.Metadata, "chat", chatJID)
		if _, ok := event.Metadata["sender_name"]; !ok {
			if name := provider.ContactName(ctx, senderJID); name != "" {
				event.Metadata["sender_name"] = name
			}
		}
		t.populatePhotoMetadata(ctx, provider, event.Metadata, "sender", senderJID)
	}
	if _, ok := event.Metadata["sender_name"]; !ok && info.PushName != "" {
		event.Metadata["sender_name"] = info.PushName
	}

	if info.MsgMetaInfo.TargetID != "" {
		event.Metadata["reply_to_message_id"] = info.MsgMetaInfo.TargetID
	}
	if !info.MsgMetaInfo.TargetSender.IsEmpty() {
		event.Metadata["reply_to_sender"] = info.MsgMetaInfo.TargetSender.String()
	}
	if !info.MsgMetaInfo.TargetChat.IsEmpty() {
		event.Metadata["reply_to_chat"] = info.MsgMetaInfo.TargetChat.String()
	}
	if info.MsgMetaInfo.ThreadMessageID != "" {
		event.Metadata["thread_message_id"] = info.MsgMetaInfo.ThreadMessageID
	}
	if !info.MsgMetaInfo.ThreadMessageSenderJID.IsEmpty() {
		event.Metadata["thread_message_sender"] = info.MsgMetaInfo.ThreadMessageSenderJID.String()
	}

	if info.DeviceSentMeta != nil && info.DeviceSentMeta.Phash != "" {
		event.Metadata["device_sent_phash"] = info.DeviceSentMeta.Phash
	}

	if info.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(info.MessageSource.AddressingMode)
	}
	if !info.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = info.MessageSource.SenderAlt.String()
	}
	if !info.MessageSource.RecipientAlt.IsEmpty() {
		recipientAlt := info.MessageSource.RecipientAlt.String()
		event.Metadata["recipient_alt"] = recipientAlt
		if info.MessageSource.RecipientAlt.Server == watypes.GroupServer {
			event.Metadata["chat_alt"] = recipientAlt
		}
	}
	if !info.MessageSource.BroadcastListOwner.IsEmpty() {
		event.Metadata["broadcast_list_owner"] = info.MessageSource.BroadcastListOwner.String()
	}

	event.Metadata["waiting_message"] = "true"
	event.Metadata["is_unavailable"] = fmt.Sprintf("%t", msg.IsUnavailable)
	if msg.UnavailableType != "" {
		event.Metadata["unavailable_type"] = string(msg.UnavailableType)
		if msg.UnavailableType == events.UnavailableTypeViewOnce {
			event.Metadata["is_view_once"] = "true"
		}
	}
	if msg.DecryptFailMode != "" {
		event.Metadata["decrypt_fail_mode"] = string(msg.DecryptFailMode)
	}

	logger.WarnContext(ctx, "captured undecryptable message",
		slog.String("event_id", eventID.String()),
		slog.String("message_id", info.ID),
		slog.Bool("is_unavailable", msg.IsUnavailable),
		slog.String("unavailable_type", string(msg.UnavailableType)),
	)

	t.logUndecryptableDebug(ctx, logger, msg)

	return event, nil
}

func (t *Transformer) logMessageDebug(ctx context.Context, logger *slog.Logger, msg *events.Message) {
	if !t.debug || msg == nil {
		return
	}

	details := map[string]interface{}{
		"message_id":   msg.Info.ID,
		"from":         msg.Info.Sender.String(),
		"chat":         msg.Info.Chat.String(),
		"timestamp":    msg.Info.Timestamp.Unix(),
		"is_view_once": msg.IsViewOnce,
		"is_ephemeral": msg.IsEphemeral,
	}

	if msg.Message != nil {
		if data, err := debugProtoMarshalOpts.Marshal(msg.Message); err == nil {
			details["proto"] = json.RawMessage(data)
		}
	}

	if msg.RawMessage != nil {
		if data, err := debugProtoMarshalOpts.Marshal(msg.RawMessage); err == nil {
			details["raw_proto"] = json.RawMessage(data)
		}
	}

	if data, err := json.Marshal(details); err == nil {
		logger.DebugContext(ctx, "whatsmeow raw message", slog.String("payload", string(data)))
	}
}

func (t *Transformer) capturePollMetadata(ctx context.Context, event *types.InternalEvent, msg *events.Message) {
	if msg == nil || msg.Message == nil {
		return
	}
	if poll := msg.Message.GetPollCreationMessage(); poll != nil {
		if messageID := msg.Info.ID; messageID != "" {
			t.storePollOptions(ctx, poll, messageID)
		}
	}
	if pollVote := msg.Message.GetPollUpdateMessage(); pollVote != nil {
		if key := pollVote.GetPollCreationMessageKey(); key != nil {
			event.Metadata["poll_message_id"] = key.GetID()
		}
		t.capturePollVoteHashes(ctx, event, msg, pollVote)
	}
}

func (t *Transformer) storePollOptions(ctx context.Context, poll *waE2E.PollCreationMessage, messageID string) {
	if t.pollStore == nil || poll == nil || messageID == "" {
		return
	}
	options := poll.GetOptions()
	if len(options) == 0 {
		return
	}
	mapping := make(map[string]string, len(options))
	names := make([]string, len(options))
	missingHash := false
	for i, opt := range options {
		name := strings.TrimSpace(opt.GetOptionName())
		if name == "" {
			continue
		}
		names[i] = name
		hash := strings.TrimSpace(opt.GetOptionHash())
		if hash == "" {
			missingHash = true
		} else {
			mapping[strings.ToLower(hash)] = name
		}
	}
	if missingHash {
		computed := whatsmeowcore.HashPollOptions(names)
		for i, hashed := range computed {
			if len(hashed) == 0 || names[i] == "" {
				continue
			}
			hash := strings.ToLower(hex.EncodeToString(hashed))
			if _, exists := mapping[hash]; !exists {
				mapping[hash] = names[i]
			}
		}
	}
	if len(mapping) == 0 {
		return
	}
	if err := t.pollStore.SaveOptions(ctx, t.instanceID, messageID, mapping); err != nil && t.debug {
		logging.ContextLogger(ctx, nil).Debug("poll option store failed",
			slog.String("instance_id", t.instanceID.String()),
			slog.String("message_id", messageID),
			slog.String("error", err.Error()))
	}
}

func (t *Transformer) capturePollVoteHashes(ctx context.Context, event *types.InternalEvent, msg *events.Message, pollVote *waE2E.PollUpdateMessage) {
	decrypter := eventctx.PollDecrypterFromContext(ctx)
	if decrypter == nil {
		return
	}
	voteMsg, err := decrypter.DecryptPollVote(ctx, msg)
	if err != nil {
		if t.debug {
			logging.ContextLogger(ctx, nil).Debug("poll vote decrypt failed",
				slog.String("instance_id", t.instanceID.String()),
				slog.String("message_id", msg.Info.ID),
				slog.String("error", err.Error()))
		}
		return
	}
	selected := voteMsg.GetSelectedOptions()
	if len(selected) == 0 {
		return
	}
	hashes := make([]string, 0, len(selected))
	for _, entry := range selected {
		if len(entry) == 0 {
			continue
		}
		hashes = append(hashes, strings.ToLower(hex.EncodeToString(entry)))
	}
	if len(hashes) == 0 {
		return
	}
	payload, err := json.Marshal(hashes)
	if err != nil {
		return
	}
	event.Metadata["poll_vote_hashes"] = string(payload)
}

func (t *Transformer) logUndecryptableDebug(ctx context.Context, logger *slog.Logger, msg *events.UndecryptableMessage) {
	if !t.debug || msg == nil {
		return
	}

	details := map[string]interface{}{
		"message_id":        msg.Info.ID,
		"from":              msg.Info.Sender.String(),
		"chat":              msg.Info.Chat.String(),
		"timestamp":         msg.Info.Timestamp.Unix(),
		"is_unavailable":    msg.IsUnavailable,
		"unavailable_type":  string(msg.UnavailableType),
		"decrypt_fail_mode": string(msg.DecryptFailMode),
	}

	if data, err := json.Marshal(details); err == nil {
		logger.DebugContext(ctx, "whatsmeow undecryptable message", slog.String("payload", string(data)))
	}
}

func (t *Transformer) populatePhotoMetadata(ctx context.Context, provider eventctx.ContactMetadataProvider, metadata map[string]string, prefix string, jid watypes.JID) {
	if jid.IsEmpty() {
		return
	}

	if detailProvider, ok := provider.(eventctx.ContactPhotoDetailProvider); ok {
		details := detailProvider.ContactPhotoDetails(ctx, jid)
		full := strings.TrimSpace(details.FullURL)
		preview := strings.TrimSpace(details.PreviewURL)
		if full != "" {
			if current, ok := metadata[prefix+"_photo"]; !ok || strings.TrimSpace(current) == "" {
				metadata[prefix+"_photo"] = full
			}
			return
		}
		if preview != "" {
			if current, ok := metadata[prefix+"_photo"]; !ok || strings.TrimSpace(current) == "" {
				metadata[prefix+"_photo"] = preview
			}
			return
		}
	}

	if _, ok := metadata[prefix+"_photo"]; ok {
		return
	}

	if photo := strings.TrimSpace(provider.ContactPhoto(ctx, jid)); photo != "" {
		metadata[prefix+"_photo"] = photo
	}
}

func (t *Transformer) transformReceipt(ctx context.Context, logger *slog.Logger, receipt *events.Receipt) (*types.InternalEvent, error) {
	eventID := uuid.New()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "receipt",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: receipt,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	event.Metadata["chat"] = receipt.Chat.String()
	event.Metadata["sender"] = receipt.Sender.String()
	event.Metadata["receipt_type"] = string(receipt.Type)
	event.Metadata["timestamp"] = fmt.Sprintf("%d", receipt.Timestamp.Unix())
	event.Metadata["from_me"] = fmt.Sprintf("%t", receipt.IsFromMe)

	if receipt.Chat.Device > 0 {
		event.Metadata["chat_device"] = fmt.Sprintf("%d", receipt.Chat.Device)
	}
	if receipt.Sender.Device > 0 {
		event.Metadata["sender_device"] = fmt.Sprintf("%d", receipt.Sender.Device)
	}

	event.Metadata["sender_jid"] = receipt.Sender.String()

	if !receipt.MessageSender.IsEmpty() {
		event.Metadata["message_sender"] = receipt.MessageSender.String()
	}

	isGroup := receipt.Chat.Server == watypes.GroupServer
	if !isGroup && !receipt.MessageSource.RecipientAlt.IsEmpty() && receipt.MessageSource.RecipientAlt.Server == watypes.GroupServer {
		isGroup = true
		event.Metadata["chat_alt"] = receipt.MessageSource.RecipientAlt.String()
	}
	event.Metadata["is_group"] = fmt.Sprintf("%t", isGroup)

	if receipt.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(receipt.MessageSource.AddressingMode)
	}
	if !receipt.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = receipt.MessageSource.SenderAlt.String()
	}
	if !receipt.MessageSource.RecipientAlt.IsEmpty() {
		event.Metadata["recipient_alt"] = receipt.MessageSource.RecipientAlt.String()
	}

	if len(receipt.MessageIDs) > 0 {
		messageIDsJSON, _ := json.Marshal(receipt.MessageIDs)
		event.Metadata["message_ids"] = string(messageIDsJSON)
	}

	if _, exists := event.Metadata["chat_pn"]; !exists {
		if pn := deriveChatPNString(ctx, lidResolver, receipt.Chat, receipt.MessageSource.RecipientAlt); pn != "" {
			event.Metadata["chat_pn"] = pn
		}
	}

	if _, exists := event.Metadata["sender_pn"]; !exists {
		if pn := resolvedPNString(ctx, lidResolver, receipt.Sender); pn != "" {
			event.Metadata["sender_pn"] = pn
		} else if pn := resolvedPNString(ctx, lidResolver, receipt.MessageSource.SenderAlt); pn != "" {
			event.Metadata["sender_pn"] = pn
		}
	}

	if !receipt.MessageSender.IsEmpty() {
		if _, exists := event.Metadata["message_sender_pn"]; !exists {
			if pn := resolvedPNString(ctx, lidResolver, receipt.MessageSender); pn != "" {
				event.Metadata["message_sender_pn"] = pn
			}
		}
	}

	if provider := eventctx.ContactProvider(ctx); provider != nil {
		senderJID := receipt.Sender.ToNonAD()
		if _, ok := event.Metadata["sender_name"]; !ok {
			if name := provider.ContactName(ctx, senderJID); name != "" {
				event.Metadata["sender_name"] = name
			}
		}
		t.populatePhotoMetadata(ctx, provider, event.Metadata, "sender", senderJID)
	}

	logger.InfoContext(ctx, "transformed receipt event",
		slog.String("event_id", eventID.String()),
		slog.String("receipt_type", string(receipt.Type)),
		slog.Int("message_count", len(receipt.MessageIDs)),
	)

	return event, nil
}

func (t *Transformer) transformChatPresence(ctx context.Context, logger *slog.Logger, presence *events.ChatPresence) (*types.InternalEvent, error) {
	eventID := uuid.New()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "chat_presence",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: presence,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	event.Metadata["chat"] = presence.Chat.String()
	event.Metadata["sender"] = presence.Sender.String()
	event.Metadata["state"] = string(presence.State)
	event.Metadata["media"] = string(presence.Media)
	event.Metadata["is_group"] = fmt.Sprintf("%t", presence.IsGroup)

	if presence.MessageSource.AddressingMode != "" {
		event.Metadata["addressing_mode"] = string(presence.MessageSource.AddressingMode)
	}
	if !presence.MessageSource.SenderAlt.IsEmpty() {
		event.Metadata["sender_alt"] = presence.MessageSource.SenderAlt.String()
	}
	if !presence.MessageSource.RecipientAlt.IsEmpty() {
		event.Metadata["recipient_alt"] = presence.MessageSource.RecipientAlt.String()
	}

	if _, exists := event.Metadata["chat_pn"]; !exists {
		if pn := deriveChatPNString(ctx, lidResolver, presence.Chat, presence.MessageSource.RecipientAlt); pn != "" {
			event.Metadata["chat_pn"] = pn
		}
	}

	if _, exists := event.Metadata["sender_pn"]; !exists {
		if pn := resolvedPNString(ctx, lidResolver, presence.Sender); pn != "" {
			event.Metadata["sender_pn"] = pn
		} else if pn := resolvedPNString(ctx, lidResolver, presence.MessageSource.SenderAlt); pn != "" {
			event.Metadata["sender_pn"] = pn
		}
	}

	// Presence events depend exclusively on WhatsApp metadata; avoid extra lookups

	logger.InfoContext(ctx, "transformed chat presence event",
		slog.String("event_id", eventID.String()),
		slog.String("state", string(presence.State)),
		slog.String("media", string(presence.Media)),
	)

	return event, nil
}

func (t *Transformer) transformPresence(ctx context.Context, logger *slog.Logger, presence *events.Presence) (*types.InternalEvent, error) {
	eventID := uuid.New()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "presence",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: presence,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	event.Metadata["from"] = presence.From.String()
	event.Metadata["unavailable"] = fmt.Sprintf("%t", presence.Unavailable)
	if !presence.LastSeen.IsZero() {
		event.Metadata["last_seen"] = fmt.Sprintf("%d", presence.LastSeen.Unix())
	}
	if _, exists := event.Metadata["from_pn"]; !exists {
		if pn := resolvedPNString(ctx, lidResolver, presence.From); pn != "" {
			event.Metadata["from_pn"] = pn
		}
	}

	logger.InfoContext(ctx, "transformed presence event",
		slog.String("event_id", eventID.String()),
		slog.Bool("unavailable", presence.Unavailable),
	)

	return event, nil
}

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

func resolvedPNString(ctx context.Context, resolver eventctx.LIDResolver, jid watypes.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	normalized := jid.ToNonAD()
	switch normalized.Server {
	case watypes.DefaultUserServer, watypes.LegacyUserServer:
		return normalized.String()
	case watypes.HiddenUserServer:
		if resolver == nil {
			return ""
		}
		if pn, ok := resolver.PNForLID(ctx, normalized); ok && !pn.IsEmpty() {
			return pn.ToNonAD().String()
		}
	}
	return ""
}

func deriveChatPNString(ctx context.Context, resolver eventctx.LIDResolver, chat watypes.JID, recipientAlt watypes.JID) string {
	if pn := resolvedPNString(ctx, resolver, chat); pn != "" {
		return pn
	}
	if pn := resolvedPNString(ctx, resolver, recipientAlt); pn != "" {
		return pn
	}
	return ""
}

func encodeParticipantPNList(ctx context.Context, resolver eventctx.LIDResolver, participants []watypes.JID) string {
	if len(participants) == 0 {
		return ""
	}
	values := make([]string, len(participants))
	hasValue := false
	for i, member := range participants {
		if pn := resolvedPNString(ctx, resolver, member); pn != "" {
			values[i] = pn
			hasValue = true
		}
	}
	if !hasValue {
		return ""
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func storeParticipantPNMetadata(ctx context.Context, resolver eventctx.LIDResolver, metadata map[string]string, key string, participants []watypes.JID) {
	if metadata == nil {
		return
	}
	if _, exists := metadata[key]; exists {
		return
	}
	if encoded := encodeParticipantPNList(ctx, resolver, participants); encoded != "" {
		metadata[key] = encoded
	}
}

type MediaInfo struct {
	MediaKey      string
	DirectPath    string
	FileSHA256    *string
	FileEncSHA256 *string
	MediaType     string
	MimeType      *string
	FileLength    *int64
	IsGIF         bool
	IsAnimated    bool
	Width         int
	Height        int
	Waveform      []byte
}

func (t *Transformer) extractMediaInfo(msg *waE2E.Message) (bool, MediaInfo) {
	var info MediaInfo

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
		if mime := img.GetMimetype(); mime == "image/gif" {
			info.IsGIF = true
			info.IsAnimated = true
		}
		if width := img.GetWidth(); width > 0 {
			info.Width = int(width)
		}
		if height := img.GetHeight(); height > 0 {
			info.Height = int(height)
		}
		return true, info
	}

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
		info.IsGIF = video.GetGifPlayback()
		if width := video.GetWidth(); width > 0 {
			info.Width = int(width)
		}
		if height := video.GetHeight(); height > 0 {
			info.Height = int(height)
		}
		return true, info
	}

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
		if waveform := audio.GetWaveform(); len(waveform) > 0 {
			info.Waveform = waveform
		}
		return true, info
	}

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
		info.IsAnimated = sticker.GetIsAnimated()
		if width := sticker.GetWidth(); width > 0 {
			info.Width = int(width)
		}
		if height := sticker.GetHeight(); height > 0 {
			info.Height = int(height)
		}
		return true, info
	}

	return false, info
}

func (t *Transformer) extractContextInfo(ctxInfo *waE2E.ContextInfo, event *types.InternalEvent) {
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
	if quotedMsg := ctxInfo.GetQuotedMessage(); quotedMsg != nil {
		if quotedJSON, err := json.Marshal(quotedMsg); err == nil {
			event.Metadata["quoted_message"] = string(quotedJSON)
		}
	}
	if mentionedJIDs := ctxInfo.GetMentionedJID(); len(mentionedJIDs) > 0 {
		event.MentionedJIDs = mentionedJIDs
		event.Metadata["mentioned_jids"] = strings.Join(mentionedJIDs, ",")
	}
	if isForwarded := ctxInfo.GetIsForwarded(); isForwarded {
		event.IsForwarded = true
		event.Metadata["is_forwarded"] = "true"
	}
	if ephemeralExpiry := ctxInfo.GetEphemeralSettingTimestamp(); ephemeralExpiry > 0 {
		event.EphemeralExpiry = ephemeralExpiry
		event.Metadata["ephemeral_expiry"] = fmt.Sprintf("%d", ephemeralExpiry)
	}
	if external := ctxInfo.GetExternalAdReply(); external != nil {
		externalPayload := map[string]interface{}{
			"title":                 external.GetTitle(),
			"body":                  external.GetBody(),
			"mediaType":             int(external.GetMediaType()),
			"thumbnailUrl":          external.GetThumbnailURL(),
			"sourceType":            external.GetSourceType(),
			"sourceId":              external.GetSourceID(),
			"sourceUrl":             external.GetSourceURL(),
			"containsAutoReply":     external.GetContainsAutoReply(),
			"renderLargerThumbnail": external.GetRenderLargerThumbnail(),
			"showAdAttribution":     external.GetShowAdAttribution(),
		}
		if raw, err := json.Marshal(externalPayload); err == nil {
			event.Metadata["external_ad_reply"] = string(raw)
		}
	}
}

func (t *Transformer) transformJoinedGroup(ctx context.Context, logger *slog.Logger, joined *events.JoinedGroup) (*types.InternalEvent, error) {
	eventID := uuid.New()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "group_joined",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: joined,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

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
	if _, ok := event.Metadata["sender_pn"]; !ok && joined.Sender != nil {
		if pn := resolvedPNString(ctx, lidResolver, *joined.Sender); pn != "" {
			event.Metadata["sender_pn"] = pn
		}
	}
	if joined.Notify != "" {
		event.Metadata["notify"] = joined.Notify
	}
	if joined.CreateKey != "" {
		event.Metadata["create_key"] = joined.CreateKey
	}

	participantJIDs := make([]watypes.JID, 0, len(joined.Participants))
	for _, member := range joined.Participants {
		if member.JID.IsEmpty() {
			continue
		}
		participantJIDs = append(participantJIDs, member.JID)
	}
	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "join_participants_pn", participantJIDs)

	logger.InfoContext(ctx, "transformed joined group event",
		slog.String("event_id", eventID.String()),
		slog.String("group_id", joined.JID.String()),
	)

	return event, nil
}

func (t *Transformer) transformGroupInfo(ctx context.Context, logger *slog.Logger, info *events.GroupInfo) (*types.InternalEvent, error) {
	eventID := uuid.New()

	lidResolver := eventctx.LIDResolverFromContext(ctx)

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "group_info",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: info,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	event.Metadata["group_id"] = info.JID.String()
	event.Metadata["timestamp"] = fmt.Sprintf("%d", info.Timestamp.Unix())

	if info.Sender != nil {
		event.Metadata["sender"] = info.Sender.String()
	}
	if info.SenderPN != nil {
		event.Metadata["sender_pn"] = info.SenderPN.String()
	}
	if _, ok := event.Metadata["sender_pn"]; !ok && info.Sender != nil {
		if pn := resolvedPNString(ctx, lidResolver, *info.Sender); pn != "" {
			event.Metadata["sender_pn"] = pn
		}
	}
	if info.Notify != "" {
		event.Metadata["notify"] = info.Notify
	}
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

	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "join_participants_pn", info.Join)
	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "leave_participants_pn", info.Leave)
	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "promote_participants_pn", info.Promote)
	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "demote_participants_pn", info.Demote)
	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "membership_request_created_pn", info.MembershipRequestsCreated)
	storeParticipantPNMetadata(ctx, lidResolver, event.Metadata, "membership_request_revoked_pn", info.MembershipRequestsRevoked)
	if info.Locked != nil {
		event.Metadata["locked"] = fmt.Sprintf("%t", info.Locked.IsLocked)
	}
	if info.Announce != nil {
		event.Metadata["announce"] = fmt.Sprintf("%t", info.Announce.IsAnnounce)
	}
	if method := strings.TrimSpace(info.MembershipRequestMethod); method != "" {
		event.Metadata["membership_request_method"] = method
	}

	logger.InfoContext(ctx, "transformed group info event",
		slog.String("event_id", eventID.String()),
		slog.String("group_id", info.JID.String()),
	)

	return event, nil
}

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

func (t *Transformer) transformHistorySync(ctx context.Context, logger *slog.Logger, history *events.HistorySync) (*types.InternalEvent, error) {
	if history.Data == nil {
		return nil, fmt.Errorf("history sync event missing Data field")
	}

	eventID := uuid.New()

	event := &types.InternalEvent{
		InstanceID: t.instanceID,
		EventID:    eventID,
		EventType:  "history_sync",
		SourceLib:  types.SourceLibWhatsmeow,
		RawPayload: history,
		Metadata:   make(map[string]string),
		CapturedAt: time.Now(),
	}

	syncTypeStr := strings.TrimPrefix(history.Data.SyncType.String(), "HistorySync_")
	syncTypeStr = strings.ToLower(syncTypeStr)
	event.Metadata["history_sync_type"] = syncTypeStr
	event.Metadata["chunk_order"] = strconv.FormatUint(uint64(history.Data.GetChunkOrder()), 10)
	event.Metadata["progress"] = strconv.FormatUint(uint64(history.Data.GetProgress()), 10)

	conversationCount := len(history.Data.GetConversations())
	event.Metadata["conversation_count"] = strconv.Itoa(conversationCount)

	statusMessageCount := len(history.Data.GetStatusV3Messages())
	event.Metadata["status_message_count"] = strconv.Itoa(statusMessageCount)

	pushnameCount := len(history.Data.GetPushnames())
	event.Metadata["pushname_count"] = strconv.Itoa(pushnameCount)

	logger.InfoContext(ctx, "transformed history sync event",
		slog.String("event_id", eventID.String()),
		slog.String("sync_type", syncTypeStr),
		slog.Uint64("chunk_order", uint64(history.Data.GetChunkOrder())),
		slog.Uint64("progress", uint64(history.Data.GetProgress())),
		slog.Int("conversation_count", conversationCount),
		slog.Int("status_message_count", statusMessageCount),
		slog.Int("pushname_count", pushnameCount),
	)

	return event, nil
}
