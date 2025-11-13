package http

import (
	"net/http"
	"time"

	"log/slog"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.mau.fi/whatsmeow/api/docs"
	"go.mau.fi/whatsmeow/api/internal/http/handlers"
	ourMiddleware "go.mau.fi/whatsmeow/api/internal/http/middleware"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type RouterDeps struct {
	Logger             *slog.Logger
	Metrics            *observability.Metrics
	SentryHandler      *sentryhttp.Handler
	InstanceHandler    *handlers.InstanceHandler
	PartnerHandler     *handlers.PartnerHandler
	HealthHandler      *handlers.HealthHandler
	MediaHandler       *handlers.MediaHandler
	MessageHandler     *handlers.MessageHandler
	GroupsHandler      *handlers.GroupsHandler
	CommunitiesHandler *handlers.CommunitiesHandler
	NewslettersHandler *handlers.NewslettersHandler
	PartnerToken       string
	DocsConfig         docs.Config
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	if deps.Logger != nil {
		r.Use(ourMiddleware.RequestLogger(deps.Logger))
	}
	if deps.Metrics != nil {
		r.Use(ourMiddleware.PrometheusMiddleware(deps.Metrics))
	}
	if deps.SentryHandler != nil {
		r.Use(deps.SentryHandler.Handle)
	}

	if deps.HealthHandler != nil {
		r.Get("/health", deps.HealthHandler.Health)
		r.Get("/ready", deps.HealthHandler.Ready)
	}

	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	r.Mount("/debug", chiMiddleware.Profiler())

	r.Route("/docs", func(dr chi.Router) {
		dr.Get("/", func(w http.ResponseWriter, req *http.Request) {
			docs.UIHandler().ServeHTTP(w, req)
		})
		dr.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) {
			docs.YAMLHandler(deps.DocsConfig).ServeHTTP(w, req)
		})
		dr.Get("/openapi.json", func(w http.ResponseWriter, req *http.Request) {
			docs.JSONHandler(deps.DocsConfig).ServeHTTP(w, req)
		})
	})

	// Inject MessageHandler into InstanceHandler to avoid route conflicts
	// Both handlers need to register routes under /instances/{instanceId}/token/{token}
	if deps.InstanceHandler != nil {
		if deps.MessageHandler != nil {
			deps.InstanceHandler.SetMessageHandler(deps.MessageHandler)
		}
		if deps.GroupsHandler != nil {
			deps.InstanceHandler.SetGroupsHandler(deps.GroupsHandler)
		}
		if deps.CommunitiesHandler != nil {
			deps.InstanceHandler.SetCommunitiesHandler(deps.CommunitiesHandler)
		}
		if deps.NewslettersHandler != nil {
			deps.InstanceHandler.SetNewslettersHandler(deps.NewslettersHandler)
		}
	}

	if deps.InstanceHandler != nil {
		deps.InstanceHandler.Register(r)
	}

	if deps.PartnerHandler != nil {
		r.Group(func(pr chi.Router) {
			pr.Use(ourMiddleware.PartnerAuth(deps.PartnerToken))
			deps.PartnerHandler.Register(pr)
		})
	}

	if deps.MediaHandler != nil {
		r.Route("/media", func(mr chi.Router) {
			mr.Get("/{instance_id}/*", deps.MediaHandler.ServeMedia)
			mr.Get("/stats", deps.MediaHandler.GetStats)
		})
	}

	// MessageHandler routes are now registered via InstanceHandler
	// No need to call Register() here to avoid duplicate route panic

	return r
}
