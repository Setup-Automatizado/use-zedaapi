package whatsmeow

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"go.mau.fi/whatsmeow/api/internal/observability"
)

// CallRejectConfig holds the call rejection settings for an instance.
type CallRejectConfig struct {
	Enabled bool
	Message string
}

// CallRejectConfigProvider retrieves call rejection settings for an instance.
// The repository returns (enabled bool, message *string, error).
type CallRejectConfigProvider interface {
	GetCallRejectConfig(ctx context.Context, instanceID uuid.UUID) (bool, *string, error)
}

// callRejectHandler handles automatic call rejection for WhatsApp instances.
type callRejectHandler struct {
	client         *whatsmeow.Client
	instanceID     uuid.UUID
	configProvider CallRejectConfigProvider
	log            *slog.Logger
}

// newCallRejectHandler creates a new call rejection handler.
func newCallRejectHandler(
	client *whatsmeow.Client,
	instanceID uuid.UUID,
	configProvider CallRejectConfigProvider,
	log *slog.Logger,
) *callRejectHandler {
	return &callRejectHandler{
		client:         client,
		instanceID:     instanceID,
		configProvider: configProvider,
		log: log.With(
			slog.String("component", "call_reject_handler"),
			slog.String("instance_id", instanceID.String()),
		),
	}
}

// HandleCallOffer processes a CallOffer event and rejects the call if configured.
func (h *callRejectHandler) HandleCallOffer(ctx context.Context, evt *events.CallOffer) {
	if evt == nil || h.client == nil {
		return
	}

	// Get call rejection config
	enabled, message, err := h.configProvider.GetCallRejectConfig(ctx, h.instanceID)
	if err != nil {
		h.log.Error("failed to get call reject config",
			slog.String("error", err.Error()))
		return
	}

	if !enabled {
		h.log.Debug("call rejection not enabled for instance")
		return
	}

	// Build config struct for internal use
	config := &CallRejectConfig{
		Enabled: enabled,
		Message: "",
	}
	if message != nil {
		config.Message = *message
	}

	// Detect if this is a desktop or mobile call
	// ZuckZapGo pattern: Desktop calls DON'T have ":" in the From JID
	// Mobile calls HAVE ":" (e.g., "user:device@server")
	hasDevicePart := strings.Contains(evt.From.String(), ":")
	isDesktopCall := !hasDevicePart

	h.log.Info("rejecting incoming call",
		slog.String("call_id", evt.CallID),
		slog.String("from", evt.From.String()),
		slog.String("call_creator", evt.CallCreator.String()),
		slog.Bool("is_desktop_call", isDesktopCall))

	// Reject the call with retries (3 attempts like ZuckZapGo)
	var rejectErr error
	for attempt := 1; attempt <= 3; attempt++ {
		rejectErr = h.client.RejectCall(ctx, evt.From, evt.CallID)
		if rejectErr == nil {
			h.log.Info("call rejected successfully",
				slog.String("call_id", evt.CallID),
				slog.Int("attempt", attempt))
			break
		}
		h.log.Warn("call rejection attempt failed",
			slog.String("call_id", evt.CallID),
			slog.Int("attempt", attempt),
			slog.String("error", rejectErr.Error()))

		if attempt < 3 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Desktop calls need additional rejection attempts (ZuckZapGo pattern)
	// Desktop calls don't have ':' in the From JID
	if !isDesktopCall && rejectErr == nil {
		time.Sleep(300 * time.Millisecond)

		for termAttempt := 1; termAttempt <= 2; termAttempt++ {
			termErr := h.client.RejectCall(ctx, evt.From, evt.CallID)
			if termErr != nil {
				h.log.Warn("additional desktop rejection attempt failed",
					slog.String("call_id", evt.CallID),
					slog.Int("term_attempt", termAttempt),
					slog.String("error", termErr.Error()))
			} else {
				h.log.Info("additional desktop rejection attempt successful",
					slog.String("call_id", evt.CallID),
					slog.Int("term_attempt", termAttempt))
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	if rejectErr != nil {
		h.log.Error("failed to reject call after all attempts",
			slog.String("call_id", evt.CallID),
			slog.String("error", rejectErr.Error()))
		observability.CaptureWorkerException(ctx, "call_reject_handler", "reject_call", h.instanceID.String(), rejectErr)
		return
	}

	// Wait before sending message (ZuckZapGo pattern)
	time.Sleep(500 * time.Millisecond)

	// Send rejection message if configured
	if config.Message != "" {
		messageJID := h.resolveMessageJID(ctx, evt)
		if messageJID.IsEmpty() {
			h.log.Warn("could not resolve JID for rejection message",
				slog.String("call_id", evt.CallID))
			return
		}

		h.sendRejectionMessage(ctx, messageJID, config.Message, evt.CallID)
	}
}

// HandleCallOfferNotice processes a CallOfferNotice event (group calls).
func (h *callRejectHandler) HandleCallOfferNotice(ctx context.Context, evt *events.CallOfferNotice) {
	if evt == nil || h.client == nil {
		return
	}

	// Get call rejection config
	enabled, message, err := h.configProvider.GetCallRejectConfig(ctx, h.instanceID)
	if err != nil {
		h.log.Error("failed to get call reject config",
			slog.String("error", err.Error()))
		return
	}

	if !enabled {
		h.log.Debug("call rejection not enabled for instance")
		return
	}

	// Build config struct for internal use
	config := &CallRejectConfig{
		Enabled: enabled,
		Message: "",
	}
	if message != nil {
		config.Message = *message
	}

	// Detect if this is a desktop or mobile call
	// ZuckZapGo pattern: Desktop calls DON'T have ":" in the From JID
	// Mobile calls HAVE ":" (e.g., "user:device@server")
	hasDevicePart := strings.Contains(evt.From.String(), ":")
	isDesktopCall := !hasDevicePart

	h.log.Info("rejecting incoming call notice",
		slog.String("call_id", evt.CallID),
		slog.String("from", evt.From.String()),
		slog.String("call_creator", evt.CallCreator.String()),
		slog.String("type", evt.Type),
		slog.Bool("is_desktop_call", isDesktopCall))

	// Reject the call with retries (3 attempts like ZuckZapGo)
	var rejectErr error
	for attempt := 1; attempt <= 3; attempt++ {
		rejectErr = h.client.RejectCall(ctx, evt.From, evt.CallID)
		if rejectErr == nil {
			h.log.Info("call notice rejected successfully",
				slog.String("call_id", evt.CallID),
				slog.Int("attempt", attempt))
			break
		}
		h.log.Warn("call notice rejection attempt failed",
			slog.String("call_id", evt.CallID),
			slog.Int("attempt", attempt),
			slog.String("error", rejectErr.Error()))

		if attempt < 3 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Desktop calls need additional rejection attempts (ZuckZapGo pattern)
	// Desktop calls don't have ':' in the From JID
	if !isDesktopCall && rejectErr == nil {
		time.Sleep(300 * time.Millisecond)

		for termAttempt := 1; termAttempt <= 2; termAttempt++ {
			termErr := h.client.RejectCall(ctx, evt.From, evt.CallID)
			if termErr != nil {
				h.log.Warn("additional desktop rejection attempt failed (notice)",
					slog.String("call_id", evt.CallID),
					slog.Int("term_attempt", termAttempt),
					slog.String("error", termErr.Error()))
			} else {
				h.log.Info("additional desktop rejection attempt successful (notice)",
					slog.String("call_id", evt.CallID),
					slog.Int("term_attempt", termAttempt))
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	if rejectErr != nil {
		h.log.Error("failed to reject call notice after all attempts",
			slog.String("call_id", evt.CallID),
			slog.String("error", rejectErr.Error()))
		observability.CaptureWorkerException(ctx, "call_reject_handler", "reject_call_notice", h.instanceID.String(), rejectErr)
		return
	}

	// Wait before sending message (ZuckZapGo pattern)
	time.Sleep(500 * time.Millisecond)

	// Send rejection message if configured
	if config.Message != "" {
		messageJID := h.resolveMessageJIDFromNotice(ctx, evt)
		if messageJID.IsEmpty() {
			h.log.Warn("could not resolve JID for rejection message",
				slog.String("call_id", evt.CallID))
			return
		}

		h.sendRejectionMessage(ctx, messageJID, config.Message, evt.CallID)
	}
}

// resolveMessageJID resolves the correct JID to send the rejection message to.
// This implements the 3-strategy approach from ZuckZapGo:
// 1. Use CallCreatorAlt if available and on DefaultUserServer
// 2. Convert LID to PN using Store.LIDs.GetPNForLID
// 3. Fallback to From.User with DefaultUserServer
func (h *callRejectHandler) resolveMessageJID(ctx context.Context, evt *events.CallOffer) types.JID {
	// Strategy 1: Use CallCreatorAlt if available and on DefaultUserServer
	// IMPORTANT: Always strip device part using ToNonAD() - desktop clients
	// include device ID (e.g., :18) which causes SendMessage to fail with
	// "message recipient must be a user JID with no device part"
	if !evt.CallCreatorAlt.IsEmpty() {
		if evt.CallCreatorAlt.Server == types.DefaultUserServer {
			nonADJID := evt.CallCreatorAlt.ToNonAD()
			h.log.Debug("using CallCreatorAlt for message JID",
				slog.String("original_jid", evt.CallCreatorAlt.String()),
				slog.String("normalized_jid", nonADJID.String()))
			return nonADJID
		}
	}

	// Strategy 2: Convert LID to PN using Store.LIDs.GetPNForLID
	if evt.CallCreator.Server == types.HiddenUserServer {
		if h.client.Store != nil && h.client.Store.LIDs != nil {
			pn, err := h.client.Store.LIDs.GetPNForLID(ctx, evt.CallCreator)
			if err == nil && !pn.IsEmpty() {
				resolvedJID := pn.ToNonAD()
				h.log.Debug("resolved LID to PN for message JID",
					slog.String("lid", evt.CallCreator.String()),
					slog.String("pn", resolvedJID.String()))
				return resolvedJID
			}
			if err != nil {
				h.log.Debug("failed to resolve LID to PN",
					slog.String("lid", evt.CallCreator.String()),
					slog.String("error", err.Error()))
			}
		}
	}

	// Strategy 3: Fallback to From.User with DefaultUserServer
	fallbackJID := types.NewJID(evt.From.User, types.DefaultUserServer)
	h.log.Debug("using fallback JID for message",
		slog.String("from_user", evt.From.User),
		slog.String("jid", fallbackJID.String()))
	return fallbackJID
}

// resolveMessageJIDFromNotice resolves the correct JID for CallOfferNotice events.
func (h *callRejectHandler) resolveMessageJIDFromNotice(ctx context.Context, evt *events.CallOfferNotice) types.JID {
	// Strategy 1: Use CallCreatorAlt if available and on DefaultUserServer
	// IMPORTANT: Always strip device part using ToNonAD() - desktop clients
	// include device ID (e.g., :18) which causes SendMessage to fail with
	// "message recipient must be a user JID with no device part"
	if !evt.CallCreatorAlt.IsEmpty() {
		if evt.CallCreatorAlt.Server == types.DefaultUserServer {
			nonADJID := evt.CallCreatorAlt.ToNonAD()
			h.log.Debug("using CallCreatorAlt for message JID (notice)",
				slog.String("original_jid", evt.CallCreatorAlt.String()),
				slog.String("normalized_jid", nonADJID.String()))
			return nonADJID
		}
	}

	// Strategy 2: Convert LID to PN using Store.LIDs.GetPNForLID
	if evt.CallCreator.Server == types.HiddenUserServer {
		if h.client.Store != nil && h.client.Store.LIDs != nil {
			pn, err := h.client.Store.LIDs.GetPNForLID(ctx, evt.CallCreator)
			if err == nil && !pn.IsEmpty() {
				resolvedJID := pn.ToNonAD()
				h.log.Debug("resolved LID to PN for message JID (notice)",
					slog.String("lid", evt.CallCreator.String()),
					slog.String("pn", resolvedJID.String()))
				return resolvedJID
			}
			if err != nil {
				h.log.Debug("failed to resolve LID to PN (notice)",
					slog.String("lid", evt.CallCreator.String()),
					slog.String("error", err.Error()))
			}
		}
	}

	// Strategy 3: Fallback to From.User with DefaultUserServer
	fallbackJID := types.NewJID(evt.From.User, types.DefaultUserServer)
	h.log.Debug("using fallback JID for message (notice)",
		slog.String("from_user", evt.From.User),
		slog.String("jid", fallbackJID.String()))
	return fallbackJID
}

// sendRejectionMessage sends the configured rejection message to the caller.
func (h *callRejectHandler) sendRejectionMessage(ctx context.Context, to types.JID, message string, callID string) {
	if to.IsEmpty() || message == "" {
		return
	}

	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	resp, err := h.client.SendMessage(ctx, to, msg)
	if err != nil {
		h.log.Error("failed to send call rejection message",
			slog.String("to", to.String()),
			slog.String("call_id", callID),
			slog.String("error", err.Error()))
		observability.CaptureWorkerException(ctx, "call_reject_handler", "send_rejection_message", h.instanceID.String(), err)
		return
	}

	h.log.Info("sent call rejection message",
		slog.String("to", to.String()),
		slog.String("call_id", callID),
		slog.String("message_id", resp.ID))
}
