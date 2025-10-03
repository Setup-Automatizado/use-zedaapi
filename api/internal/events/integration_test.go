package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/database"
	"go.mau.fi/whatsmeow/api/internal/events"
	"go.mau.fi/whatsmeow/api/internal/observability"
	"github.com/prometheus/client_golang/prometheus"
)

// TestEventOrchestrator_Initialization tests basic orchestrator creation
func TestEventOrchestrator_Initialization(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Load test config
	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config")

	// Create database pool
	pool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	require.NoError(t, err, "failed to create database pool")
	defer pool.Close()

	// Create metrics
	metrics := observability.NewMetrics("test", prometheus.NewRegistry())

	// Create orchestrator
	orchestrator, err := events.NewOrchestrator(ctx, cfg, pool, metrics)
	require.NoError(t, err, "failed to create orchestrator")
	require.NotNil(t, orchestrator, "orchestrator should not be nil")

	// Cleanup
	defer orchestrator.Stop(ctx)
}

// TestEventOrchestrator_InstanceLifecycle tests instance registration and unregistration
func TestEventOrchestrator_InstanceLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup
	cfg, err := config.Load()
	require.NoError(t, err)

	pool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	require.NoError(t, err)
	defer pool.Close()

	metrics := observability.NewMetrics("test", prometheus.NewRegistry())

	orchestrator, err := events.NewOrchestrator(ctx, cfg, pool, metrics)
	require.NoError(t, err)
	defer orchestrator.Stop(ctx)

	// Test instance registration
	instanceID := uuid.New()

	// Register instance
	err = orchestrator.RegisterInstance(ctx, instanceID)
	require.NoError(t, err, "failed to register instance")

	// Verify instance is registered
	assert.True(t, orchestrator.IsInstanceRegistered(instanceID), "instance should be registered")

	// Verify handler exists
	handler, exists := orchestrator.GetHandler(instanceID)
	assert.True(t, exists, "handler should exist")
	assert.NotNil(t, handler, "handler should not be nil")

	// Unregister instance
	err = orchestrator.UnregisterInstance(ctx, instanceID)
	require.NoError(t, err, "failed to unregister instance")

	// Verify instance is unregistered
	assert.False(t, orchestrator.IsInstanceRegistered(instanceID), "instance should not be registered")
}

// TestIntegrationHelper_Creation tests integration helper creation
func TestIntegrationHelper_Creation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup
	cfg, err := config.Load()
	require.NoError(t, err)

	pool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	require.NoError(t, err)
	defer pool.Close()

	metrics := observability.NewMetrics("test", prometheus.NewRegistry())

	orchestrator, err := events.NewOrchestrator(ctx, cfg, pool, metrics)
	require.NoError(t, err)
	defer orchestrator.Stop(ctx)

	// Create integration helper
	integration := events.NewIntegrationHelper(ctx, orchestrator)
	require.NotNil(t, integration, "integration helper should not be nil")

	// Test instance lifecycle methods
	instanceID := uuid.New()

	// OnInstanceConnect should register
	err = integration.OnInstanceConnect(ctx, instanceID)
	require.NoError(t, err, "OnInstanceConnect should succeed")
	assert.True(t, orchestrator.IsInstanceRegistered(instanceID), "instance should be registered")

	// OnInstanceDisconnect should flush
	err = integration.OnInstanceDisconnect(ctx, instanceID)
	require.NoError(t, err, "OnInstanceDisconnect should succeed")

	// OnInstanceRemove should unregister
	err = integration.OnInstanceRemove(ctx, instanceID)
	require.NoError(t, err, "OnInstanceRemove should succeed")
	assert.False(t, orchestrator.IsInstanceRegistered(instanceID), "instance should be unregistered")
}

// TestEventOrchestrator_FlushInstance tests buffer flushing
func TestEventOrchestrator_FlushInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup
	cfg, err := config.Load()
	require.NoError(t, err)

	pool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	require.NoError(t, err)
	defer pool.Close()

	metrics := observability.NewMetrics("test", prometheus.NewRegistry())

	orchestrator, err := events.NewOrchestrator(ctx, cfg, pool, metrics)
	require.NoError(t, err)
	defer orchestrator.Stop(ctx)

	// Register instance
	instanceID := uuid.New()
	err = orchestrator.RegisterInstance(ctx, instanceID)
	require.NoError(t, err)

	// Flush instance
	err = orchestrator.FlushInstance(instanceID)
	assert.NoError(t, err, "flush should succeed")

	// Cleanup
	err = orchestrator.UnregisterInstance(ctx, instanceID)
	require.NoError(t, err)
}

// TestEventOrchestrator_DoubleRegistration tests that double registration fails
func TestEventOrchestrator_DoubleRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup
	cfg, err := config.Load()
	require.NoError(t, err)

	pool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	require.NoError(t, err)
	defer pool.Close()

	metrics := observability.NewMetrics("test", prometheus.NewRegistry())

	orchestrator, err := events.NewOrchestrator(ctx, cfg, pool, metrics)
	require.NoError(t, err)
	defer orchestrator.Stop(ctx)

	// Register instance
	instanceID := uuid.New()
	err = orchestrator.RegisterInstance(ctx, instanceID)
	require.NoError(t, err)

	// Try to register again - should fail
	err = orchestrator.RegisterInstance(ctx, instanceID)
	assert.Error(t, err, "double registration should fail")

	// Cleanup
	err = orchestrator.UnregisterInstance(ctx, instanceID)
	require.NoError(t, err)
}

// TestEventOrchestrator_StopWithActiveInstances tests graceful shutdown
func TestEventOrchestrator_StopWithActiveInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup
	cfg, err := config.Load()
	require.NoError(t, err)

	pool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	require.NoError(t, err)
	defer pool.Close()

	metrics := observability.NewMetrics("test", prometheus.NewRegistry())

	orchestrator, err := events.NewOrchestrator(ctx, cfg, pool, metrics)
	require.NoError(t, err)

	// Register multiple instances
	instanceIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, id := range instanceIDs {
		err = orchestrator.RegisterInstance(ctx, id)
		require.NoError(t, err)
	}

	// Stop orchestrator (should handle active instances gracefully)
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// This should not panic or hang
	orchestrator.Stop(stopCtx)

	// All instances should be unregistered after stop
	for _, id := range instanceIDs {
		assert.False(t, orchestrator.IsInstanceRegistered(id), "instance should be unregistered after stop")
	}
}
