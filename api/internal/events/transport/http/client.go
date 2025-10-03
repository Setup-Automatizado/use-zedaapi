package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Config holds HTTP transport configuration
type Config struct {
	// Timeout is the maximum time to wait for webhook delivery
	Timeout time.Duration

	// MaxIdleConns controls the maximum number of idle connections
	MaxIdleConns int

	// MaxIdleConnsPerHost controls idle connections per host
	MaxIdleConnsPerHost int

	// MaxConnsPerHost limits total connections per host
	MaxConnsPerHost int

	// IdleConnTimeout is the maximum time an idle connection will remain idle
	IdleConnTimeout time.Duration

	// TLSHandshakeTimeout specifies the maximum amount of time waiting for TLS handshake
	TLSHandshakeTimeout time.Duration

	// ExpectContinueTimeout specifies the amount of time to wait for a server's first
	// response headers after fully writing the request headers
	ExpectContinueTimeout time.Duration

	// DisableKeepAlives disables HTTP keep-alives
	DisableKeepAlives bool

	// DisableCompression disables compression
	DisableCompression bool

	// MaxResponseHeaderBytes specifies a limit on response header bytes
	MaxResponseHeaderBytes int64

	// InsecureSkipVerify controls whether TLS certificates are verified
	// WARNING: Only use for development/testing
	InsecureSkipVerify bool

	// UserAgent is the User-Agent header to send
	UserAgent string

	// MaxRetries is the maximum number of HTTP-level retries for transient failures
	MaxRetries int

	// RetryWaitMin is the minimum time to wait between retries
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum time to wait between retries
	RetryWaitMax time.Duration
}

// DefaultConfig returns a default HTTP client configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout:                30 * time.Second,
		MaxIdleConns:           100,
		MaxIdleConnsPerHost:    10,
		MaxConnsPerHost:        50,
		IdleConnTimeout:        90 * time.Second,
		TLSHandshakeTimeout:    10 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		DisableKeepAlives:      false,
		DisableCompression:     false,
		MaxResponseHeaderBytes: 1 << 20, // 1MB
		InsecureSkipVerify:     false,
		UserAgent:              "FunnelChat-Webhook/1.0",
		MaxRetries:             3,
		RetryWaitMin:           1 * time.Second,
		RetryWaitMax:           30 * time.Second,
	}
}

// NewHTTPClient creates a configured HTTP client for webhook delivery
func NewHTTPClient(config *Config) *http.Client {
	if config == nil {
		config = DefaultConfig()
	}

	// Custom dialer for connection control
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Custom transport with connection pooling and timeouts
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		DisableCompression:    config.DisableCompression,
		MaxResponseHeaderBytes: config.MaxResponseHeaderBytes,
	}

	// TLS configuration
	if config.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects to 10
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}
