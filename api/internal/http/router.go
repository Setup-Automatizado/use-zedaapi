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

// RouterDeps collects dependencies for building the HTTP router.
type RouterDeps struct {
	Logger          *slog.Logger
	Metrics         *observability.Metrics
	SentryHandler   *sentryhttp.Handler
	InstanceHandler *handlers.InstanceHandler
	PartnerHandler  *handlers.PartnerHandler
	PartnerToken    string
}

// NewRouter wires routes and middlewares for the API service.
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

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	r.Mount("/debug", chiMiddleware.Profiler())

	r.Route("/docs", func(dr chi.Router) {
		dr.Get("/", func(w http.ResponseWriter, req *http.Request) {
			docs.UIHandler().ServeHTTP(w, req)
		})
		dr.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) {
			docs.YAMLHandler().ServeHTTP(w, req)
		})
		dr.Get("/openapi.json", func(w http.ResponseWriter, req *http.Request) {
			docs.JSONHandler().ServeHTTP(w, req)
		})
	})

	if deps.InstanceHandler != nil {
		deps.InstanceHandler.Register(r)
	}

	if deps.PartnerHandler != nil {
		r.Group(func(pr chi.Router) {
			pr.Use(ourMiddleware.PartnerAuth(deps.PartnerToken))
			deps.PartnerHandler.Register(pr)
		})
	}

	return r
}
