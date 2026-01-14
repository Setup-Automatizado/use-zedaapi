package http

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

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
	RetryAfter      time.Duration // Extracted from Retry-After header for 429 responses
}

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

type ResponseHandler struct {
	maxResponseSize int64
}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{
		maxResponseSize: 1024 * 1024 * 100,
	}
}

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

	result.Retryable, result.ErrorType = classifyHTTPStatus(resp.StatusCode)

	// Extract Retry-After header for rate-limited responses (HTTP 429)
	if resp.StatusCode == 429 {
		result.RetryAfter = h.parseRetryAfter(resp.Header.Get("Retry-After"))
	}

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

// parseRetryAfter parses the Retry-After header which can be either:
// - A number of seconds (e.g., "120")
// - An HTTP-date (e.g., "Wed, 21 Oct 2015 07:28:00 GMT")
func (h *ResponseHandler) parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	// Try parsing as seconds first (most common for 429 responses)
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date
	if t, err := http.ParseTime(value); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	return 0
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

type HTTPTransport struct {
	client          *http.Client
	config          *Config
	responseHandler *ResponseHandler
	metrics         *observability.Metrics
}

func NewHTTPTransport(config *Config) *HTTPTransport {
	return NewHTTPTransportWithMetrics(config, nil)
}

func NewHTTPTransportWithMetrics(config *Config, metrics *observability.Metrics) *HTTPTransport {
	if config == nil {
		config = DefaultConfig()
	}

	return &HTTPTransport{
		client:          NewHTTPClient(config),
		config:          config,
		responseHandler: NewResponseHandler(),
		metrics:         metrics,
	}
}

func (t *HTTPTransport) Deliver(ctx context.Context, request *DeliveryRequest) (*DeliveryResult, error) {
	logger := logging.ContextLogger(ctx, nil)

	if err := t.validateRequest(request); err != nil {
		return &DeliveryResult{
			Success:      false,
			Retryable:    false,
			ErrorMessage: err.Error(),
			ErrorType:    ErrorTypeClient,
		}, err
	}

	result := t.deliverWithRetry(ctx, request, logger)

	t.logDeliveryResult(logger, request, result)

	return result, nil
}

func (t *HTTPTransport) deliverWithRetry(ctx context.Context, request *DeliveryRequest, logger *slog.Logger) *DeliveryResult {
	var lastResult *DeliveryResult

	maxAttempts := t.config.MaxRetries
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return &DeliveryResult{
				Success:      false,
				Retryable:    false,
				ErrorMessage: fmt.Sprintf("context cancelled: %v", ctx.Err()),
				ErrorType:    ErrorTypeTimeout,
			}
		}

		// Create a fresh HTTP request for each attempt to avoid body reader exhaustion.
		// The body reader (bytes.Reader) is consumed on each request and cannot be reused.
		httpReq, err := t.prepareRequest(ctx, request)
		if err != nil {
			logger.Error("failed to prepare HTTP request",
				slog.String("error", err.Error()),
				slog.String("event_id", request.EventID),
				slog.String("endpoint", request.Endpoint),
				slog.Int("attempt", attempt))

			return &DeliveryResult{
				Success:      false,
				Retryable:    false,
				ErrorMessage: fmt.Sprintf("failed to prepare request: %v", err),
				ErrorType:    ErrorTypeSerialization,
			}
		}

		start := time.Now()
		resp, err := t.client.Do(httpReq)
		duration := time.Since(start)

		lastResult = t.responseHandler.HandleHTTPResponse(resp, duration, err)

		if lastResult.Success {
			logger.Debug("webhook delivered successfully",
				slog.String("event_id", request.EventID),
				slog.Int("attempt", attempt),
				slog.Int("status_code", lastResult.StatusCode),
				slog.Duration("duration", duration))

			// Record success metrics
			if t.metrics != nil {
				t.metrics.TransportDeliveries.WithLabelValues(request.InstanceID, "http", "success").Inc()
				t.metrics.TransportDuration.WithLabelValues(request.InstanceID, "http").Observe(duration.Seconds())
			}
			return lastResult
		}

		if !lastResult.Retryable {
			logger.Warn("webhook delivery failed (permanent)",
				slog.String("event_id", request.EventID),
				slog.Int("attempt", attempt),
				slog.String("error", lastResult.ErrorMessage),
				slog.String("error_type", lastResult.ErrorType))

			// Record permanent failure metrics
			if t.metrics != nil {
				t.metrics.TransportDeliveries.WithLabelValues(request.InstanceID, "http", "failed").Inc()
				t.metrics.TransportErrors.WithLabelValues(request.InstanceID, "http", lastResult.ErrorType).Inc()
			}
			return lastResult
		}

		if attempt < maxAttempts {
			// Use Retry-After header if available (for 429 responses), otherwise use exponential backoff.
			// This respects rate-limit directives from the webhook server.
			waitDuration := t.calculateBackoff(attempt)
			if lastResult.RetryAfter > 0 {
				waitDuration = lastResult.RetryAfter
				// Cap the Retry-After to avoid excessively long waits
				if waitDuration > t.config.RetryWaitMax {
					waitDuration = t.config.RetryWaitMax
				}
			}

			logAttrs := []any{
				slog.String("event_id", request.EventID),
				slog.Int("attempt", attempt),
				slog.Int("max_attempts", maxAttempts),
				slog.String("error", lastResult.ErrorMessage),
				slog.Duration("wait", waitDuration),
			}
			if lastResult.RetryAfter > 0 {
				logAttrs = append(logAttrs, slog.Duration("retry_after_header", lastResult.RetryAfter))
			}
			logger.Warn("webhook delivery failed, retrying", logAttrs...)

			// Record retry metrics
			if t.metrics != nil {
				t.metrics.TransportRetries.WithLabelValues(request.InstanceID, "http", strconv.Itoa(attempt)).Inc()
			}

			timer := time.NewTimer(waitDuration)
			select {
			case <-timer.C:
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

	logger.Error("webhook delivery failed after all retries",
		slog.String("event_id", request.EventID),
		slog.Int("total_attempts", maxAttempts),
		slog.String("final_error", lastResult.ErrorMessage))

	return lastResult
}

func (t *HTTPTransport) calculateBackoff(attempt int) time.Duration {
	backoff := t.config.RetryWaitMin * time.Duration(1<<uint(attempt-1))

	if backoff > t.config.RetryWaitMax {
		backoff = t.config.RetryWaitMax
	}

	return backoff
}

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

func (t *HTTPTransport) prepareRequest(ctx context.Context, request *DeliveryRequest) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, request.Endpoint, bytes.NewReader(request.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", t.config.UserAgent)

	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	httpReq.Header.Set("X-FunnelChat-Event-ID", request.EventID)
	httpReq.Header.Set("X-FunnelChat-Event-Type", request.EventType)
	httpReq.Header.Set("X-FunnelChat-Instance-ID", request.InstanceID)
	httpReq.Header.Set("X-FunnelChat-Delivery-Attempt", fmt.Sprintf("%d", request.Attempt))

	return httpReq, nil
}

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

func (t *HTTPTransport) Name() string {
	return "http"
}

func (t *HTTPTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}
