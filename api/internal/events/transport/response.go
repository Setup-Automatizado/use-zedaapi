package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ResponseHandler processes HTTP responses and creates DeliveryResults
type ResponseHandler struct {
	maxResponseSize int64 // Maximum response body size to read (default: 1MB)
}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{
		maxResponseSize: 1024 * 1024, // 1MB default
	}
}

// HandleHTTPResponse processes an HTTP response and returns a DeliveryResult
func (h *ResponseHandler) HandleHTTPResponse(resp *http.Response, duration time.Duration, requestErr error) *DeliveryResult {
	result := &DeliveryResult{
		Duration: duration,
	}

	// Handle request-level errors (timeout, connection, etc)
	if requestErr != nil {
		result.Success = false
		result.ErrorMessage = requestErr.Error()
		result.ErrorType, result.Retryable = h.classifyRequestError(requestErr)
		return result
	}

	// Handle HTTP response
	if resp == nil {
		result.Success = false
		result.ErrorMessage = "nil response received"
		result.ErrorType = ErrorTypeUnknown
		result.Retryable = false
		return result
	}

	result.StatusCode = resp.StatusCode
	result.ResponseHeaders = resp.Header

	// Read response body (with size limit)
	if resp.Body != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, h.maxResponseSize))
		if err == nil {
			result.Response = body
		} else {
			result.ErrorMessage = fmt.Sprintf("error reading response body: %v", err)
		}
	}

	// Classify based on status code
	result.Retryable, result.ErrorType = ClassifyHTTPStatus(resp.StatusCode)

	// Success for 2xx status codes
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

// classifyRequestError classifies request-level errors for retry decisions
func (h *ResponseHandler) classifyRequestError(err error) (errorType string, retryable bool) {
	if err == nil {
		return "", false
	}

	errStr := err.Error()

	// Timeout errors - retryable
	if isTimeoutError(err) {
		return ErrorTypeTimeout, true
	}

	// Connection errors - retryable
	if isConnectionError(errStr) {
		return ErrorTypeConnection, true
	}

	// TLS/certificate errors - not retryable
	if isTLSError(errStr) {
		return ErrorTypeClient, false
	}

	// DNS errors - retryable (might be temporary)
	if isDNSError(errStr) {
		return ErrorTypeConnection, true
	}

	// Unknown errors - not retryable to be safe
	return ErrorTypeUnknown, false
}

// isTimeoutError checks if error is a timeout
func isTimeoutError(err error) bool {
	type timeout interface {
		Timeout() bool
	}

	if e, ok := err.(timeout); ok {
		return e.Timeout()
	}

	return false
}

// isConnectionError checks if error is connection-related
func isConnectionError(errStr string) bool {
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no such host",
		"network is unreachable",
		"host is down",
	}

	for _, connErr := range connectionErrors {
		if contains(errStr, connErr) {
			return true
		}
	}

	return false
}

// isTLSError checks if error is TLS/certificate related
func isTLSError(errStr string) bool {
	tlsErrors := []string{
		"certificate",
		"tls",
		"x509",
		"ssl",
	}

	for _, tlsErr := range tlsErrors {
		if contains(errStr, tlsErr) {
			return true
		}
	}

	return false
}

// isDNSError checks if error is DNS-related
func isDNSError(errStr string) bool {
	dnsErrors := []string{
		"no such host",
		"dns",
		"name resolution",
	}

	for _, dnsErr := range dnsErrors {
		if contains(errStr, dnsErr) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	// Simple case-insensitive substring check
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}

	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// ParseWebhookResponse attempts to parse a webhook response as JSON
// This is optional validation - not all webhooks return structured responses
func ParseWebhookResponse(body []byte) (map[string]interface{}, error) {
	if len(body) == 0 {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}
