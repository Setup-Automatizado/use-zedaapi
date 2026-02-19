package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	"go.mau.fi/whatsmeow/types"
)

// PollVoteProcessor handles sending poll vote messages through WhatsApp
type PollVoteProcessor struct {
	log         *slog.Logger
	echoEmitter *echo.Emitter
}

// NewPollVoteProcessor creates a new poll vote processor instance
func NewPollVoteProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *PollVoteProcessor {
	return &PollVoteProcessor{
		log:         log.With(slog.String("processor", "poll_vote")),
		echoEmitter: echoEmitter,
	}
}

// Process sends a poll vote via WhatsApp
func (p *PollVoteProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.PollVoteContent == nil {
		return fmt.Errorf("poll_vote_content is required for poll_vote messages")
	}

	content := args.PollVoteContent

	if content.PollID == "" {
		return fmt.Errorf("poll_id is required")
	}
	if content.PollSender == "" {
		return fmt.Errorf("poll_sender is required")
	}
	if len(content.Options) == 0 {
		return fmt.Errorf("at least one poll option is required")
	}

	// Parse chat JID
	chatJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Parse poll sender JID
	senderJID, err := types.ParseJID(content.PollSender)
	if err != nil {
		return fmt.Errorf("invalid poll sender: %w", err)
	}

	// Simulate typing indicator if DelayTyping is set
	if args.DelayTyping > 0 {
		if err := p.simulateTyping(client, chatJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Construct MessageInfo for the poll message
	pollInfo := &types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:     chatJID,
			Sender:   senderJID,
			IsFromMe: senderJID.User == client.Store.ID.User,
			IsGroup:  content.IsGroup,
		},
		ID:        types.MessageID(content.PollID),
		Timestamp: time.Now(),
	}

	// Build the poll vote message
	voteMsg, err := client.BuildPollVote(ctx, pollInfo, content.Options)
	if err != nil {
		return fmt.Errorf("build poll vote: %w", err)
	}

	// Send the poll vote
	resp, err := client.SendMessage(ctx, chatJID, voteMsg, BuildSendExtra(args))
	if err != nil {
		return fmt.Errorf("send poll vote: %w", err)
	}

	p.log.Info("poll vote sent successfully",
		slog.String("message_id", resp.ID),
		slog.String("phone", args.Phone),
		slog.String("poll_id", content.PollID),
		slog.Int("options_count", len(content.Options)))

	args.WhatsAppMessageID = resp.ID

	// Emit API echo event
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      chatJID,
			Message:           voteMsg,
			Timestamp:         resp.Timestamp,
			MessageType:       "poll_vote",
			ZaapID:            args.ZaapID,
			HasMedia:          false,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

func (p *PollVoteProcessor) simulateTyping(client *wameow.Client, jid types.JID, delayMs int64) error {
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
		return err
	}
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	if err := client.SendChatPresence(context.Background(), jid, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
		return err
	}
	return nil
}
