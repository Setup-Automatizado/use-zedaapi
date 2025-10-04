package eventctx

import (
	"context"

	"go.mau.fi/whatsmeow/types"
)

type contextKey string

const contactProviderKey contextKey = "contact-metadata-provider"

type ContactMetadataProvider interface {
	ContactName(ctx context.Context, jid types.JID) string
	ContactPhoto(ctx context.Context, jid types.JID) string
}

func WithContactProvider(ctx context.Context, provider ContactMetadataProvider) context.Context {
	if provider == nil {
		return ctx
	}
	return context.WithValue(ctx, contactProviderKey, provider)
}

func ContactProvider(ctx context.Context) ContactMetadataProvider {
	if provider, ok := ctx.Value(contactProviderKey).(ContactMetadataProvider); ok {
		return provider
	}
	return nil
}
