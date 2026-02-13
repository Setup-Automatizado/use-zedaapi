package whatsmeow

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"

	wameow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/logging"
)

// ProxyRepository provides proxy configuration for instances.
type ProxyRepository interface {
	GetProxyConfig(ctx context.Context, id uuid.UUID) (*ProxyConfig, error)
	ListInstancesWithProxy(ctx context.Context) ([]ProxyInstance, error)
	UpdateProxyHealthStatus(ctx context.Context, id uuid.UUID, status string, failures int) error
}

// ProxyConfig mirrors instances.ProxyConfig for the registry layer.
type ProxyConfig struct {
	ProxyURL       *string
	Enabled        bool
	NoWebsocket    bool
	OnlyLogin      bool
	NoMedia        bool
	HealthStatus   string
	HealthFailures int
}

// ProxyInstance is a lightweight proxy-enabled instance reference.
type ProxyInstance struct {
	InstanceID     uuid.UUID
	ProxyURL       string
	HealthStatus   string
	HealthFailures int
}

// SetProxyRepository sets the proxy repository for looking up per-instance proxy configs.
func (r *ClientRegistry) SetProxyRepository(repo ProxyRepository) {
	r.proxyRepo = repo
}

// SetQueuePauser sets the optional queue pauser for pausing/resuming during proxy swaps.
func (r *ClientRegistry) SetQueuePauser(pauser QueuePauser) {
	r.queuePauser = pauser
}

// ApplyProxy sets the proxy on an existing client. The proxy takes effect on the next
// Connect() call. If the client is currently connected, a reconnect is needed.
func (r *ClientRegistry) ApplyProxy(ctx context.Context, instanceID uuid.UUID, proxyURL string, noWebsocket, onlyLogin, noMedia bool) error {
	r.mu.RLock()
	state, ok := r.clients[instanceID]
	r.mu.RUnlock()

	if !ok || state == nil || state.client == nil {
		return nil // No live client, proxy will be applied on next EnsureClient
	}

	logger := logging.ContextLogger(ctx, r.log).With(
		slog.String("component", "proxy_registry"),
		slog.String("instance_id", instanceID.String()),
	)

	opts := buildProxyOptions(noWebsocket, onlyLogin, noMedia)
	if err := state.client.SetProxyAddress(proxyURL, opts...); err != nil {
		return fmt.Errorf("set proxy address: %w", err)
	}

	// Track proxy on client state
	r.mu.Lock()
	state.proxyURL = proxyURL
	r.mu.Unlock()

	logger.Info("proxy applied to client (effective on next connect)",
		slog.String("proxy", sanitizeProxyForLog(proxyURL)))

	// If clearing proxy, ensure queue is resumed (instance runs direct)
	if proxyURL == "" && r.queuePauser != nil {
		if resumeErr := r.queuePauser.ResumeInstance(ctx, instanceID); resumeErr != nil {
			logger.Warn("failed to resume queue after proxy clear",
				slog.String("error", resumeErr.Error()))
		}
	}

	return nil
}

// SwapProxy hot-swaps the proxy on an active connection by:
// 1. Setting the new proxy on the client
// 2. Disconnecting the current WebSocket
// 3. Reconnecting through the new proxy
// The WhatsApp session (device store) is preserved.
func (r *ClientRegistry) SwapProxy(ctx context.Context, instanceID uuid.UUID, proxyURL string, noWebsocket, onlyLogin, noMedia bool) error {
	r.mu.RLock()
	state, ok := r.clients[instanceID]
	r.mu.RUnlock()

	if !ok || state == nil || state.client == nil {
		return fmt.Errorf("instance %s: no active client", instanceID)
	}

	client := state.client
	logger := logging.ContextLogger(ctx, r.log).With(
		slog.String("component", "proxy_registry"),
		slog.String("instance_id", instanceID.String()),
	)

	// Pause message queue for this instance during swap
	if r.queuePauser != nil {
		if pauseErr := r.queuePauser.PauseInstance(ctx, instanceID, "proxy_swap"); pauseErr != nil {
			logger.Warn("failed to pause queue for proxy swap",
				slog.String("error", pauseErr.Error()))
		}
		defer func() {
			if resumeErr := r.queuePauser.ResumeInstance(ctx, instanceID); resumeErr != nil {
				logger.Warn("failed to resume queue after proxy swap",
					slog.String("error", resumeErr.Error()))
			}
		}()
	}

	// 1. Apply new proxy to client transport
	opts := buildProxyOptions(noWebsocket, onlyLogin, noMedia)
	if err := client.SetProxyAddress(proxyURL, opts...); err != nil {
		return fmt.Errorf("set proxy address for swap: %w", err)
	}

	// Track proxy URL
	r.mu.Lock()
	state.proxyURL = proxyURL
	r.mu.Unlock()

	wasConnected := client.IsConnected()
	if !wasConnected {
		logger.Info("proxy set on disconnected client, will take effect on next connect")
		return nil
	}

	// 2. Disconnect current WebSocket (session preserved in device store)
	logger.Info("disconnecting for proxy swap",
		slog.String("new_proxy", sanitizeProxyForLog(proxyURL)))

	client.Disconnect()

	// Small delay to allow clean disconnection
	time.Sleep(200 * time.Millisecond)

	// 3. Reconnect through new proxy
	if err := client.Connect(); err != nil {
		logger.Error("reconnect after proxy swap failed",
			slog.String("error", err.Error()))

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "proxy_registry")
			scope.SetTag("instance_id", instanceID.String())
			scope.SetContext("proxy_swap", map[string]interface{}{
				"new_proxy": sanitizeProxyForLog(proxyURL),
			})
			sentry.CaptureException(fmt.Errorf("reconnect after proxy swap: %w", err))
		})

		return fmt.Errorf("reconnect after proxy swap: %w", err)
	}

	logger.Info("proxy swap completed, client reconnected",
		slog.String("proxy", sanitizeProxyForLog(proxyURL)))

	return nil
}

// GetClientProxy returns the current proxy URL for an instance, if any.
func (r *ClientRegistry) GetClientProxy(instanceID uuid.UUID) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.clients[instanceID]
	if !ok || state == nil {
		return "", false
	}
	return state.proxyURL, state.proxyURL != ""
}

// applyProxyFromConfig loads proxy config from the repository and applies it to a whatsmeow client.
// Called during instantiateClientWithLock before the client state is stored.
func (r *ClientRegistry) applyProxyFromConfig(ctx context.Context, instanceID uuid.UUID, client *wameow.Client) string {
	if r.proxyRepo == nil {
		return ""
	}

	cfg, err := r.proxyRepo.GetProxyConfig(ctx, instanceID)
	if err != nil {
		r.log.Debug("no proxy config for instance",
			slog.String("component", "proxy_registry"),
			slog.String("instance_id", instanceID.String()),
			slog.String("error", err.Error()))
		return ""
	}

	if !cfg.Enabled || cfg.ProxyURL == nil || *cfg.ProxyURL == "" {
		return ""
	}

	// Health check gate: don't apply unhealthy proxies
	if cfg.HealthStatus == "unhealthy" {
		r.log.Warn("skipping unhealthy proxy for instance",
			slog.String("component", "proxy_registry"),
			slog.String("instance_id", instanceID.String()),
			slog.String("proxy", sanitizeProxyForLog(*cfg.ProxyURL)),
			slog.Int("failures", cfg.HealthFailures))
		return ""
	}

	opts := buildProxyOptions(cfg.NoWebsocket, cfg.OnlyLogin, cfg.NoMedia)
	if proxyErr := client.SetProxyAddress(*cfg.ProxyURL, opts...); proxyErr != nil {
		r.log.Warn("failed to set proxy from config",
			slog.String("component", "proxy_registry"),
			slog.String("instance_id", instanceID.String()),
			slog.String("proxy", sanitizeProxyForLog(*cfg.ProxyURL)),
			slog.String("error", proxyErr.Error()))
		return ""
	}

	r.log.Info("proxy applied from config",
		slog.String("component", "proxy_registry"),
		slog.String("instance_id", instanceID.String()),
		slog.String("proxy", sanitizeProxyForLog(*cfg.ProxyURL)))

	return *cfg.ProxyURL
}

// buildProxyOptions creates SetProxyOptions from boolean flags.
func buildProxyOptions(noWebsocket, onlyLogin, noMedia bool) []wameow.SetProxyOptions {
	if !noWebsocket && !onlyLogin && !noMedia {
		return nil
	}
	return []wameow.SetProxyOptions{{
		NoWebsocket: noWebsocket,
		OnlyLogin:   onlyLogin,
		NoMedia:     noMedia,
	}}
}

// resolveProxyURL returns the proxy URL from config if available and healthy.
func (r *ClientRegistry) resolveProxyURL(ctx context.Context, instanceID uuid.UUID) string {
	if r.proxyRepo == nil {
		return ""
	}
	cfg, err := r.proxyRepo.GetProxyConfig(ctx, instanceID)
	if err != nil || !cfg.Enabled || cfg.ProxyURL == nil || *cfg.ProxyURL == "" {
		return ""
	}
	if cfg.HealthStatus == "unhealthy" {
		return ""
	}
	return *cfg.ProxyURL
}

// sanitizeProxyForLog removes credentials from proxy URL for safe logging.
func sanitizeProxyForLog(rawURL string) string {
	if rawURL == "" {
		return "<none>"
	}
	// Simple sanitization: mask everything between :// and @
	for i := 0; i < len(rawURL); i++ {
		if i+3 < len(rawURL) && rawURL[i:i+3] == "://" {
			atIdx := -1
			for j := i + 3; j < len(rawURL); j++ {
				if rawURL[j] == '@' {
					atIdx = j
					break
				}
			}
			if atIdx != -1 {
				return rawURL[:i+3] + "***:***" + rawURL[atIdx:]
			}
		}
	}
	return rawURL
}
