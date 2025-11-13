package whatsmeow

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type stubLock struct{}

func (stubLock) Refresh(context.Context, int) error { return nil }
func (stubLock) Release(context.Context) error      { return nil }
func (stubLock) GetValue() string                   { return "" }

type stubLockManager struct{}

func (stubLockManager) Acquire(context.Context, string, int) (locks.Lock, bool, error) {
	return stubLock{}, true, nil
}

type stubInstanceRepo struct {
	links []StoreLink
}

func (s *stubInstanceRepo) ListInstancesWithStoreJID(context.Context) ([]StoreLink, error) {
	return s.links, nil
}

func (s *stubInstanceRepo) UpdateConnectionStatus(context.Context, uuid.UUID, bool, string, *string, *string) error {
	return nil
}

func (s *stubInstanceRepo) GetConnectionState(context.Context, uuid.UUID) (*ConnectionState, error) {
	return nil, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestClientRegistryAutoConnectRetries(t *testing.T) {
	t.Parallel()

	reg := &ClientRegistry{
		log: newTestLogger(),
		autoConnectCfg: AutoConnectConfig{
			Enabled:        true,
			MaxAttempts:    2,
			InitialBackoff: 5 * time.Millisecond,
			MaxBackoff:     5 * time.Millisecond,
		},
		clients: make(map[uuid.UUID]*clientState),
	}

	instanceID := uuid.New()
	store := "storeJID"
	reg.clients[instanceID] = &clientState{storeJID: &store, client: &whatsmeow.Client{}}

	var (
		mu        sync.Mutex
		attempts  int
		connected bool
		done      = make(chan struct{}, 1)
	)

	reg.isConnectedOverride = func(*whatsmeow.Client) bool {
		mu.Lock()
		defer mu.Unlock()
		return connected
	}

	reg.connectOverride = func(*whatsmeow.Client) error {
		mu.Lock()
		defer mu.Unlock()
		attempts++
		if attempts == 1 {
			return errors.New("temporary failure")
		}
		connected = true
		done <- struct{}{}
		return nil
	}

	reg.maybeStartAutoConnect(instanceID)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("auto-connect did not succeed in time")
	}

	mu.Lock()
	defer mu.Unlock()
	if attempts != 2 {
		t.Fatalf("expected 2 connect attempts, got %d", attempts)
	}
	if !connected {
		t.Fatal("expected client to be marked connected")
	}
}

func TestReconcileOrphanedInstancesSuccess(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New()
	store := "storeJID"

	reg := &ClientRegistry{
		log:         newTestLogger(),
		repo:        &stubInstanceRepo{links: []StoreLink{{ID: instanceID, StoreJID: store}}},
		lockManager: stubLockManager{},
		clients:     make(map[uuid.UUID]*clientState),
	}

	promReg := prometheus.NewRegistry()
	reg.obsMetrics = observability.NewMetrics("test", promReg)

	var (
		ensureCalls  int
		connectCalls int
		connected    bool
	)

	reg.isConnectedOverride = func(*whatsmeow.Client) bool {
		return connected
	}

	reg.connectOverride = func(*whatsmeow.Client) error {
		connectCalls++
		connected = true
		return nil
	}

	reg.ensureWithLockOverride = func(ctx context.Context, info InstanceInfo, _ locks.Lock) (*whatsmeow.Client, bool, error) {
		ensureCalls++
		client := &whatsmeow.Client{}
		reg.mu.Lock()
		reg.clients[info.ID] = &clientState{client: client, storeJID: info.StoreJID}
		reg.mu.Unlock()
		return client, false, nil
	}

	reg.reconcileOrphanedInstances(context.Background(), newTestLogger())

	if ensureCalls != 1 {
		t.Fatalf("expected ensure to be called once, got %d", ensureCalls)
	}
	if connectCalls != 1 {
		t.Fatalf("expected connect to be called once, got %d", connectCalls)
	}

	if got := testutil.ToFloat64(reg.obsMetrics.ReconciliationAttempts.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected success counter to be 1, got %v", got)
	}
	if got := testutil.ToFloat64(reg.obsMetrics.OrphanedInstances); got != 0 {
		t.Fatalf("expected orphaned gauge to be 0, got %v", got)
	}
}
