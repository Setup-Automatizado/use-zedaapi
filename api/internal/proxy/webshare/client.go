package webshare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client handles HTTP communication with Webshare API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	planID     string // Webshare plan ID to target (static residential vs rotating)
	mode       string // "direct" for static/ISP or "backbone" for rotating residential
	limiter    *RateLimiter
	log        *slog.Logger
}

// NewClient creates a new Webshare API client.
func NewClient(apiKey, baseURL, planID, mode string, log *slog.Logger) *Client {
	if baseURL == "" {
		baseURL = "https://proxy.webshare.io/api/v2"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if mode == "" {
		mode = "direct"
	}
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		planID:     planID,
		mode:       mode,
		limiter:    NewRateLimiter(),
		log:        log.With(slog.String("component", "webshare_client")),
	}
}

// ListProxies fetches proxies with optional country filter and pagination.
func (c *Client) ListProxies(ctx context.Context, countryCodes []string, page, pageSize int) (*ProxyListResponse, error) {
	if err := c.limiter.WaitProxyList(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	params := url.Values{}
	params.Set("mode", c.mode)
	if c.planID != "" {
		params.Set("plan_id", c.planID)
	}
	if len(countryCodes) > 0 {
		params.Set("country_code__in", strings.Join(countryCodes, ","))
	}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	endpoint := fmt.Sprintf("%s/proxy/list/?%s", c.baseURL, params.Encode())
	var result ProxyListResponse
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, fmt.Errorf("list proxies: %w", err)
	}
	return &result, nil
}

// ReplaceProxy requests replacement of a failed proxy.
func (c *Client) ReplaceProxy(ctx context.Context, proxyAddress string, port int) (*ReplaceResponse, error) {
	if err := c.limiter.WaitGeneral(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	body := ReplaceRequest{ProxyAddress: proxyAddress, Port: port}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal replace request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/proxy/list/replaced/", c.baseURL)
	var result ReplaceResponse
	if err := c.doRequest(ctx, http.MethodPost, endpoint, bodyJSON, &result); err != nil {
		return nil, fmt.Errorf("replace proxy: %w", err)
	}
	return &result, nil
}

// ReplaceProxyByID requests replacement of a failed proxy using its Webshare ID.
func (c *Client) ReplaceProxyByID(ctx context.Context, proxyID string) (*ReplaceResponse, error) {
	if err := c.limiter.WaitGeneral(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	body := ReplaceByIDRequest{ProxyID: proxyID}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal replace request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/proxy/list/replaced/", c.baseURL)
	var result ReplaceResponse
	if err := c.doRequest(ctx, http.MethodPost, endpoint, bodyJSON, &result); err != nil {
		return nil, fmt.Errorf("replace proxy by ID: %w", err)
	}
	return &result, nil
}

// doRequest executes an HTTP request with authentication, 429 retry, and error handling.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body []byte, target any) error {
	const maxRetries = 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = strings.NewReader(string(body))
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Token "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("execute request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}

		// Handle 429 rate limit with exponential backoff retry.
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt >= maxRetries {
				return fmt.Errorf("webshare API rate limited: exhausted %d retries", maxRetries)
			}
			backoff := c.parseRetryAfter(resp, attempt)
			c.log.Warn("rate limited by Webshare API, retrying",
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", maxRetries),
				slog.Duration("backoff", backoff))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			continue
		}

		if resp.StatusCode >= 400 {
			var apiErr APIError
			if jsonErr := json.Unmarshal(respBody, &apiErr); jsonErr == nil && apiErr.Detail != "" {
				return fmt.Errorf("webshare API error %d: %s", resp.StatusCode, apiErr.Detail)
			}
			return fmt.Errorf("webshare API error %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 200)]))
		}

		if target != nil {
			if err := json.Unmarshal(respBody, target); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("webshare API: exhausted %d retries", maxRetries)
}

// parseRetryAfter extracts the backoff duration from the Retry-After header,
// falling back to exponential backoff: 15s, 30s, 60s, 60s, 60s.
func (c *Client) parseRetryAfter(resp *http.Response, attempt int) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if seconds, err := strconv.Atoi(ra); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	backoff := 15 * time.Second * (1 << uint(attempt))
	if backoff > 60*time.Second {
		backoff = 60 * time.Second
	}
	return backoff
}

// Close releases resources.
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
