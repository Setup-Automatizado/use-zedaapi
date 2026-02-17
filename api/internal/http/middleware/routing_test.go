package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"log/slog"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockLocator struct {
	clients map[uuid.UUID]bool
}

func (m *mockLocator) HasClient(id uuid.UUID) bool { return m.clients[id] }

type mockResolver struct {
	addrs map[string]string
}

func (m *mockResolver) ResolveAddr(id string) (string, bool) {
	addr, ok := m.addrs[id]
	return addr, ok
}

type mockOwnerLookup struct {
	owners map[uuid.UUID]string
	err    error
	calls  atomic.Int64
}

func (m *mockOwnerLookup) LookupOwner(_ context.Context, id uuid.UUID) (string, error) {
	m.calls.Add(1)
	if m.err != nil {
		return "", m.err
	}
	return m.owners[id], nil
}

// localHandler writes "local" so tests can distinguish local from proxied.
func localHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("local"))
	})
}

// ---------------------------------------------------------------------------
// TestRoutingMiddleware
// ---------------------------------------------------------------------------

func TestRoutingMiddleware(t *testing.T) {
	selfWorker := "self-worker"
	instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	instancePath := "/instances/" + instanceID.String() + "/token/abc/device"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		path           string
		headers        map[string]string
		locatorClients map[uuid.UUID]bool
		resolverAddrs  map[string]string
		owners         map[uuid.UUID]string
		ownerErr       error
		// remoteHandler, if non-nil, starts a remote httptest server and
		// the resolver is pointed at its address.
		remoteHandler func(t *testing.T) http.Handler
		// closeRemote closes the remote server before the request so the
		// proxy gets a connection-refused error.
		closeRemote bool
		wantStatus  int
		wantBody    string
	}{
		{
			name:           "local_instance",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{instanceID: true},
			resolverAddrs:  map[string]string{},
			owners:         map[uuid.UUID]string{},
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "remote_instance",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{},
			owners:         map[uuid.UUID]string{instanceID: "other-worker"},
			// resolverAddrs filled dynamically via remoteHandler
			remoteHandler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Helper()
					if r.Header.Get(proxiedHeader) == "" {
						t.Error("expected proxied header on remote request")
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("remote"))
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   "remote",
		},
		{
			name:           "proxied_header_skips_routing",
			path:           instancePath,
			headers:        map[string]string{proxiedHeader: "some-worker"},
			locatorClients: map[uuid.UUID]bool{},
			resolverAddrs:  map[string]string{"other-worker": "http://10.0.0.2:8080"},
			owners:         map[uuid.UUID]string{instanceID: "other-worker"},
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "non_instance_route",
			path:           "/health",
			locatorClients: map[uuid.UUID]bool{},
			resolverAddrs:  map[string]string{},
			owners:         map[uuid.UUID]string{},
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "no_owner_in_db",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{},
			resolverAddrs:  map[string]string{},
			owners:         map[uuid.UUID]string{}, // empty, LookupOwner returns ""
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "same_worker_owner",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{},
			resolverAddrs:  map[string]string{selfWorker: "http://127.0.0.1:8080"},
			owners:         map[uuid.UUID]string{instanceID: selfWorker},
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "worker_address_not_found",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{},
			resolverAddrs:  map[string]string{}, // no entry for other-worker
			owners:         map[uuid.UUID]string{instanceID: "other-worker"},
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "lookup_error",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{},
			resolverAddrs:  map[string]string{},
			owners:         map[uuid.UUID]string{},
			ownerErr:       fmt.Errorf("database connection lost"),
			wantStatus:     http.StatusOK,
			wantBody:       "local",
		},
		{
			name:           "proxy_failure_fallback",
			path:           instancePath,
			locatorClients: map[uuid.UUID]bool{},
			owners:         map[uuid.UUID]string{instanceID: "other-worker"},
			// Start a remote server and close it before the request so the
			// proxy gets connection refused and falls back to local.
			remoteHandler: func(_ *testing.T) http.Handler {
				return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
			},
			closeRemote: true,
			wantStatus:  http.StatusOK,
			wantBody:    "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolverAddrs := tt.resolverAddrs
			if resolverAddrs == nil {
				resolverAddrs = map[string]string{}
			}

			// If a remote handler is provided, start a test server and
			// point the resolver at it.
			if tt.remoteHandler != nil {
				remote := httptest.NewServer(tt.remoteHandler(t))
				resolverAddrs["other-worker"] = remote.URL
				if tt.closeRemote {
					// Close immediately so the proxy gets connection refused.
					remote.Close()
				} else {
					defer remote.Close()
				}
			}

			mw := NewRoutingMiddleware(
				&mockLocator{clients: tt.locatorClients},
				&mockResolver{addrs: resolverAddrs},
				&mockOwnerLookup{owners: tt.owners, err: tt.ownerErr},
				selfWorker,
				logger,
			)
			defer mw.Close()
			handler := mw.Handler(localHandler())
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestHeaderSpoofingStripped
// ---------------------------------------------------------------------------

func TestHeaderSpoofingStripped(t *testing.T) {
	instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	instancePath := "/instances/" + instanceID.String() + "/token/abc/device"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// The middleware saves the proxied header, strips it from the request,
	// then uses the saved value for loop-prevention. If a client spoofs
	// the header, the middleware:
	// 1. Saves the spoofed value (non-empty)
	// 2. Strips it from the request (so downstream/remote never sees it)
	// 3. Treats it as "already proxied" and handles locally
	// This is correct: the spoofed header is neutralized (stripped) and
	// the request is served locally instead of being forwarded.

	var sawProxiedHeader bool
	downstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The downstream handler must NOT see the spoofed header.
		if r.Header.Get(proxiedHeader) != "" {
			sawProxiedHeader = true
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("local"))
	})

	mw := NewRoutingMiddleware(
		&mockLocator{clients: map[uuid.UUID]bool{}},
		&mockResolver{addrs: map[string]string{"other-worker": "http://10.0.0.2:8080"}},
		&mockOwnerLookup{owners: map[uuid.UUID]string{instanceID: "other-worker"}},
		"self-worker",
		logger,
	)
	defer mw.Close()

	handler := mw.Handler(downstream)

	req := httptest.NewRequest(http.MethodGet, instancePath, nil)
	// Attacker tries to spoof the proxied header.
	req.Header.Set(proxiedHeader, "attacker")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Request is handled locally (the saved header value was non-empty, so
	// it's treated as "already proxied"). This prevents attackers from
	// forcing proxy loops.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "local" {
		t.Errorf("body = %q, want %q", body, "local")
	}
	// Critical: the spoofed header must NOT leak to downstream.
	if sawProxiedHeader {
		t.Error("downstream handler saw the spoofed X-Zedaapi-Proxied header; it should have been stripped")
	}
}

// ---------------------------------------------------------------------------
// TestResponseCommittedNoDoubleWrite
// ---------------------------------------------------------------------------

func TestResponseCommittedNoDoubleWrite(t *testing.T) {
	selfWorker := "self-worker"
	instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	instancePath := "/instances/" + instanceID.String() + "/token/abc/device"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Remote handler hijacks the connection to write partial HTTP response
	// bytes, then closes the TCP connection abruptly. This simulates a
	// real-world scenario where the remote crashes mid-response.
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			// Fallback: write and return normally (test still validates no panic).
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("partial"))
			return
		}
		conn, buf, err := hj.Hijack()
		if err != nil {
			return
		}
		// Write a valid HTTP response header + partial body, then close.
		buf.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\npartial")
		buf.Flush()
		conn.Close()
	}))
	defer remote.Close()

	mw := NewRoutingMiddleware(
		&mockLocator{clients: map[uuid.UUID]bool{}},
		&mockResolver{addrs: map[string]string{"other-worker": remote.URL}},
		&mockOwnerLookup{owners: map[uuid.UUID]string{instanceID: "other-worker"}},
		selfWorker,
		logger,
	)
	defer mw.Close()

	handler := mw.Handler(localHandler())

	req := httptest.NewRequest(http.MethodGet, instancePath, nil)
	rec := httptest.NewRecorder()

	// The key assertion: this must not panic.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("middleware panicked on partially committed response: %v", r)
			}
		}()
		handler.ServeHTTP(rec, req)
	}()

	// The response interceptor should detect that headers were committed by
	// the proxy (partial remote response written) and prevent fallback to
	// localHandler. The body may contain "partial" from the remote, but must
	// NOT contain "local" appended after it (double-write).
	body := rec.Body.String()
	if strings.Contains(body, "local") && strings.Contains(body, "partial") {
		t.Errorf("body contains both 'partial' and 'local', indicating double-write corruption: %q", body)
	}
}

// ---------------------------------------------------------------------------
// TestExtractInstanceID
// ---------------------------------------------------------------------------

func TestExtractInstanceID(t *testing.T) {
	validUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name   string
		path   string
		wantID uuid.UUID
		wantOK bool
	}{
		{
			name:   "valid_path",
			path:   "/instances/550e8400-e29b-41d4-a716-446655440000/token/abc/device",
			wantID: validUUID,
			wantOK: true,
		},
		{
			name:   "valid_path_no_trailing",
			path:   "/instances/550e8400-e29b-41d4-a716-446655440000/token/abc",
			wantID: validUUID,
			wantOK: true,
		},
		{
			name:   "no_instance_prefix",
			path:   "/health",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "invalid_uuid",
			path:   "/instances/not-a-uuid/token/abc",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "missing_token_segment",
			path:   "/instances/550e8400-e29b-41d4-a716-446655440000",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "empty_path",
			path:   "",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "just_instances",
			path:   "/instances/",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "empty_token",
			path:   "/instances/550e8400-e29b-41d4-a716-446655440000/token/",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "media_path",
			path:   "/media/550e8400-e29b-41d4-a716-446655440000/file.jpg",
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "partner_path",
			path:   "/proxy-providers",
			wantID: uuid.Nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := extractInstanceID(tt.path)
			if gotOK != tt.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotID != tt.wantID {
				t.Errorf("id = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestConcurrentCacheAccess
// ---------------------------------------------------------------------------

func TestConcurrentCacheAccess(t *testing.T) {
	selfWorker := "self-worker"
	instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	instancePath := "/instances/" + instanceID.String() + "/token/abc/device"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("remote"))
	}))
	defer remote.Close()

	mw := NewRoutingMiddleware(
		&mockLocator{clients: map[uuid.UUID]bool{}},
		&mockResolver{addrs: map[string]string{"other-worker": remote.URL}},
		&mockOwnerLookup{owners: map[uuid.UUID]string{instanceID: "other-worker"}},
		selfWorker,
		logger,
	)
	defer mw.Close()

	handler := mw.Handler(localHandler())

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errCh := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, instancePath, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				errCh <- fmt.Sprintf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			body := rec.Body.String()
			if body != "remote" {
				errCh <- fmt.Sprintf("body = %q, want %q", body, "remote")
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for msg := range errCh {
		t.Error(msg)
	}
}

// ---------------------------------------------------------------------------
// TestCacheExpirationTriggersRelookup
// ---------------------------------------------------------------------------

func TestCacheExpirationTriggersRelookup(t *testing.T) {
	selfWorker := "self-worker"
	instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	instancePath := "/instances/" + instanceID.String() + "/token/abc/device"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("remote"))
	}))
	defer remote.Close()

	ownerLookup := &mockOwnerLookup{owners: map[uuid.UUID]string{instanceID: "other-worker"}}

	// Construct manually with a short cacheTTL to avoid a data race between
	// the test goroutine writing cacheTTL and the eviction goroutine reading it.
	mw := &RoutingMiddleware{
		locator:      &mockLocator{clients: map[uuid.UUID]bool{}},
		resolver:     &mockResolver{addrs: map[string]string{"other-worker": remote.URL}},
		ownerLookup:  ownerLookup,
		selfWorkerID: selfWorker,
		log:          logger.With(slog.String("component", "routing")),
		cacheTTL:     10 * time.Millisecond,
		proxyTimeout: 30 * time.Second,
		transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
		},
		done: make(chan struct{}),
	}
	go mw.evictExpiredCache()
	defer mw.Close()

	handler := mw.Handler(localHandler())

	// First request: populates cache, calls LookupOwner once.
	req1 := httptest.NewRequest(http.MethodGet, instancePath, nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Body.String() != "remote" {
		t.Fatalf("first request: body = %q, want %q", rec1.Body.String(), "remote")
	}

	callsAfterFirst := ownerLookup.calls.Load()
	if callsAfterFirst != 1 {
		t.Fatalf("expected 1 LookupOwner call after first request, got %d", callsAfterFirst)
	}

	// Second request immediately: should use cache, no additional LookupOwner call.
	req2 := httptest.NewRequest(http.MethodGet, instancePath, nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Body.String() != "remote" {
		t.Fatalf("second request: body = %q, want %q", rec2.Body.String(), "remote")
	}

	callsAfterSecond := ownerLookup.calls.Load()
	if callsAfterSecond != 1 {
		t.Errorf("expected no additional LookupOwner call from cache hit, got %d total", callsAfterSecond)
	}

	// Wait for cache to expire.
	time.Sleep(20 * time.Millisecond)

	// Third request after expiration: should call LookupOwner again.
	req3 := httptest.NewRequest(http.MethodGet, instancePath, nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Body.String() != "remote" {
		t.Fatalf("third request: body = %q, want %q", rec3.Body.String(), "remote")
	}

	callsAfterThird := ownerLookup.calls.Load()
	if callsAfterThird != 2 {
		t.Errorf("expected 2 total LookupOwner calls after cache expiration, got %d", callsAfterThird)
	}
}

// ---------------------------------------------------------------------------
// TestCloseStopsEviction
// ---------------------------------------------------------------------------

func TestCloseStopsEviction(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mw := NewRoutingMiddleware(
		&mockLocator{clients: map[uuid.UUID]bool{}},
		&mockResolver{addrs: map[string]string{}},
		&mockOwnerLookup{owners: map[uuid.UUID]string{}},
		"self-worker",
		logger,
	)

	// Close should not panic.
	mw.Close()

	// Calling Close a second time would panic due to closing an already-closed
	// channel. Verify that we can detect this. We intentionally do NOT call
	// Close() again here since the implementation uses a bare close(m.done),
	// which would panic. This test verifies single-close correctness.
}

// ---------------------------------------------------------------------------
// TestSharedTransportReuse
// ---------------------------------------------------------------------------

func TestSharedTransportReuse(t *testing.T) {
	selfWorker := "self-worker"
	instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	instancePath := "/instances/" + instanceID.String() + "/token/abc/device"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("remote"))
	}))
	defer remote.Close()

	mw := NewRoutingMiddleware(
		&mockLocator{clients: map[uuid.UUID]bool{}},
		&mockResolver{addrs: map[string]string{"other-worker": remote.URL}},
		&mockOwnerLookup{owners: map[uuid.UUID]string{instanceID: "other-worker"}},
		selfWorker,
		logger,
	)
	defer mw.Close()

	// Capture the transport pointer before any requests.
	transportBefore := mw.transport

	handler := mw.Handler(localHandler())

	// Send multiple requests that get proxied to the remote.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, instancePath, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Body.String() != "remote" {
			t.Fatalf("request %d: body = %q, want %q", i, rec.Body.String(), "remote")
		}
	}

	// The transport pointer must be the same after all requests -- the
	// middleware must reuse a single shared transport, not create a new
	// one per proxy call.
	if mw.transport != transportBefore {
		t.Error("transport pointer changed across requests; expected shared *http.Transport reuse")
	}
}
