package proxy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
)

// PoolConfig holds configuration for the PoolManager.
type PoolConfig struct {
	SyncInterval         time.Duration
	DefaultCountryCodes  []string
	AssignmentRetryDelay time.Duration
	MaxAssignmentRetries int
}

// PoolConfigWithDefaults returns a copy with zero values replaced by defaults.
func (c PoolConfig) WithDefaults() PoolConfig {
	if c.SyncInterval <= 0 {
		c.SyncInterval = 5 * time.Minute
	}
	if c.AssignmentRetryDelay <= 0 {
		c.AssignmentRetryDelay = 200 * time.Millisecond
	}
	if c.MaxAssignmentRetries <= 0 {
		c.MaxAssignmentRetries = 3
	}
	return c
}

// PoolManager orchestrates proxy pool operations: assignment, release, swap, sync.
type PoolManager struct {
	repo            *PoolRepository
	registry        RegistrySwapper
	instanceUpdater InstanceProxyUpdater
	lockManager     locks.Manager
	providers       map[uuid.UUID]ProxyProvider
	providersMu     sync.RWMutex
	cfg             PoolConfig
	log             *slog.Logger
	metrics         *PoolMetrics

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewPoolManager creates a new PoolManager.
func NewPoolManager(repo *PoolRepository, registry RegistrySwapper, lockMgr locks.Manager, cfg PoolConfig, log *slog.Logger, metrics *PoolMetrics) *PoolManager {
	return &PoolManager{
		repo:        repo,
		registry:    registry,
		lockManager: lockMgr,
		providers:   make(map[uuid.UUID]ProxyProvider),
		cfg:         cfg.WithDefaults(),
		log:         log.With(slog.String("component", "pool_manager")),
		metrics:     metrics,
	}
}

// SetInstanceUpdater sets the instance proxy updater for persisting pool proxy
// assignments to the instances table. This ensures the proxy is applied on restarts.
func (m *PoolManager) SetInstanceUpdater(updater InstanceProxyUpdater) {
	m.instanceUpdater = updater
}

// GetRegistry returns the registry swapper for proxy operations.
func (m *PoolManager) GetRegistry() RegistrySwapper { return m.registry }

// GetInstanceUpdater returns the instance proxy updater.
func (m *PoolManager) GetInstanceUpdater() InstanceProxyUpdater { return m.instanceUpdater }

// RegisterProvider adds a proxy provider to the in-memory map.
func (m *PoolManager) RegisterProvider(providerID uuid.UUID, provider ProxyProvider) {
	m.providersMu.Lock()
	defer m.providersMu.Unlock()
	m.providers[providerID] = provider
	m.log.Info("provider registered",
		slog.String("provider_id", providerID.String()),
		slog.String("type", string(provider.Type())))
}

// UnregisterProvider removes a provider from the map and closes it.
func (m *PoolManager) UnregisterProvider(providerID uuid.UUID) {
	m.providersMu.Lock()
	provider, ok := m.providers[providerID]
	if ok {
		delete(m.providers, providerID)
	}
	m.providersMu.Unlock()

	if ok {
		if err := provider.Close(); err != nil {
			m.log.Warn("failed to close provider",
				slog.String("provider_id", providerID.String()),
				slog.String("error", err.Error()))
		}
		m.log.Info("provider unregistered", slog.String("provider_id", providerID.String()))
	}
}

// getProvider retrieves a provider from the in-memory map.
func (m *PoolManager) getProvider(providerID uuid.UUID) (ProxyProvider, bool) {
	m.providersMu.RLock()
	defer m.providersMu.RUnlock()
	p, ok := m.providers[providerID]
	return p, ok
}

// Start begins the periodic sync loop.
func (m *PoolManager) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)

	m.wg.Add(1)
	go m.syncLoop(ctx)

	m.log.Info("pool manager started",
		slog.Duration("sync_interval", m.cfg.SyncInterval),
		slog.Int("max_assignment_retries", m.cfg.MaxAssignmentRetries))
}

// Stop gracefully stops the pool manager and closes all providers.
func (m *PoolManager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()

	m.providersMu.Lock()
	for id, provider := range m.providers {
		if err := provider.Close(); err != nil {
			m.log.Warn("failed to close provider on shutdown",
				slog.String("provider_id", id.String()),
				slog.String("error", err.Error()))
		}
	}
	m.providers = make(map[uuid.UUID]ProxyProvider)
	m.providersMu.Unlock()

	m.log.Info("pool manager stopped")
}

func (m *PoolManager) syncLoop(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.cfg.SyncInterval)
	defer ticker.Stop()

	// Run immediately on start.
	m.syncAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.syncAll(ctx)
		}
	}
}

func (m *PoolManager) syncAll(ctx context.Context) {
	logger := logging.ContextLogger(ctx, m.log)

	providers, err := m.repo.ListProviders(ctx)
	if err != nil {
		logger.Error("list providers for sync", slog.String("error", err.Error()))
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "pool_manager")
			sentry.CaptureException(fmt.Errorf("list providers for sync: %w", err))
		})
		return
	}

	for _, rec := range providers {
		if !rec.Enabled {
			continue
		}
		if syncErr := m.SyncProvider(ctx, rec.ID); syncErr != nil {
			logger.Warn("provider sync failed",
				slog.String("provider_id", rec.ID.String()),
				slog.String("provider_name", rec.Name),
				slog.String("error", syncErr.Error()))
		}
	}
}

// SyncProvider synchronizes the proxy pool from a single provider.
// It fetches proxies page-by-page and upserts each page immediately to the
// database, so proxies become available without waiting for the full sync.
func (m *PoolManager) SyncProvider(ctx context.Context, providerID uuid.UUID) error {
	logger := logging.ContextLogger(ctx, m.log).With(
		slog.String("provider_id", providerID.String()))

	provider, ok := m.getProvider(providerID)
	if !ok {
		logger.Debug("provider not registered in memory, skipping sync")
		return nil
	}

	record, err := m.repo.GetProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("get provider record: %w", err)
	}

	start := time.Now()

	// Stream pages: fetch one page at a time, upsert immediately, track IDs.
	const pageSize = 100
	var (
		activeExternalIDs []string
		totalSynced       int
		page              = 1
		syncErr           error
	)

	for {
		select {
		case <-ctx.Done():
			syncErr = ctx.Err()
			goto done
		default:
		}

		entries, _, fetchErr := provider.ListProxies(ctx, ProxyFilter{Page: page, PageSize: pageSize})
		if fetchErr != nil {
			if totalSynced == 0 {
				// Complete failure on first page.
				duration := time.Since(start).Seconds()
				errStr := fetchErr.Error()
				_ = m.repo.UpdateProviderSyncStatus(ctx, providerID, &errStr, record.ProxyCount)
				if m.metrics != nil {
					m.metrics.RecordSync(providerID.String(), "error", duration)
				}
				sentry.WithScope(func(scope *sentry.Scope) {
					scope.SetTag("component", "pool_manager")
					scope.SetTag("provider_id", providerID.String())
					sentry.CaptureException(fmt.Errorf("sync provider %s page %d: %w", providerID, page, fetchErr))
				})
				return fmt.Errorf("sync provider page %d: %w", page, fetchErr)
			}
			// Partial failure: keep what we have.
			syncErr = fmt.Errorf("interrupted at page %d: %w", page, fetchErr)
			logger.Warn("sync interrupted, keeping persisted entries",
				slog.Int("page", page),
				slog.Int("persisted", totalSynced),
				slog.String("error", fetchErr.Error()))
			break
		}

		if len(entries) == 0 {
			break
		}

		// Upsert this page immediately.
		for _, entry := range entries {
			if _, upsertErr := m.repo.UpsertPoolProxy(ctx, providerID, entry, record.MaxInstancesPerProxy); upsertErr != nil {
				logger.Warn("upsert pool proxy failed",
					slog.String("external_id", entry.ExternalID),
					slog.String("error", upsertErr.Error()))
				continue
			}
			totalSynced++
			if entry.ExternalID != "" {
				activeExternalIDs = append(activeExternalIDs, entry.ExternalID)
			}
		}

		// Update sync status every 10 pages so the UI shows progress.
		if page%10 == 0 {
			_ = m.repo.UpdateProviderSyncStatus(ctx, providerID, nil, totalSynced)
			logger.Debug("sync progress",
				slog.Int("page", page),
				slog.Int("persisted", totalSynced))
		}

		if len(entries) < pageSize {
			break // Last page.
		}
		page++
	}

done:
	isPartialSync := syncErr != nil

	// Only retire proxies on full sync. On partial sync, unfetched pages may
	// contain valid proxies that would be incorrectly retired.
	var retired int64
	if !isPartialSync {
		var retireErr error
		retired, retireErr = m.repo.RetirePoolProxies(ctx, providerID, activeExternalIDs)
		if retireErr != nil {
			logger.Warn("retire stale pool proxies failed", slog.String("error", retireErr.Error()))
		} else if retired > 0 {
			logger.Info("retired stale pool proxies", slog.Int64("count", retired))
		}
	}

	// Final sync status update.
	var syncErrStr *string
	if isPartialSync {
		s := fmt.Sprintf("partial sync: %d entries persisted, %s", totalSynced, syncErr.Error())
		syncErrStr = &s
	}
	if err := m.repo.UpdateProviderSyncStatus(ctx, providerID, syncErrStr, totalSynced); err != nil {
		logger.Warn("update provider sync status failed", slog.String("error", err.Error()))
	}

	duration := time.Since(start).Seconds()
	syncStatus := "success"
	if isPartialSync {
		syncStatus = "partial"
	}
	if m.metrics != nil {
		m.metrics.RecordSync(providerID.String(), syncStatus, duration)
	}

	logger.Info("provider sync completed",
		slog.String("status", syncStatus),
		slog.Int("synced", totalSynced),
		slog.Int("pages", page),
		slog.Int64("retired", retired),
		slog.Float64("duration_s", duration))

	m.updatePoolSizeMetrics(ctx)

	if isPartialSync {
		return fmt.Errorf("partial sync (%d entries persisted): %w", totalSynced, syncErr)
	}
	return nil
}

// AssignProxy assigns a pool proxy to an instance with retry logic for concurrent safety.
func (m *PoolManager) AssignProxy(ctx context.Context, instanceID uuid.UUID, opts AssignOptions) (*AssignmentRecord, error) {
	logger := logging.ContextLogger(ctx, m.log).With(
		slog.String("instance_id", instanceID.String()))

	for attempt := 0; attempt < m.cfg.MaxAssignmentRetries; attempt++ {
		// 1. Check if instance already has an active assignment.
		existing, err := m.repo.GetActiveAssignment(ctx, instanceID)
		if err == nil && existing != nil {
			logger.Debug("instance already has active assignment",
				slog.String("assignment_id", existing.ID.String()))
			return existing, nil
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("check active assignment: %w", err)
		}

		// 2. Determine country codes.
		countryCodes := opts.CountryCodes
		if len(countryCodes) == 0 {
			countryCodes = m.cfg.DefaultCountryCodes
		}

		// 3. Find available proxy (concurrent-safe via FOR UPDATE SKIP LOCKED).
		poolProxy, err := m.repo.FindAvailableProxy(ctx, countryCodes, opts.ProviderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if attempt < m.cfg.MaxAssignmentRetries-1 {
					logger.Debug("no available proxy, retrying",
						slog.Int("attempt", attempt+1))
					time.Sleep(m.cfg.AssignmentRetryDelay)
					continue
				}
				if m.metrics != nil {
					m.metrics.RecordAssignment("unknown", "no_proxy")
				}
				return nil, fmt.Errorf("no available proxy found after %d attempts", m.cfg.MaxAssignmentRetries)
			}
			return nil, fmt.Errorf("find available proxy: %w", err)
		}

		// 4. Increment assigned count.
		if err := m.repo.IncrementAssignedCount(ctx, poolProxy.ID); err != nil {
			logger.Debug("increment assigned count failed, retrying",
				slog.String("pool_proxy_id", poolProxy.ID.String()),
				slog.String("error", err.Error()))
			continue
		}

		// 5. Create assignment record.
		assignment, err := m.repo.CreateAssignment(ctx, poolProxy.ID, instanceID, opts.GroupID, "auto")
		if err != nil {
			// Rollback the increment.
			if decErr := m.repo.DecrementAssignedCount(ctx, poolProxy.ID); decErr != nil {
				logger.Error("rollback decrement assigned count failed",
					slog.String("pool_proxy_id", poolProxy.ID.String()),
					slog.String("error", decErr.Error()))
			}
			logger.Debug("create assignment failed, retrying",
				slog.String("error", err.Error()))
			continue
		}

		// 6. Persist proxy URL to instances table so it survives restarts.
		if m.instanceUpdater != nil {
			if updateErr := m.instanceUpdater.UpdateProxyURL(ctx, instanceID, &poolProxy.ProxyURL, true); updateErr != nil {
				logger.Warn("failed to persist pool proxy to instance config",
					slog.String("error", updateErr.Error()))
			}
		}

		// 7. Apply proxy to live client via registry (hot-swap if connected).
		if m.registry != nil {
			if swapErr := m.registry.SwapProxy(ctx, instanceID, poolProxy.ProxyURL, opts.NoWebsocket, opts.OnlyLogin, opts.NoMedia); swapErr != nil {
				// Fallback to SetProxyAddress (will take effect on next connect).
				if applyErr := m.registry.ApplyProxy(ctx, instanceID, poolProxy.ProxyURL, opts.NoWebsocket, opts.OnlyLogin, opts.NoMedia); applyErr != nil {
					logger.Warn("failed to apply pool proxy to live client",
						slog.String("proxy_url", sanitizeURL(poolProxy.ProxyURL)),
						slog.String("error", applyErr.Error()))
				}
			}
		}

		// Populate proxy URL on the returned record.
		assignment.ProxyURL = poolProxy.ProxyURL

		// 8. Record metrics.
		if m.metrics != nil {
			m.metrics.RecordAssignment(poolProxy.ProviderID.String(), "success")
		}

		logger.Info("proxy assigned from pool",
			slog.String("assignment_id", assignment.ID.String()),
			slog.String("pool_proxy_id", poolProxy.ID.String()),
			slog.String("proxy_url", sanitizeURL(poolProxy.ProxyURL)))

		m.updatePoolSizeMetrics(ctx)

		return assignment, nil
	}

	if m.metrics != nil {
		m.metrics.RecordAssignment("unknown", "exhausted_retries")
	}
	return nil, fmt.Errorf("proxy assignment failed after %d attempts", m.cfg.MaxAssignmentRetries)
}

// ReleaseProxy releases the active pool proxy assignment for an instance.
func (m *PoolManager) ReleaseProxy(ctx context.Context, instanceID uuid.UUID, reason string) error {
	logger := logging.ContextLogger(ctx, m.log).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("reason", reason))

	// 1. Get active assignment.
	assignment, err := m.repo.GetActiveAssignment(ctx, instanceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("no active assignment to release")
			return nil
		}
		return fmt.Errorf("get active assignment: %w", err)
	}

	// 2. Deactivate assignment.
	if err := m.repo.DeactivateAssignment(ctx, instanceID, reason); err != nil {
		return fmt.Errorf("deactivate assignment: %w", err)
	}

	// 3. Decrement assigned count on pool proxy.
	if err := m.repo.DecrementAssignedCount(ctx, assignment.PoolProxyID); err != nil {
		logger.Warn("decrement assigned count failed",
			slog.String("pool_proxy_id", assignment.PoolProxyID.String()),
			slog.String("error", err.Error()))
	}

	// 4. Clear proxy from instances table.
	if m.instanceUpdater != nil {
		if clearErr := m.instanceUpdater.ClearProxyConfig(ctx, instanceID); clearErr != nil {
			logger.Warn("failed to clear pool proxy from instance config",
				slog.String("error", clearErr.Error()))
		}
	}

	// 5. Clear proxy from live client.
	if m.registry != nil {
		if applyErr := m.registry.ApplyProxy(ctx, instanceID, "", false, false, false); applyErr != nil {
			logger.Warn("failed to clear proxy from live client",
				slog.String("error", applyErr.Error()))
		}
	}

	// 6. Record metrics.
	if m.metrics != nil {
		m.metrics.RecordRelease(reason)
	}

	logger.Info("proxy released",
		slog.String("assignment_id", assignment.ID.String()),
		slog.String("pool_proxy_id", assignment.PoolProxyID.String()))

	m.updatePoolSizeMetrics(ctx)

	return nil
}

// updatePoolSizeMetrics queries pool stats and updates the Prometheus pool size gauge.
func (m *PoolManager) updatePoolSizeMetrics(ctx context.Context) {
	if m.metrics == nil {
		return
	}
	stats, err := m.repo.GetPoolStats(ctx)
	if err != nil {
		return
	}
	for _, ps := range stats.ByProvider {
		pid := ps.ProviderID.String()
		m.metrics.SetPoolSize(pid, "available", float64(ps.Available))
		m.metrics.SetPoolSize(pid, "assigned", float64(ps.Assigned))
		m.metrics.SetPoolSize(pid, "unhealthy", float64(ps.Unhealthy))
	}
}

// SwapProxy releases the current proxy and assigns a new one for the instance.
func (m *PoolManager) SwapProxy(ctx context.Context, instanceID uuid.UUID, opts AssignOptions) (*AssignmentRecord, error) {
	logger := logging.ContextLogger(ctx, m.log).With(
		slog.String("instance_id", instanceID.String()))

	// 1. Release current proxy.
	if err := m.ReleaseProxy(ctx, instanceID, "swap"); err != nil {
		return nil, fmt.Errorf("release for swap: %w", err)
	}

	// 2. Assign new proxy.
	assignment, err := m.AssignProxy(ctx, instanceID, opts)
	if err != nil {
		return nil, fmt.Errorf("assign after swap: %w", err)
	}

	// 3. Hot-swap on the live client if registry supports it.
	if m.registry != nil && assignment != nil {
		poolProxy, lookupErr := m.repo.GetPoolProxyByID(ctx, assignment.PoolProxyID)
		if lookupErr == nil && poolProxy != nil {
			if swapErr := m.registry.SwapProxy(ctx, instanceID, poolProxy.ProxyURL, opts.NoWebsocket, opts.OnlyLogin, opts.NoMedia); swapErr != nil {
				logger.Warn("hot-swap failed, proxy assigned but client not reconnected",
					slog.String("error", swapErr.Error()))
			}
		} else {
			logger.Warn("could not look up pool proxy for hot-swap",
				slog.String("pool_proxy_id", assignment.PoolProxyID.String()))
		}
	}

	if m.metrics != nil {
		providerID := "unknown"
		if assignment != nil {
			providerID = assignment.PoolProxyID.String()
		}
		m.metrics.RecordSwap(providerID, "success")
	}

	logger.Info("proxy swapped")

	return assignment, nil
}

// getPoolProxyByID retrieves a pool proxy by its ID from the database.
func (m *PoolManager) getPoolProxyByID(ctx context.Context, id uuid.UUID) (*PoolProxyRecord, error) {
	return m.repo.GetPoolProxyByID(ctx, id)
}

// SwapGroup swaps the proxy for all instances in a proxy group.
func (m *PoolManager) SwapGroup(ctx context.Context, groupID uuid.UUID) error {
	logger := logging.ContextLogger(ctx, m.log).With(
		slog.String("group_id", groupID.String()))

	// 1. Get the group.
	group, err := m.repo.GetGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	// 2. Find a new proxy for the group.
	var countryCodes []string
	if group.CountryCode != nil {
		countryCodes = []string{*group.CountryCode}
	}

	newProxy, err := m.repo.FindAvailableProxy(ctx, countryCodes, group.ProviderID)
	if err != nil {
		return fmt.Errorf("find available proxy for group: %w", err)
	}

	// 3. Get all active assignments with the group's current pool_proxy_id.
	var assignments []AssignmentRecord
	if group.PoolProxyID != nil {
		assignments, err = m.repo.GetAssignmentsByPoolProxy(ctx, *group.PoolProxyID)
		if err != nil {
			return fmt.Errorf("get group assignments: %w", err)
		}
	}

	// 4. For each active assignment: release old, assign the new proxy.
	var swapErrors int
	for _, a := range assignments {
		if a.Status != AssignmentStatusActive {
			continue
		}

		// Release old assignment.
		if releaseErr := m.ReleaseProxy(ctx, a.InstanceID, "group_swap"); releaseErr != nil {
			logger.Warn("release failed during group swap",
				slog.String("instance_id", a.InstanceID.String()),
				slog.String("error", releaseErr.Error()))
			swapErrors++
			continue
		}

		// Increment count on the new proxy.
		if incErr := m.repo.IncrementAssignedCount(ctx, newProxy.ID); incErr != nil {
			logger.Warn("increment assigned count failed during group swap",
				slog.String("instance_id", a.InstanceID.String()),
				slog.String("error", incErr.Error()))
			swapErrors++
			continue
		}

		// Create new assignment pointing to the same new proxy.
		_, createErr := m.repo.CreateAssignment(ctx, newProxy.ID, a.InstanceID, &groupID, "group_swap")
		if createErr != nil {
			if decErr := m.repo.DecrementAssignedCount(ctx, newProxy.ID); decErr != nil {
				logger.Error("rollback decrement during group swap",
					slog.String("error", decErr.Error()))
			}
			logger.Warn("create assignment failed during group swap",
				slog.String("instance_id", a.InstanceID.String()),
				slog.String("error", createErr.Error()))
			swapErrors++
			continue
		}

		// Hot-swap on the live client.
		if m.registry != nil {
			if swapErr := m.registry.SwapProxy(ctx, a.InstanceID, newProxy.ProxyURL, false, false, false); swapErr != nil {
				logger.Warn("hot-swap failed during group swap",
					slog.String("instance_id", a.InstanceID.String()),
					slog.String("error", swapErr.Error()))
			}
		}
	}

	// 5. Update group's pool_proxy_id to the new proxy.
	if err := m.repo.UpdateGroupProxy(ctx, groupID, &newProxy.ID); err != nil {
		return fmt.Errorf("update group proxy: %w", err)
	}

	logger.Info("group swap completed",
		slog.Int("instances_swapped", len(assignments)-swapErrors),
		slog.Int("errors", swapErrors),
		slog.String("new_proxy_id", newProxy.ID.String()))

	if swapErrors > 0 {
		return fmt.Errorf("group swap completed with %d errors out of %d instances", swapErrors, len(assignments))
	}
	return nil
}

// GetProviderForPoolProxy retrieves the provider ID and ProxyProvider for a given pool proxy.
func (m *PoolManager) GetProviderForPoolProxy(ctx context.Context, poolProxyID uuid.UUID) (uuid.UUID, ProxyProvider, error) {
	poolProxy, err := m.getPoolProxyByID(ctx, poolProxyID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("get pool proxy: %w", err)
	}

	provider, ok := m.getProvider(poolProxy.ProviderID)
	if !ok {
		return poolProxy.ProviderID, nil, fmt.Errorf("provider %s not registered", poolProxy.ProviderID)
	}

	return poolProxy.ProviderID, provider, nil
}

// BulkAssignProxies assigns pool proxies to all unassigned instances. Uses a Redis
// distributed lock to prevent concurrent bulk-assign operations across multiple workers.
// The lock is refreshed every 50 assignments to keep the TTL alive for large batches.
func (m *PoolManager) BulkAssignProxies(ctx context.Context, req BulkAssignRequest) (*BulkAssignResult, error) {
	logger := logging.ContextLogger(ctx, m.log)

	// Acquire Redis distributed lock (5 min TTL, refreshed during iteration).
	const lockKey = "pool:bulk_assign"
	const lockTTL = 300 // 5 minutes

	lock, acquired, err := m.lockManager.Acquire(ctx, lockKey, lockTTL)
	if err != nil {
		return nil, fmt.Errorf("acquire bulk assign lock: %w", err)
	}
	if !acquired {
		return nil, fmt.Errorf("bulk assign already in progress by another worker")
	}
	defer func() {
		if relErr := lock.Release(ctx); relErr != nil {
			logger.Warn("failed to release bulk assign lock", slog.String("error", relErr.Error()))
		}
	}()

	// Find all unassigned instances.
	instanceIDs, err := m.repo.ListUnassignedInstanceIDs(ctx, req.InstanceIDs)
	if err != nil {
		return nil, fmt.Errorf("list unassigned instances: %w", err)
	}

	result := &BulkAssignResult{
		Total: len(instanceIDs),
	}

	if result.Total == 0 {
		logger.Info("bulk assign: no unassigned instances found")
		return result, nil
	}

	logger.Info("bulk assign started",
		slog.Int("total_instances", result.Total))

	opts := AssignOptions{
		ProviderID:   req.ProviderID,
		CountryCodes: req.CountryCodes,
	}
	if len(opts.CountryCodes) == 0 {
		opts.CountryCodes = m.cfg.DefaultCountryCodes
	}

	const refreshEvery = 50

	for i, instanceID := range instanceIDs {
		// Refresh Redis lock periodically to prevent expiry on large batches.
		if i > 0 && i%refreshEvery == 0 {
			if refreshErr := lock.Refresh(ctx, lockTTL); refreshErr != nil {
				logger.Warn("failed to refresh bulk assign lock, continuing",
					slog.Int("progress", i),
					slog.String("error", refreshErr.Error()))
			}
		}

		_, assignErr := m.AssignProxy(ctx, instanceID, opts)
		if assignErr != nil {
			if errors.Is(assignErr, pgx.ErrNoRows) || errors.Is(assignErr, ErrProxyCapacityFull) {
				result.Skipped++
			} else {
				result.Failed++
				if len(result.Errors) < 20 { // Cap error messages to avoid huge responses
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", instanceID, assignErr.Error()))
				}
			}
			continue
		}
		result.Assigned++
	}

	m.updatePoolSizeMetrics(ctx)

	logger.Info("bulk assign completed",
		slog.Int("total", result.Total),
		slog.Int("assigned", result.Assigned),
		slog.Int("skipped", result.Skipped),
		slog.Int("failed", result.Failed))

	return result, nil
}
