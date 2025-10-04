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
	"go.mau.fi/whatsmeow/api/internal/events"
	"go.mau.fi/whatsmeow/api/internal/events/capture"
	"go.mau.fi/whatsmeow/api/internal/events/dispatch"
	"go.mau.fi/whatsmeow/api/internal/events/media"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	transporthttp "go.mau.fi/whatsmeow/api/internal/events/transport/http"
	"go.mau.fi/whatsmeow/api/internal/events/types"
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

type instanceLookupAdapter struct {
	repo *instances.Repository
}

func (a *instanceLookupAdapter) StoreJID(ctx context.Context, id uuid.UUID) (string, error) {
	inst, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if inst.StoreJID == nil {
		return "", nil
	}
	return *inst.StoreJID, nil
}

type webhookResolverAdapter struct {
	repo *instances.Repository
}

func (a *webhookResolverAdapter) Resolve(ctx context.Context, id uuid.UUID) (*capture.ResolvedWebhookConfig, error) {
	inst, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	webhook, err := a.repo.GetWebhookConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	deref := func(ptr *string) string {
		if ptr == nil {
			return ""
		}
		return *ptr
	}

	return &capture.ResolvedWebhookConfig{
		DeliveryURL:         deref(webhook.DeliveryURL),
		ReceivedURL:         deref(webhook.ReceivedURL),
		ReceivedDeliveryURL: deref(webhook.ReceivedDeliveryURL),
		MessageStatusURL:    deref(webhook.MessageStatusURL),
		DisconnectedURL:     deref(webhook.DisconnectedURL),
		ChatPresenceURL:     deref(webhook.ChatPresenceURL),
		ConnectedURL:        deref(webhook.ConnectedURL),
		NotifySentByMe:      webhook.NotifySentByMe,
		StoreJID:            inst.StoreJID,
		ClientToken:         inst.ClientToken,
	}, nil
}

func storeJIDEnricher() capture.MetadataEnricher {
	return func(cfg *capture.ResolvedWebhookConfig, event *types.InternalEvent) {
		if cfg == nil || cfg.StoreJID == nil || *cfg.StoreJID == "" {
			return
		}
		if event.Metadata == nil {
			event.Metadata = make(map[string]string)
		}
		if _, ok := event.Metadata["store_jid"]; !ok {
			event.Metadata["store_jid"] = *cfg.StoreJID
		}
	}
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

	repo := instances.NewRepository(pgPool)

	resolver := &webhookResolverAdapter{repo: repo}
	metadataEnricher := storeJIDEnricher()

	eventOrchestrator, err := events.NewOrchestrator(ctx, cfg, pgPool, resolver, metadataEnricher, metrics)
	if err != nil {
		logger.Error("failed to create event orchestrator", slog.String("error", err.Error()))
		os.Exit(1)
	}

	eventIntegration := events.NewIntegrationHelper(ctx, eventOrchestrator)

	logger.Info("event system initialized",
		slog.Int("buffer_size", cfg.Events.BufferSize),
		slog.Int("batch_size", cfg.Events.BatchSize),
	)

	outboxRepo := persistence.NewOutboxRepository(pgPool)
	dlqRepo := persistence.NewDLQRepository(pgPool)

	httpTransportConfig := transporthttp.DefaultConfig()
	httpTransportConfig.Timeout = cfg.Events.WebhookTimeout
	httpTransportConfig.MaxRetries = cfg.Events.WebhookMaxRetries
	transportRegistry := transport.NewRegistry(httpTransportConfig)
	defer transportRegistry.Close()

	dispatchCoordinator := dispatch.NewCoordinator(
		&cfg,
		pgPool,
		outboxRepo,
		dlqRepo,
		transportRegistry,
		&instanceLookupAdapter{repo: repo},
		metrics,
	)

	if err := dispatchCoordinator.Start(ctx); err != nil {
		logger.Error("failed to start dispatch coordinator", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		logger.Info("phase 2.3: stopping dispatch coordinator")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		dispatchCoordinator.Stop(shutdownCtx)
		shutdownCancel()
		logger.Info("dispatch coordinator stopped")
	}()

	logger.Info("dispatch system initialized",
		slog.Int("workers", dispatchCoordinator.GetWorkerCount()))

	mediaRepo := persistence.NewMediaRepository(pgPool)
	mediaCoordinator := media.NewMediaCoordinator(
		&cfg,
		mediaRepo,
		outboxRepo,
		metrics,
	)

	logger.Info("media processing system initialized",
		slog.Int("max_workers", cfg.Workers.Media))

	var mediaHTTPHandler *handlers.MediaHandler
	if localStorage, err := media.NewLocalMediaStorage(ctx, &cfg, metrics); err != nil {
		logger.Warn("local media storage disabled",
			slog.String("error", err.Error()))
	} else {
		mediaHTTPHandler = handlers.NewMediaHandler(localStorage, metrics, logger)
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

	lockManager.SetMetrics(locks.CircuitBreakerMetricsCallbacks{
		LockSuccess:  func() { metrics.LockAcquisitions.WithLabelValues("success").Inc() },
		LockFailure:  func() { metrics.LockAcquisitions.WithLabelValues("failure").Inc() },
		CircuitState: func(state float64) { metrics.CircuitBreakerState.Set(state) },
		ReacquireAttempt: func(instanceID, result string) {
			metrics.LockReacquisitionAttempts.WithLabelValues(instanceID, result).Inc()
		},
		ReacquireFallback: func(instanceID, circuitState string) {
			metrics.LockReacquisitionFallbacks.WithLabelValues(instanceID, circuitState).Inc()
		},
	})

	defer lockManager.StopHealthCheck()

	pairCallback := func(ctx context.Context, id uuid.UUID, jid string) error {
		return repo.UpdateStoreJID(ctx, id, &jid)
	}
	resetCallback := func(ctx context.Context, id uuid.UUID, _ string) error {
		return repo.UpdateStoreJID(ctx, id, nil)
	}

	repoAdapter := &repositoryAdapter{repo: repo}
	registry, err := whatsmeow.NewClientRegistry(ctx, cfg.WhatsmeowStore.DSN, cfg.WhatsmeowStore.LogLevel, lockManager, repoAdapter, logger, pairCallback, resetCallback, eventIntegration, dispatchCoordinator, mediaCoordinator)
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

	registry.SetMetrics(whatsmeow.ClientRegistryMetrics{
		SplitBrainDetected: func() { metrics.SplitBrainDetected.Inc() },
		SplitBrainInvalidLock: func(instanceID string) {
			metrics.SplitBrainInvalidLocks.WithLabelValues(instanceID).Inc()
		},
	})
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
		MediaHandler:    mediaHTTPHandler,
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

	logger.Info("releasing redis locks")
	released := registry.ReleaseAllLocks()
	logger.Info("redis locks released", slog.Int("count", released))

	logger.Info("disconnecting all whatsapp clients")
	disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	disconnected := registry.DisconnectAll(disconnectCtx)
	disconnectCancel()
	if disconnected > 0 {
		logger.Info("disconnected all clients", slog.Int("count", disconnected))
	} else {
		logger.Info("no active clients to disconnect")
	}

	logger.Info("stopping event orchestrator")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	eventOrchestrator.Stop(shutdownCtx)
	shutdownCancel()

	if cfg.Sentry.DSN != "" {
		logger.Info("flushing sentry events")
		sentry.Flush(5 * time.Second)
	}

	logger.Info("shutdown complete")
}
