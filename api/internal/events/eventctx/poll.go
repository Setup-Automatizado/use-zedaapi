package eventctx

import (
	"context"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

type PollDecrypter interface {
	DecryptPollVote(ctx context.Context, msg *events.Message) (*waE2E.PollVoteMessage, error)
}

const pollDecrypterKey contextKey = "poll-decrypter"

func WithPollDecrypter(ctx context.Context, decrypter PollDecrypter) context.Context {
	if decrypter == nil {
		return ctx
	}
	return context.WithValue(ctx, pollDecrypterKey, decrypter)
}

func PollDecrypterFromContext(ctx context.Context) PollDecrypter {
	if ctx == nil {
		return nil
	}
	if decrypter, ok := ctx.Value(pollDecrypterKey).(PollDecrypter); ok {
		return decrypter
	}
	return nil
}
