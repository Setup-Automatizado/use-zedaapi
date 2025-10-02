package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go.mau.fi/whatsmeow/api/internal/observability"
)

func PrometheusMiddleware(metrics *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if metrics == nil {
				next.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(rw, r)
			status := rw.Status()
			if status == 0 {
				status = 200
			}
			routePattern := chi.RouteContext(r.Context())
			path := r.URL.Path
			if routePattern != nil {
				if rp := routePattern.RoutePattern(); rp != "" {
					path = rp
				}
			}
			labels := []string{r.Method, path, strconv.Itoa(status)}
			metrics.HTTPRequests.WithLabelValues(labels...).Inc()
			metrics.HTTPDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
		})
	}
}
