package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/version"
)

type componentStatus struct {
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
	CircuitState string `json:"circuit_state,omitempty"`
}

type readinessResponse struct {
	Ready      bool                       `json:"ready"`
	ObservedAt time.Time                  `json:"observed_at"`
	Checks     map[string]componentStatus `json:"checks"`
}

// NATSHealthChecker provides NATS health status for readiness checks.
type NATSHealthChecker interface {
	IsConnected() bool
}

type HealthHandler struct {
	db          *pgxpool.Pool
	lockManager locks.Manager
	natsClient  NATSHealthChecker

	healthCheckMetric func(component, status string)
}

func NewHealthHandler(db *pgxpool.Pool, lockManager locks.Manager) *HealthHandler {
	return &HealthHandler{
		db:          db,
		lockManager: lockManager,
	}
}

// SetNATSClient enables NATS health checking in the readiness probe.
func (h *HealthHandler) SetNATSClient(client NATSHealthChecker) {
	h.natsClient = client
}

func (h *HealthHandler) SetMetrics(healthCheckMetric func(component, status string)) {
	h.healthCheckMetric = healthCheckMetric
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	versionInfo := version.Get()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"service":    "zedaapi",
		"version":    versionInfo.Version,
		"build_time": versionInfo.BuildTime,
		"git_commit": versionInfo.GitCommit,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	logger := logging.ContextLogger(r.Context(), nil)

	dbStatus, dbErr := h.checkDatabase(ctx)
	redisStatus, redisErr := h.checkRedis(ctx)

	checks := map[string]componentStatus{
		"database": dbStatus,
		"redis":    redisStatus,
	}
	ready := dbStatus.Status == "healthy" && redisStatus.Status == "healthy"

	if h.natsClient != nil {
		natsStatus := h.checkNATS()
		checks["nats"] = natsStatus
		if natsStatus.Status != "healthy" {
			ready = false
		}
	}

	if dbErr != nil {
		logger.Error("database health check failed",
			slog.String("error", dbErr.Error()),
			slog.String("status", dbStatus.Status))
		captureHealthCheckFailure("database", dbStatus, dbErr)
	}
	if redisErr != nil {
		logger.Error("redis health check failed",
			slog.String("error", redisErr.Error()),
			slog.String("status", redisStatus.Status),
			slog.String("circuit_state", redisStatus.CircuitState))
		captureHealthCheckFailure("redis", redisStatus, redisErr)
	}

	response := readinessResponse{
		Ready:      ready,
		ObservedAt: time.Now().UTC(),
		Checks:     checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) (componentStatus, error) {
	result := componentStatus{Status: "healthy"}
	start := time.Now()
	defer func() {
		result.DurationMs = time.Since(start).Milliseconds()
	}()

	if h.db == nil {
		err := fmt.Errorf("database not configured")
		result.Status = "unhealthy"
		result.Error = err.Error()
		h.recordMetric("database", result.Status)
		return result, err
	}

	err := h.db.Ping(ctx)
	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
	}

	h.recordMetric("database", result.Status)
	return result, err
}

func (h *HealthHandler) checkRedis(ctx context.Context) (componentStatus, error) {
	result := componentStatus{Status: "healthy"}
	start := time.Now()
	defer func() {
		result.DurationMs = time.Since(start).Milliseconds()
	}()

	if h.lockManager == nil {
		err := fmt.Errorf("redis not configured")
		result.Status = "unhealthy"
		result.Error = err.Error()
		h.recordMetric("redis", result.Status)
		return result, err
	}

	if provider, ok := h.lockManager.(interface{ GetState() locks.CircuitState }); ok {
		result.CircuitState = provider.GetState().String()
	}

	testKey := "health:check:test"
	lock, acquired, err := h.lockManager.Acquire(ctx, testKey, 5)

	switch {
	case err != nil:
		result.Status = "unhealthy"
		result.Error = err.Error()
	case !acquired:
		result.Status = "degraded"
		result.Error = "lock acquisition unsuccessful"
		err = errors.New(result.Error)
	case lock != nil && lock.GetValue() == "":
		result.Status = "degraded"
		result.Error = "fallback lock in use"
		err = errors.New(result.Error)
	}

	if lock != nil {
		_ = lock.Release(context.Background())
	}

	h.recordMetric("redis", result.Status)
	if err != nil && !errors.Is(err, context.Canceled) {
		return result, fmt.Errorf("redis check failed: %w", err)
	}
	return result, err
}

func (h *HealthHandler) checkNATS() componentStatus {
	result := componentStatus{Status: "healthy"}
	start := time.Now()
	defer func() {
		result.DurationMs = time.Since(start).Milliseconds()
	}()

	if h.natsClient == nil {
		result.Status = "unhealthy"
		result.Error = "nats client not configured"
		h.recordMetric("nats", result.Status)
		return result
	}

	if !h.natsClient.IsConnected() {
		result.Status = "unhealthy"
		result.Error = "nats not connected"
	}

	h.recordMetric("nats", result.Status)
	return result
}

func (h *HealthHandler) recordMetric(component, status string) {
	if h.healthCheckMetric != nil {
		h.healthCheckMetric(component, status)
	}
}

func captureHealthCheckFailure(component string, status componentStatus, err error) {
	if err == nil {
		return
	}
	if hub := sentry.CurrentHub(); hub == nil || hub.Client() == nil {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("component", component)
		scope.SetLevel(sentry.LevelWarning)
		scope.SetContext("healthcheck", map[string]any{
			"status":        status.Status,
			"duration_ms":   status.DurationMs,
			"error":         status.Error,
			"circuit_state": status.CircuitState,
		})
		sentry.CaptureException(err)
	})
}
