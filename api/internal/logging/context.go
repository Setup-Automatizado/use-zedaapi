package logging

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerKey contextKey = "logging.logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) (*slog.Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok || logger == nil {
		return nil, false
	}
	return logger, true
}

func ContextLogger(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if logger, ok := FromContext(ctx); ok {
		return logger
	}
	if fallback != nil {
		return fallback
	}
	return slog.Default()
}

func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	logger, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}
	return WithLogger(ctx, logger.With(args...))
}
