package eventctx

import (
	"context"

	"go.mau.fi/whatsmeow/types"
)

type contextKey string

const (
	contactProviderKey contextKey = "contact-metadata-provider"
	lidResolverKey     contextKey = "lid-resolver"
)

type ContactMetadataProvider interface {
	ContactName(ctx context.Context, jid types.JID) string
	ContactPhoto(ctx context.Context, jid types.JID) string
}

type ContactPhotoDetails struct {
	PreviewURL string
	PreviewID  string
	FullURL    string
	FullID     string
}

type ContactPhotoDetailProvider interface {
	ContactPhotoDetails(ctx context.Context, jid types.JID) ContactPhotoDetails
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

type LIDResolver interface {
	PNForLID(ctx context.Context, lid types.JID) (types.JID, bool)
}

func WithLIDResolver(ctx context.Context, resolver LIDResolver) context.Context {
	if resolver == nil {
		return ctx
	}
	return context.WithValue(ctx, lidResolverKey, resolver)
}

func LIDResolverFromContext(ctx context.Context) LIDResolver {
	if resolver, ok := ctx.Value(lidResolverKey).(LIDResolver); ok {
		return resolver
	}
	return nil
}
