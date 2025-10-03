package transport

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow/api/internal/events/transport/http"
)

// httpTransportWrapper wraps http.HTTPTransport to implement Transport interface
// This avoids import cycle
type httpTransportWrapper struct {
	impl *http.HTTPTransport
}

func (w *httpTransportWrapper) Deliver(ctx context.Context, request *DeliveryRequest) (*DeliveryResult, error) {
	httpReq := &http.DeliveryRequest{
		Endpoint:    request.Endpoint,
		Payload:     request.Payload,
		Headers:     request.Headers,
		EventID:     request.EventID,
		EventType:   request.EventType,
		InstanceID:  request.InstanceID,
		Attempt:     request.Attempt,
		MaxAttempts: request.MaxAttempts,
	}

	httpRes, err := w.impl.Deliver(ctx, httpReq)
	if err != nil && httpRes == nil {
		return nil, err
	}

	result := &DeliveryResult{
		Success:         httpRes.Success,
		StatusCode:      httpRes.StatusCode,
		Response:        httpRes.Response,
		ResponseHeaders: httpRes.ResponseHeaders,
		Duration:        httpRes.Duration,
		Retryable:       httpRes.Retryable,
		ErrorMessage:    httpRes.ErrorMessage,
		ErrorType:       httpRes.ErrorType,
	}

	return result, err
}

func (w *httpTransportWrapper) Name() string {
	return "http"
}

func (w *httpTransportWrapper) Close() error {
	return w.impl.Close()
}

// Registry manages transport instances and provides factory methods
type Registry struct {
	mu         sync.RWMutex
	transports map[TransportType]Transport
	httpConfig *http.Config
}

// NewRegistry creates a new transport registry
func NewRegistry(httpConfig *http.Config) *Registry {
	return &Registry{
		transports: make(map[TransportType]Transport),
		httpConfig: httpConfig,
	}
}

// GetTransport returns a transport instance for the given type
// Creates the transport if it doesn't exist (singleton pattern)
func (r *Registry) GetTransport(transportType TransportType) (Transport, error) {
	r.mu.RLock()
	if transport, exists := r.transports[transportType]; exists {
		r.mu.RUnlock()
		return transport, nil
	}
	r.mu.RUnlock()

	// Create transport if it doesn't exist
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if transport, exists := r.transports[transportType]; exists {
		return transport, nil
	}

	transport, err := r.createTransport(transportType)
	if err != nil {
		return nil, err
	}

	r.transports[transportType] = transport
	return transport, nil
}

// createTransport creates a new transport instance based on type
func (r *Registry) createTransport(transportType TransportType) (Transport, error) {
	switch transportType {
	case TransportTypeHTTP:
		if r.httpConfig == nil {
			return nil, fmt.Errorf("HTTP transport config not provided")
		}
		return &httpTransportWrapper{
			impl: http.NewHTTPTransport(r.httpConfig),
		}, nil

	case TransportTypeRabbitMQ:
		return nil, fmt.Errorf("RabbitMQ transport not implemented yet")

	case TransportTypeSQS:
		return nil, fmt.Errorf("SQS transport not implemented yet")

	case TransportTypeNATS:
		return nil, fmt.Errorf("NATS transport not implemented yet")

	case TransportTypeKafka:
		return nil, fmt.Errorf("Kafka transport not implemented yet")

	default:
		return nil, fmt.Errorf("unknown transport type: %s", transportType)
	}
}

// Close closes all registered transports
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var closeErrors []error

	for transportType, transport := range r.transports {
		if err := transport.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close %s transport: %w", transportType, err))
		}
	}

	// Clear the map
	r.transports = make(map[TransportType]Transport)

	if len(closeErrors) > 0 {
		return fmt.Errorf("errors closing transports: %v", closeErrors)
	}

	return nil
}

// RegisterTransport allows manual registration of a transport instance
// Useful for testing or custom transport implementations
func (r *Registry) RegisterTransport(transportType TransportType, transport Transport) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transports[transportType] = transport
}
