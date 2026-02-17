package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"

	"github.com/google/uuid"
)

// InstanceLocator checks if an instance client exists on this process.
type InstanceLocator interface {
	HasClient(instanceID uuid.UUID) bool
}

// WorkerResolver resolves a worker ID to its HTTP advertise address.
type WorkerResolver interface {
	ResolveAddr(workerID string) (string, bool)
}

// InstanceOwnerLookup finds which worker currently owns an instance.
type InstanceOwnerLookup interface {
	LookupOwner(ctx context.Context, instanceID uuid.UUID) (string, error)
}

const (
	proxiedHeader  = "X-Zedaapi-Proxied"
	instancePrefix = "/instances/"
	tokenSegment   = "/token/"
)

type ownerCacheEntry struct {
	workerID  string
	expiresAt time.Time
}

// responseInterceptor tracks whether headers have been written to the
// underlying ResponseWriter so the proxy fallback path can avoid
// double-writes on a partially committed response.
type responseInterceptor struct {
	http.ResponseWriter
	headerWritten bool
}

func (ri *responseInterceptor) WriteHeader(code int) {
	ri.headerWritten = true
	ri.ResponseWriter.WriteHeader(code)
}

func (ri *responseInterceptor) Write(b []byte) (int, error) {
	ri.headerWritten = true
	return ri.ResponseWriter.Write(b)
}

// RoutingMiddleware routes instance-scoped HTTP requests to the replica
// that owns the WhatsApp client for that instance.
type RoutingMiddleware struct {
	locator      InstanceLocator
	resolver     WorkerResolver
	ownerLookup  InstanceOwnerLookup
	selfWorkerID string
	log          *slog.Logger
	cacheTTL     time.Duration
	proxyTimeout time.Duration
	cache        sync.Map // uuid.UUID -> ownerCacheEntry
	transport    *http.Transport
	done         chan struct{}
}

// NewRoutingMiddleware creates a RoutingMiddleware that proxies requests
// to the replica owning the target instance.
func NewRoutingMiddleware(
	locator InstanceLocator,
	resolver WorkerResolver,
	ownerLookup InstanceOwnerLookup,
	selfWorkerID string,
	log *slog.Logger,
) *RoutingMiddleware {
	if log == nil {
		log = slog.Default()
	}
	m := &RoutingMiddleware{
		locator:      locator,
		resolver:     resolver,
		ownerLookup:  ownerLookup,
		selfWorkerID: selfWorkerID,
		log:          log.With(slog.String("component", "routing")),
		cacheTTL:     5 * time.Second,
		proxyTimeout: 30 * time.Second,
		transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
		},
		done: make(chan struct{}),
	}
	go m.evictExpiredCache()
	return m
}

// Handler returns an http.Handler middleware that intercepts instance-scoped
// requests and proxies them to the owning replica when needed.
func (m *RoutingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Save and strip the proxied header to prevent spoofing by external
		// clients and leakage to downstream handlers.
		proxied := r.Header.Get(proxiedHeader)
		r.Header.Del(proxiedHeader)

		// Already proxied — handle locally to prevent infinite loops.
		if proxied != "" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract instanceId from URL path.
		instanceID, ok := extractInstanceID(r.URL.Path)
		if !ok {
			// Not an instance route (e.g. /health, /metrics) — pass through.
			next.ServeHTTP(w, r)
			return
		}

		// Fast path: instance client lives on this process.
		if m.locator.HasClient(instanceID) {
			next.ServeHTTP(w, r)
			return
		}

		// Look up owner worker ID (with short-lived cache).
		ownerID, err := m.cachedLookupOwner(r.Context(), instanceID)
		if err != nil || ownerID == "" || ownerID == m.selfWorkerID {
			// Error, no owner, or we are the owner — handle locally.
			next.ServeHTTP(w, r)
			return
		}

		// Resolve owner's HTTP address from worker registry.
		addr, found := m.resolver.ResolveAddr(ownerID)
		if !found || addr == "" {
			// Worker offline or no address published — handle locally.
			next.ServeHTTP(w, r)
			return
		}

		m.log.Debug("routing to owner",
			slog.String("instance_id", instanceID.String()),
			slog.String("target_worker", ownerID),
			slog.String("target_addr", addr),
		)
		m.proxyOrFallback(w, r, addr, next)
	})
}

// extractInstanceID parses /instances/{uuid}/token/{token}/... paths.
// It does not depend on the Chi router and runs before route matching.
func extractInstanceID(path string) (uuid.UUID, bool) {
	if !strings.HasPrefix(path, instancePrefix) {
		return uuid.Nil, false
	}
	rest := path[len(instancePrefix):]
	idx := strings.Index(rest, tokenSegment)
	if idx <= 0 {
		return uuid.Nil, false
	}
	// Ensure there's a non-empty token after /token/
	if idx+len(tokenSegment) >= len(rest) {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(rest[:idx])
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (m *RoutingMiddleware) cachedLookupOwner(ctx context.Context, instanceID uuid.UUID) (string, error) {
	if entry, ok := m.cache.Load(instanceID); ok {
		ce := entry.(ownerCacheEntry)
		if time.Now().Before(ce.expiresAt) {
			return ce.workerID, nil
		}
	}
	ownerID, err := m.ownerLookup.LookupOwner(ctx, instanceID)
	if err != nil {
		return "", err
	}
	m.cache.Store(instanceID, ownerCacheEntry{
		workerID:  ownerID,
		expiresAt: time.Now().Add(m.cacheTTL),
	})
	return ownerID, nil
}

// proxyOrFallback attempts to reverse-proxy the request to the target replica.
// On transport errors, it falls back to handling the request locally.
func (m *RoutingMiddleware) proxyOrFallback(w http.ResponseWriter, r *http.Request, targetAddr string, next http.Handler) {
	target, err := url.Parse(targetAddr)
	if err != nil {
		next.ServeHTTP(w, r)
		return
	}

	// Buffer body for potential replay on fallback.
	var bodyBuf []byte
	if r.Body != nil && r.Body != http.NoBody {
		bodyBuf, err = io.ReadAll(r.Body)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBuf))
	}

	var transportFailed atomic.Bool

	ri := &responseInterceptor{ResponseWriter: w}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			req.Header.Set(proxiedHeader, m.selfWorkerID)
		},
		Transport: m.transport,
		ErrorHandler: func(_ http.ResponseWriter, _ *http.Request, proxyErr error) {
			// Do NOT write to ResponseWriter — leave it clean for fallback.
			transportFailed.Store(true)
			m.log.Warn("proxy transport error, falling back to local",
				slog.String("target", targetAddr),
				slog.String("error", proxyErr.Error()),
			)
		},
	}

	proxy.ServeHTTP(ri, r)

	if transportFailed.Load() {
		if ri.headerWritten {
			m.log.Error("proxy error after response committed, cannot fall back",
				slog.String("target", targetAddr),
			)
			return
		}
		// Transport error — target unreachable. Fall back to local handler.
		if bodyBuf != nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBuf))
		}
		next.ServeHTTP(w, r)
	}
}

// evictExpiredCache periodically removes stale entries from the owner cache.
func (m *RoutingMiddleware) evictExpiredCache() {
	ticker := time.NewTicker(m.cacheTTL * 10)
	defer ticker.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			now := time.Now()
			m.cache.Range(func(k, v any) bool {
				if now.After(v.(ownerCacheEntry).expiresAt) {
					m.cache.Delete(k)
				}
				return true
			})
		}
	}
}

// Close shuts down the eviction goroutine and releases transport resources.
func (m *RoutingMiddleware) Close() {
	close(m.done)
	m.transport.CloseIdleConnections()
}
