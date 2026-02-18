package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
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
	"go.mau.fi/whatsmeow/api/internal/events/echo"
	"go.mau.fi/whatsmeow/api/internal/events/media"
	eventsnats "go.mau.fi/whatsmeow/api/internal/events/nats"
	"go.mau.fi/whatsmeow/api/internal/events/persistence"
	"go.mau.fi/whatsmeow/api/internal/events/pollstore"
	"go.mau.fi/whatsmeow/api/internal/events/transport"
	transporthttp "go.mau.fi/whatsmeow/api/internal/events/transport/http"
	"go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/api/internal/groups"
	apihandler "go.mau.fi/whatsmeow/api/internal/http"
	"go.mau.fi/whatsmeow/api/internal/http/handlers"
	ourmiddleware "go.mau.fi/whatsmeow/api/internal/http/middleware"
	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/messages/queue"
	natsclient "go.mau.fi/whatsmeow/api/internal/nats"
	"go.mau.fi/whatsmeow/api/internal/newsletters"
	"go.mau.fi/whatsmeow/api/internal/observability"
	proxycheck "go.mau.fi/whatsmeow/api/internal/proxy"
	"go.mau.fi/whatsmeow/api/internal/proxy/webshare"
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
		HistorySyncURL:      deref(webhook.HistorySyncURL),
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

// dispatchWebhookResolverAdapter bridges instances.Repository to dispatch.WebhookResolver.
type dispatchWebhookResolverAdapter struct {
	repo        *instances.Repository
	clientToken string
}

func (a *dispatchWebhookResolverAdapter) ResolveWebhook(ctx context.Context, id uuid.UUID) (*dispatch.ResolvedWebhook, error) {
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
	return &dispatch.ResolvedWebhook{
		DeliveryURL:         deref(webhook.DeliveryURL),
		ReceivedURL:         deref(webhook.ReceivedURL),
		ReceivedDeliveryURL: deref(webhook.ReceivedDeliveryURL),
		MessageStatusURL:    deref(webhook.MessageStatusURL),
		DisconnectedURL:     deref(webhook.DisconnectedURL),
		ChatPresenceURL:     deref(webhook.ChatPresenceURL),
		ConnectedURL:        deref(webhook.ConnectedURL),
		HistorySyncURL:      deref(webhook.HistorySyncURL),
		NotifySentByMe:      webhook.NotifySentByMe,
		StoreJID:            inst.StoreJID,
		ClientToken:         a.clientToken,
		IsBusiness:          inst.BusinessDevice,
	}, nil
}

// proxyHealthRepoAdapter bridges instances.Repository to proxy.Repository.
type proxyHealthRepoAdapter struct {
	repo *instances.Repository
}

func (a *proxyHealthRepoAdapter) ListInstancesWithProxy(ctx context.Context) ([]proxycheck.ProxyInstance, error) {
	items, err := a.repo.ListInstancesWithProxy(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]proxycheck.ProxyInstance, len(items))
	for i, item := range items {
		result[i] = proxycheck.ProxyInstance{
			InstanceID:     item.InstanceID,
			ProxyURL:       item.ProxyURL,
			HealthStatus:   item.HealthStatus,
			HealthFailures: item.HealthFailures,
		}
	}
	return result, nil
}

func (a *proxyHealthRepoAdapter) UpdateProxyHealthStatus(ctx context.Context, id uuid.UUID, status string, failures int) error {
	return a.repo.UpdateProxyHealthStatus(ctx, id, status, failures)
}

func (a *proxyHealthRepoAdapter) InsertProxyHealthLog(ctx context.Context, entry proxycheck.HealthLog) error {
	return a.repo.InsertProxyHealthLog(ctx, instances.ProxyHealthLog{
		InstanceID:   entry.InstanceID,
		ProxyURL:     entry.ProxyURL,
		Status:       entry.Status,
		LatencyMs:    entry.LatencyMs,
		ErrorMessage: entry.ErrorMessage,
		CheckedAt:    entry.CheckedAt,
	})
}

func (a *proxyHealthRepoAdapter) CleanupProxyHealthLogs(ctx context.Context, retention time.Duration) (int64, error) {
	return a.repo.CleanupProxyHealthLogs(ctx, retention)
}

// proxyRepoAdapter bridges instances.Repository to whatsmeow.ProxyRepository.
type proxyRepoAdapter struct {
	repo *instances.Repository
}

func (a *proxyRepoAdapter) GetProxyConfig(ctx context.Context, id uuid.UUID) (*whatsmeow.ProxyConfig, error) {
	cfg, err := a.repo.GetProxyConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	return &whatsmeow.ProxyConfig{
		ProxyURL:       cfg.ProxyURL,
		Enabled:        cfg.Enabled,
		NoWebsocket:    cfg.NoWebsocket,
		OnlyLogin:      cfg.OnlyLogin,
		NoMedia:        cfg.NoMedia,
		HealthStatus:   cfg.HealthStatus,
		HealthFailures: cfg.HealthFailures,
	}, nil
}

func (a *proxyRepoAdapter) ListInstancesWithProxy(ctx context.Context) ([]whatsmeow.ProxyInstance, error) {
	items, err := a.repo.ListInstancesWithProxy(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]whatsmeow.ProxyInstance, len(items))
	for i, item := range items {
		result[i] = whatsmeow.ProxyInstance{
			InstanceID:     item.InstanceID,
			ProxyURL:       item.ProxyURL,
			HealthStatus:   item.HealthStatus,
			HealthFailures: item.HealthFailures,
		}
	}
	return result, nil
}

func (a *proxyRepoAdapter) UpdateProxyHealthStatus(ctx context.Context, id uuid.UUID, status string, failures int) error {
	return a.repo.UpdateProxyHealthStatus(ctx, id, status, failures)
}

// registrySwapperAdapter bridges *whatsmeow.ClientRegistry to proxy.RegistrySwapper.
type registrySwapperAdapter struct {
	registry *whatsmeow.ClientRegistry
}

func (a *registrySwapperAdapter) ApplyProxy(ctx context.Context, instanceID uuid.UUID, proxyURL string, noWebsocket, onlyLogin, noMedia bool) error {
	return a.registry.ApplyProxy(ctx, instanceID, proxyURL, noWebsocket, onlyLogin, noMedia)
}

func (a *registrySwapperAdapter) SwapProxy(ctx context.Context, instanceID uuid.UUID, proxyURL string, noWebsocket, onlyLogin, noMedia bool) error {
	return a.registry.SwapProxy(ctx, instanceID, proxyURL, noWebsocket, onlyLogin, noMedia)
}

// instanceProxyUpdaterAdapter bridges *instances.Repository to proxy.InstanceProxyUpdater.
type instanceProxyUpdaterAdapter struct {
	repo *instances.Repository
}

func (a *instanceProxyUpdaterAdapter) UpdateProxyURL(ctx context.Context, id uuid.UUID, proxyURL *string, enabled bool) error {
	cfg := instances.ProxyConfig{
		ProxyURL: proxyURL,
		Enabled:  enabled,
	}
	return a.repo.UpdateProxyConfig(ctx, id, cfg)
}

func (a *instanceProxyUpdaterAdapter) ClearProxyConfig(ctx context.Context, id uuid.UUID) error {
	return a.repo.ClearProxyConfig(ctx, id)
}

// mediaClientProviderAdapter provides WhatsApp clients to NATS media workers.
// The registry reference is set after registry creation via SetRegistry().
type mediaClientProviderAdapter struct {
	registry *whatsmeow.ClientRegistry
}

func (a *mediaClientProviderAdapter) GetClient(instanceID uuid.UUID) (*wameow.Client, bool) {
	if a.registry == nil {
		return nil, false
	}
	return a.registry.GetClient(instanceID)
}

// mediaTaskPublisherAdapter bridges media.NATSMediaPublisher to capture.MediaTaskPublisher.
// This allows the NATSEventWriter to publish media tasks without importing media directly.
type mediaTaskPublisherAdapter struct {
	pub *media.NATSMediaPublisher
}

func (a *mediaTaskPublisherAdapter) PublishMediaTask(ctx context.Context, info capture.MediaTaskInfo) error {
	return a.pub.PublishTask(ctx, media.MediaTask{
		InstanceID:  info.InstanceID,
		EventID:     info.EventID,
		MediaKey:    info.MediaKey,
		DirectPath:  info.DirectPath,
		MediaType:   info.MediaType,
		MimeType:    info.MimeType,
		FileLength:  info.FileLength,
		PublishedAt: time.Now(),
		Payload:     info.Payload,
	})
}

// routingLocatorAdapter bridges *whatsmeow.ClientRegistry to middleware.InstanceLocator.
type routingLocatorAdapter struct {
	registry *whatsmeow.ClientRegistry
}

func (a *routingLocatorAdapter) HasClient(instanceID uuid.UUID) bool {
	_, ok := a.registry.GetClient(instanceID)
	return ok
}

// routingOwnerAdapter bridges *instances.Repository to middleware.InstanceOwnerLookup.
type routingOwnerAdapter struct {
	repo *instances.Repository
}

func (a *routingOwnerAdapter) LookupOwner(ctx context.Context, instanceID uuid.UUID) (string, error) {
	inst, err := a.repo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, instances.ErrInstanceNotFound) {
			return "", nil
		}
		return "", err
	}
	if inst.WorkerID == nil {
		return "", nil
	}
	return *inst.WorkerID, nil
}

// resolveAdvertiseAddr determines the HTTP address this replica can be reached at.
// It resolves the hostname to an IP (for Docker Swarm / K8s overlay networks) and
// combines it with the port extracted from the listen address.
func resolveAdvertiseAddr(hostname, listenAddr string, logger *slog.Logger) string {
	_, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		port = "8080"
	}

	ips, err := net.LookupHost(hostname)
	if err != nil || len(ips) == 0 {
		logger.Warn("could not resolve hostname, using hostname directly",
			slog.String("hostname", hostname),
			slog.Any("error", err),
		)
		return "http://" + hostname + ":" + port
	}

	ip := ips[0]
	if strings.Contains(ip, ":") {
		ip = "[" + ip + "]"
	}
	return "http://" + ip + ":" + port
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
	logger.Info("starting Zé da API", slog.String("env", cfg.AppEnv))
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

	// Initialize NATS client if enabled
	var natsClient *natsclient.Client
	var natsMetrics *natsclient.NATSMetrics
	if cfg.NATS.Enabled {
		natsMetrics = natsclient.NewNATSMetrics(cfg.Prometheus.Namespace, prometheus.DefaultRegisterer)
		natsCfg := natsclient.Config{
			URL:                   cfg.NATS.URL,
			Token:                 cfg.NATS.Token,
			ConnectTimeout:        cfg.NATS.ConnectTimeout,
			ReconnectWait:         cfg.NATS.ReconnectWait,
			MaxReconnects:         cfg.NATS.MaxReconnects,
			PublishTimeout:        cfg.NATS.PublishTimeout,
			DrainTimeout:          cfg.NATS.DrainTimeout,
			StreamMessageQueue:    "MESSAGE_QUEUE",
			StreamWhatsAppEvents:  "WHATSAPP_EVENTS",
			StreamMediaProcessing: "MEDIA_PROCESSING",
			StreamDLQ:             "DLQ",
		}
		natsClient = natsclient.NewClient(natsCfg, logger, natsMetrics)
		if err := natsClient.Connect(ctx); err != nil {
			logger.Error("nats connect failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer func() {
			logger.Info("draining NATS connection")
			if err := natsClient.Drain(cfg.NATS.DrainTimeout); err != nil {
				logger.Warn("nats drain error", slog.String("error", err.Error()))
			}
		}()

		if err := natsclient.EnsureAllStreams(ctx, natsClient.JetStream(), natsCfg, logger); err != nil {
			logger.Error("nats ensure streams failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("NATS JetStream initialized", slog.String("url", cfg.NATS.URL))
	}

	// Create NATS KV bucket for media results (used to coordinate media URL injection).
	// The CompletionHandler writes results here; the dispatch worker reads them.
	var mediaResultKV *natsclient.MediaResultKV
	if cfg.NATS.Enabled && natsClient != nil {
		var kvErr error
		mediaResultKV, kvErr = natsclient.EnsureMediaResultsBucket(ctx, natsClient.JetStream())
		if kvErr != nil {
			logger.Warn("media results KV bucket creation failed (media URLs will not be injected into webhooks)",
				slog.String("error", kvErr.Error()))
		} else {
			logger.Info("NATS media results KV bucket created")
		}
	}

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

	// Inject NATS event writer if NATS is enabled
	if cfg.NATS.Enabled && natsClient != nil {
		eventPublisher := eventsnats.NewEventPublisher(natsClient, logger, metrics)
		natsWriter := capture.NewNATSEventWriter(eventPublisher, &cfg, metrics, logger)

		// Wire media task publisher so events with HasMedia also queue a media task
		mediaPublisher := media.NewNATSMediaPublisher(natsClient, logger, metrics)
		natsWriter.SetMediaTaskPublisher(&mediaTaskPublisherAdapter{pub: mediaPublisher})

		eventOrchestrator.SetEventWriter(natsWriter)
		logger.Info("NATS event writer injected into orchestrator (with media task publisher)")
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

	var dispatchCoordinator dispatch.DispatchCoordinator
	if cfg.NATS.Enabled && natsClient != nil {
		// NATS-based dispatch coordinator
		eventCfg := eventsnats.DefaultNATSEventConfig()
		eventCfg.WebhookTimeout = cfg.Events.WebhookTimeout

		natsDispatch := dispatch.NewNATSDispatchCoordinator(&dispatch.NATSDispatchCoordinatorConfig{
			NATSClient:        natsClient,
			Config:            &cfg,
			EventConfig:       eventCfg,
			TransportRegistry: transportRegistry,
			WebhookResolver:   &dispatchWebhookResolverAdapter{repo: repo, clientToken: cfg.Client.AuthToken},
			PollStore:         pollStore,
			MediaResults:      mediaResultKV,
			Metrics:           metrics,
			NATSMetrics:       natsMetrics,
			Logger:            logger,
		})
		dispatchCoordinator = natsDispatch
		logger.Info("NATS dispatch coordinator created")
	} else {
		// PostgreSQL-based dispatch coordinator
		pgDispatch := dispatch.NewCoordinator(
			&cfg,
			pgPool,
			outboxRepo,
			dlqRepo,
			transportRegistry,
			&instanceLookupAdapter{repo: repo},
			pollStore,
			metrics,
		)
		dispatchCoordinator = pgDispatch
		logger.Info("PostgreSQL dispatch coordinator created")
	}

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
			"zedaapi",
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

	// mediaClientProvider is used by NATS media workers to get WhatsApp clients.
	// It wraps the registry, which is created later. SetRegistry() is called after registry init.
	mediaClientProv := &mediaClientProviderAdapter{}

	var mediaCoordinator media.MediaCoordinatorProvider
	var natsMediaCoord *media.NATSMediaCoordinator
	if cfg.NATS.Enabled && natsClient != nil {
		var natsMediaErr error
		natsMediaCoord, natsMediaErr = media.NewNATSMediaCoordinator(ctx, &media.NATSMediaCoordinatorConfig{
			NATSClient:     natsClient,
			ClientProvider: mediaClientProv,
			Config:         &cfg,
			MediaConfig:    media.DefaultNATSMediaConfig(),
			Metrics:        metrics,
			NATSMetrics:    natsMetrics,
			Logger:         logger,
		})
		if natsMediaErr != nil {
			logger.Error("failed to create NATS media coordinator", slog.String("error", natsMediaErr.Error()))
			os.Exit(1)
		}
		mediaCoordinator = natsMediaCoord
		logger.Info("NATS media coordinator initialized")

		defer func() {
			logger.Info("stopping NATS media coordinator")
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := natsMediaCoord.Stop(stopCtx); err != nil {
				logger.Error("NATS media coordinator stop error", slog.String("error", err.Error()))
			}
			stopCancel()
			logger.Info("NATS media coordinator stopped")
		}()

		// Start media completion handler to consume media.done events.
		// The callback writes results to the NATS KV store so that the dispatch
		// worker can inject media URLs into webhook payloads before delivery.
		var completionCallback func(ctx context.Context, result media.MediaResult)
		if mediaResultKV != nil {
			completionCallback = func(ctx context.Context, result media.MediaResult) {
				if err := mediaResultKV.Put(ctx, result.EventID.String(), result.Success, result.MediaURL, result.Error); err != nil {
					logger.Error("failed to store media result in KV",
						slog.String("event_id", result.EventID.String()),
						slog.Bool("success", result.Success),
						slog.String("error", err.Error()))
				} else {
					logger.Debug("media result stored in KV",
						slog.String("event_id", result.EventID.String()),
						slog.Bool("success", result.Success),
						slog.String("media_url", result.MediaURL))
				}
			}
		}
		completionHandler := media.NewCompletionHandler(media.CompletionHandlerConfig{
			NATSClient: natsClient,
			Metrics:    metrics,
			Logger:     logger,
			Callback:   completionCallback,
		})
		if completionErr := completionHandler.Start(ctx); completionErr != nil {
			logger.Warn("media completion handler start failed (non-fatal)",
				slog.String("error", completionErr.Error()))
		} else {
			logger.Info("media completion handler started")
			defer completionHandler.Stop()
		}
	} else {
		mediaCoordinator = media.NewMediaCoordinator(
			&cfg,
			mediaRepo,
			outboxRepo,
			metrics,
		)
		logger.Info("PostgreSQL media coordinator initialized")
	}

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
	registry.SetProxyRepository(&proxyRepoAdapter{repo: repo})
	// Wire the client provider for NATS media workers now that registry is available
	mediaClientProv.registry = registry
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

	advertiseAddr := resolveAdvertiseAddr(registry.WorkerHostname(), cfg.HTTP.Addr, logger)
	logger.Info("resolved advertise address",
		slog.String("hostname", registry.WorkerHostname()),
		slog.String("advertise_addr", advertiseAddr),
	)

	workerDirectory := workers.NewRegistry(
		pgPool,
		registry.WorkerID(),
		registry.WorkerHostname(),
		cfg.AppEnv,
		workers.Config{
			HeartbeatInterval: cfg.WorkerRegistry.HeartbeatInterval,
			Expiry:            cfg.WorkerRegistry.Expiry,
			AdvertiseAddr:     advertiseAddr,
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

	// Initialize proxy health checker
	proxyMetrics := proxycheck.NewMetrics(cfg.Prometheus.Namespace, prometheus.DefaultRegisterer)
	proxyHealthChecker := proxycheck.NewHealthChecker(
		&proxyHealthRepoAdapter{repo: repo},
		proxycheck.Config{
			HealthCheckInterval:    cfg.Proxy.HealthCheckInterval,
			HealthCheckTimeout:     cfg.Proxy.HealthCheckTimeout,
			MaxConsecutiveFailures: cfg.Proxy.MaxConsecutiveFailures,
			DLQOnUnhealthy:         cfg.Proxy.DLQOnUnhealthy,
			PauseQueueOnUnhealthy:  cfg.Proxy.PauseQueueOnUnhealthy,
			LogRetentionPeriod:     cfg.Proxy.LogRetentionPeriod,
			CleanupInterval:        cfg.Proxy.CleanupInterval,
		},
		logger,
		proxyMetrics,
	)
	proxyHealthChecker.SetUnhealthyCallback(func(ctx context.Context, instanceID uuid.UUID, proxyURL string, failures int) {
		logger.Error("proxy unhealthy callback triggered",
			slog.String("component", "proxy_health_checker"),
			slog.String("instance_id", instanceID.String()),
			slog.Int("failures", failures))
	})
	proxyHealthChecker.SetRecoveredCallback(func(ctx context.Context, instanceID uuid.UUID, proxyURL string) {
		logger.Info("proxy recovered callback triggered",
			slog.String("component", "proxy_health_checker"),
			slog.String("instance_id", instanceID.String()))
	})
	proxyHealthChecker.Start(ctx)
	defer func() {
		logger.Info("stopping proxy health checker")
		proxyHealthChecker.Stop()
		logger.Info("proxy health checker stopped")
	}()

	logger.Info("proxy health checker initialized",
		slog.Duration("interval", cfg.Proxy.HealthCheckInterval),
		slog.Int("max_failures", cfg.Proxy.MaxConsecutiveFailures))

	// Initialize proxy pool system if enabled
	var autoHealer *proxycheck.AutoHealer
	var poolHandler *proxycheck.PoolHandler
	if cfg.ProxyPool.Enabled {
		poolRepo := proxycheck.NewPoolRepository(pgPool)
		poolMetrics := proxycheck.NewPoolMetrics(cfg.Prometheus.Namespace, prometheus.DefaultRegisterer)

		countryCodes := strings.Split(cfg.ProxyPool.DefaultCountryCodes, ",")
		for i := range countryCodes {
			countryCodes[i] = strings.TrimSpace(countryCodes[i])
		}

		poolCfg := proxycheck.PoolConfig{
			SyncInterval:         cfg.ProxyPool.SyncInterval,
			DefaultCountryCodes:  countryCodes,
			AssignmentRetryDelay: cfg.ProxyPool.AssignmentRetryDelay,
			MaxAssignmentRetries: cfg.ProxyPool.MaxAssignmentRetries,
		}

		registrySwapper := &registrySwapperAdapter{registry: registry}
		poolManager := proxycheck.NewPoolManager(poolRepo, registrySwapper, lockManager, poolCfg, logger, poolMetrics)
		poolManager.SetInstanceUpdater(&instanceProxyUpdaterAdapter{repo: repo})

		// Register Webshare provider if API key is configured
		if cfg.ProxyPool.WebshareAPIKey != "" {
			// Create provider record in database (idempotent)
			providerRecord, err := poolRepo.CreateProvider(ctx, proxycheck.CreateProviderRequest{
				Name:                 "webshare",
				ProviderType:         "webshare",
				Enabled:              true,
				Priority:             100,
				APIKey:               cfg.ProxyPool.WebshareAPIKey,
				APIEndpoint:          cfg.ProxyPool.WebshareEndpoint,
				MaxInstancesPerProxy: 1,
				CountryCodes:         countryCodes,
				RateLimitRPM:         240,
			})
			if err != nil {
				// Provider may already exist, try to find it
				providers, listErr := poolRepo.ListProviders(ctx)
				if listErr != nil {
					logger.Error("failed to list providers for webshare lookup", slog.String("error", listErr.Error()))
				} else {
					for _, p := range providers {
						if p.Name == "webshare" {
							providerRecord = &p
							break
						}
					}
				}
			}

			if providerRecord != nil {
				wsProvider := webshare.NewProvider(cfg.ProxyPool.WebshareAPIKey, cfg.ProxyPool.WebshareEndpoint, cfg.ProxyPool.WebsharePlanID, cfg.ProxyPool.WebshareMode, logger)
				poolManager.RegisterProvider(providerRecord.ID, wsProvider)
				logger.Info("webshare provider registered",
					slog.String("provider_id", providerRecord.ID.String()))
			}
		}

		// Wire auto-healer to existing health checker callbacks
		autoHealer = proxycheck.NewAutoHealer(poolManager, poolRepo, lockManager, logger, poolMetrics)
		proxyHealthChecker.SetUnhealthyCallback(autoHealer.OnProxyUnhealthy)
		proxyHealthChecker.SetRecoveredCallback(autoHealer.OnProxyRecovered)

		poolManager.Start(ctx)
		defer func() {
			logger.Info("stopping proxy pool manager")
			poolManager.Stop()
			logger.Info("proxy pool manager stopped")
		}()

		poolService := proxycheck.NewPoolService(poolManager, poolRepo, logger)
		poolHandler = proxycheck.NewPoolHandler(poolService, logger)

		logger.Info("proxy pool system initialized",
			slog.Duration("sync_interval", cfg.ProxyPool.SyncInterval),
			slog.Bool("webshare_enabled", cfg.ProxyPool.WebshareAPIKey != ""))
	}

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
	var messageCoordinator queue.QueueCoordinator
	var messageHandler *handlers.MessageHandler
	if cfg.MessageQueue.Enabled {
		// Initialize API echo emitter for webhook notifications (fromMe=true, fromApi=true)
		var echoEmitter *echo.Emitter
		if cfg.APIEcho.Enabled {
			echoEmitter = echo.NewEmitter(ctx, &echo.EmitterConfig{
				InstanceID:  uuid.Nil, // Shared emitter; instance ID comes from EchoRequest
				EventRouter: eventOrchestrator.GetEventRouter(),
				Metrics:     metrics,
				Enabled:     cfg.APIEcho.Enabled,
				StoreJID:    "", // Instance-specific; comes from EchoRequest
			})
			logger.Info("API echo emitter initialized for message queue")
		}

		if cfg.NATS.Enabled && natsClient != nil {
			// NATS-based message queue coordinator
			natsCfgQueue := queue.NATSConfig{
				MaxAttempts:          cfg.MessageQueue.MaxAttempts,
				InitialBackoff:       cfg.MessageQueue.InitialBackoff,
				MaxBackoff:           cfg.MessageQueue.MaxBackoff,
				BackoffMultiplier:    cfg.MessageQueue.BackoffMultiplier,
				DisconnectRetryDelay: cfg.MessageQueue.DisconnectRetryDelay,
				ProxyRetryDelay:      cfg.MessageQueue.ProxyRetryDelay,
			}

			natsCoord, natsErr := queue.NewNATSCoordinator(ctx, &queue.NATSCoordinatorConfig{
				NATSClient:     natsClient,
				ClientRegistry: &queueClientRegistryAdapter{registry: registry},
				Processor:      queue.NewWhatsAppMessageProcessor(logger, echoEmitter),
				Config:         natsCfgQueue,
				Logger:         logger,
				Metrics:        metrics,
				NATSMetrics:    natsMetrics,
				EchoEmitter:    echoEmitter,
				RedisClient:    redisClient,
			})
			if natsErr != nil {
				logger.Error("failed to initialize NATS message queue coordinator", slog.String("error", natsErr.Error()))
				os.Exit(1)
			}
			messageCoordinator = natsCoord

			logger.Info("NATS message queue coordinator initialized",
				slog.Int("max_attempts", cfg.MessageQueue.MaxAttempts))
		} else {
			// PostgreSQL-based message queue coordinator
			queueConfig := &queue.Config{
				PollInterval:         cfg.MessageQueue.PollInterval,
				BatchSize:            1, // CRITICAL: Must be 1 for FIFO ordering
				MaxAttempts:          cfg.MessageQueue.MaxAttempts,
				InitialBackoff:       cfg.MessageQueue.InitialBackoff,
				MaxBackoff:           cfg.MessageQueue.MaxBackoff,
				BackoffMultiplier:    cfg.MessageQueue.BackoffMultiplier,
				DisconnectRetryDelay: cfg.MessageQueue.DisconnectRetryDelay,
				ProxyRetryDelay:      cfg.MessageQueue.ProxyRetryDelay,
				CompletedRetention:   cfg.MessageQueue.CompletedRetention,
				FailedRetention:      cfg.MessageQueue.FailedRetention,
				WorkersPerInstance:   1, // CRITICAL: Must be 1 for FIFO ordering
			}

			coordinatorConfig := &queue.CoordinatorConfig{
				Pool:            pgPool,
				ClientRegistry:  &queueClientRegistryAdapter{registry: registry},
				Processor:       queue.NewWhatsAppMessageProcessor(logger, echoEmitter),
				Config:          queueConfig,
				Logger:          logger,
				CleanupInterval: cfg.MessageQueue.CleanupInterval,
				CleanupTimeout:  cfg.MessageQueue.CleanupTimeout,
				Metrics:         metrics,
				EchoEmitter:     echoEmitter,
			}

			pgCoord, pgErr := queue.NewCoordinator(ctx, coordinatorConfig)
			if pgErr != nil {
				logger.Error("failed to initialize message queue coordinator", slog.String("error", pgErr.Error()))
				os.Exit(1)
			}
			messageCoordinator = pgCoord

			logger.Info("PostgreSQL message queue coordinator initialized",
				slog.Duration("poll_interval", cfg.MessageQueue.PollInterval),
				slog.Int("max_attempts", cfg.MessageQueue.MaxAttempts))
		}

		// Wire proxy-queue integration: pause message processing during proxy failures/swaps
		if cfg.ProxyPool.Enabled && messageCoordinator != nil {
			proxyHealthChecker.SetQueuePauser(messageCoordinator)
			if autoHealer != nil {
				autoHealer.SetQueuePauser(messageCoordinator)
			}
			registry.SetQueuePauser(messageCoordinator)
			logger.Info("proxy-queue integration wired: message queue will pause during proxy failures and swaps",
				slog.Bool("pause_on_unhealthy", cfg.Proxy.PauseQueueOnUnhealthy))
		}

		// Wire queue coordinator for cleanup during instance reset
		if messageCoordinator != nil {
			registry.SetQueueCoordinator(messageCoordinator)
		}

		// Create message handler for HTTP endpoints
		messageHandler = handlers.NewMessageHandler(
			messageCoordinator,
			instanceService,
			contactsService,
			chatsService,
			registry,
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

	if natsClient != nil {
		healthHandler.SetNATSClient(natsClient)
	}

	healthHandler.SetMetrics(func(component, status string) {
		metrics.HealthChecks.WithLabelValues(component, status).Inc()
	})

	// Instance-aware routing middleware: proxies requests to the replica
	// that owns the WhatsApp client for the target instance.
	routingMW := ourmiddleware.NewRoutingMiddleware(
		&routingLocatorAdapter{registry: registry},
		workerDirectory,
		&routingOwnerAdapter{repo: repo},
		registry.WorkerID(),
		logger,
	)
	defer routingMW.Close()

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
		PoolHandler:        poolHandler,
		PartnerToken:       cfg.Partner.AuthToken,
		DocsConfig:         docs.Config{BaseURL: cfg.HTTP.BaseURL},
		RoutingMiddleware:  routingMW.Handler,
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
