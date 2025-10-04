package transport

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTransportFailed  = errors.New("transport delivery failed")
	ErrTransportTimeout = errors.New("transport delivery timeout")
	ErrInvalidEndpoint  = errors.New("invalid endpoint")
	ErrPermanentFailure = errors.New("permanent failure")
)

type Transport interface {
	Deliver(ctx context.Context, request *DeliveryRequest) (*DeliveryResult, error)
	Name() string
	Close() error
}

type DeliveryRequest struct {
	Endpoint    string
	Payload     []byte
	Headers     map[string]string
	EventID     string
	EventType   string
	InstanceID  string
	Attempt     int
	MaxAttempts int
}

type DeliveryResult struct {
	Success         bool
	StatusCode      int
	Response        []byte
	ResponseHeaders map[string][]string
	Duration        time.Duration
	Retryable       bool
	ErrorMessage    string
	ErrorType       string
}

type TransportType string

const (
	TransportTypeHTTP     TransportType = "http"
	TransportTypeRabbitMQ TransportType = "rabbitmq"
	TransportTypeSQS      TransportType = "sqs"
	TransportTypeNATS     TransportType = "nats"
	TransportTypeKafka    TransportType = "kafka"
)

const (
	ErrorTypeTimeout       = "timeout"
	ErrorTypeConnection    = "connection"
	ErrorTypeServer        = "server"
	ErrorTypeClient        = "client"
	ErrorTypeSerialization = "serialization"
	ErrorTypeUnknown       = "unknown"
)

func IsRetryableError(result *DeliveryResult) bool {
	if result == nil {
		return false
	}

	return result.Retryable
}

func ClassifyHTTPStatus(statusCode int) (retryable bool, errorType string) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return false, ""

	case statusCode >= 400 && statusCode < 500:
		if statusCode == 408 || statusCode == 429 {
			return true, ErrorTypeTimeout
		}
		return false, ErrorTypeClient

	case statusCode >= 500 && statusCode < 600:
		return true, ErrorTypeServer

	default:
		return false, ErrorTypeUnknown
	}
}
