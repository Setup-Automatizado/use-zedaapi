package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// Init configures the global Sentry SDK and returns an HTTP handler middleware.
func Init(dsn, environment, release string) (*sentryhttp.Handler, error) {
	if dsn == "" {
		return nil, nil
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: environment,
		Release:     release,
	}); err != nil {
		return nil, err
	}
	return sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: true,
		Timeout:         5 * time.Second,
	}), nil
}
