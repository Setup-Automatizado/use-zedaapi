package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"golang.org/x/net/proxy"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// Config holds proxy health checker configuration.
type Config struct {
	HealthCheckInterval    time.Duration // Default: 60s
	HealthCheckTimeout     time.Duration // Default: 10s
	MaxConsecutiveFailures int           // Default: 3
	DLQOnUnhealthy         bool          // Default: true
	PauseQueueOnUnhealthy  bool          // Default: true
	LogRetentionPeriod     time.Duration // Default: 7 days
	CleanupInterval        time.Duration // Default: 1h
}

// WithDefaults returns a copy of Config with zero values replaced by defaults.
func (c Config) WithDefaults() Config {
	if c.HealthCheckInterval <= 0 {
		c.HealthCheckInterval = 60 * time.Second
	}
	if c.HealthCheckTimeout <= 0 {
		c.HealthCheckTimeout = 10 * time.Second
	}
	if c.MaxConsecutiveFailures <= 0 {
		c.MaxConsecutiveFailures = 3
	}
	if c.LogRetentionPeriod <= 0 {
		c.LogRetentionPeriod = 7 * 24 * time.Hour
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = time.Hour
	}
	return c
}

// Repository interface for proxy health operations.
type Repository interface {
	ListInstancesWithProxy(ctx context.Context) ([]ProxyInstance, error)
	UpdateProxyHealthStatus(ctx context.Context, id uuid.UUID, status string, failures int) error
	InsertProxyHealthLog(ctx context.Context, log HealthLog) error
	CleanupProxyHealthLogs(ctx context.Context, retention time.Duration) (int64, error)
}

// ProxyInstance represents an instance with proxy configuration.
type ProxyInstance struct {
	InstanceID     uuid.UUID
	ProxyURL       string
	HealthStatus   string
	HealthFailures int
}

// HealthLog represents a single proxy health check result.
type HealthLog struct {
	InstanceID   uuid.UUID
	ProxyURL     string
	Status       string
	LatencyMs    *int
	ErrorMessage *string
	CheckedAt    time.Time
}

// UnhealthyCallback is called when a proxy transitions to unhealthy state.
type UnhealthyCallback func(ctx context.Context, instanceID uuid.UUID, proxyURL string, failures int)

// RecoveredCallback is called when a proxy recovers from unhealthy state.
type RecoveredCallback func(ctx context.Context, instanceID uuid.UUID, proxyURL string)

// HealthChecker monitors proxy health for all instances with configured proxies.
type HealthChecker struct {
	repo    Repository
	cfg     Config
	log     *slog.Logger
	metrics *Metrics

	onUnhealthy UnhealthyCallback
	onRecovered RecoveredCallback
	queuePauser InstanceQueuePauser

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewHealthChecker creates a new proxy health checker.
func NewHealthChecker(repo Repository, cfg Config, log *slog.Logger, metrics *Metrics) *HealthChecker {
	return &HealthChecker{
		repo:    repo,
		cfg:     cfg.WithDefaults(),
		log:     log.With(slog.String("component", "proxy_health_checker")),
		metrics: metrics,
	}
}

// SetUnhealthyCallback sets the callback invoked when a proxy crosses the failure threshold.
func (h *HealthChecker) SetUnhealthyCallback(cb UnhealthyCallback) { h.onUnhealthy = cb }

// SetRecoveredCallback sets the callback invoked when a previously-unhealthy proxy recovers.
func (h *HealthChecker) SetRecoveredCallback(cb RecoveredCallback) { h.onRecovered = cb }

// SetQueuePauser sets the optional queue pauser for pausing/resuming message processing.
func (h *HealthChecker) SetQueuePauser(pauser InstanceQueuePauser) { h.queuePauser = pauser }

// Start begins the periodic health check loop.
func (h *HealthChecker) Start(ctx context.Context) {
	ctx, h.cancel = context.WithCancel(ctx)

	h.wg.Add(2)
	go h.checkLoop(ctx)
	go h.cleanupLoop(ctx)

	h.log.Info("proxy health checker started",
		slog.Duration("interval", h.cfg.HealthCheckInterval),
		slog.Int("max_failures", h.cfg.MaxConsecutiveFailures))
}

// Stop gracefully stops the health checker and waits for goroutines to finish.
func (h *HealthChecker) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
	h.wg.Wait()
	h.log.Info("proxy health checker stopped")
}

func (h *HealthChecker) checkLoop(ctx context.Context) {
	defer h.wg.Done()
	ticker := time.NewTicker(h.cfg.HealthCheckInterval)
	defer ticker.Stop()

	// Run immediately on start.
	h.checkAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkAll(ctx)
		}
	}
}

func (h *HealthChecker) cleanupLoop(ctx context.Context) {
	defer h.wg.Done()
	ticker := time.NewTicker(h.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleaned, err := h.repo.CleanupProxyHealthLogs(ctx, h.cfg.LogRetentionPeriod)
			if err != nil {
				logger := logging.ContextLogger(ctx, h.log)
				logger.Error("cleanup proxy health logs", slog.String("error", err.Error()))
			} else if cleaned > 0 {
				h.log.Info("cleaned up proxy health logs", slog.Int64("removed", cleaned))
			}
		}
	}
}

func (h *HealthChecker) checkAll(ctx context.Context) {
	logger := logging.ContextLogger(ctx, h.log)

	instances, err := h.repo.ListInstancesWithProxy(ctx)
	if err != nil {
		logger.Error("list instances with proxy", slog.String("error", err.Error()))
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "proxy_health_checker")
			sentry.CaptureException(fmt.Errorf("list instances with proxy: %w", err))
		})
		return
	}

	if len(instances) == 0 {
		return
	}

	logger.Debug("checking proxy health", slog.Int("instances", len(instances)))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Max 10 concurrent health checks

	for _, inst := range instances {
		wg.Add(1)
		sem <- struct{}{}
		go func(p ProxyInstance) {
			defer wg.Done()
			defer func() { <-sem }()
			h.checkOne(ctx, p)
		}(inst)
	}

	wg.Wait()
}

func (h *HealthChecker) checkOne(ctx context.Context, inst ProxyInstance) {
	checkCtx, cancel := context.WithTimeout(ctx, h.cfg.HealthCheckTimeout)
	defer cancel()

	start := time.Now()
	reachable, checkErr := testProxyReachability(checkCtx, inst.ProxyURL, h.cfg.HealthCheckTimeout)
	latency := int(time.Since(start).Milliseconds())

	logger := logging.ContextLogger(ctx, h.log).With(
		slog.String("instance_id", inst.InstanceID.String()),
		slog.String("proxy", sanitizeURL(inst.ProxyURL)),
		slog.Int("latency_ms", latency),
	)

	status := "healthy"
	var errMsg *string
	newFailures := 0

	if !reachable {
		status = "unhealthy"
		if checkErr != nil {
			msg := checkErr.Error()
			errMsg = &msg
		}
		newFailures = inst.HealthFailures + 1
		logger.Warn("proxy health check failed",
			slog.Int("consecutive_failures", newFailures),
			slog.String("error", stringFromPtr(errMsg)))
	} else {
		if inst.HealthStatus == "unhealthy" {
			logger.Info("proxy recovered", slog.Int("previous_failures", inst.HealthFailures))
			if h.onRecovered != nil {
				h.onRecovered(ctx, inst.InstanceID, inst.ProxyURL)
			}
			// Resume message queue when proxy recovers
			if h.queuePauser != nil {
				if resumeErr := h.queuePauser.ResumeInstance(ctx, inst.InstanceID); resumeErr != nil {
					logger.Warn("failed to resume queue after proxy recovery",
						slog.String("error", resumeErr.Error()))
				}
			}
		}
		newFailures = 0
	}

	// Record health check metrics.
	if h.metrics != nil {
		h.metrics.RecordHealthCheck(inst.InstanceID.String(), status, float64(latency)/1000.0)
	}

	// Log to database.
	logEntry := HealthLog{
		InstanceID:   inst.InstanceID,
		ProxyURL:     inst.ProxyURL,
		Status:       status,
		LatencyMs:    &latency,
		ErrorMessage: errMsg,
		CheckedAt:    time.Now().UTC(),
	}
	if logErr := h.repo.InsertProxyHealthLog(ctx, logEntry); logErr != nil {
		logger.Error("insert health log", slog.String("error", logErr.Error()))
	}

	// Determine final status based on failure threshold.
	finalStatus := "healthy"
	if newFailures >= h.cfg.MaxConsecutiveFailures {
		finalStatus = "unhealthy"
	} else if newFailures > 0 {
		finalStatus = "unknown" // degraded but not yet unhealthy
	}

	// Update database.
	if updateErr := h.repo.UpdateProxyHealthStatus(ctx, inst.InstanceID, finalStatus, newFailures); updateErr != nil {
		logger.Error("update proxy health status", slog.String("error", updateErr.Error()))
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "proxy_health_checker")
			scope.SetTag("instance_id", inst.InstanceID.String())
			sentry.CaptureException(fmt.Errorf("update proxy health status: %w", updateErr))
		})
		return
	}

	// Trigger unhealthy callback when threshold crossed.
	if finalStatus == "unhealthy" && inst.HealthStatus != "unhealthy" {
		logger.Error("proxy marked unhealthy - threshold exceeded",
			slog.Int("failures", newFailures),
			slog.Int("threshold", h.cfg.MaxConsecutiveFailures))

		// Pause message queue for this instance if configured
		if h.cfg.PauseQueueOnUnhealthy && h.queuePauser != nil {
			if pauseErr := h.queuePauser.PauseInstance(ctx, inst.InstanceID, "proxy_unhealthy"); pauseErr != nil {
				logger.Warn("failed to pause queue for unhealthy proxy",
					slog.String("error", pauseErr.Error()))
			}
		}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "proxy_health_checker")
			scope.SetTag("instance_id", inst.InstanceID.String())
			scope.SetContext("proxy", map[string]any{
				"proxy_url":     sanitizeURL(inst.ProxyURL),
				"failures":      newFailures,
				"threshold":     h.cfg.MaxConsecutiveFailures,
				"latency_ms":    latency,
				"error_message": stringFromPtr(errMsg),
			})
			sentry.CaptureMessage(fmt.Sprintf("Proxy marked unhealthy for instance %s", inst.InstanceID.String()))
		})

		if h.onUnhealthy != nil {
			h.onUnhealthy(ctx, inst.InstanceID, inst.ProxyURL, newFailures)
		}
	}
}

// TestProxy performs an on-demand health check for a specific proxy URL.
func (h *HealthChecker) TestProxy(ctx context.Context, proxyURL string) (bool, int, error) {
	start := time.Now()
	reachable, err := testProxyReachability(ctx, proxyURL, h.cfg.HealthCheckTimeout)
	latency := int(time.Since(start).Milliseconds())
	return reachable, latency, err
}

// testProxyReachability tests if a proxy can reach WhatsApp servers.
func testProxyReachability(ctx context.Context, proxyURL string, timeout time.Duration) (bool, error) {
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return false, fmt.Errorf("invalid proxy URL: %w", err)
	}

	target := "web.whatsapp.com:443"

	switch parsed.Scheme {
	case "socks5":
		dialer, dialErr := proxy.FromURL(parsed, &net.Dialer{Timeout: timeout})
		if dialErr != nil {
			return false, fmt.Errorf("socks5 setup: %w", dialErr)
		}
		ctxDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			return false, fmt.Errorf("socks5 proxy does not support context dialing")
		}
		conn, connErr := ctxDialer.DialContext(ctx, "tcp", target)
		if connErr != nil {
			return false, connErr
		}
		conn.Close()
		return true, nil

	case "http", "https":
		dialer := &net.Dialer{Timeout: timeout}
		proxyConn, connErr := dialer.DialContext(ctx, "tcp", parsed.Host)
		if connErr != nil {
			return false, fmt.Errorf("connect to proxy: %w", connErr)
		}
		defer proxyConn.Close()

		connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target)
		if _, writeErr := proxyConn.Write([]byte(connectReq)); writeErr != nil {
			return false, fmt.Errorf("send CONNECT: %w", writeErr)
		}

		buf := make([]byte, 1024)
		n, readErr := proxyConn.Read(buf)
		if readErr != nil {
			return false, fmt.Errorf("read CONNECT response: %w", readErr)
		}
		response := string(buf[:n])
		if len(response) >= 12 && response[9:12] == "200" {
			return true, nil
		}
		return false, fmt.Errorf("proxy rejected: %s", response[:min(len(response), 80)])

	default:
		return false, fmt.Errorf("unsupported proxy scheme: %s", parsed.Scheme)
	}
}

func sanitizeURL(rawURL string) string {
	if rawURL == "" {
		return "<none>"
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid>"
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("***", "***")
	}
	return parsed.String()
}

func stringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
