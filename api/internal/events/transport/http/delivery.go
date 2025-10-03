package http

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// DeliveryRequest contains all information needed for event delivery
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

// DeliveryResult contains the outcome of a delivery attempt
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

// Error types
const (
	ErrorTypeTimeout       = "timeout"
	ErrorTypeConnection    = "connection"
	ErrorTypeServer        = "server"
	ErrorTypeClient        = "client"
	ErrorTypeSerialization = "serialization"
	ErrorTypeUnknown       = "unknown"
)

var (
	ErrInvalidEndpoint = fmt.Errorf("invalid endpoint")
)

// ResponseHandler processes HTTP responses
type ResponseHandler struct {
	maxResponseSize int64
}

// NewResponseHandler creates a response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{
		maxResponseSize: 1024 * 1024, // 1MB
	}
}

// HandleHTTPResponse processes an HTTP response
func (h *ResponseHandler) HandleHTTPResponse(resp *http.Response, duration time.Duration, requestErr error) *DeliveryResult {
	result := &DeliveryResult{
		Duration: duration,
	}

	if requestErr != nil {
		result.Success = false
		result.ErrorMessage = requestErr.Error()
		result.ErrorType, result.Retryable = h.classifyRequestError(requestErr)
		return result
	}

	if resp == nil {
		result.Success = false
		result.ErrorMessage = "nil response received"
		result.ErrorType = ErrorTypeUnknown
		result.Retryable = false
		return result
	}

	result.StatusCode = resp.StatusCode
	result.ResponseHeaders = resp.Header

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Classify based on status code
	result.Retryable, result.ErrorType = classifyHTTPStatus(resp.StatusCode)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.ErrorType = ""
		result.Retryable = false
	} else {
		result.Success = false
		if result.ErrorMessage == "" {
			result.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
	}

	return result
}

func (h *ResponseHandler) classifyRequestError(err error) (errorType string, retryable bool) {
	if err == nil {
		return "", false
	}

	type timeout interface {
		Timeout() bool
	}

	if e, ok := err.(timeout); ok && e.Timeout() {
		return ErrorTypeTimeout, true
	}

	return ErrorTypeUnknown, false
}

func classifyHTTPStatus(statusCode int) (retryable bool, errorType string) {
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

// HTTPTransport implements the Transport interface for HTTP webhooks
type HTTPTransport struct {
	client          *http.Client
	config          *Config
	responseHandler *ResponseHandler
}

// NewHTTPTransport creates a new HTTP transport instance
func NewHTTPTransport(config *Config) *HTTPTransport {
	if config == nil {
		config = DefaultConfig()
	}

	return &HTTPTransport{
		client:          NewHTTPClient(config),
		config:          config,
		responseHandler: NewResponseHandler(),
	}
}

// Deliver sends an HTTP POST request with the event payload
func (t *HTTPTransport) Deliver(ctx context.Context, request *DeliveryRequest) (*DeliveryResult, error) {
	logger := logging.ContextLogger(ctx, nil)

	// Validate request
	if err := t.validateRequest(request); err != nil {
		return &DeliveryResult{
			Success:      false,
			Retryable:    false,
			ErrorMessage: err.Error(),
			ErrorType:    ErrorTypeClient,
		}, err
	}

	// Prepare HTTP request
	httpReq, err := t.prepareRequest(ctx, request)
	if err != nil {
		logger.Error("failed to prepare HTTP request",
			slog.String("error", err.Error()),
			slog.String("event_id", request.EventID),
			slog.String("endpoint", request.Endpoint))

		return &DeliveryResult{
			Success:      false,
			Retryable:    false,
			ErrorMessage: fmt.Sprintf("failed to prepare request: %v", err),
			ErrorType:    ErrorTypeSerialization,
		}, err
	}

	// Perform delivery with retry
	result := t.deliverWithRetry(ctx, httpReq, request, logger)

	// Log result
	t.logDeliveryResult(logger, request, result)

	return result, nil
}

// deliverWithRetry performs HTTP delivery with internal retry logic
func (t *HTTPTransport) deliverWithRetry(ctx context.Context, httpReq *http.Request, request *DeliveryRequest, logger *slog.Logger) *DeliveryResult {
	var lastResult *DeliveryResult

	maxAttempts := t.config.MaxRetries
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			return &DeliveryResult{
				Success:      false,
				Retryable:    false,
				ErrorMessage: fmt.Sprintf("context cancelled: %v", ctx.Err()),
				ErrorType:    ErrorTypeTimeout,
			}
		}

		// Perform single delivery attempt
		start := time.Now()
		resp, err := t.client.Do(httpReq)
		duration := time.Since(start)

		// Handle response
		lastResult = t.responseHandler.HandleHTTPResponse(resp, duration, err)

		// Success - return immediately
		if lastResult.Success {
			logger.Debug("webhook delivered successfully",
				slog.String("event_id", request.EventID),
				slog.Int("attempt", attempt),
				slog.Int("status_code", lastResult.StatusCode),
				slog.Duration("duration", duration))
			return lastResult
		}

		// Permanent failure - don't retry
		if !lastResult.Retryable {
			logger.Warn("webhook delivery failed (permanent)",
				slog.String("event_id", request.EventID),
				slog.Int("attempt", attempt),
				slog.String("error", lastResult.ErrorMessage),
				slog.String("error_type", lastResult.ErrorType))
			return lastResult
		}

		// Retry logic - only if not last attempt
		if attempt < maxAttempts {
			waitDuration := t.calculateBackoff(attempt)
			logger.Warn("webhook delivery failed, retrying",
				slog.String("event_id", request.EventID),
				slog.Int("attempt", attempt),
				slog.Int("max_attempts", maxAttempts),
				slog.String("error", lastResult.ErrorMessage),
				slog.Duration("wait", waitDuration))

			// Wait before retry
			timer := time.NewTimer(waitDuration)
			select {
			case <-timer.C:
				// Continue to next attempt
			case <-ctx.Done():
				timer.Stop()
				return &DeliveryResult{
					Success:      false,
					Retryable:    false,
					ErrorMessage: fmt.Sprintf("context cancelled during retry wait: %v", ctx.Err()),
					ErrorType:    ErrorTypeTimeout,
				}
			}
		}
	}

	// All attempts exhausted
	logger.Error("webhook delivery failed after all retries",
		slog.String("event_id", request.EventID),
		slog.Int("total_attempts", maxAttempts),
		slog.String("final_error", lastResult.ErrorMessage))

	return lastResult
}

// calculateBackoff calculates exponential backoff duration for retry
func (t *HTTPTransport) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: min * (2 ^ (attempt - 1))
	backoff := t.config.RetryWaitMin * time.Duration(1<<uint(attempt-1))

	// Cap at max wait
	if backoff > t.config.RetryWaitMax {
		backoff = t.config.RetryWaitMax
	}

	return backoff
}

// validateRequest validates the delivery request
func (t *HTTPTransport) validateRequest(request *DeliveryRequest) error {
	if request == nil {
		return fmt.Errorf("nil delivery request")
	}

	if request.Endpoint == "" {
		return ErrInvalidEndpoint
	}

	if len(request.Payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	return nil
}

// prepareRequest creates an HTTP request from DeliveryRequest
func (t *HTTPTransport) prepareRequest(ctx context.Context, request *DeliveryRequest) (*http.Request, error) {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, request.Endpoint, bytes.NewReader(request.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", t.config.UserAgent)

	// Set custom headers from webhook config
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set FunnelChat-specific headers for tracking
	httpReq.Header.Set("X-FunnelChat-Event-ID", request.EventID)
	httpReq.Header.Set("X-FunnelChat-Event-Type", request.EventType)
	httpReq.Header.Set("X-FunnelChat-Instance-ID", request.InstanceID)
	httpReq.Header.Set("X-FunnelChat-Delivery-Attempt", fmt.Sprintf("%d", request.Attempt))

	return httpReq, nil
}

// logDeliveryResult logs the delivery result with appropriate level
func (t *HTTPTransport) logDeliveryResult(logger *slog.Logger, request *DeliveryRequest, result *DeliveryResult) {
	attrs := []any{
		slog.String("event_id", request.EventID),
		slog.String("event_type", request.EventType),
		slog.String("instance_id", request.InstanceID),
		slog.String("endpoint", request.Endpoint),
		slog.Int("attempt", request.Attempt),
		slog.Int("max_attempts", request.MaxAttempts),
		slog.Bool("success", result.Success),
		slog.Int("status_code", result.StatusCode),
		slog.Duration("duration", result.Duration),
		slog.Bool("retryable", result.Retryable),
	}

	if !result.Success {
		attrs = append(attrs,
			slog.String("error", result.ErrorMessage),
			slog.String("error_type", result.ErrorType))
	}

	if result.Success {
		logger.InfoContext(context.Background(), "webhook delivered", attrs...)
	} else if result.Retryable {
		logger.WarnContext(context.Background(), "webhook delivery failed (retryable)", attrs...)
	} else {
		logger.ErrorContext(context.Background(), "webhook delivery failed (permanent)", attrs...)
	}
}

// Name returns the transport name
func (t *HTTPTransport) Name() string {
	return "http"
}

// Close cleans up HTTP transport resources
func (t *HTTPTransport) Close() error {
	// HTTP client doesn't require explicit cleanup
	// Connection pool will be garbage collected
	t.client.CloseIdleConnections()
	return nil
}
