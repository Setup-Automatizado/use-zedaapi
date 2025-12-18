package sentry

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

var sentryEnabled atomic.Bool

func Init(dsn, environment, release string) (*sentryhttp.Handler, error) {
	if dsn == "" {
		sentryEnabled.Store(false)
		return nil, nil
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: environment,
		Release:     release,
	}); err != nil {
		sentryEnabled.Store(false)
		return nil, err
	}
	sentryEnabled.Store(true)
	return sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: true,
		Timeout:         5 * time.Second,
	}), nil
}

func Enabled() bool {
	return sentryEnabled.Load()
}

func CaptureLifecycleEvent(phase string, tags map[string]string, extras map[string]any) {
	if !Enabled() {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("event", "lifecycle")
		scope.SetTag("lifecycle_phase", phase)
		scope.SetLevel(sentry.LevelInfo)
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		for k, v := range extras {
			scope.SetExtra(k, v)
		}
		sentry.CaptureMessage(fmt.Sprintf("api.lifecycle.%s", phase))
	})
}

func Flush(timeout time.Duration) {
	if !Enabled() {
		return
	}
	sentry.Flush(timeout)
}
