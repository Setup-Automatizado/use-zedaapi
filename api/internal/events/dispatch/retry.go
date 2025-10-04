package dispatch

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

func CalculateNextAttempt(attemptCount int, retryDelays []time.Duration) time.Time {
	if attemptCount >= len(retryDelays) {
		return time.Now().Add(retryDelays[len(retryDelays)-1])
	}

	delay := retryDelays[attemptCount]
	return time.Now().Add(delay)
}

func ShouldRetry(attemptCount, maxAttempts int, err error) bool {
	if err == nil {
		return false
	}

	if attemptCount >= maxAttempts {
		return false
	}

	return ClassifyError(err) == ErrorTypeRetryable
}

type ErrorType int

const (
	ErrorTypeRetryable ErrorType = iota
	ErrorTypePermanent
)

func ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypePermanent
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrorTypeRetryable
		}
		return ErrorTypeRetryable
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ErrorTypeRetryable
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return ErrorTypePermanent
	}

	if isConnectionError(err) {
		return ErrorTypeRetryable
	}

	if isTLSError(err) {
		return ErrorTypePermanent
	}

	return ErrorTypeRetryable
}

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

func ClassifyHTTPStatus(statusCode int) ErrorType {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return ErrorTypePermanent

	case statusCode == 408:
		return ErrorTypeRetryable

	case statusCode == 429:
		return ErrorTypeRetryable

	case statusCode >= 400 && statusCode < 500:
		return ErrorTypePermanent

	case statusCode >= 500 && statusCode < 600:
		return ErrorTypeRetryable

	default:
		return ErrorTypePermanent
	}
}

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

func GetRetrySchedule(retryDelays []time.Duration) []string {
	schedule := make([]string, len(retryDelays))
	for i, delay := range retryDelays {
		schedule[i] = fmt.Sprintf("Attempt %d: %s", i+1, FormatRetryDelay(delay))
	}
	return schedule
}

func CalculateTotalRetryDuration(retryDelays []time.Duration) time.Duration {
	var total time.Duration
	for _, delay := range retryDelays {
		total += delay
	}
	return total
}
