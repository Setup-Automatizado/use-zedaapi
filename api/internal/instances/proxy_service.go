package instances

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"golang.org/x/net/proxy"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// GetProxy retrieves the proxy configuration for an instance.
func (s *Service) GetProxy(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) (*ProxyConfig, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}
	return s.repo.GetProxyConfig(ctx, id)
}

// UpdateProxy sets or updates the proxy configuration for an instance.
// It validates the proxy URL format and optionally tests connectivity before applying.
func (s *Service) UpdateProxy(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, cfg ProxyConfig) (*ProxyConfig, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}

	if cfg.ProxyURL != nil && *cfg.ProxyURL != "" {
		if err := validateProxyURL(*cfg.ProxyURL); err != nil {
			return nil, err
		}
	}

	cfg.HealthStatus = "unknown"
	cfg.HealthFailures = 0

	if err := s.repo.UpdateProxyConfig(ctx, id, cfg); err != nil {
		return nil, err
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "proxy_service"),
		slog.String("instance_id", id.String()),
	)

	// Apply proxy to live client if connected
	if cfg.Enabled && cfg.ProxyURL != nil && *cfg.ProxyURL != "" {
		if applyErr := s.registry.ApplyProxy(ctx, id, *cfg.ProxyURL, cfg.NoWebsocket, cfg.OnlyLogin, cfg.NoMedia); applyErr != nil {
			logger.Warn("failed to apply proxy to live client",
				slog.String("error", applyErr.Error()))
		} else {
			logger.Info("proxy applied to live client")
		}
	}

	return s.repo.GetProxyConfig(ctx, id)
}

// RemoveProxy clears the proxy configuration and removes it from the live client.
func (s *Service) RemoveProxy(ctx context.Context, id uuid.UUID, clientToken, instanceToken string) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return ErrUnauthorized
	}

	if err := s.repo.ClearProxyConfig(ctx, id); err != nil {
		return err
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "proxy_service"),
		slog.String("instance_id", id.String()),
	)

	// Remove proxy from live client (set to nil = use default/env proxy)
	if applyErr := s.registry.ApplyProxy(ctx, id, "", false, false, false); applyErr != nil {
		logger.Warn("failed to clear proxy from live client",
			slog.String("error", applyErr.Error()))
	} else {
		logger.Info("proxy cleared from live client")
	}

	return nil
}

// TestProxy validates a proxy URL and tests its connectivity without applying it.
func (s *Service) TestProxy(ctx context.Context, id uuid.UUID, clientToken, instanceToken, proxyURL string) (*ProxyTestResult, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}

	if err := validateProxyURL(proxyURL); err != nil {
		return nil, err
	}

	return testProxyConnectivity(ctx, proxyURL)
}

// SwapProxy hot-swaps the proxy on an active WhatsApp connection.
// The client is disconnected and immediately reconnected through the new proxy.
// The WhatsApp session (device store) is preserved - only the transport changes.
func (s *Service) SwapProxy(ctx context.Context, id uuid.UUID, clientToken, instanceToken, newProxyURL string) (*ProxyConfig, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}

	if newProxyURL != "" {
		if err := validateProxyURL(newProxyURL); err != nil {
			return nil, err
		}

		// Test the new proxy before swapping
		result, testErr := testProxyConnectivity(ctx, newProxyURL)
		if testErr != nil {
			return nil, fmt.Errorf("test new proxy: %w", testErr)
		}
		if !result.Reachable {
			return nil, fmt.Errorf("%w: %s", ErrProxyUnreachable, result.Error)
		}
	}

	logger := logging.ContextLogger(ctx, s.log).With(
		slog.String("component", "proxy_service"),
		slog.String("instance_id", id.String()),
	)

	// Get current proxy config to preserve options
	currentCfg, err := s.repo.GetProxyConfig(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update database with new proxy URL
	var proxyPtr *string
	enabled := false
	if newProxyURL != "" {
		proxyPtr = &newProxyURL
		enabled = true
	}

	newCfg := ProxyConfig{
		ProxyURL:    proxyPtr,
		Enabled:     enabled,
		NoWebsocket: currentCfg.NoWebsocket,
		OnlyLogin:   currentCfg.OnlyLogin,
		NoMedia:     currentCfg.NoMedia,
	}

	if err := s.repo.UpdateProxyConfig(ctx, id, newCfg); err != nil {
		return nil, err
	}

	// Hot-swap: disconnect and reconnect through new proxy
	if swapErr := s.registry.SwapProxy(ctx, id, newProxyURL, newCfg.NoWebsocket, newCfg.OnlyLogin, newCfg.NoMedia); swapErr != nil {
		logger.Error("proxy hot-swap failed",
			slog.String("error", swapErr.Error()))

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "proxy_service")
			scope.SetTag("instance_id", id.String())
			scope.SetContext("proxy_swap", map[string]interface{}{
				"new_proxy_url": sanitizeProxyURL(newProxyURL),
			})
			sentry.CaptureException(fmt.Errorf("proxy hot-swap failed: %w", swapErr))
		})

		return nil, fmt.Errorf("proxy hot-swap failed: %w", swapErr)
	}

	logger.Info("proxy hot-swapped successfully",
		slog.String("new_proxy", sanitizeProxyURL(newProxyURL)))

	return s.repo.GetProxyConfig(ctx, id)
}

// GetProxyHealthLogs retrieves recent health check logs for an instance.
func (s *Service) GetProxyHealthLogs(ctx context.Context, id uuid.UUID, clientToken, instanceToken string, limit int) ([]ProxyHealthLog, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.tokensMatch(inst, clientToken, instanceToken) {
		return nil, ErrUnauthorized
	}
	return s.repo.GetProxyHealthLogs(ctx, id, limit)
}

// validateProxyURL checks that the proxy URL has a valid scheme, host, and port.
func validateProxyURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidProxyURL, err.Error())
	}
	switch parsed.Scheme {
	case "http", "https", "socks5":
		// valid
	default:
		return fmt.Errorf("%w: scheme %q not supported", ErrInvalidProxyURL, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidProxyURL)
	}
	// Validate port range (1-65535).
	if portStr := parsed.Port(); portStr != "" {
		port, portErr := strconv.Atoi(portStr)
		if portErr != nil || port <= 0 || port > 65535 {
			return fmt.Errorf("%w: port %s is invalid (must be 1-65535)", ErrInvalidProxyURL, portStr)
		}
	}
	return nil
}

// testProxyConnectivity tests whether a proxy can reach WhatsApp servers.
func testProxyConnectivity(ctx context.Context, proxyURL string) (*ProxyTestResult, error) {
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	result := &ProxyTestResult{}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}

	target := "web.whatsapp.com:443"

	switch parsed.Scheme {
	case "socks5":
		dialer, dialErr := proxy.FromURL(parsed, &net.Dialer{Timeout: 10 * time.Second})
		if dialErr != nil {
			result.Error = fmt.Sprintf("socks5 setup: %s", dialErr.Error())
			return result, nil
		}
		ctxDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			result.Error = "socks5 proxy does not support context dialing"
			return result, nil
		}
		conn, connErr := ctxDialer.DialContext(testCtx, "tcp", target)
		if connErr != nil {
			result.Error = connErr.Error()
			return result, nil
		}
		conn.Close()

	case "http", "https":
		// For HTTP proxies, test by establishing a TCP connection through the proxy via CONNECT
		proxyDialer := &net.Dialer{Timeout: 10 * time.Second}
		proxyConn, connErr := proxyDialer.DialContext(testCtx, "tcp", parsed.Host)
		if connErr != nil {
			result.Error = fmt.Sprintf("connect to proxy: %s", connErr.Error())
			return result, nil
		}
		defer proxyConn.Close()

		// Send HTTP CONNECT request
		connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", target, target)
		if parsed.User != nil {
			connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", parsed.User.String())
		}
		connectReq += "\r\n"

		if _, err := proxyConn.Write([]byte(connectReq)); err != nil {
			result.Error = fmt.Sprintf("send CONNECT: %s", err.Error())
			return result, nil
		}

		// Read response (just check for 200)
		buf := make([]byte, 1024)
		n, readErr := proxyConn.Read(buf)
		if readErr != nil {
			result.Error = fmt.Sprintf("read CONNECT response: %s", readErr.Error())
			return result, nil
		}
		response := string(buf[:n])
		if len(response) < 12 || response[9:12] != "200" {
			result.Error = fmt.Sprintf("proxy rejected CONNECT: %s", response[:min(len(response), 50)])
			return result, nil
		}
	}

	result.Reachable = true
	result.LatencyMs = int(time.Since(start).Milliseconds())
	return result, nil
}

// sanitizeProxyURL removes credentials from proxy URL for safe logging.
func sanitizeProxyURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid>"
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("***", "***")
	}
	return parsed.String()
}
