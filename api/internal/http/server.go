package http

import (
	"context"
	"net/http"
	"time"

	"log/slog"
)

// Server wraps net/http.Server with graceful shutdown helpers.
type Server struct {
	srv *http.Server
	log *slog.Logger
}

// NewServer builds a configured HTTP server.
func NewServer(handler http.Handler, addr string, readHeaderTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxHeaderBytes int, log *slog.Logger) *Server {
	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
	return &Server{srv: s, log: log}
}

// Run starts the HTTP server and blocks until the context is cancelled or the server exits.
func (s *Server) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)
	go func() {
		s.log.Info("http server starting", slog.String("addr", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.log.Error("http shutdown", slog.String("error", err.Error()))
		}
		return nil
	case err := <-serverErr:
		return err
	}
}
