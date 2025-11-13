package whatsmeow

import (
	"context"
	"fmt"
	"log/slog"

	"go.mau.fi/whatsmeow"
	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

type storePollDecrypter struct {
	client *whatsmeow.Client
	log    *slog.Logger
}

func newPollDecrypter(client *whatsmeow.Client, log *slog.Logger) eventctx.PollDecrypter {
	if client == nil {
		return nil
	}
	return &storePollDecrypter{client: client, log: log}
}

func (d *storePollDecrypter) DecryptPollVote(ctx context.Context, msg *events.Message) (*waE2E.PollVoteMessage, error) {
	if d == nil || d.client == nil {
		return nil, fmt.Errorf("poll decrypter not configured")
	}
	return d.client.DecryptPollVote(ctx, msg)
}
