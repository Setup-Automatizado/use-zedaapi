package whatsmeow

import (
	"context"
	"log/slog"

	"go.mau.fi/whatsmeow"
	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/types"
)

type storeLIDResolver struct {
	client *whatsmeow.Client
	log    *slog.Logger
}

func newStoreLIDResolver(client *whatsmeow.Client, log *slog.Logger) eventctx.LIDResolver {
	if client == nil || client.Store == nil || client.Store.LIDs == nil {
		return nil
	}
	return &storeLIDResolver{client: client, log: log}
}

func (r *storeLIDResolver) PNForLID(ctx context.Context, lid types.JID) (types.JID, bool) {
	if r == nil || r.client == nil || r.client.Store == nil || r.client.Store.LIDs == nil {
		return types.JID{}, false
	}
	if lid.Server != types.HiddenUserServer {
		return types.JID{}, false
	}

	pn, err := r.client.Store.LIDs.GetPNForLID(ctx, lid)
	if err != nil {
		if r.log != nil {
			r.log.Debug("failed to resolve PN for LID",
				slog.String("lid", lid.String()),
				slog.String("error", err.Error()))
		}
		return types.JID{}, false
	}
	if pn.IsEmpty() {
		return types.JID{}, false
	}
	return pn.ToNonAD(), true
}
