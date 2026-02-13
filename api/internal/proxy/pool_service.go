package proxy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// captureServiceError logs and captures an error in Sentry with component+operation tags.
func captureServiceError(err error, operation string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("component", "pool_service")
		scope.SetTag("operation", operation)
		sentry.CaptureException(err)
	})
}

// PoolService provides business logic for proxy pool operations.
type PoolService struct {
	manager *PoolManager
	repo    *PoolRepository
	log     *slog.Logger
}

// NewPoolService creates a new PoolService.
func NewPoolService(manager *PoolManager, repo *PoolRepository, log *slog.Logger) *PoolService {
	return &PoolService{
		manager: manager,
		repo:    repo,
		log:     log.With(slog.String("component", "pool_service")),
	}
}

// CreateProvider validates and creates a new proxy provider.
func (s *PoolService) CreateProvider(ctx context.Context, req CreateProviderRequest) (*ProviderRecord, error) {
	logger := logging.ContextLogger(ctx, s.log)

	if req.Name == "" {
		return nil, fmt.Errorf("provider name is required")
	}
	if req.ProviderType == "" {
		return nil, fmt.Errorf("provider type is required")
	}

	rec, err := s.repo.CreateProvider(ctx, req)
	if err != nil {
		captureServiceError(fmt.Errorf("create provider %q: %w", req.Name, err), "create_provider")
		return nil, fmt.Errorf("create provider: %w", err)
	}

	logger.Info("provider created",
		slog.String("provider_id", rec.ID.String()),
		slog.String("name", rec.Name))

	return rec, nil
}

// GetProvider retrieves a proxy provider by ID.
func (s *PoolService) GetProvider(ctx context.Context, id uuid.UUID) (*ProviderRecord, error) {
	return s.repo.GetProvider(ctx, id)
}

// ListProviders returns all proxy providers.
func (s *PoolService) ListProviders(ctx context.Context) ([]ProviderRecord, error) {
	return s.repo.ListProviders(ctx)
}

// UpdateProvider applies a partial update to a proxy provider.
func (s *PoolService) UpdateProvider(ctx context.Context, id uuid.UUID, req UpdateProviderRequest) (*ProviderRecord, error) {
	logger := logging.ContextLogger(ctx, s.log)

	rec, err := s.repo.UpdateProvider(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("update provider: %w", err)
	}

	logger.Info("provider updated",
		slog.String("provider_id", id.String()))

	return rec, nil
}

// DeleteProvider unregisters and deletes a proxy provider.
func (s *PoolService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	logger := logging.ContextLogger(ctx, s.log)

	s.manager.UnregisterProvider(id)

	if err := s.repo.DeleteProvider(ctx, id); err != nil {
		captureServiceError(fmt.Errorf("delete provider %s: %w", id, err), "delete_provider")
		return fmt.Errorf("delete provider: %w", err)
	}

	logger.Info("provider deleted",
		slog.String("provider_id", id.String()))

	return nil
}

// TriggerSync triggers a sync for a specific provider.
func (s *PoolService) TriggerSync(ctx context.Context, id uuid.UUID) error {
	logger := logging.ContextLogger(ctx, s.log)

	if err := s.manager.SyncProvider(ctx, id); err != nil {
		captureServiceError(fmt.Errorf("trigger sync provider %s: %w", id, err), "trigger_sync")
		return fmt.Errorf("trigger sync: %w", err)
	}

	logger.Info("provider sync triggered",
		slog.String("provider_id", id.String()))

	return nil
}

// GetPoolStats returns aggregate pool statistics.
func (s *PoolService) GetPoolStats(ctx context.Context) (*PoolStats, error) {
	return s.repo.GetPoolStats(ctx)
}

// ListPoolProxies returns a paginated list of pool proxies with optional filters.
func (s *PoolService) ListPoolProxies(ctx context.Context, providerID *uuid.UUID, status *string, limit, offset int) ([]PoolProxyRecord, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPoolProxies(ctx, providerID, status, limit, offset)
}

// AssignPoolProxy assigns a pool proxy to an instance.
func (s *PoolService) AssignPoolProxy(ctx context.Context, instanceID uuid.UUID, req AssignPoolProxyRequest) (*AssignmentRecord, error) {
	logger := logging.ContextLogger(ctx, s.log)

	opts := AssignOptions{
		ProviderID:   req.ProviderID,
		CountryCodes: req.CountryCodes,
		NoWebsocket:  req.NoWebsocket,
		OnlyLogin:    req.OnlyLogin,
		NoMedia:      req.NoMedia,
	}

	assignment, err := s.manager.AssignProxy(ctx, instanceID, opts)
	if err != nil {
		captureServiceError(fmt.Errorf("assign pool proxy to %s: %w", instanceID, err), "assign_pool_proxy")
		return nil, fmt.Errorf("assign pool proxy: %w", err)
	}

	logger.Info("pool proxy assigned",
		slog.String("instance_id", instanceID.String()),
		slog.String("assignment_id", assignment.ID.String()))

	return assignment, nil
}

// ReleasePoolProxy releases the active pool proxy for an instance.
func (s *PoolService) ReleasePoolProxy(ctx context.Context, instanceID uuid.UUID) error {
	logger := logging.ContextLogger(ctx, s.log)

	if err := s.manager.ReleaseProxy(ctx, instanceID, "manual"); err != nil {
		return fmt.Errorf("release pool proxy: %w", err)
	}

	logger.Info("pool proxy released",
		slog.String("instance_id", instanceID.String()))

	return nil
}

// GetPoolAssignment retrieves the active pool proxy assignment for an instance.
func (s *PoolService) GetPoolAssignment(ctx context.Context, instanceID uuid.UUID) (*AssignmentRecord, error) {
	return s.repo.GetActiveAssignment(ctx, instanceID)
}

// CreateGroup creates a new proxy group.
func (s *PoolService) CreateGroup(ctx context.Context, name string, providerID *uuid.UUID, maxInstances int, countryCode *string) (*GroupRecord, error) {
	logger := logging.ContextLogger(ctx, s.log)

	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if maxInstances <= 0 {
		maxInstances = 10
	}

	rec, err := s.repo.CreateGroup(ctx, name, providerID, maxInstances, countryCode)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	logger.Info("proxy group created",
		slog.String("group_id", rec.ID.String()),
		slog.String("name", name))

	return rec, nil
}

// ListGroups returns all proxy groups.
func (s *PoolService) ListGroups(ctx context.Context) ([]GroupRecord, error) {
	return s.repo.ListGroups(ctx)
}

// DeleteGroup deletes a proxy group.
func (s *PoolService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	logger := logging.ContextLogger(ctx, s.log)

	if err := s.repo.DeleteGroup(ctx, id); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	logger.Info("proxy group deleted",
		slog.String("group_id", id.String()))

	return nil
}

// BulkAssignPoolProxies triggers a bulk assignment of pool proxies to unassigned instances.
func (s *PoolService) BulkAssignPoolProxies(ctx context.Context, req BulkAssignRequest) (*BulkAssignResult, error) {
	logger := logging.ContextLogger(ctx, s.log)

	result, err := s.manager.BulkAssignProxies(ctx, req)
	if err != nil {
		captureServiceError(fmt.Errorf("bulk assign proxies: %w", err), "bulk_assign")
		return nil, fmt.Errorf("bulk assign: %w", err)
	}

	logger.Info("bulk assign completed",
		slog.Int("total", result.Total),
		slog.Int("assigned", result.Assigned),
		slog.Int("skipped", result.Skipped),
		slog.Int("failed", result.Failed))

	return result, nil
}

// AssignToGroup assigns an instance to a proxy group.
func (s *PoolService) AssignToGroup(ctx context.Context, instanceID uuid.UUID, groupID uuid.UUID) (*AssignmentRecord, error) {
	logger := logging.ContextLogger(ctx, s.log)

	group, err := s.repo.GetGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}

	opts := AssignOptions{
		ProviderID: group.ProviderID,
		GroupID:    &groupID,
	}
	if group.CountryCode != nil {
		opts.CountryCodes = []string{*group.CountryCode}
	}

	assignment, err := s.manager.AssignProxy(ctx, instanceID, opts)
	if err != nil {
		return nil, fmt.Errorf("assign to group: %w", err)
	}

	logger.Info("instance assigned to group",
		slog.String("instance_id", instanceID.String()),
		slog.String("group_id", groupID.String()),
		slog.String("assignment_id", assignment.ID.String()))

	return assignment, nil
}
