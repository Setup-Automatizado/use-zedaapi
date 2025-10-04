package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
		body, err := io.ReadAll(io.LimitReader(resp.Body, h.maxResponseSize))
		if err == nil {
			result.Response = body
		} else {
			result.ErrorMessage = fmt.Sprintf("error reading response body: %v", err)
		}
	}

	result.Retryable, result.ErrorType = ClassifyHTTPStatus(resp.StatusCode)

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

	errStr := err.Error()

	if isTimeoutError(err) {
		return ErrorTypeTimeout, true
	}
	if isConnectionError(errStr) {
		return ErrorTypeConnection, true
	}
	if isTLSError(errStr) {
		return ErrorTypeClient, false
	}
	if isDNSError(errStr) {
		return ErrorTypeConnection, true
	}

	return ErrorTypeUnknown, false
}

func isTimeoutError(err error) bool {
	type timeout interface {
		Timeout() bool
	}

	if e, ok := err.(timeout); ok {
		return e.Timeout()
	}

	return false
}

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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
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
