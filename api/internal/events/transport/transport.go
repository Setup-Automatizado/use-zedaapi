package transport

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrTransportFailed indicates a transport delivery failure
	ErrTransportFailed = errors.New("transport delivery failed")

	// ErrTransportTimeout indicates a timeout during delivery
	ErrTransportTimeout = errors.New("transport delivery timeout")

	// ErrInvalidEndpoint indicates an invalid endpoint configuration
	ErrInvalidEndpoint = errors.New("invalid endpoint")

	// ErrPermanentFailure indicates a non-retryable failure (4xx errors)
	ErrPermanentFailure = errors.New("permanent failure")
)

// Transport defines the interface for event delivery mechanisms
type Transport interface {
	// Deliver sends an event to the configured endpoint
	// Returns DeliveryResult with success/failure details
	Deliver(ctx context.Context, request *DeliveryRequest) (*DeliveryResult, error)

	// Name returns the transport type name
	Name() string

	// Close cleans up transport resources
	Close() error
}

// DeliveryRequest contains all information needed for event delivery
type DeliveryRequest struct {
	// Endpoint is the destination URL or address
	Endpoint string

	// Payload is the event data to deliver (JSON)
	Payload []byte

	// Headers are custom HTTP headers (for HTTP transport)
	Headers map[string]string

	// EventID for tracking and logging
	EventID string

	// EventType for routing and classification
	EventType string

	// InstanceID for context and logging
	InstanceID string

	// Attempt is the current delivery attempt number (1-based)
	Attempt int

	// MaxAttempts is the maximum number of attempts allowed
	MaxAttempts int
}

// DeliveryResult contains the outcome of a delivery attempt
type DeliveryResult struct {
	// Success indicates if delivery was successful
	Success bool

	// StatusCode is the HTTP status code (for HTTP transport)
	StatusCode int

	// Response is the raw response body
	Response []byte

	// ResponseHeaders contains response headers (for HTTP transport)
	ResponseHeaders map[string][]string

	// Duration is the total delivery time
	Duration time.Duration

	// Retryable indicates if the failure can be retried
	// false for 4xx errors (client errors), true for 5xx/timeouts
	Retryable bool

	// ErrorMessage contains error details if delivery failed
	ErrorMessage string

	// ErrorType classifies the error (timeout, connection, server, client)
	ErrorType string
}

// TransportType represents the type of transport mechanism
type TransportType string

const (
	// TransportTypeHTTP represents HTTP/HTTPS webhook delivery
	TransportTypeHTTP TransportType = "http"

	// TransportTypeRabbitMQ represents RabbitMQ message delivery (future)
	TransportTypeRabbitMQ TransportType = "rabbitmq"

	// TransportTypeSQS represents AWS SQS message delivery (future)
	TransportTypeSQS TransportType = "sqs"

	// TransportTypeNATS represents NATS message delivery (future)
	TransportTypeNATS TransportType = "nats"

	// TransportTypeKafka represents Kafka message delivery (future)
	TransportTypeKafka TransportType = "kafka"
)

// ErrorType constants for error classification
const (
	ErrorTypeTimeout    = "timeout"
	ErrorTypeConnection = "connection"
	ErrorTypeServer     = "server"     // 5xx errors
	ErrorTypeClient     = "client"     // 4xx errors
	ErrorTypeSerialization = "serialization"
	ErrorTypeUnknown    = "unknown"
)

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(result *DeliveryResult) bool {
	if result == nil {
		return false
	}

	// Explicit retryable flag takes precedence
	return result.Retryable
}

// ClassifyHTTPStatus classifies HTTP status codes for retry decisions
func ClassifyHTTPStatus(statusCode int) (retryable bool, errorType string) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		// Success
		return false, ""

	case statusCode >= 400 && statusCode < 500:
		// Client errors (4xx) - permanent failures, don't retry
		// Exception: 408 Request Timeout, 429 Too Many Requests are retryable
		if statusCode == 408 || statusCode == 429 {
			return true, ErrorTypeTimeout
		}
		return false, ErrorTypeClient

	case statusCode >= 500 && statusCode < 600:
		// Server errors (5xx) - temporary failures, retry
		return true, ErrorTypeServer

	default:
		// Unknown status codes - don't retry
		return false, ErrorTypeUnknown
	}
}
