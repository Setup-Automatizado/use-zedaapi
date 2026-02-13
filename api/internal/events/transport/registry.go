package transport

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow/api/internal/events/transport/http"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

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

type Registry struct {
	mu         sync.RWMutex
	transports map[TransportType]Transport
	httpConfig *http.Config
	metrics    *observability.Metrics
}

func NewRegistry(httpConfig *http.Config, metrics *observability.Metrics) *Registry {
	return &Registry{
		transports: make(map[TransportType]Transport),
		httpConfig: httpConfig,
		metrics:    metrics,
	}
}

func (r *Registry) GetTransport(transportType TransportType) (Transport, error) {
	r.mu.RLock()
	if transport, exists := r.transports[transportType]; exists {
		r.mu.RUnlock()
		return transport, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *Registry) createTransport(transportType TransportType) (Transport, error) {
	switch transportType {
	case TransportTypeHTTP:
		if r.httpConfig == nil {
			return nil, fmt.Errorf("HTTP transport config not provided")
		}
		return &httpTransportWrapper{
			impl: http.NewHTTPTransportWithMetrics(r.httpConfig, r.metrics),
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

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var closeErrors []error

	for transportType, transport := range r.transports {
		if err := transport.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close %s transport: %w", transportType, err))
		}
	}

	r.transports = make(map[TransportType]Transport)

	if len(closeErrors) > 0 {
		return fmt.Errorf("errors closing transports: %v", closeErrors)
	}

	return nil
}

func (r *Registry) RegisterTransport(transportType TransportType, transport Transport) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transports[transportType] = transport
}
