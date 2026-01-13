package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	wameow "go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/docs"
	"go.mau.fi/whatsmeow/api/internal/chats"
	"go.mau.fi/whatsmeow/api/internal/communities"
	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/contacts"
	"go.mau.fi/whatsmeow/api/internal/database"
	"go.mau.fi/whatsmeow/api/internal/events"
	"go.mau.fi/whatsmeow/api/internal/events/capture"
	"go.mau.fi/whatsmeow/api/internal/events/dispatch"
	"go.mau.fi/whatsmeow/api/internal/events/media"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	transporthttp "go.mau.fi/whatsmeow/api/internal/events/transport/http"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/groups"
	apihandler "go.mau.fi/whatsmeow/api/internal/http"
	"go.mau.fi/whatsmeow/api/internal/http/handlers"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
	"go.mau.fi/whatsmeow/api/internal/newsletters"
	"go.mau.fi/whatsmeow/api/internal/observability"
	redisinit "go.mau.fi/whatsmeow/api/internal/redis"
	sentryinit "go.mau.fi/whatsmeow/api/internal/sentry"
	"go.mau.fi/whatsmeow/api/internal/statuscache"
	"go.mau.fi/whatsmeow/api/internal/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/workers"
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

func (a *repositoryAdapter) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, connected bool, status string, workerID *string, desiredWorkerID *string) error {
	return a.repo.UpdateConnectionStatus(ctx, id, connected, status, workerID, desiredWorkerID)
}

func (a *repositoryAdapter) GetConnectionState(ctx context.Context, id uuid.UUID) (*whatsmeow.ConnectionState, error) {
	state, err := a.repo.GetConnectionState(ctx, id)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil
	}

	return &whatsmeow.ConnectionState{
		Connected:       state.Connected,
		Status:          state.Status,
		LastConnectedAt: state.LastConnectedAt,
		WorkerID:        state.WorkerID,
		DesiredWorkerID: state.DesiredWorkerID,
	}, nil
}

func (a *repositoryAdapter) GetCallRejectConfig(ctx context.Context, id uuid.UUID) (bool, *string, error) {
	return a.repo.GetCallRejectConfig(ctx, id)
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

func (a *instanceLookupAdapter) IsBusiness(ctx context.Context, id uuid.UUID) (bool, error) {
	inst, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return inst.BusinessDevice, nil
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
	}, nil
}

type queueClientRegistryAdapter struct {
	registry *whatsmeow.ClientRegistry
}

func (a *queueClientRegistryAdapter) GetClient(instanceID string) (*wameow.Client, bool) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, false
	}
	return a.registry.GetClient(id)
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

	if cfg.Document.MuPDFVersion != "" {
		fitz.FzVersion = cfg.Document.MuPDFVersion
	}

	logger := logging.New(cfg.Log.Level)
	logger.Info("starting FunnelChat API", slog.String("env", cfg.AppEnv))
	if cfg.Document.MuPDFVersion != "" {
		logger.Debug("configured MuPDF runtime",
			slog.String("version", cfg.Document.MuPDFVersion))
	}

	sentryHandler, err := sentryinit.Init(cfg.Sentry.DSN, cfg.Sentry.Environment, cfg.Sentry.Release)
	if err != nil {
		logger.Error("sentry init failed", slog.String("error", err.Error()))
	}

	if sentryinit.Enabled() {
		hostname, _ := os.Hostname()
		tags := map[string]string{
			"environment": cfg.Sentry.Environment,
			"app_env":     cfg.AppEnv,
		}
		extras := map[string]any{
			"hostname":             hostname,
			"http_addr":            cfg.HTTP.Addr,
			"prometheus_namespace": cfg.Prometheus.Namespace,
			"mu_pdf_version":       cfg.Document.MuPDFVersion,
		}
		sentryinit.CaptureLifecycleEvent("startup", tags, extras)
		defer func() {
			sentryinit.CaptureLifecycleEvent("shutdown", tags, extras)
			sentryinit.Flush(5 * time.Second)
		}()
	}

	metrics := observability.NewMetrics(cfg.Prometheus.Namespace, prometheus.DefaultRegisterer)

	// Ensure all required databases exist before connecting
	// This allows automatic database creation in fresh environments (dev, CI/CD, deploy)
	databases := map[string]string{
		"application": cfg.Postgres.DSN,
		"whatsmeow":   cfg.WhatsmeowStore.DSN,
	}
	if err := database.EnsureMultipleDatabases(ctx, databases, logger); err != nil {
		logger.Error("ensure databases exist", slog.String("error", err.Error()))
		os.Exit(1)
	}

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

	pollStore := pollstore.NewRedisStore(redisClient, 365*24*time.Hour)
	redisLockManager := locks.NewRedisManager(redisClient)

	repo := instances.NewRepository(pgPool)

	resolver := &webhookResolverAdapter{repo: repo}
	metadataEnricher := storeJIDEnricher()

	eventOrchestrator, err := events.NewOrchestrator(ctx, cfg, pgPool, resolver, metadataEnricher, pollStore, metrics)
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
	transportRegistry := transport.NewRegistry(httpTransportConfig, metrics)
	defer transportRegistry.Close()

	dispatchCoordinator := dispatch.NewCoordinator(
		&cfg,
		pgPool,
		outboxRepo,
		dlqRepo,
		transportRegistry,
		&instanceLookupAdapter{repo: repo},
		pollStore,
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

	// Initialize status cache system if enabled
	var statusCacheService *statuscache.ServiceImpl
	if cfg.StatusCache.Enabled {
		statusCacheRepo := statuscache.NewRedisRepository(
			redisClient,
			cfg.StatusCache.TTL,
			"funnelchat",
		)

		// Create webhook dispatcher for flush operations
		// This will look up the instance's webhook URL and send via HTTP transport
		webhookDispatcher := statuscache.NewFlushDispatcher(func(ctx context.Context, instanceID string, payload []byte) error {
			// Look up instance webhook configuration
			instID, parseErr := uuid.Parse(instanceID)
			if parseErr != nil {
				return fmt.Errorf("invalid instance ID: %w", parseErr)
			}

			webhookConfig, err := repo.GetWebhookConfig(ctx, instID)
			if err != nil {
				return fmt.Errorf("failed to get webhook config: %w", err)
			}

			// Use message status URL, fallback to delivery URL
			var targetURL string
			if webhookConfig.MessageStatusURL != nil && *webhookConfig.MessageStatusURL != "" {
				targetURL = *webhookConfig.MessageStatusURL
			} else if webhookConfig.DeliveryURL != nil && *webhookConfig.DeliveryURL != "" {
				targetURL = *webhookConfig.DeliveryURL
			} else {
				// No webhook URL configured, skip silently
				return nil
			}

			// Get HTTP transport
			trans, err := transportRegistry.GetTransport(transport.TransportTypeHTTP)
			if err != nil {
				return fmt.Errorf("failed to get transport: %w", err)
			}

			// Build request
			request := &transport.DeliveryRequest{
				Endpoint:    targetURL,
				Payload:     payload,
				Headers:     map[string]string{"Content-Type": "application/json"},
				EventID:     fmt.Sprintf("flush-%s-%d", instanceID, time.Now().UnixNano()),
				EventType:   "receipt",
				InstanceID:  instanceID,
				Attempt:     1,
				MaxAttempts: 3,
			}

			// Deliver webhook
			result, err := trans.Deliver(ctx, request)
			if err != nil {
				return fmt.Errorf("delivery failed: %w", err)
			}
			if !result.Success {
				return fmt.Errorf("delivery unsuccessful: %s", result.ErrorMessage)
			}

			return nil
		})

		statusCacheService = statuscache.NewService(
			statusCacheRepo,
			&cfg,
			webhookDispatcher,
			metrics,
			logger,
		)

		if err := statusCacheService.Start(ctx); err != nil {
			logger.Error("failed to start status cache service", slog.String("error", err.Error()))
		} else {
			logger.Info("status cache service started",
				slog.Duration("ttl", cfg.StatusCache.TTL),
				slog.Bool("suppress_webhooks", cfg.StatusCache.SuppressWebhooks),
			)

			// Create interceptor and wire to dispatch coordinator
			statusInterceptor := statuscache.NewInterceptor(statusCacheService, &cfg, logger)
			dispatchCoordinator.SetStatusInterceptor(statusInterceptor)
			logger.Info("status cache interceptor configured for dispatch coordinator")
		}

		defer func() {
			if statusCacheService != nil {
				logger.Info("stopping status cache service")
				stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := statusCacheService.Stop(stopCtx); err != nil {
					logger.Error("status cache service stop error", slog.String("error", err.Error()))
				}
				stopCancel()
				logger.Info("status cache service stopped")
			}
		}()
	}

	mediaRepo := persistence.NewMediaRepository(pgPool)
	mediaCoordinator := media.NewMediaCoordinator(
		&cfg,
		mediaRepo,
		outboxRepo,
		metrics,
	)

	logger.Info("media processing system initialized",
		slog.Int("max_workers", cfg.Workers.Media))

	var (
		mediaHTTPHandler *handlers.MediaHandler
		localStorage     *media.LocalMediaStorage
	)
	if ls, err := media.NewLocalMediaStorage(ctx, &cfg, metrics); err != nil {
		logger.Warn("local media storage disabled",
			slog.String("error", err.Error()))
	} else {
		localStorage = ls
		mediaHTTPHandler = handlers.NewMediaHandler(localStorage, metrics, logger)
	}

	var mediaReaper *media.MediaReaper
	if cfg.Media.CleanupInterval > 0 {
		s3CleanupUploader, err := media.NewS3Uploader(ctx, &cfg, metrics)
		if err != nil {
			logger.Warn("media cleanup disabled: s3 uploader init failed",
				slog.String("error", err.Error()))
		} else {
			reaperCfg := media.MediaReaperConfig{
				Interval:       cfg.Media.CleanupInterval,
				BatchSize:      cfg.Media.CleanupBatchSize,
				S3Retention:    cfg.Media.S3Retention,
				LocalRetention: cfg.Media.LocalRetention,
			}
			mediaReaper, err = media.NewMediaReaper(mediaRepo, s3CleanupUploader, localStorage, metrics, logger, redisLockManager, reaperCfg)
			if err != nil {
				logger.Warn("media cleanup disabled",
					slog.String("error", err.Error()))
			} else {
				mediaReaper.Start(ctx)
				defer func() {
					stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					if err := mediaReaper.Stop(stopCtx); err != nil {
						logger.Warn("media reaper stop error",
							slog.String("error", err.Error()))
					}
				}()
			}
		}
	}

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
	contactMetadataCfg := whatsmeow.ContactMetadataConfig{
		CacheCapacity:   cfg.ContactMetadata.CacheCapacity,
		NameTTL:         cfg.ContactMetadata.NameTTL,
		PhotoTTL:        cfg.ContactMetadata.PhotoTTL,
		ErrorTTL:        cfg.ContactMetadata.ErrorTTL,
		PrefetchWorkers: cfg.ContactMetadata.PrefetchWorkers,
		FetchQueueSize:  cfg.ContactMetadata.FetchQueueSize,
	}
	autoConnectCfg := whatsmeow.AutoConnectConfig{
		Enabled:        cfg.Connection.AutoConnectPaired,
		MaxAttempts:    cfg.Connection.MaxAttempts,
		InitialBackoff: cfg.Connection.InitialBackoff,
		MaxBackoff:     cfg.Connection.MaxBackoff,
	}
	registry, err := whatsmeow.NewClientRegistry(
		ctx,
		cfg.WhatsmeowStore.DSN,
		cfg.WhatsmeowStore.LogLevel,
		lockManager,
		repoAdapter,
		logger,
		pairCallback,
		resetCallback,
		eventIntegration,
		dispatchCoordinator,
		mediaCoordinator,
		cfg.Events.HandlerTimeout,
		contactMetadataCfg,
		redisClient,
		autoConnectCfg,
		whatsmeow.LockConfig{
			KeyPrefix:       cfg.RedisLock.KeyPrefix,
			TTL:             cfg.RedisLock.TTL,
			RefreshInterval: cfg.RedisLock.RefreshInterval,
		},
		metrics,
	)
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

	workerDirectory := workers.NewRegistry(
		pgPool,
		registry.WorkerID(),
		registry.WorkerHostname(),
		cfg.AppEnv,
		workers.Config{
			HeartbeatInterval: cfg.WorkerRegistry.HeartbeatInterval,
			Expiry:            cfg.WorkerRegistry.Expiry,
		},
		logger,
	)

	if err := workerDirectory.Start(ctx); err != nil {
		logger.Error("worker directory start", slog.String("error", err.Error()))
		os.Exit(1)
	}
	registry.AttachWorkerDirectory(workerDirectory, cfg.WorkerRegistry.RebalanceInterval)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
		workerDirectory.Stop(stopCtx)
		stopCancel()
	}()

	instanceService := instances.NewService(repo, registry, cfg.Client.AuthToken, logger)
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

	if cfg.Reconciliation.Enabled {
		logger.Info("starting reconciliation worker",
			slog.Duration("interval", cfg.Reconciliation.Interval))
		registry.StartReconciliationWorker(ctx, cfg.Reconciliation.Interval)
	}

	registry.SetMetrics(whatsmeow.ClientRegistryMetrics{
		SplitBrainDetected: func() { metrics.SplitBrainDetected.Inc() },
		SplitBrainInvalidLock: func(instanceID string) {
			metrics.SplitBrainInvalidLocks.WithLabelValues(instanceID).Inc()
		},
	})
	registry.StartSplitBrainDetection()

	contactsProvider := contacts.NewRegistryClientProvider(registry, repo, logger)
	contactsService := contacts.NewService(contactsProvider, logger)

	chatsProvider := chats.NewRegistryClientProvider(registry, repo, logger)
	chatsService := chats.NewService(chatsProvider, logger)

	// Initialize message queue coordinator
	var messageCoordinator *queue.Coordinator
	var messageHandler *handlers.MessageHandler
	if cfg.MessageQueue.Enabled {
		queueConfig := &queue.Config{
			PollInterval:         cfg.MessageQueue.PollInterval,
			BatchSize:            1, // CRITICAL: Must be 1 for FIFO ordering
			MaxAttempts:          cfg.MessageQueue.MaxAttempts,
			InitialBackoff:       cfg.MessageQueue.InitialBackoff,
			MaxBackoff:           cfg.MessageQueue.MaxBackoff,
			BackoffMultiplier:    cfg.MessageQueue.BackoffMultiplier,
			DisconnectRetryDelay: cfg.MessageQueue.DisconnectRetryDelay,
			CompletedRetention:   cfg.MessageQueue.CompletedRetention,
			FailedRetention:      cfg.MessageQueue.FailedRetention,
			WorkersPerInstance:   1, // CRITICAL: Must be 1 for FIFO ordering
		}

		coordinatorConfig := &queue.CoordinatorConfig{
			Pool:            pgPool,
			ClientRegistry:  &queueClientRegistryAdapter{registry: registry},
			Processor:       queue.NewWhatsAppMessageProcessor(logger),
			Config:          queueConfig,
			Logger:          logger,
			CleanupInterval: cfg.MessageQueue.CleanupInterval,
			CleanupTimeout:  cfg.MessageQueue.CleanupTimeout,
			Metrics:         metrics,
		}

		messageCoordinator, err = queue.NewCoordinator(ctx, coordinatorConfig)
		if err != nil {
			logger.Error("failed to initialize message queue coordinator", slog.String("error", err.Error()))
			os.Exit(1)
		}

		logger.Info("message queue coordinator initialized",
			slog.Duration("poll_interval", cfg.MessageQueue.PollInterval),
			slog.Int("max_attempts", cfg.MessageQueue.MaxAttempts))

		// Create message handler for HTTP endpoints
		messageHandler = handlers.NewMessageHandler(
			messageCoordinator,
			instanceService,
			contactsService,
			chatsService,
			logger,
			cfg.Shutdown.QueueDrainTimeout,
		)

		// Cleanup on shutdown
		defer func() {
			logger.Info("stopping message queue coordinator")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := messageCoordinator.Stop(shutdownCtx); err != nil {
				logger.Error("message queue coordinator stop error", slog.String("error", err.Error()))
			}
			shutdownCancel()
			logger.Info("message queue coordinator stopped")
		}()
	}

	groupsProvider := groups.NewRegistryClientProvider(registry, repo, logger)
	groupsService := groups.NewService(groupsProvider, logger)

	communitiesProvider := communities.NewRegistryClientProvider(registry, repo, logger)
	communitiesService := communities.NewService(communitiesProvider, groupsService, logger)

	newslettersProvider := newsletters.NewRegistryClientProvider(registry, repo, logger)
	newslettersService := newsletters.NewService(newslettersProvider, logger)

	instanceHandler := handlers.NewInstanceHandler(instanceService, logger)
	partnerHandler := handlers.NewPartnerHandler(instanceService, logger)
	healthHandler := handlers.NewHealthHandler(pgPool, lockManager)
	groupsHandler := handlers.NewGroupsHandler(instanceService, groupsService, communitiesService, metrics, logger)
	communitiesHandler := handlers.NewCommunitiesHandler(instanceService, communitiesService, metrics, logger)
	newslettersHandler := handlers.NewNewslettersHandler(instanceService, newslettersService, metrics, logger)

	// Create StatusCacheHandler if service is enabled
	var statusCacheHTTPHandler *handlers.StatusCacheHandler
	if statusCacheService != nil {
		statusCacheHTTPHandler = handlers.NewStatusCacheHandler(statusCacheService, logger)
	}

	// Create PrivacyHandler for privacy settings endpoints
	var privacyHandler *handlers.PrivacyHandler
	if messageCoordinator != nil {
		privacyHandler = handlers.NewPrivacyHandler(messageCoordinator, instanceService, logger)
	}

	healthHandler.SetMetrics(func(component, status string) {
		metrics.HealthChecks.WithLabelValues(component, status).Inc()
	})

	router := apihandler.NewRouter(apihandler.RouterDeps{
		Logger:             logger,
		Metrics:            metrics,
		SentryHandler:      sentryHandler,
		InstanceHandler:    instanceHandler,
		PartnerHandler:     partnerHandler,
		HealthHandler:      healthHandler,
		MediaHandler:       mediaHTTPHandler,
		MessageHandler:     messageHandler,
		GroupsHandler:      groupsHandler,
		CommunitiesHandler: communitiesHandler,
		NewslettersHandler: newslettersHandler,
		StatusCacheHandler: statusCacheHTTPHandler,
		PrivacyHandler:     privacyHandler,
		PartnerToken:       cfg.Partner.AuthToken,
		DocsConfig:         docs.Config{BaseURL: cfg.HTTP.BaseURL},
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

	if messageCoordinator != nil {
		logger.Info("draining message queue before shutdown")
		drainCtx, drainCancel := context.WithTimeout(context.Background(), cfg.Shutdown.QueueDrainTimeout)
		if err := messageCoordinator.DrainQueue(drainCtx); err != nil {
			logger.Error("queue drain error", slog.String("error", err.Error()))
		}
		drainCancel()
	}

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
