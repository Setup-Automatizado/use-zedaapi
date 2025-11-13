package whatsmeow

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	eventtypes "go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (t *Transformer) transformNewsletterJoin(ctx context.Context, logger *slog.Logger, join *events.NewsletterJoin) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "newsletter_join", join)
	populateNewsletterMetadata(event, &join.NewsletterMetadata)
	logger.InfoContext(ctx, "transformed newsletter join",
		slog.String("event_id", event.EventID.String()),
		slog.String("newsletter_id", join.ID.String()),
	)
	return event, nil
}

func (t *Transformer) transformNewsletterLeave(ctx context.Context, logger *slog.Logger, leave *events.NewsletterLeave) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "newsletter_leave", leave)
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	if !leave.ID.IsEmpty() {
		event.Metadata["newsletter_jid"] = leave.ID.String()
	}
	event.Metadata["newsletter_role"] = string(leave.Role)
	if event.Metadata["timestamp"] == "" {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	}
	logger.InfoContext(ctx, "transformed newsletter leave",
		slog.String("event_id", event.EventID.String()),
		slog.String("newsletter_id", leave.ID.String()),
		slog.String("role", string(leave.Role)),
	)
	return event, nil
}

func (t *Transformer) transformNewsletterMuteChange(ctx context.Context, logger *slog.Logger, mute *events.NewsletterMuteChange) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "newsletter_mute_change", mute)
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	if !mute.ID.IsEmpty() {
		event.Metadata["newsletter_jid"] = mute.ID.String()
	}
	event.Metadata["newsletter_mute"] = string(mute.Mute)
	if event.Metadata["timestamp"] == "" {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	}
	logger.InfoContext(ctx, "transformed newsletter mute change",
		slog.String("event_id", event.EventID.String()),
		slog.String("newsletter_id", mute.ID.String()),
		slog.String("mute", string(mute.Mute)),
	)
	return event, nil
}

func (t *Transformer) transformNewsletterLiveUpdate(ctx context.Context, logger *slog.Logger, live *events.NewsletterLiveUpdate) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "newsletter_live_update", live)
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	if !live.JID.IsEmpty() {
		event.Metadata["newsletter_jid"] = live.JID.String()
		event.Metadata["chat"] = live.JID.String()
		event.Metadata["from"] = live.JID.String()
		event.Metadata["sender"] = live.JID.String()
	}
	event.Metadata["from_me"] = "false"
	event.Metadata["is_group"] = "false"
	event.Metadata["newsletter_message_count"] = fmt.Sprintf("%d", len(live.Messages))
	if len(live.Messages) > 0 {
		message := live.Messages[0]
		event.Metadata["message_id"] = message.MessageID
		event.Metadata["timestamp"] = fmt.Sprintf("%d", message.Timestamp.Unix())
	}
	logger.InfoContext(ctx, "transformed newsletter live update",
		slog.String("event_id", event.EventID.String()),
		slog.String("newsletter_id", live.JID.String()),
		slog.Int("message_count", len(live.Messages)),
	)
	return event, nil
}

func populateNewsletterMetadata(event *eventtypes.InternalEvent, meta *types.NewsletterMetadata) {
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	if meta == nil {
		return
	}
	if !meta.ID.IsEmpty() {
		event.Metadata["newsletter_jid"] = meta.ID.String()
	}
	if name := strings.TrimSpace(meta.ThreadMeta.Name.Text); name != "" {
		event.Metadata["newsletter_name"] = name
	}
	if meta.ThreadMeta.Picture != nil {
		if url := strings.TrimSpace(meta.ThreadMeta.Picture.URL); url != "" {
			event.Metadata["newsletter_photo"] = url
		}
	}
	if meta.ViewerMeta != nil {
		if meta.ViewerMeta.Role != "" {
			event.Metadata["newsletter_role"] = string(meta.ViewerMeta.Role)
		}
		if meta.ViewerMeta.Mute != "" {
			event.Metadata["newsletter_mute"] = string(meta.ViewerMeta.Mute)
		}
	}
	if event.Metadata["timestamp"] == "" {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	}
}

func (t *Transformer) transformPushName(ctx context.Context, logger *slog.Logger, push *events.PushName) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "push_name", push)
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata["contact_jid"] = push.JID.String()
	event.Metadata["old_name"] = push.OldPushName
	event.Metadata["new_name"] = push.NewPushName
	if push.Message != nil {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", push.Message.Timestamp.Unix())
	} else if event.Metadata["timestamp"] == "" {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	}
	logger.InfoContext(ctx, "transformed push name",
		slog.String("event_id", event.EventID.String()),
		slog.String("jid", push.JID.String()),
		slog.String("name", push.NewPushName),
	)
	return event, nil
}

func (t *Transformer) transformBusinessName(ctx context.Context, logger *slog.Logger, bn *events.BusinessName) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "business_name", bn)
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata["contact_jid"] = bn.JID.String()
	event.Metadata["old_name"] = bn.OldBusinessName
	event.Metadata["new_name"] = bn.NewBusinessName
	if bn.Message != nil {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", bn.Message.Timestamp.Unix())
	} else if event.Metadata["timestamp"] == "" {
		event.Metadata["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	}
	logger.InfoContext(ctx, "transformed business name",
		slog.String("event_id", event.EventID.String()),
		slog.String("jid", bn.JID.String()),
		slog.String("name", bn.NewBusinessName),
	)
	return event, nil
}

func (t *Transformer) transformUserAbout(ctx context.Context, logger *slog.Logger, about *events.UserAbout) (*eventtypes.InternalEvent, error) {
	event := newInternalEvent(t.instanceID, "user_about", about)
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata["contact_jid"] = about.JID.String()
	event.Metadata["about_status"] = about.Status
	event.Metadata["timestamp"] = fmt.Sprintf("%d", about.Timestamp.Unix())
	logger.InfoContext(ctx, "transformed user about",
		slog.String("event_id", event.EventID.String()),
		slog.String("jid", about.JID.String()),
	)
	return event, nil
}
