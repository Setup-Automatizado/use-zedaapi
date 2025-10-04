package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Timeout                time.Duration
	MaxIdleConns           int
	MaxIdleConnsPerHost    int
	MaxConnsPerHost        int
	IdleConnTimeout        time.Duration
	TLSHandshakeTimeout    time.Duration
	ExpectContinueTimeout  time.Duration
	DisableKeepAlives      bool
	DisableCompression     bool
	MaxResponseHeaderBytes int64
	InsecureSkipVerify     bool
	UserAgent              string
	MaxRetries             int
	RetryWaitMin           time.Duration
	RetryWaitMax           time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:                60 * time.Second,
		MaxIdleConns:           100,
		MaxIdleConnsPerHost:    100,
		MaxConnsPerHost:        50,
		IdleConnTimeout:        120 * time.Second,
		TLSHandshakeTimeout:    10 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		DisableKeepAlives:      false,
		DisableCompression:     false,
		MaxResponseHeaderBytes: 1 << 20 * 100,
		InsecureSkipVerify:     false,
		UserAgent:              "FunnelChat-Webhook/1.0",
		MaxRetries:             3,
		RetryWaitMin:           2 * time.Second,
		RetryWaitMax:           30 * time.Second,
	}
}

func NewHTTPClient(config *Config) *http.Client {
	if config == nil {
		config = DefaultConfig()
	}

	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                  http.ProxyFromEnvironment,
		DialContext:            dialer.DialContext,
		ForceAttemptHTTP2:      true,
		MaxIdleConns:           config.MaxIdleConns,
		MaxIdleConnsPerHost:    config.MaxIdleConnsPerHost,
		MaxConnsPerHost:        config.MaxConnsPerHost,
		IdleConnTimeout:        config.IdleConnTimeout,
		TLSHandshakeTimeout:    config.TLSHandshakeTimeout,
		ExpectContinueTimeout:  config.ExpectContinueTimeout,
		DisableKeepAlives:      config.DisableKeepAlives,
		DisableCompression:     config.DisableCompression,
		MaxResponseHeaderBytes: config.MaxResponseHeaderBytes,
	}

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
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}
