package queue

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// ContextInfoBuilder helps build WhatsApp ContextInfo for messages
// Supports: mentions, reply-to, ephemeral messages, private answers
type ContextInfoBuilder struct {
	client       *wameow.Client
	recipientJID types.JID
	args         *SendMessageArgs
	log          *slog.Logger
}

// NewContextInfoBuilder creates a new ContextInfo builder
func NewContextInfoBuilder(client *wameow.Client, recipientJID types.JID, args *SendMessageArgs, log *slog.Logger) *ContextInfoBuilder {
	return &ContextInfoBuilder{
		client:       client,
		recipientJID: recipientJID,
		args:         args,
		log:          log,
	}
}

// Build constructs the complete ContextInfo with all requested features
// Returns nil if no context features are needed
func (b *ContextInfoBuilder) Build(ctx context.Context) (*waProto.ContextInfo, error) {
	contextInfo := &waProto.ContextInfo{}
	hasContent := false

	// 1. Add mentions (mentioned, groupMentioned, mentionedAll)
	if err := b.addMentions(ctx, contextInfo); err != nil {
		b.log.Warn("failed to add mentions",
			slog.String("error", err.Error()))
		// Non-fatal, continue
	} else if len(contextInfo.MentionedJID) > 0 {
		hasContent = true
	}

	// 2. Add reply-to (messageId)
	if b.args.ReplyToMessageID != "" {
		stanzaID := b.args.ReplyToMessageID
		participant := b.recipientJID.String()
		contextInfo.StanzaID = &stanzaID
		contextInfo.Participant = &participant
		contextInfo.QuotedMessage = &waProto.Message{}
		hasContent = true
	}

	// 3. Add ephemeral/duration (disappearing messages)
	if b.args.Duration != nil && *b.args.Duration > 0 {
		expiration := uint32(*b.args.Duration)
		contextInfo.Expiration = &expiration
		hasContent = true

		b.log.Debug("configured ephemeral message",
			slog.Int("duration_seconds", *b.args.Duration))
	}

	// 4. Add private answer (for groups)
	if b.args.PrivateAnswer && b.recipientJID.Server == types.GroupServer {
		// For private answer in groups, we need to set the participant
		// to the message sender (not ourselves)
		// This is typically used when replying to a message in a group privately
		if b.args.ReplyToMessageID != "" {
			// Participant already set for reply-to
			b.log.Debug("private answer enabled with reply-to")
		} else {
			b.log.Warn("private answer requires reply-to message id")
		}
		hasContent = true
	}

	// Return nil if no context info was added
	if !hasContent {
		return nil, nil
	}

	return contextInfo, nil
}

// addMentions processes all mention types and adds to ContextInfo
func (b *ContextInfoBuilder) addMentions(ctx context.Context, contextInfo *waProto.ContextInfo) error {
	var mentionedJIDs []string

	// 1. Handle individual mentions (mentioned array)
	if len(b.args.Mentioned) > 0 {
		for _, phone := range b.args.Mentioned {
			jid := normalizePhoneToJID(phone, "@s.whatsapp.net")
			mentionedJIDs = append(mentionedJIDs, jid)
		}

		b.log.Debug("added individual mentions",
			slog.Int("count", len(b.args.Mentioned)))
	}

	// 2. Handle group mentions (groupMentioned array)
	if len(b.args.GroupMentioned) > 0 {
		for _, groupMention := range b.args.GroupMentioned {
			jid := normalizePhoneToJID(groupMention.Phone, "@g.us")
			mentionedJIDs = append(mentionedJIDs, jid)
		}

		b.log.Debug("added group mentions",
			slog.Int("count", len(b.args.GroupMentioned)))
	}

	// 3. Handle mention all (mentionedAll flag)
	if b.args.MentionedAll {
		if b.recipientJID.Server == types.GroupServer {
			// Fetch group participants
			groupInfo, err := b.client.GetGroupInfo(context.Background(), b.recipientJID)
			if err != nil {
				return fmt.Errorf("get group info for mentionedAll: %w", err)
			}

			// Add all participants to mentions
			for _, participant := range groupInfo.Participants {
				mentionedJIDs = append(mentionedJIDs, participant.JID.String())
			}

			b.log.Debug("added all group participants",
				slog.Int("count", len(groupInfo.Participants)))
		} else {
			b.log.Warn("mentionedAll only works for group chats",
				slog.String("recipient", b.recipientJID.String()))
		}
	}

	// Set mentions in ContextInfo
	if len(mentionedJIDs) > 0 {
		contextInfo.MentionedJID = mentionedJIDs

		b.log.Info("built mentions for message",
			slog.Int("total_mentions", len(mentionedJIDs)))
	}

	return nil
}

// normalizePhoneToJID ensures phone number has correct JID format
func normalizePhoneToJID(phone, defaultSuffix string) string {
	if strings.Contains(phone, "@") {
		return phone
	}
	return phone + defaultSuffix
}
