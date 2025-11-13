package handlers

import (
	"github.com/getsentry/sentry-go"
)

func captureHandlerError(feature, operation, instanceID string, err error) {
	if err == nil {
		return
	}
	hub := sentry.CurrentHub()
	if hub == nil || hub.Client() == nil {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("feature", feature)
		scope.SetTag("operation", operation)
		if instanceID != "" {
			scope.SetTag("instance_id", instanceID)
		}
		scope.SetLevel(sentry.LevelError)
		sentry.CaptureException(err)
	})
}
