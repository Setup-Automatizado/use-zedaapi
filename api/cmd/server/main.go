package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/database"
	apihandler "go.mau.fi/whatsmeow/api/internal/http"
	"go.mau.fi/whatsmeow/api/internal/http/handlers"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
	redisinit "go.mau.fi/whatsmeow/api/internal/redis"
	sentryinit "go.mau.fi/whatsmeow/api/internal/sentry"
	"go.mau.fi/whatsmeow/api/internal/whatsmeow"
	"go.mau.fi/whatsmeow/api/migrations"
)

type repositoryAdapter struct {
	repo *instances.Repository
}

func (a *repositoryAdapter) ListInstancesWithStoreJID(ctx context.Context) ([]whatsmeow.StoreLink, error) {
	links, err := a.repo.ListInstancesWithStoreJID(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]whatsmeow.StoreLink, len(links))
	for i, link := range links {
		result[i] = whatsmeow.StoreLink{
			ID:       link.ID,
			StoreJID: link.StoreJID,
		}
	}
	return result, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	for _, path := range []string{"api/.env", ".env"} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	logger := logging.New(cfg.Log.Level)
	logger.Info("starting FunnelChat API", slog.String("env", cfg.AppEnv))

	sentryHandler, err := sentryinit.Init(cfg.Sentry.DSN, cfg.Sentry.Environment, cfg.Sentry.Release)
	if err != nil {
		logger.Error("sentry init failed", slog.String("error", err.Error()))
	}

	metrics := observability.NewMetrics(cfg.Prometheus.Namespace, prometheus.DefaultRegisterer)

	pgPool, err := database.NewPool(ctx, cfg.Postgres.DSN, cfg.Postgres.MaxConns)
	if err != nil {
		logger.Error("postgres connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pgPool.Close()

	if err := migrations.Apply(ctx, pgPool, logger); err != nil {
		logger.Error("apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisClient := redisinit.NewClient(redisinit.Config{
		Addr:       cfg.Redis.Addr,
		Username:   cfg.Redis.Username,
		Password:   cfg.Redis.Password,
		DB:         cfg.Redis.DB,
		TLSEnabled: cfg.Redis.TLSEnabled,
	})
	defer redisClient.Close()

	redisLockManager := locks.NewRedisManager(redisClient)

	cbConfig := locks.DefaultCircuitBreakerConfig()
	lockManager := locks.NewCircuitBreakerManager(redisLockManager, cbConfig)

	lockManager.OnStateChange(func(old, new locks.CircuitState) {
		logger.Warn("lock manager circuit breaker state changed",
			slog.String("old_state", old.String()),
			slog.String("new_state", new.String()))
	})

	lockManager.SetMetrics(
		func() { metrics.LockAcquisitions.WithLabelValues("success").Inc() },
		func() { metrics.LockAcquisitions.WithLabelValues("failure").Inc() },
		func(state float64) { metrics.CircuitBreakerState.Set(state) },
	)

	defer lockManager.StopHealthCheck()

	repo := instances.NewRepository(pgPool)

	pairCallback := func(ctx context.Context, id uuid.UUID, jid string) error {
		return repo.UpdateStoreJID(ctx, id, &jid)
	}
	resetCallback := func(ctx context.Context, id uuid.UUID, _ string) error {
		return repo.UpdateStoreJID(ctx, id, nil)
	}

	repoAdapter := &repositoryAdapter{repo: repo}
	registry, err := whatsmeow.NewClientRegistry(ctx, cfg.WhatsmeowStore.DSN, cfg.WhatsmeowStore.LogLevel, lockManager, repoAdapter, logger, pairCallback, resetCallback)
	if err != nil {
		logger.Error("whatsmeow registry", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		logger.Info("phase 2: closing whatsmeow registry")
		registryCloseDone := make(chan error, 1)
		go func() {
			registryCloseDone <- registry.Close()
		}()

		select {
		case err := <-registryCloseDone:
			if err != nil {
				logger.Error("registry close failed", slog.String("error", err.Error()))
			} else {
				logger.Info("registry closed successfully")
			}
		case <-time.After(5 * time.Second):
			logger.Warn("registry close timeout after 5 seconds - continuing shutdown")
		}
	}()

	instanceService := instances.NewService(repo, registry, logger)
	reconcileCtx, reconcileCancel := context.WithTimeout(ctx, 30*time.Second)
	cleaned, reconcileErr := instanceService.ReconcileDetachedStores(reconcileCtx)
	reconcileCancel()
	if reconcileErr != nil {
		logger.Error("reconcile stores", slog.String("error", reconcileErr.Error()))
	} else if len(cleaned) > 0 {
		logger.Info("reconciled stores", slog.Int("count", len(cleaned)))
	}

	connectCtx, connectCancel := context.WithTimeout(ctx, 60*time.Second)
	connected, skipped, connectErr := registry.ConnectExistingClients(connectCtx)
	connectCancel()
	if connectErr != nil {
		logger.Error("connect existing clients", slog.String("error", connectErr.Error()))
	} else if connected > 0 {
		logger.Info("connected existing clients", slog.Int("connected", connected), slog.Int("skipped", skipped))
	}

	registry.SetSplitBrainMetrics(func() { metrics.SplitBrainDetected.Inc() })
	registry.StartSplitBrainDetection()

	instanceHandler := handlers.NewInstanceHandler(instanceService, logger)
	partnerHandler := handlers.NewPartnerHandler(instanceService, logger)
	healthHandler := handlers.NewHealthHandler(pgPool, lockManager)

	healthHandler.SetMetrics(func(component, status string) {
		metrics.HealthChecks.WithLabelValues(component, status).Inc()
	})

	router := apihandler.NewRouter(apihandler.RouterDeps{
		Logger:          logger,
		Metrics:         metrics,
		SentryHandler:   sentryHandler,
		InstanceHandler: instanceHandler,
		PartnerHandler:  partnerHandler,
		HealthHandler:   healthHandler,
		PartnerToken:    cfg.Partner.AuthToken,
	})

	server := apihandler.NewServer(
		router,
		cfg.HTTP.Addr,
		cfg.HTTP.ReadHeaderTimeout,
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
		cfg.HTTP.MaxHeaderBytes,
		logger,
	)

	if err := server.Run(ctx); err != nil {
		logger.Error("http server stopped", slog.String("error", err.Error()))
	}

	emergencyTimeout := time.AfterFunc(45*time.Second, func() {
		logger.Error("EMERGENCY TIMEOUT: Forcing exit after 45 seconds")
		os.Exit(1)
	})
	defer emergencyTimeout.Stop()

	logger.Info("starting graceful shutdown sequence")

	logger.Info("phase 0: releasing redis locks")
	released := registry.ReleaseAllLocks()
	logger.Info("phase 0 complete: redis locks released", slog.Int("count", released))

	logger.Info("phase 1: disconnecting all whatsapp clients")
	disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	disconnected := registry.DisconnectAll(disconnectCtx)
	disconnectCancel()
	if disconnected > 0 {
		logger.Info("phase 1: disconnected all clients", slog.Int("count", disconnected))
	} else {
		logger.Info("phase 1: no active clients to disconnect")
	}

	if cfg.Sentry.DSN != "" {
		logger.Info("phase 3: flushing sentry events")
		sentry.Flush(5 * time.Second)
	}

	logger.Info("shutdown complete")
}
