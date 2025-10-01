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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

	lockManager := locks.NewRedisManager(redisClient)

	repo := instances.NewRepository(pgPool)

	pairCallback := func(ctx context.Context, id uuid.UUID, jid string) error {
		return repo.UpdateStoreJID(ctx, id, &jid)
	}
	resetCallback := func(ctx context.Context, id uuid.UUID, _ string) error {
		return repo.UpdateStoreJID(ctx, id, nil)
	}

	registry, err := whatsmeow.NewClientRegistry(ctx, cfg.WhatsmeowStore.DSN, cfg.WhatsmeowStore.LogLevel, lockManager, logger, pairCallback, resetCallback)
	if err != nil {
		logger.Error("whatsmeow registry", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer registry.Close()

	instanceService := instances.NewService(repo, registry, logger)
	reconcileCtx, reconcileCancel := context.WithTimeout(ctx, 30*time.Second)
	cleaned, reconcileErr := instanceService.ReconcileDetachedStores(reconcileCtx)
	reconcileCancel()
	if reconcileErr != nil {
		logger.Error("reconcile stores", slog.String("error", reconcileErr.Error()))
	} else if len(cleaned) > 0 {
		logger.Info("reconciled stores", slog.Int("count", len(cleaned)))
	}
	instanceHandler := handlers.NewInstanceHandler(instanceService, logger)
	partnerHandler := handlers.NewPartnerHandler(instanceService, logger)

	router := apihandler.NewRouter(apihandler.RouterDeps{
		Logger:          logger,
		Metrics:         metrics,
		SentryHandler:   sentryHandler,
		InstanceHandler: instanceHandler,
		PartnerHandler:  partnerHandler,
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

	if cfg.Sentry.DSN != "" {
		sentry.Flush(5 * time.Second)
	}

	logger.Info("shutdown complete")
}
