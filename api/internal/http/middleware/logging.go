package middleware

import (
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// RequestLogger logs basic request/response information using slog and injects a contextual logger.
func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	if base == nil {
		base = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqLogger := base
			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				reqLogger = reqLogger.With(slog.String("request_id", reqID))
			}
			if host := r.Host; host != "" {
				reqLogger = reqLogger.With(slog.String("host", host))
			}
			if remoteAddr := strings.TrimSpace(r.RemoteAddr); remoteAddr != "" {
				reqLogger = reqLogger.With(slog.String("remote_addr", remoteAddr))
			}

			ctx := logging.WithLogger(r.Context(), reqLogger)

			rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(start)
			reqLogger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.Status()),
				slog.Int("bytes", rw.BytesWritten()),
				slog.Duration("duration", duration),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}
