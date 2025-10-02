package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go.mau.fi/whatsmeow/api/internal/locks"
)

type HealthHandler struct {
	db          *pgxpool.Pool
	lockManager locks.Manager

	healthCheckMetric func(component, status string)
}

func NewHealthHandler(db *pgxpool.Pool, lockManager locks.Manager) *HealthHandler {
	return &HealthHandler{
		db:          db,
		lockManager: lockManager,
	}
}

func (h *HealthHandler) SetMetrics(healthCheckMetric func(component, status string)) {
	h.healthCheckMetric = healthCheckMetric
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "whatsapp-api",
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	response := map[string]interface{}{
		"ready":  true,
		"checks": make(map[string]interface{}),
	}

	if err := h.checkDatabase(ctx); err != nil {
		response["ready"] = false
		response["checks"].(map[string]interface{})["database"] = map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		response["checks"].(map[string]interface{})["database"] = map[string]string{
			"status": "healthy",
		}
	}

	if err := h.checkRedis(ctx); err != nil {
		response["ready"] = false
		response["checks"].(map[string]interface{})["redis"] = map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		response["checks"].(map[string]interface{})["redis"] = map[string]string{
			"status": "healthy",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if response["ready"].(bool) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) error {
	if h.db == nil {
		if h.healthCheckMetric != nil {
			h.healthCheckMetric("database", "unhealthy")
		}
		return fmt.Errorf("database not configured")
	}
	err := h.db.Ping(ctx)

	if h.healthCheckMetric != nil {
		status := "healthy"
		if err != nil {
			status = "unhealthy"
		}
		h.healthCheckMetric("database", status)
	}

	return err
}

func (h *HealthHandler) checkRedis(ctx context.Context) error {
	if h.lockManager == nil {
		if h.healthCheckMetric != nil {
			h.healthCheckMetric("redis", "unhealthy")
		}
		return fmt.Errorf("redis not configured")
	}

	testKey := "health:check:test"
	lock, acquired, err := h.lockManager.Acquire(ctx, testKey, 5)

	if h.healthCheckMetric != nil {
		status := "healthy"
		if err != nil {
			status = "unhealthy"
		}
		h.healthCheckMetric("redis", status)
	}

	if err != nil {
		return fmt.Errorf("redis unavailable: %w", err)
	}
	if acquired && lock != nil {
		lock.Release(ctx)
	}
	return nil
}
