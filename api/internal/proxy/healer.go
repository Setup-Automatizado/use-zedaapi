package proxy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
)

// AutoHealer automatically recovers from proxy failures for pool-managed instances.
// It hooks into the HealthChecker's UnhealthyCallback and RecoveredCallback.
// Uses Redis distributed locks to prevent concurrent healing of the same instance.
type AutoHealer struct {
	poolManager *PoolManager
	poolRepo    *PoolRepository
	lockManager locks.Manager
	log         *slog.Logger
	metrics     *PoolMetrics
	queuePauser InstanceQueuePauser
}

// NewAutoHealer creates a new AutoHealer.
func NewAutoHealer(poolManager *PoolManager, poolRepo *PoolRepository, lockMgr locks.Manager, log *slog.Logger, metrics *PoolMetrics) *AutoHealer {
	return &AutoHealer{
		poolManager: poolManager,
		poolRepo:    poolRepo,
		lockManager: lockMgr,
		log:         log.With(slog.String("component", "auto_healer")),
		metrics:     metrics,
	}
}

// SetQueuePauser sets the optional queue pauser for resuming message processing after healing.
func (h *AutoHealer) SetQueuePauser(pauser InstanceQueuePauser) { h.queuePauser = pauser }

// OnProxyUnhealthy is the callback for HealthChecker.SetUnhealthyCallback.
// It triggers auto-healing for pool-managed proxies; manual proxy instances are skipped.
func (h *AutoHealer) OnProxyUnhealthy(ctx context.Context, instanceID uuid.UUID, proxyURL string, failures int) {
	logger := logging.ContextLogger(ctx, h.log).With(
		slog.String("instance_id", instanceID.String()),
		slog.Int("failures", failures),
	)

	// 1. Check if this instance has a pool assignment (manual proxy -> clear & fallback to direct).
	assignment, err := h.poolRepo.GetActiveAssignment(ctx, instanceID)
	if err != nil || assignment == nil {
		// Manual proxy instance: clear the broken proxy so instance runs direct
		logger.Warn("manual proxy unhealthy, clearing proxy config for direct connection")

		if registry := h.poolManager.GetRegistry(); registry != nil {
			if applyErr := registry.ApplyProxy(ctx, instanceID, "", false, false, false); applyErr != nil {
				logger.Error("failed to clear manual proxy from live client",
					slog.String("error", applyErr.Error()))
			}
		}
		if updater := h.poolManager.GetInstanceUpdater(); updater != nil {
			if clearErr := updater.ClearProxyConfig(ctx, instanceID); clearErr != nil {
				logger.Error("failed to clear manual proxy from database",
					slog.String("error", clearErr.Error()))
			}
		}

		// Resume queue so messages flow through direct connection
		if h.queuePauser != nil {
			_ = h.queuePauser.ResumeInstance(ctx, instanceID)
		}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "auto_healer")
			scope.SetTag("instance_id", instanceID.String())
			sentry.CaptureMessage(fmt.Sprintf("Manual proxy cleared for instance %s - using direct connection", instanceID))
		})
		return
	}

	// 2. Prevent concurrent healing of same instance via Redis distributed lock.
	lockKey := fmt.Sprintf("pool:heal:%s", instanceID)
	lock, acquired, lockErr := h.lockManager.Acquire(ctx, lockKey, 120) // 2 min TTL
	if lockErr != nil {
		logger.Warn("failed to acquire healing lock", slog.String("error", lockErr.Error()))
		return
	}
	if !acquired {
		logger.Debug("instance already being healed by another worker, skipping")
		return
	}
	defer func() {
		if relErr := lock.Release(ctx); relErr != nil {
			logger.Warn("failed to release healing lock", slog.String("error", relErr.Error()))
		}
	}()

	start := time.Now()
	logger.Info("starting auto-heal for unhealthy pool proxy",
		slog.String("pool_proxy_id", assignment.PoolProxyID.String()))

	// 3. Check if this is a group assignment.
	if assignment.GroupID != nil {
		if groupErr := h.healGroup(ctx, *assignment.GroupID); groupErr != nil {
			logger.Error("group healing failed", slog.String("error", groupErr.Error()))
			h.recordHealingResult(instanceID, "failure", time.Since(start))
			return
		}
		h.recordHealingResult(instanceID, "success", time.Since(start))
		logger.Info("group auto-heal completed successfully", slog.Duration("duration", time.Since(start)))
		return
	}

	// 4. Healing cascade for individual assignment.
	if healErr := h.healInstance(ctx, instanceID, assignment); healErr != nil {
		logger.Error("instance healing failed", slog.String("error", healErr.Error()))
		h.recordHealingResult(instanceID, "failure", time.Since(start))

		// Last resort: release proxy so instance runs without proxy.
		logger.Warn("releasing proxy as last resort - instance will run without proxy")
		if releaseErr := h.poolManager.ReleaseProxy(ctx, instanceID, "heal_exhausted"); releaseErr != nil {
			logger.Error("failed to release proxy during healing", slog.String("error", releaseErr.Error()))
		}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "auto_healer")
			scope.SetTag("instance_id", instanceID.String())
			sentry.CaptureMessage(fmt.Sprintf("All proxy healing attempts exhausted for instance %s", instanceID))
		})

		// Resume queue - instance will run without proxy
		if h.queuePauser != nil {
			if resumeErr := h.queuePauser.ResumeInstance(ctx, instanceID); resumeErr != nil {
				logger.Warn("failed to resume queue after healing exhausted",
					slog.String("error", resumeErr.Error()))
			}
		}
		return
	}

	h.recordHealingResult(instanceID, "success", time.Since(start))

	// Resume queue after successful healing
	if h.queuePauser != nil {
		if resumeErr := h.queuePauser.ResumeInstance(ctx, instanceID); resumeErr != nil {
			logger.Warn("failed to resume queue after healing",
				slog.String("error", resumeErr.Error()))
		}
	}

	logger.Info("auto-heal completed successfully", slog.Duration("duration", time.Since(start)))
}

// healInstance attempts a three-step healing cascade for an individual instance:
//
//	Step A: Ask the provider to replace the proxy (if supported).
//	Step B: Swap to a different proxy from the same provider.
//	Step C: Swap to a proxy from any other provider.
func (h *AutoHealer) healInstance(ctx context.Context, instanceID uuid.UUID, assignment *AssignmentRecord) error {
	logger := logging.ContextLogger(ctx, h.log).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("pool_proxy_id", assignment.PoolProxyID.String()),
	)

	// Resolve provider for the current pool proxy.
	providerID, provider, provErr := h.poolManager.GetProviderForPoolProxy(ctx, assignment.PoolProxyID)

	// Step A: Try provider.ReplaceProxy if the provider is available and supports it.
	if provErr == nil && provider != nil {
		if replaceErr := h.tryProviderReplace(ctx, instanceID, assignment.PoolProxyID, providerID, provider); replaceErr == nil {
			logger.Info("healed via provider replace")
			return nil
		} else if !errors.Is(replaceErr, ErrReplaceNotSupported) {
			logger.Debug("provider replace failed, trying swap",
				slog.String("error", replaceErr.Error()))
		}
	}

	// Step B: Swap to different proxy from same provider.
	if provErr == nil {
		opts := AssignOptions{ProviderID: &providerID}
		if _, swapErr := h.poolManager.SwapProxy(ctx, instanceID, opts); swapErr == nil {
			logger.Info("healed via same-provider swap",
				slog.String("provider_id", providerID.String()))
			return nil
		} else {
			logger.Debug("same-provider swap failed",
				slog.String("provider_id", providerID.String()),
				slog.String("error", swapErr.Error()))
		}
	}

	// Step C: Try other providers by priority.
	providers, listErr := h.poolRepo.ListProviders(ctx)
	if listErr != nil {
		return fmt.Errorf("list providers for healing: %w", listErr)
	}
	for _, p := range providers {
		if p.ID == providerID || !p.Enabled {
			continue
		}
		otherID := p.ID
		otherOpts := AssignOptions{ProviderID: &otherID}
		if _, swapErr := h.poolManager.SwapProxy(ctx, instanceID, otherOpts); swapErr == nil {
			logger.Info("healed via other-provider swap",
				slog.String("provider_id", otherID.String()))
			return nil
		}
	}

	return fmt.Errorf("all healing attempts exhausted for instance %s", instanceID)
}

// tryProviderReplace asks the provider to replace the failed proxy and, on success,
// upserts the new proxy and swaps the instance to it.
func (h *AutoHealer) tryProviderReplace(ctx context.Context, instanceID, poolProxyID, providerID uuid.UUID, provider ProxyProvider) error {
	// Look up the pool proxy record to get the external ID.
	poolProxy, err := h.poolManager.getPoolProxyByID(ctx, poolProxyID)
	if err != nil || poolProxy == nil || poolProxy.ExternalID == nil {
		return fmt.Errorf("cannot resolve external ID for pool proxy %s: %w", poolProxyID, err)
	}

	result, replaceErr := provider.ReplaceProxy(ctx, *poolProxy.ExternalID)
	if replaceErr != nil {
		return replaceErr
	}
	if result == nil || !result.Success || result.NewProxy == nil {
		return fmt.Errorf("provider replace unsuccessful for proxy %s", poolProxyID)
	}

	// Upsert the replacement proxy into the pool.
	if _, upsertErr := h.poolRepo.UpsertPoolProxy(ctx, providerID, *result.NewProxy, poolProxy.MaxAssignments); upsertErr != nil {
		return fmt.Errorf("upsert replacement proxy: %w", upsertErr)
	}

	// Swap the instance to pick up the newly upserted proxy (prefer same provider).
	opts := AssignOptions{ProviderID: &providerID}
	if _, swapErr := h.poolManager.SwapProxy(ctx, instanceID, opts); swapErr != nil {
		return fmt.Errorf("swap to replacement proxy: %w", swapErr)
	}

	return nil
}

// healGroup delegates to PoolManager.SwapGroup to swap all instances in the group.
func (h *AutoHealer) healGroup(ctx context.Context, groupID uuid.UUID) error {
	return h.poolManager.SwapGroup(ctx, groupID)
}

// OnProxyRecovered is the callback for HealthChecker.SetRecoveredCallback.
// It updates the pool proxy health status when a pool-managed proxy recovers.
func (h *AutoHealer) OnProxyRecovered(ctx context.Context, instanceID uuid.UUID, proxyURL string) {
	logger := logging.ContextLogger(ctx, h.log).With(
		slog.String("instance_id", instanceID.String()),
	)

	assignment, err := h.poolRepo.GetActiveAssignment(ctx, instanceID)
	if err != nil || assignment == nil {
		// Not a pool proxy, nothing to update.
		return
	}

	if updateErr := h.poolRepo.UpdatePoolProxyHealth(ctx, assignment.PoolProxyID, "healthy", 0); updateErr != nil {
		if !errors.Is(updateErr, pgx.ErrNoRows) {
			logger.Error("failed to update pool proxy health on recovery",
				slog.String("pool_proxy_id", assignment.PoolProxyID.String()),
				slog.String("error", updateErr.Error()))
		}
	} else {
		logger.Info("pool proxy health restored",
			slog.String("pool_proxy_id", assignment.PoolProxyID.String()))
	}
}

// recordHealingResult records healing metrics if metrics are available.
func (h *AutoHealer) recordHealingResult(instanceID uuid.UUID, status string, duration time.Duration) {
	if h.metrics != nil {
		h.metrics.RecordHealing(instanceID.String(), status, duration.Seconds())
	}
}
