package observability

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

type AsyncContextOptions struct {
	Logger     *slog.Logger
	Component  string
	Worker     string
	InstanceID string
	Extra      []slog.Attr
}

func AsyncContext(opts AsyncContextOptions) context.Context {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	attrs := make([]any, 0, 3+len(opts.Extra))
	if opts.Component != "" {
		attrs = append(attrs, slog.String("component", opts.Component))
	}
	if opts.Worker != "" {
		attrs = append(attrs, slog.String("worker", opts.Worker))
	}
	if opts.InstanceID != "" {
		attrs = append(attrs, slog.String("instance_id", opts.InstanceID))
	}
	if len(opts.Extra) > 0 {
		for _, attr := range opts.Extra {
			attrs = append(attrs, attr)
		}
	}
	return logging.WithLogger(context.Background(), logger.With(attrs...))
}

func CaptureWorkerException(ctx context.Context, component, worker, instanceID string, err error) {
	if err == nil {
		return
	}
	if hub := sentry.CurrentHub(); hub == nil || hub.Client() == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		if component != "" {
			scope.SetTag("component", component)
		}
		if worker != "" {
			scope.SetTag("worker", worker)
		}
		if instanceID != "" {
			scope.SetTag("instance_id", instanceID)
		}
		scope.SetContext("worker", map[string]any{
			"component":   component,
			"worker":      worker,
			"instance_id": instanceID,
		})
		sentry.CaptureException(err)
	})
}
