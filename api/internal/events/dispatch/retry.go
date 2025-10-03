package dispatch

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// CalculateNextAttempt calculates the next retry attempt time using exponential backoff
func CalculateNextAttempt(attemptCount int, retryDelays []time.Duration) time.Time {
	if attemptCount >= len(retryDelays) {
		// Use last delay if we exceed configured delays
		return time.Now().Add(retryDelays[len(retryDelays)-1])
	}

	delay := retryDelays[attemptCount]
	return time.Now().Add(delay)
}

// ShouldRetry determines if an error is retryable
func ShouldRetry(attemptCount, maxAttempts int, err error) bool {
	if err == nil {
		return false
	}

	if attemptCount >= maxAttempts {
		return false
	}

	return ClassifyError(err) == ErrorTypeRetryable
}

// ErrorType represents the classification of an error
type ErrorType int

const (
	// ErrorTypeRetryable indicates a temporary error that should be retried
	ErrorTypeRetryable ErrorType = iota
	// ErrorTypePermanent indicates a permanent error that should not be retried
	ErrorTypePermanent
)

// ClassifyError classifies an error as retryable or permanent
func ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypePermanent // No error, shouldn't retry
	}

	// Network errors are typically retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrorTypeRetryable // Timeout
		}
		return ErrorTypeRetryable // Network error
	}

	// DNS errors are retryable
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ErrorTypeRetryable
	}

	// URL parse errors are permanent
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return ErrorTypePermanent
	}

	// Connection errors are retryable
	if isConnectionError(err) {
		return ErrorTypeRetryable
	}

	// TLS errors are permanent
	if isTLSError(err) {
		return ErrorTypePermanent
	}

	// Default to retryable for unknown errors (conservative approach)
	return ErrorTypeRetryable
}

// isConnectionError checks if error is a connection-related error
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no route to host",
		"network is unreachable",
		"connection timed out",
		"i/o timeout",
		"EOF",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}

	return false
}

// isTLSError checks if error is a TLS/certificate error
func isTLSError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	tlsErrors := []string{
		"tls",
		"certificate",
		"x509",
		"handshake failure",
		"bad certificate",
		"unknown authority",
	}

	for _, tlsErr := range tlsErrors {
		if strings.Contains(errStr, tlsErr) {
			return true
		}
	}

	return false
}

// ClassifyHTTPStatus classifies an HTTP status code as retryable or permanent
func ClassifyHTTPStatus(statusCode int) ErrorType {
	switch {
	case statusCode >= 200 && statusCode < 300:
		// Success - no retry needed
		return ErrorTypePermanent

	case statusCode == 408: // Request Timeout
		return ErrorTypeRetryable

	case statusCode == 429: // Too Many Requests
		return ErrorTypeRetryable

	case statusCode >= 400 && statusCode < 500:
		// Client errors (4xx) are typically permanent
		// The request is malformed or unauthorized
		return ErrorTypePermanent

	case statusCode >= 500 && statusCode < 600:
		// Server errors (5xx) are typically retryable
		// The server is temporarily unavailable
		return ErrorTypeRetryable

	default:
		// Unknown status codes - default to permanent
		return ErrorTypePermanent
	}
}

// FormatRetryDelay formats retry delay for human-readable output
func FormatRetryDelay(delay time.Duration) string {
	if delay < time.Second {
		return fmt.Sprintf("%dms", delay.Milliseconds())
	}
	if delay < time.Minute {
		return fmt.Sprintf("%ds", int(delay.Seconds()))
	}
	if delay < time.Hour {
		return fmt.Sprintf("%dm", int(delay.Minutes()))
	}
	return fmt.Sprintf("%dh", int(delay.Hours()))
}

// GetRetrySchedule returns a human-readable retry schedule
func GetRetrySchedule(retryDelays []time.Duration) []string {
	schedule := make([]string, len(retryDelays))
	for i, delay := range retryDelays {
		schedule[i] = fmt.Sprintf("Attempt %d: %s", i+1, FormatRetryDelay(delay))
	}
	return schedule
}

// CalculateTotalRetryDuration calculates the total time for all retry attempts
func CalculateTotalRetryDuration(retryDelays []time.Duration) time.Duration {
	var total time.Duration
	for _, delay := range retryDelays {
		total += delay
	}
	return total
}
