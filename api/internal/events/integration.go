package events

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// IntegrationHelper provides helper methods for integrating events with ClientRegistry
type IntegrationHelper struct {
	log          *slog.Logger
	orchestrator *Orchestrator
}

// NewIntegrationHelper creates a new integration helper
func NewIntegrationHelper(ctx context.Context, orchestrator *Orchestrator) *IntegrationHelper {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_integration"),
	)

	return &IntegrationHelper{
		log:          log,
		orchestrator: orchestrator,
	}
}

// WrapEventHandler wraps the existing ClientRegistry.wrapEventHandler
// This should be called inside the existing wrapEventHandler function
func (h *IntegrationHelper) WrapEventHandler(ctx context.Context, instanceID uuid.UUID, evt interface{}) {
	// Forward event to orchestrator
	if err := h.orchestrator.HandleEvent(ctx, instanceID, evt); err != nil {
		h.log.ErrorContext(ctx, "failed to handle event",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()),
		)
	}
}

// OnInstanceConnect is called when an instance connects
func (h *IntegrationHelper) OnInstanceConnect(ctx context.Context, instanceID uuid.UUID) error {
	// Register instance with event system if not already registered
	if !h.orchestrator.IsInstanceRegistered(instanceID) {
		if err := h.orchestrator.RegisterInstance(ctx, instanceID); err != nil {
			h.log.ErrorContext(ctx, "failed to register instance",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()),
			)
			return err
		}
	}

	return nil
}

// OnInstanceDisconnect is called when an instance disconnects
func (h *IntegrationHelper) OnInstanceDisconnect(ctx context.Context, instanceID uuid.UUID) error {
	// Flush buffer before unregistering
	if err := h.orchestrator.FlushInstance(instanceID); err != nil {
		h.log.WarnContext(ctx, "failed to flush instance buffer",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()),
		)
	}

	return nil
}

// OnInstanceRemove is called when an instance is removed from registry
func (h *IntegrationHelper) OnInstanceRemove(ctx context.Context, instanceID uuid.UUID) error {
	// Unregister from event system
	if h.orchestrator.IsInstanceRegistered(instanceID) {
		if err := h.orchestrator.UnregisterInstance(ctx, instanceID); err != nil {
			h.log.ErrorContext(ctx, "failed to unregister instance",
				slog.String("instance_id", instanceID.String()),
				slog.String("error", err.Error()),
			)
			return err
		}
	}

	return nil
}
