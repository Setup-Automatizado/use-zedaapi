package encoding

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/proto/waWeb"
	whatsmeowtypes "go.mau.fi/whatsmeow/types"
	whatsmeowevents "go.mau.fi/whatsmeow/types/events"
)

const (
	payloadTypeMessage       = "message"
	payloadTypeReceipt       = "receipt"
	payloadTypeChatPresence  = "chat_presence"
	payloadTypePresence      = "presence"
	payloadTypeConnected     = "connected"
	payloadTypeDisconnected  = "disconnected"
	payloadTypeJoinedGroup   = "joined_group"
	payloadTypeGroupInfo     = "group_info"
	payloadTypePicture       = "picture"
	payloadTypeUndecryptable = "undecryptable"
)

var (
	protoMarshalOpts   = protojson.MarshalOptions{EmitUnpopulated: true, UseEnumNumbers: false}
	protoUnmarshalOpts = protojson.UnmarshalOptions{DiscardUnknown: true}
)

func init() {
	gob.Register(&whatsmeowevents.Message{})
	gob.Register(&whatsmeowevents.Receipt{})
	gob.Register(&whatsmeowevents.ChatPresence{})
	gob.Register(&whatsmeowevents.Presence{})
	gob.Register(&whatsmeowevents.Connected{})
	gob.Register(&whatsmeowevents.Disconnected{})
	gob.Register(&whatsmeowevents.JoinedGroup{})
	gob.Register(&whatsmeowevents.GroupInfo{})
	gob.Register(&whatsmeowevents.Picture{})
	gob.Register(&whatsmeowevents.UndecryptableMessage{})
	gob.Register(uuid.UUID{})
}

func EncodeInternalEvent(event *types.InternalEvent) (string, error) {
	if event == nil {
		return "", fmt.Errorf("internal event is nil")
	}

	persisted, err := toPersistedInternalEvent(event)
	if err != nil {
		return "", err
	}

	raw, err := json.Marshal(persisted)
	if err != nil {
		return "", fmt.Errorf("marshal internal event: %w", err)
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}

func DecodeInternalEvent(encoded string) (*types.InternalEvent, error) {
	if encoded == "" {
		return nil, fmt.Errorf("encoded internal event payload is empty")
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode internal event: %w", err)
	}

	var persisted persistedInternalEvent
	if err := json.Unmarshal(raw, &persisted); err == nil && persisted.EventID != "" {
		return fromPersistedInternalEvent(&persisted)
	}

	decoder := gob.NewDecoder(bytes.NewReader(raw))
	var legacy types.InternalEvent
	if err := decoder.Decode(&legacy); err != nil {
		return nil, fmt.Errorf("decode internal event: %w", err)
	}

	return &legacy, nil
}

type persistedInternalEvent struct {
	InstanceID      string            `json:"instance_id"`
	EventID         string            `json:"event_id"`
	EventType       string            `json:"event_type"`
	SourceLib       string            `json:"source_lib"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CapturedAt      time.Time         `json:"captured_at"`
	RawPayload      *persistedPayload `json:"raw_payload,omitempty"`
	HasMedia        bool              `json:"has_media"`
	MediaKey        string            `json:"media_key,omitempty"`
	DirectPath      string            `json:"direct_path,omitempty"`
	FileSHA256      *string           `json:"file_sha256,omitempty"`
	FileEncSHA256   *string           `json:"file_enc_sha256,omitempty"`
	MediaType       string            `json:"media_type,omitempty"`
	MimeType        *string           `json:"mime_type,omitempty"`
	FileLength      *int64            `json:"file_length,omitempty"`
	MediaIsGIF      bool              `json:"media_is_gif,omitempty"`
	MediaIsAnimated bool              `json:"media_is_animated,omitempty"`
	MediaWidth      int               `json:"media_width,omitempty"`
	MediaHeight     int               `json:"media_height,omitempty"`
	MediaWaveform   []byte            `json:"media_waveform,omitempty"`
	QuotedMessageID string            `json:"quoted_message_id,omitempty"`
	QuotedSender    string            `json:"quoted_sender,omitempty"`
	QuotedRemoteJID string            `json:"quoted_remote_jid,omitempty"`
	MentionedJIDs   []string          `json:"mentioned_jids,omitempty"`
	IsForwarded     bool              `json:"is_forwarded,omitempty"`
	EphemeralExpiry int64             `json:"ephemeral_expiry,omitempty"`
	TransportType   string            `json:"transport_type,omitempty"`
	TransportConfig json.RawMessage   `json:"transport_config,omitempty"`
}

type persistedPayload struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type serializableMessage struct {
	Info                  whatsmeowtypes.MessageInfo             `json:"info"`
	Message               json.RawMessage                        `json:"message,omitempty"`
	RawMessage            json.RawMessage                        `json:"raw_message,omitempty"`
	SourceWebMessage      json.RawMessage                        `json:"source_web_message,omitempty"`
	UnavailableRequestID  string                                 `json:"unavailable_request_id,omitempty"`
	RetryCount            int                                    `json:"retry_count,omitempty"`
	NewsletterMeta        *whatsmeowevents.NewsletterMessageMeta `json:"newsletter_meta,omitempty"`
	IsEphemeral           bool                                   `json:"is_ephemeral,omitempty"`
	IsViewOnce            bool                                   `json:"is_view_once,omitempty"`
	IsViewOnceV2          bool                                   `json:"is_view_once_v2,omitempty"`
	IsViewOnceV2Extension bool                                   `json:"is_view_once_v2_extension,omitempty"`
	IsDocumentWithCaption bool                                   `json:"is_document_with_caption,omitempty"`
	IsLottieSticker       bool                                   `json:"is_lottie_sticker,omitempty"`
	IsBotInvoke           bool                                   `json:"is_bot_invoke,omitempty"`
	IsEdit                bool                                   `json:"is_edit,omitempty"`
}

func toPersistedInternalEvent(event *types.InternalEvent) (*persistedInternalEvent, error) {
	payload, err := encodeRawPayload(event.RawPayload)
	if err != nil {
		return nil, err
	}

	return &persistedInternalEvent{
		InstanceID:      event.InstanceID.String(),
		EventID:         event.EventID.String(),
		EventType:       event.EventType,
		SourceLib:       string(event.SourceLib),
		Metadata:        cloneStringMap(event.Metadata),
		CapturedAt:      event.CapturedAt,
		RawPayload:      payload,
		HasMedia:        event.HasMedia,
		MediaKey:        event.MediaKey,
		DirectPath:      event.DirectPath,
		FileSHA256:      cloneStringPtr(event.FileSHA256),
		FileEncSHA256:   cloneStringPtr(event.FileEncSHA256),
		MediaType:       event.MediaType,
		MimeType:        cloneStringPtr(event.MimeType),
		FileLength:      cloneInt64Ptr(event.FileLength),
		MediaIsGIF:      event.MediaIsGIF,
		MediaIsAnimated: event.MediaIsAnimated,
		MediaWidth:      event.MediaWidth,
		MediaHeight:     event.MediaHeight,
		MediaWaveform:   cloneBytes(event.MediaWaveform),
		QuotedMessageID: event.QuotedMessageID,
		QuotedSender:    event.QuotedSender,
		QuotedRemoteJID: event.QuotedRemoteJID,
		MentionedJIDs:   cloneStringSlice(event.MentionedJIDs),
		IsForwarded:     event.IsForwarded,
		EphemeralExpiry: event.EphemeralExpiry,
		TransportType:   event.TransportType,
		TransportConfig: cloneRaw(event.TransportConfig),
	}, nil
}

func fromPersistedInternalEvent(p *persistedInternalEvent) (*types.InternalEvent, error) {
	instanceID, err := uuid.Parse(p.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("parse instance_id: %w", err)
	}

	eventID, err := uuid.Parse(p.EventID)
	if err != nil {
		return nil, fmt.Errorf("parse event_id: %w", err)
	}

	payload, err := decodeRawPayload(p.RawPayload)
	if err != nil {
		return nil, err
	}

	return &types.InternalEvent{
		InstanceID:      instanceID,
		EventID:         eventID,
		EventType:       p.EventType,
		SourceLib:       types.SourceLib(p.SourceLib),
		Metadata:        cloneStringMap(p.Metadata),
		CapturedAt:      p.CapturedAt,
		RawPayload:      payload,
		HasMedia:        p.HasMedia,
		MediaKey:        p.MediaKey,
		DirectPath:      p.DirectPath,
		FileSHA256:      cloneStringPtr(p.FileSHA256),
		FileEncSHA256:   cloneStringPtr(p.FileEncSHA256),
		MediaType:       p.MediaType,
		MimeType:        cloneStringPtr(p.MimeType),
		FileLength:      cloneInt64Ptr(p.FileLength),
		MediaIsGIF:      p.MediaIsGIF,
		MediaIsAnimated: p.MediaIsAnimated,
		MediaWidth:      p.MediaWidth,
		MediaHeight:     p.MediaHeight,
		MediaWaveform:   cloneBytes(p.MediaWaveform),
		QuotedMessageID: p.QuotedMessageID,
		QuotedSender:    p.QuotedSender,
		QuotedRemoteJID: p.QuotedRemoteJID,
		MentionedJIDs:   cloneStringSlice(p.MentionedJIDs),
		IsForwarded:     p.IsForwarded,
		EphemeralExpiry: p.EphemeralExpiry,
		TransportType:   p.TransportType,
		TransportConfig: cloneRaw(p.TransportConfig),
	}, nil
}

func encodeRawPayload(payload interface{}) (*persistedPayload, error) {
	if payload == nil {
		return nil, nil
	}

	switch evt := payload.(type) {
	case *whatsmeowevents.Message:
		msg, err := serializeMessage(evt)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("marshal message payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypeMessage, Data: data}, nil
	case *whatsmeowevents.Receipt:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal receipt payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypeReceipt, Data: data}, nil
	case *whatsmeowevents.ChatPresence:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal chat_presence payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypeChatPresence, Data: data}, nil
	case *whatsmeowevents.Presence:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal presence payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypePresence, Data: data}, nil
	case *whatsmeowevents.Connected:
		return &persistedPayload{Type: payloadTypeConnected}, nil
	case *whatsmeowevents.Disconnected:
		return &persistedPayload{Type: payloadTypeDisconnected}, nil
	case *whatsmeowevents.JoinedGroup:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal joined_group payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypeJoinedGroup, Data: data}, nil
	case *whatsmeowevents.GroupInfo:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal group_info payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypeGroupInfo, Data: data}, nil
	case *whatsmeowevents.Picture:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal picture payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypePicture, Data: data}, nil
	case *whatsmeowevents.UndecryptableMessage:
		data, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal undecryptable payload: %w", err)
		}
		return &persistedPayload{Type: payloadTypeUndecryptable, Data: data}, nil
	default:
		return nil, fmt.Errorf("unsupported raw payload type %T", payload)
	}
}

func decodeRawPayload(payload *persistedPayload) (interface{}, error) {
	if payload == nil {
		return nil, nil
	}

	switch payload.Type {
	case payloadTypeMessage:
		var ser serializableMessage
		if err := json.Unmarshal(payload.Data, &ser); err != nil {
			return nil, fmt.Errorf("unmarshal message payload: %w", err)
		}
		return deserializeMessage(&ser)
	case payloadTypeReceipt:
		var evt whatsmeowevents.Receipt
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal receipt payload: %w", err)
		}
		return &evt, nil
	case payloadTypeChatPresence:
		var evt whatsmeowevents.ChatPresence
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal chat_presence payload: %w", err)
		}
		return &evt, nil
	case payloadTypePresence:
		var evt whatsmeowevents.Presence
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal presence payload: %w", err)
		}
		return &evt, nil
	case payloadTypeConnected:
		return &whatsmeowevents.Connected{}, nil
	case payloadTypeDisconnected:
		return &whatsmeowevents.Disconnected{}, nil
	case payloadTypeJoinedGroup:
		var evt whatsmeowevents.JoinedGroup
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal joined_group payload: %w", err)
		}
		return &evt, nil
	case payloadTypeGroupInfo:
		var evt whatsmeowevents.GroupInfo
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal group_info payload: %w", err)
		}
		return &evt, nil
	case payloadTypePicture:
		var evt whatsmeowevents.Picture
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal picture payload: %w", err)
		}
		return &evt, nil
	case payloadTypeUndecryptable:
		var evt whatsmeowevents.UndecryptableMessage
		if err := json.Unmarshal(payload.Data, &evt); err != nil {
			return nil, fmt.Errorf("unmarshal undecryptable payload: %w", err)
		}
		return &evt, nil
	default:
		return nil, fmt.Errorf("unsupported raw payload encoding type %q", payload.Type)
	}
}

func serializeMessage(evt *whatsmeowevents.Message) (*serializableMessage, error) {
	msgJSON, err := marshalProto(evt.Message)
	if err != nil {
		return nil, fmt.Errorf("marshal message proto: %w", err)
	}

	rawJSON, err := marshalProto(evt.RawMessage)
	if err != nil {
		return nil, fmt.Errorf("marshal raw message proto: %w", err)
	}

	var sourceJSON json.RawMessage
	if evt.SourceWebMsg != nil {
		if evt.SourceWebMsg.Key != nil && evt.SourceWebMsg.Key.ID != nil && len(evt.SourceWebMsg.Key.GetID()) > 0 {
			var marshalErr error
			sourceJSON, marshalErr = marshalProto(evt.SourceWebMsg)
			if marshalErr != nil {
				return nil, fmt.Errorf("marshal source web message proto: %w", marshalErr)
			}
		}
	}

	ser := &serializableMessage{
		Info:                  evt.Info,
		Message:               msgJSON,
		RawMessage:            rawJSON,
		SourceWebMessage:      sourceJSON,
		UnavailableRequestID:  string(evt.UnavailableRequestID),
		RetryCount:            evt.RetryCount,
		NewsletterMeta:        cloneNewsletterMeta(evt.NewsletterMeta),
		IsEphemeral:           evt.IsEphemeral,
		IsViewOnce:            evt.IsViewOnce,
		IsViewOnceV2:          evt.IsViewOnceV2,
		IsViewOnceV2Extension: evt.IsViewOnceV2Extension,
		IsDocumentWithCaption: evt.IsDocumentWithCaption,
		IsLottieSticker:       evt.IsLottieSticker,
		IsBotInvoke:           evt.IsBotInvoke,
		IsEdit:                evt.IsEdit,
	}

	return ser, nil
}

func deserializeMessage(ser *serializableMessage) (*whatsmeowevents.Message, error) {
	evt := &whatsmeowevents.Message{
		Info:                  ser.Info,
		UnavailableRequestID:  whatsmeowtypes.MessageID(ser.UnavailableRequestID),
		RetryCount:            ser.RetryCount,
		NewsletterMeta:        cloneNewsletterMeta(ser.NewsletterMeta),
		IsEphemeral:           ser.IsEphemeral,
		IsViewOnce:            ser.IsViewOnce,
		IsViewOnceV2:          ser.IsViewOnceV2,
		IsViewOnceV2Extension: ser.IsViewOnceV2Extension,
		IsDocumentWithCaption: ser.IsDocumentWithCaption,
		IsLottieSticker:       ser.IsLottieSticker,
		IsBotInvoke:           ser.IsBotInvoke,
		IsEdit:                ser.IsEdit,
	}

	if len(ser.Message) > 0 {
		evt.Message = &waE2E.Message{}
		if err := protoUnmarshalOpts.Unmarshal(ser.Message, evt.Message); err != nil {
			return nil, fmt.Errorf("unmarshal message proto: %w", err)
		}
	}

	if len(ser.RawMessage) > 0 {
		evt.RawMessage = &waE2E.Message{}
		if err := protoUnmarshalOpts.Unmarshal(ser.RawMessage, evt.RawMessage); err != nil {
			return nil, fmt.Errorf("unmarshal raw message proto: %w", err)
		}
	}

	if len(ser.SourceWebMessage) > 0 {
		evt.SourceWebMsg = &waWeb.WebMessageInfo{}
		if err := protoUnmarshalOpts.Unmarshal(ser.SourceWebMessage, evt.SourceWebMsg); err != nil {
			return nil, fmt.Errorf("unmarshal source web message proto: %w", err)
		}
	}

	return evt, nil
}

func marshalProto(msg proto.Message) (json.RawMessage, error) {
	if msg == nil {
		return nil, nil
	}
	data, err := protoMarshalOpts.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func cloneRaw(src json.RawMessage) json.RawMessage {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func cloneStringPtr(src *string) *string {
	if src == nil {
		return nil
	}
	cloned := *src
	return &cloned
}

func cloneInt64Ptr(src *int64) *int64 {
	if src == nil {
		return nil
	}
	cloned := *src
	return &cloned
}

func cloneNewsletterMeta(src *whatsmeowevents.NewsletterMessageMeta) *whatsmeowevents.NewsletterMessageMeta {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}
