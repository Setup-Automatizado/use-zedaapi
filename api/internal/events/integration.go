package events

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

type IntegrationHelper struct {
	log          *slog.Logger
	orchestrator *Orchestrator
}

func NewIntegrationHelper(ctx context.Context, orchestrator *Orchestrator) *IntegrationHelper {
	log := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "event_integration"),
	)

	return &IntegrationHelper{
		log:          log,
		orchestrator: orchestrator,
	}
}

func (h *IntegrationHelper) EnsureRegistered(ctx context.Context, instanceID uuid.UUID) error {
	if h == nil || h.orchestrator == nil {
		return nil
	}

	if !h.orchestrator.IsInstanceRegistered(instanceID) {
		if err := h.orchestrator.RegisterInstance(ctx, instanceID); err != nil {
			return err
		}
	}

	return nil
}

func (h *IntegrationHelper) WrapEventHandler(ctx context.Context, instanceID uuid.UUID, evt interface{}) {
	if err := h.orchestrator.HandleEvent(ctx, instanceID, evt); err != nil {
		h.log.ErrorContext(ctx, "failed to handle event",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()),
		)
	}
}

func (h *IntegrationHelper) OnInstanceConnect(ctx context.Context, instanceID uuid.UUID) error {
	if err := h.EnsureRegistered(ctx, instanceID); err != nil {
		h.log.ErrorContext(ctx, "failed to register instance",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (h *IntegrationHelper) OnInstanceDisconnect(ctx context.Context, instanceID uuid.UUID) error {
	if err := h.orchestrator.FlushInstance(instanceID); err != nil {
		h.log.WarnContext(ctx, "failed to flush instance buffer",
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()),
		)
	}

	return nil
}

func (h *IntegrationHelper) OnInstanceRemove(ctx context.Context, instanceID uuid.UUID) error {
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
