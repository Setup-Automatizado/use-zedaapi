package whatsmeow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"log/slog"

	retry "github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	pq "github.com/lib/pq"

	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"go.mau.fi/whatsmeow/api/internal/locks"
)

type StoreLink struct {
	ID       uuid.UUID
	StoreJID string
}

type InstanceRepository interface {
	ListInstancesWithStoreJID(ctx context.Context) ([]StoreLink, error)
}

type InstanceInfo struct {
	ID            uuid.UUID
	Name          string
	SessionName   string
	ClientToken   string
	InstanceToken string
	StoreJID      *string
}

type StatusSnapshot struct {
	Connected     bool
	StoreJID      *string
	LastConnected time.Time
	AutoReconnect bool
	WorkerID      string
}

type ClientRegistry struct {
	log           *slog.Logger
	workerID      string
	container     *sqlstore.Container
	lockManager   locks.Manager
	repo          InstanceRepository
	pairCallback  func(context.Context, uuid.UUID, string) error
	resetCallback func(context.Context, uuid.UUID, string) error
	logLevel      string

	mu            sync.RWMutex
	clients       map[uuid.UUID]*clientState
	creationLocks sync.Map

	activeLocks   map[uuid.UUID]locks.Lock
	activeLocksMu sync.RWMutex

	splitBrainTicker  *time.Ticker
	splitBrainStop    chan struct{}
	splitBrainRunning bool
	splitBrainMu      sync.Mutex

	splitBrainCounter func()
}

type clientState struct {
	client         *whatsmeow.Client
	lastConnected  time.Time
	storeJID       *string
	pairing        *pairingSession
	wasNewInstance bool
	createdAt      time.Time
	primarySession bool

	lock              locks.Lock
	lockRefreshCancel context.CancelFunc
	lockMode          string
}

type noOpLock struct{}

func (l *noOpLock) Refresh(ctx context.Context, ttlSeconds int) error {
	return nil
}

func (l *noOpLock) Release(ctx context.Context) error {
	return nil
}

func (l *noOpLock) GetValue() string {
	return ""
}

type contextKey string

const lockHeldContextKey contextKey = "whatsmeow-lock-held"

const (
	pairingSessionTTL  = 20 * time.Second
	newInstanceTimeout = 10 * time.Minute
)

type pairingSession struct {
	cancel context.CancelFunc
	timer  *time.Timer
}

var ErrInstanceAlreadyPaired = errors.New("instance already paired")

func NewClientRegistry(
	ctx context.Context,
	dsn string,
	logLevel string,
	lockManager locks.Manager,
	repo InstanceRepository,
	logger *slog.Logger,
	pairCallback func(context.Context, uuid.UUID, string) error,
	resetCallback func(context.Context, uuid.UUID, string) error,
) (*ClientRegistry, error) {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	if logLevel == "" {
		logLevel = "INFO"
	}

	sqlstore.PostgresArrayWrapper = pq.Array

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open whatsmeow store: %w", err)
	}
	db.SetMaxOpenConns(32)
	db.SetConnMaxLifetime(30 * time.Minute)

	container := sqlstore.NewWithDB(db, "postgres", waLog.Stdout("whatsmeow", logLevel, false))
	if err := container.Upgrade(ctx); err != nil {
		return nil, fmt.Errorf("upgrade whatsmeow schema: %w", err)
	}

	hostname, _ := os.Hostname()
	return &ClientRegistry{
		log:           logger,
		workerID:      hostname,
		container:     container,
		lockManager:   lockManager,
		repo:          repo,
		pairCallback:  pairCallback,
		resetCallback: resetCallback,
		logLevel:      logLevel,
		clients:       make(map[uuid.UUID]*clientState),
		activeLocks:   make(map[uuid.UUID]locks.Lock),
	}, nil
}

func (r *ClientRegistry) EnsureClientWithLock(ctx context.Context, info InstanceInfo, externalLock locks.Lock) (*whatsmeow.Client, bool, error) {
	r.mu.RLock()
	state, ok := r.clients[info.ID]
	if ok && state != nil && state.client != nil {
		if state.client.Store != nil && state.client.Store.ID != nil {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}

		if state.wasNewInstance && time.Since(state.createdAt) < newInstanceTimeout {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}
	}
	needsReset := ok
	r.mu.RUnlock()

	lockInterface, _ := r.creationLocks.LoadOrStore(info.ID, &sync.Mutex{})
	instanceLock := lockInterface.(*sync.Mutex)
	instanceLock.Lock()
	defer instanceLock.Unlock()

	r.mu.RLock()
	state, ok = r.clients[info.ID]
	if ok && state != nil && state.client != nil {
		if state.client.Store != nil && state.client.Store.ID != nil {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}
		if state.wasNewInstance && time.Since(state.createdAt) < newInstanceTimeout {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}
	}
	r.mu.RUnlock()

	if needsReset {
		r.ResetClient(info.ID, "store_missing")
	}

	client, storeReset, lock, err := r.instantiateClientWithLock(ctx, info, externalLock)
	if err != nil {
		return nil, false, err
	}

	wasNew := info.StoreJID == nil || *info.StoreJID == ""

	lockMode := "none"
	if lock != nil {
		lockValue := lock.GetValue()
		if lockValue == "" {
			lockMode = "local"
		} else {
			lockMode = "redis"
		}
	}

	clientState := &clientState{
		client:         client,
		storeJID:       info.StoreJID,
		wasNewInstance: wasNew,
		createdAt:      time.Now().UTC(),
		lock:           lock,
		lockMode:       lockMode,
	}
	if storeReset {
		clientState.storeJID = nil
	}

	if lock != nil {
		r.activeLocksMu.Lock()
		r.activeLocks[info.ID] = lock
		r.activeLocksMu.Unlock()
	}

	r.mu.Lock()
	r.clients[info.ID] = clientState
	r.mu.Unlock()

	if lock != nil {
		cancelFunc := r.startLockRefresh(info.ID, lock, lockMode)
		r.mu.Lock()
		if state, exists := r.clients[info.ID]; exists {
			state.lockRefreshCancel = cancelFunc
		}
		r.mu.Unlock()

		r.log.Info("persistent lock established for client",
			slog.String("instanceId", info.ID.String()),
			slog.String("lockMode", lockMode))
	}

	if storeReset && !needsReset {
		go r.notifyReset(info.ID, "store_missing")
	}

	return client, true, nil
}

func (r *ClientRegistry) EnsureClient(ctx context.Context, info InstanceInfo) (*whatsmeow.Client, bool, error) {
	r.mu.RLock()
	state, ok := r.clients[info.ID]
	if ok && state != nil && state.client != nil {
		if state.client.Store != nil && state.client.Store.ID != nil {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}

		if state.wasNewInstance && time.Since(state.createdAt) < newInstanceTimeout {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}
	}
	needsReset := ok
	r.mu.RUnlock()

	lockInterface, _ := r.creationLocks.LoadOrStore(info.ID, &sync.Mutex{})
	instanceLock := lockInterface.(*sync.Mutex)

	instanceLock.Lock()
	defer instanceLock.Unlock()

	r.mu.RLock()
	state, ok = r.clients[info.ID]
	if ok && state != nil && state.client != nil {
		if state.client.Store != nil && state.client.Store.ID != nil {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}

		if state.wasNewInstance && time.Since(state.createdAt) < newInstanceTimeout {
			client := state.client
			r.mu.RUnlock()
			return client, false, nil
		}
	}
	r.mu.RUnlock()

	if needsReset {
		r.ResetClient(info.ID, "store_missing")
	}

	client, storeReset, lock, err := r.instantiateClientWithLock(ctx, info, nil)
	if err != nil {
		return nil, false, err
	}

	wasNew := info.StoreJID == nil || *info.StoreJID == ""

	lockMode := "none"
	if lock != nil {
		lockValue := lock.GetValue()
		if lockValue == "" {
			lockMode = "local"
		} else {
			lockMode = "redis"
		}
	}

	clientState := &clientState{
		client:         client,
		storeJID:       info.StoreJID,
		wasNewInstance: wasNew,
		createdAt:      time.Now().UTC(),
		lock:           lock,
		lockMode:       lockMode,
	}
	if storeReset {
		clientState.storeJID = nil
	}

	if lock != nil {
		r.activeLocksMu.Lock()
		r.activeLocks[info.ID] = lock
		r.activeLocksMu.Unlock()
	}

	r.mu.Lock()
	r.clients[info.ID] = clientState
	r.mu.Unlock()

	if lock != nil {
		cancelFunc := r.startLockRefresh(info.ID, lock, lockMode)
		r.log.Debug("started lock refresh goroutine",
			slog.String("instanceId", info.ID.String()),
			slog.String("lockMode", lockMode))
		r.mu.Lock()
		if state, exists := r.clients[info.ID]; exists {
			state.lockRefreshCancel = cancelFunc
		}
		r.mu.Unlock()

		r.log.Info("persistent lock established for client",
			slog.String("instanceId", info.ID.String()),
			slog.String("lockMode", lockMode))
	}

	if storeReset && !needsReset {
		go r.notifyReset(info.ID, "store_missing")
	}

	return client, true, nil
}

func (r *ClientRegistry) startLockRefresh(instanceID uuid.UUID, lock locks.Lock, lockMode string) context.CancelFunc {
	if lock == nil {
		r.log.Debug("no lock to refresh (lockManager not configured)", slog.String("instanceId", instanceID.String()))
		return func() {}
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		r.log.Debug("started lock refresh goroutine",
			slog.String("instanceId", instanceID.String()),
			slog.String("lockMode", lockMode))

		for {
			select {
			case <-ctx.Done():
				r.log.Debug("stopping lock refresh goroutine",
					slog.String("instanceId", instanceID.String()))
				return

			case <-ticker.C:
				r.mu.RLock()
				currentState, stateExists := r.clients[instanceID]
				currentLockMode := ""
				if stateExists && currentState != nil {
					currentLockMode = currentState.lockMode
				}
				r.mu.RUnlock()

				if currentLockMode == "local" && stateExists {
					type circuitStateChecker interface {
						GetState() locks.CircuitState
					}

					if checker, ok := r.lockManager.(circuitStateChecker); ok {
						circuitState := checker.GetState()

						r.log.Debug("checking circuit breaker state for lock reacquisition",
							slog.String("instanceId", instanceID.String()),
							slog.String("currentLockMode", currentLockMode),
							slog.Int("circuitState", int(circuitState)))

						if circuitState == 0 {
							r.log.Info("Redis recovered - attempting to reacquire real lock",
								slog.String("instanceId", instanceID.String()))

							lockKey := fmt.Sprintf("funnelchat:instance:%s", instanceID.String())

							var newLock locks.Lock
							var acquired bool
							var acquireErr error

							for attempt := 1; attempt <= 3; attempt++ {
								acquireCtx, acquireCancel := context.WithTimeout(context.Background(), 3*time.Second)
								newLock, acquired, acquireErr = r.lockManager.Acquire(acquireCtx, lockKey, 30)
								acquireCancel()

								if acquireErr == nil && acquired && newLock != nil {
									lockToken := newLock.GetValue()

									if lockToken == "" {
										r.log.Debug("lock reacquisition returned fallback lock - circuit breaker not ready",
											slog.String("instanceId", instanceID.String()),
											slog.Int("attempt", attempt),
											slog.String("reason", "empty_token"))

										if attempt < 3 {
											backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
											r.log.Debug("waiting before retry",
												slog.String("instanceId", instanceID.String()),
												slog.String("backoff", backoff.String()))
											time.Sleep(backoff)
										}
										continue
									}

									r.log.Info("successfully reacquired Redis lock after recovery",
										slog.String("instanceId", instanceID.String()),
										slog.Int("attempt", attempt),
										slog.String("lockToken", func() string {
											if len(lockToken) > 8 {
												return lockToken[:8] + "..."
											}
											return lockToken
										}()))

									r.mu.Lock()
									r.activeLocksMu.Lock()
									if state, exists := r.clients[instanceID]; exists && state != nil {
										if state.lock != nil {
											state.lock.Release(context.Background())
										}

										state.lock = newLock
										state.lockMode = "redis"

										r.activeLocks[instanceID] = newLock
									}
									r.activeLocksMu.Unlock()
									r.mu.Unlock()

									lock = newLock
									break
								}

								if attempt < 3 {
									backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
									r.log.Debug("lock reacquisition failed, retrying",
										slog.String("instanceId", instanceID.String()),
										slog.Int("attempt", attempt),
										slog.String("backoff", backoff.String()),
										slog.String("error", func() string {
											if acquireErr != nil {
												return acquireErr.Error()
											}
											return "lock not acquired"
										}()))
									time.Sleep(backoff)
								}
							}

							if acquireErr != nil || !acquired || newLock == nil {
								r.log.Warn("failed to reacquire Redis lock after recovery (all attempts)",
									slog.String("instanceId", instanceID.String()),
									slog.String("error", func() string {
										if acquireErr != nil {
											return acquireErr.Error()
										}
										if !acquired {
											return "lock held by another replica"
										}
										return "unknown error"
									}()))
							}
						}
					}
				}

				currentToken := lock.GetValue()

				if currentToken == "" {
					r.log.Debug("skipping lock refresh - have noOpLock (fallback mode)",
						slog.String("instanceId", instanceID.String()),
						slog.String("lockMode", currentLockMode))
					continue
				}

				refreshCtx, refreshCancel := context.WithTimeout(context.Background(), 2*time.Second)
				err := lock.Refresh(refreshCtx, 30) // Refresh for 30 seconds
				refreshCancel()

				if err != nil {
					errStr := err.Error()
					isConnectionError := strings.Contains(errStr, "connection refused") ||
						strings.Contains(errStr, "connection reset") ||
						strings.Contains(errStr, "i/o timeout") ||
						strings.Contains(errStr, "EOF") ||
						strings.Contains(errStr, "broken pipe") ||
						strings.Contains(errStr, "network is unreachable") ||
						strings.Contains(errStr, "no route to host") ||
						strings.Contains(errStr, "context deadline exceeded")

					if isConnectionError {
						r.log.Warn("lock refresh failed due to Redis connection issue - switching to local mode",
							slog.String("instanceId", instanceID.String()),
							slog.String("error", err.Error()))

						r.mu.Lock()
						if state, exists := r.clients[instanceID]; exists && state != nil {
							state.lockMode = "local"
						}
						r.mu.Unlock()

						continue
					}

					type circuitStateChecker interface {
						GetState() locks.CircuitState
					}

					var circuitOpen bool
					if checker, ok := r.lockManager.(circuitStateChecker); ok {
						state := checker.GetState()
						circuitOpen = (state != 0)
					}

					if circuitOpen {
						r.log.Debug("lock refresh failed but circuit breaker open - continuing",
							slog.String("instanceId", instanceID.String()),
							slog.String("error", err.Error()))
						r.mu.Lock()
						if state, exists := r.clients[instanceID]; exists && state != nil {
							state.lockMode = "local"
						}
						r.mu.Unlock()
					} else {
						r.log.Error("CRITICAL: lock refresh failed with circuit breaker CLOSED - SPLIT-BRAIN DETECTED",
							slog.String("instanceId", instanceID.String()),
							slog.String("error", err.Error()))

						r.mu.Lock()
						r.activeLocksMu.Lock()

						state, exists := r.clients[instanceID]
						delete(r.clients, instanceID)
						delete(r.activeLocks, instanceID)

						r.activeLocksMu.Unlock()
						r.mu.Unlock()

						if exists && state != nil {
							if state.lockRefreshCancel != nil {
								state.lockRefreshCancel()
							}

							if state.lock != nil {
								if releaseErr := state.lock.Release(context.Background()); releaseErr != nil {
									r.log.Warn("failed to release lock after lock loss detection",
										slog.String("instanceId", instanceID.String()),
										slog.String("error", releaseErr.Error()))
								}
							}

							if state.client != nil {
								state.client.Disconnect()
								r.log.Warn("forcefully disconnected client due to lock loss",
									slog.String("instanceId", instanceID.String()))
							}
						}

						return
					}
				} else {
					r.log.Debug("lock refreshed successfully",
						slog.String("instanceId", instanceID.String()))

					r.mu.Lock()
					if state, exists := r.clients[instanceID]; exists && state != nil {
						state.lockMode = "redis"
					}
					r.mu.Unlock()
				}
			}
		}
	}()

	return cancel
}

func (r *ClientRegistry) withDatabaseRetry(ctx context.Context, operation func() error) error {
	return retry.Do(
		operation,
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(100*time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			if err == nil {
				return false
			}
			if errors.Is(err, context.DeadlineExceeded) ||
				errors.Is(err, context.Canceled) {
				return false
			}
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "08000", "08003", "08006", "08001", "08004",
					"53000", "53100", "53200", "53300", "53400",
					"57P03", "40001":
					return true
				default:
					return false
				}
			}
			errStr := err.Error()
			return strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "connection reset") ||
				strings.Contains(errStr, "timeout") ||
				strings.Contains(errStr, "temporary failure") ||
				strings.Contains(errStr, "no such host")
		}),
		retry.OnRetry(func(attempt uint, err error) {
			r.log.Warn("database operation failed, retrying",
				slog.Uint64("attempt", uint64(attempt)),
				slog.String("error", err.Error()))
		}),
	)
}

func (r *ClientRegistry) instantiateClientWithLock(ctx context.Context, info InstanceInfo, externalLock locks.Lock) (*whatsmeow.Client, bool, locks.Lock, error) {
	var lock locks.Lock = externalLock
	var err error
	lockAlreadyHeld := externalLock != nil
	if ctx != nil {
		if held, ok := ctx.Value(lockHeldContextKey).(bool); ok && held {
			lockAlreadyHeld = true
		}
	}

	if r.lockManager == nil {
		r.log.Warn("lock manager not configured - multi-replica coordination disabled",
			slog.String("instanceId", info.ID.String()))
	} else if !lockAlreadyHeld {
		key := fmt.Sprintf("funnelchat:instance:%s", info.ID.String())
		var acquired bool
		lock, acquired, err = r.lockManager.Acquire(ctx, key, 30)
		if err != nil {
			return nil, false, nil, fmt.Errorf("failed to acquire redis lock (redis unavailable): %w", err)
		}
		if !acquired {
			return nil, false, nil, fmt.Errorf("instance lock already held by another replica")
		}
		r.log.Debug("redis lock acquired", slog.String("instanceId", info.ID.String()))
	}

	var deviceStore *store.Device
	storeReset := false
	if info.StoreJID != nil {
		jid, parseErr := types.ParseJID(*info.StoreJID)
		if parseErr == nil {
			err = r.withDatabaseRetry(ctx, func() error {
				var getErr error
				deviceStore, getErr = r.container.GetDevice(ctx, jid)
				return getErr
			})
			if err != nil {
				return nil, false, lock, fmt.Errorf("load device store: %w", err)
			}
			if deviceStore == nil {
				storeReset = true
			}
		} else {
			storeReset = true
		}
	}

	if deviceStore == nil {
		deviceStore = r.container.NewDevice()
		if info.StoreJID != nil {
			storeReset = true
		}
	}

	store.SetOSInfo("macOS", [3]uint32{10, 0, 0})
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_SAFARI.Enum()

	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("instance-"+info.ID.String(), logLevelOrDefault(r.logLevel), false))
	client.EnableAutoReconnect = true
	client.AddEventHandler(r.wrapEventHandler(info.ID))

	return client, storeReset, lock, nil
}

func logLevelOrDefault(level string) string {
	if level == "" {
		return "INFO"
	}
	return level
}

func (r *ClientRegistry) registerPairingSession(instanceID uuid.UUID, cancel context.CancelFunc) {
	var previous *pairingSession
	var session *pairingSession

	r.mu.Lock()
	state, ok := r.clients[instanceID]
	if ok {
		previous = state.pairing
		session = &pairingSession{cancel: cancel}
		state.pairing = session
	}
	r.mu.Unlock()

	if previous != nil {
		if previous.timer != nil {
			previous.timer.Stop()
		}
		if previous.cancel != nil {
			previous.cancel()
		}
	}

	if !ok {
		cancel()
		return
	}

	timer := time.AfterFunc(pairingSessionTTL, func() {
		r.cleanupPairingSession(instanceID, "timeout")
	})

	r.mu.Lock()
	if state, ok := r.clients[instanceID]; ok && state.pairing == session {
		state.pairing.timer = timer
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()
	timer.Stop()
	cancel()
}

func (r *ClientRegistry) cleanupPairingSession(instanceID uuid.UUID, reason string) {
	var session *pairingSession
	var client *whatsmeow.Client

	r.mu.Lock()
	if state, ok := r.clients[instanceID]; ok {
		session = state.pairing
		state.pairing = nil
		if reason == "success" {
			client = state.client
		}
	}
	r.mu.Unlock()

	if session == nil {
		return
	}
	if session.timer != nil && reason != "timeout" {
		session.timer.Stop()
	}
	if session.cancel != nil {
		session.cancel()
	}

	switch reason {
	case "success":
		r.log.Debug("pairing session completed", slog.String("instanceId", instanceID.String()))
		if client != nil {
			go r.reconnectAfterPair(instanceID, client)
		}
	case "timeout":
		r.log.Warn("pairing session timed out", slog.String("instanceId", instanceID.String()))
	case "manual":
		r.log.Info("pairing session canceled", slog.String("instanceId", instanceID.String()))
	}
}

func (r *ClientRegistry) notifyReset(instanceID uuid.UUID, reason string) {
	if r.resetCallback == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.resetCallback(ctx, instanceID, reason); err != nil {
		r.log.Error(
			"reset callback",
			slog.String("instanceId", instanceID.String()),
			slog.String("reason", reason),
			slog.String("error", err.Error()),
		)
	}
}

func (r *ClientRegistry) ResetClient(instanceID uuid.UUID, reason string) bool {
	return r.resetClient(instanceID, reason, true)
}

func (r *ClientRegistry) RemoveClient(instanceID uuid.UUID, reason string) bool {
	return r.resetClient(instanceID, reason, false)
}

func (r *ClientRegistry) resetClient(instanceID uuid.UUID, reason string, deleteStore bool) bool {
	r.cleanupPairingSession(instanceID, "manual")

	r.mu.Lock()
	state, ok := r.clients[instanceID]
	if !ok {
		r.mu.Unlock()
		return false
	}
	delete(r.clients, instanceID)
	r.mu.Unlock()

	if state == nil {
		return false
	}

	r.log.Info(
		"resetting client",
		slog.String("instanceId", instanceID.String()),
		slog.String("reason", reason),
	)

	if state.lockRefreshCancel != nil {
		r.log.Debug("stopping lock refresh goroutine",
			slog.String("instanceId", instanceID.String()))
		state.lockRefreshCancel()
	}

	if state.lock != nil {
		r.log.Debug("releasing redis lock",
			slog.String("instanceId", instanceID.String()),
			slog.String("lockMode", state.lockMode))

		if err := state.lock.Release(context.Background()); err != nil {
			r.log.Warn("failed to release redis lock during client reset",
				slog.String("instanceId", instanceID.String()),
				slog.String("error", err.Error()))
		}

		r.activeLocksMu.Lock()
		delete(r.activeLocks, instanceID)
		r.activeLocksMu.Unlock()
	}

	if state.client != nil {
		state.client.EnableAutoReconnect = false
		state.client.Disconnect()
		if deleteStore && state.client.Store != nil && state.client.Store.ID != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := state.client.Store.Delete(ctx); err != nil {
				r.log.Error(
					"delete store device",
					slog.String("instanceId", instanceID.String()),
					slog.String("reason", reason),
					slog.String("error", err.Error()),
				)
			}
			cancel()
		}
	}

	if deleteStore {
		go r.notifyReset(instanceID, reason)
	}

	return true
}

func (r *ClientRegistry) reconnectAfterPair(instanceID uuid.UUID, client *whatsmeow.Client) {
	if client == nil {
		return
	}

	time.Sleep(200 * time.Millisecond)

	if client.IsLoggedIn() {
		return
	}

	if client.IsConnected() {
		r.log.Debug("websocket already connected after pairing, waiting for authentication",
			slog.String("instanceId", instanceID.String()))
		return
	}

	const maxAttempts = 3
	backoff := 200 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if client.IsLoggedIn() {
			return
		}
		if client.IsConnected() {
			r.log.Debug("websocket connected during retry, waiting for authentication",
				slog.String("instanceId", instanceID.String()))
			return
		}
		if err := client.Connect(); err != nil {
			lastErr = err
			r.log.Warn(
				"connect after pairing",
				slog.String("instanceId", instanceID.String()),
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()),
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		r.log.Info("instance connected after pairing", slog.String("instanceId", instanceID.String()))
		return
	}

	if lastErr != nil {
		r.log.Error("failed to connect after pairing", slog.String("instanceId", instanceID.String()), slog.String("error", lastErr.Error()))
	}
}

func (r *ClientRegistry) ensurePrimarySignalSession(instanceID uuid.UUID, client *whatsmeow.Client) {
	if client == nil || client.Store == nil {
		return
	}

	r.log.Debug("ensuring primary signal session", slog.String("instanceId", instanceID.String()))

	jid := client.Store.GetJID()
	if jid.IsEmpty() {
		return
	}
	primaryPN := jid.ToNonAD()

	lid := client.Store.GetLID()
	if lid.IsEmpty() || lid.Server != types.HiddenUserServer {
		return
	}
	primaryLID := lid
	primaryLID.Device = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.StoreLIDPNMapping(ctx, primaryLID, primaryPN)

	hasSession, err := client.Store.ContainsSession(ctx, primaryLID.SignalAddress())
	if err != nil {
		r.log.Debug("failed to check primary session", slog.String("instanceId", instanceID.String()), slog.String("error", err.Error()))
		return
	}
	if hasSession {
		r.mu.Lock()
		if state, ok := r.clients[instanceID]; ok {
			state.primarySession = true
		}
		r.mu.Unlock()
		return
	}

	prekeys, err := client.DangerousInternals().FetchPreKeys(ctx, []types.JID{primaryLID})
	if err != nil {
		r.log.Debug("fetch prekeys failed", slog.String("instanceId", instanceID.String()), slog.String("error", err.Error()))
		return
	}

	respVal := reflect.ValueOf(prekeys).MapIndex(reflect.ValueOf(primaryLID))
	if !respVal.IsValid() {
		r.log.Debug("primary prekey response missing", slog.String("instanceId", instanceID.String()))
		return
	}

	respPtr := reflect.New(respVal.Type())
	respPtr.Elem().Set(respVal)
	bundleField := respPtr.Elem().FieldByName("bundle")
	errField := respPtr.Elem().FieldByName("err")

	var fetchErr error
	if errField.IsValid() && !errField.IsNil() {
		fetchErr = *(*error)(unsafe.Pointer(errField.UnsafeAddr()))
	}
	if fetchErr != nil {
		r.log.Debug("primary prekey error", slog.String("instanceId", instanceID.String()), slog.String("error", fetchErr.Error()))
		return
	}

	if !bundleField.IsValid() || bundleField.IsNil() {
		r.log.Debug("primary prekey bundle empty", slog.String("instanceId", instanceID.String()))
		return
	}
	bundle := (*prekey.Bundle)(unsafe.Pointer(bundleField.Pointer()))

	_, _, encErr := client.DangerousInternals().EncryptMessageForDevice(ctx, []byte{0x00}, primaryLID, bundle, nil)
	if encErr != nil {
		r.log.Debug("failed to prime primary session", slog.String("instanceId", instanceID.String()), slog.String("error", encErr.Error()))
		return
	}

	if ok, _ := client.Store.ContainsSession(ctx, primaryLID.SignalAddress()); ok {
		r.log.Debug("primary session established", slog.String("instanceId", instanceID.String()))
		r.mu.Lock()
		if state, ok := r.clients[instanceID]; ok {
			state.primarySession = true
		}
		r.mu.Unlock()
	}
}

func (r *ClientRegistry) wrapEventHandler(instanceID uuid.UUID) func(evt interface{}) {
	return func(evt interface{}) {
		switch e := evt.(type) {
		case *events.Connected:
			r.mu.Lock()
			var (
				client       *whatsmeow.Client
				sessionReady bool
			)
			if state, ok := r.clients[instanceID]; ok {
				state.lastConnected = time.Now().UTC()
				if state.client != nil && state.client.Store != nil && state.client.Store.ID != nil {
					jid := state.client.Store.ID.String()
					state.storeJID = &jid
				}
				client = state.client
				sessionReady = state.primarySession
			}
			r.mu.Unlock()
			if client != nil && !sessionReady {
				r.ensurePrimarySignalSession(instanceID, client)
			}
		case *events.PairSuccess:
			jid := e.ID.String()
			r.mu.Lock()
			if state, ok := r.clients[instanceID]; ok {
				state.storeJID = &jid
				state.wasNewInstance = false
				state.primarySession = false
			}
			r.mu.Unlock()
			r.cleanupPairingSession(instanceID, "success")
			if r.pairCallback != nil {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := r.pairCallback(ctx, instanceID, jid); err != nil {
						r.log.Error("pair callback", slog.String("instanceId", instanceID.String()), slog.String("error", err.Error()))
					}
				}()
			}
		case *events.Disconnected:
			r.log.Warn("instance disconnected", slog.String("instanceId", instanceID.String()))
		case *events.LoggedOut:
			reason := "logged_out"
			if desc := e.Reason.String(); desc != "" {
				reason = "logged_out:" + desc
			}
			r.log.Warn("instance logged out", slog.String("instanceId", instanceID.String()), slog.String("reason", reason))
			r.ResetClient(instanceID, reason)
		case *events.PairError:
			if e.Error != nil && (errors.Is(e.Error, whatsmeow.ErrPairInvalidDeviceIdentityHMAC) || errors.Is(e.Error, whatsmeow.ErrPairInvalidDeviceSignature)) {
				reason := "pair_error"
				r.log.Warn(
					"pairing error reset",
					slog.String("instanceId", instanceID.String()),
					slog.String("error", e.Error.Error()),
				)
				r.ResetClient(instanceID, reason)
			}
		default:
			_ = e
		}
	}
}

func (r *ClientRegistry) Status(info InstanceInfo) StatusSnapshot {
	r.mu.RLock()
	state, ok := r.clients[info.ID]
	r.mu.RUnlock()
	if !ok {
		return StatusSnapshot{Connected: false, StoreJID: info.StoreJID, WorkerID: r.workerID}
	}

	var storeJID *string
	if state.storeJID != nil {
		storeJID = state.storeJID
	} else if state.client != nil && state.client.Store != nil && state.client.Store.ID != nil {
		jid := state.client.Store.ID.String()
		storeJID = &jid
	}

	return StatusSnapshot{
		Connected:     state.client != nil && state.client.IsLoggedIn(),
		StoreJID:      storeJID,
		LastConnected: state.lastConnected,
		AutoReconnect: state.client != nil && state.client.EnableAutoReconnect,
		WorkerID:      r.workerID,
	}
}

func (r *ClientRegistry) Restart(ctx context.Context, info InstanceInfo) error {
	client, _, err := r.EnsureClient(ctx, info)
	if err != nil {
		return err
	}
	client.Disconnect()

	time.Sleep(100 * time.Millisecond)

	if client.IsConnected() {
		r.log.Debug("client auto-reconnected after disconnect",
			slog.String("instanceId", info.ID.String()))
		return nil
	}

	return client.Connect()
}

func (r *ClientRegistry) Disconnect(ctx context.Context, info InstanceInfo) error {
	r.log.Info("performing complete disconnect with cleanup",
		slog.String("instanceId", info.ID.String()))

	r.cleanupPairingSession(info.ID, "manual")

	r.mu.Lock()
	state, ok := r.clients[info.ID]
	if !ok {
		r.mu.Unlock()
		r.log.Debug("instance not found in registry during disconnect",
			slog.String("instanceId", info.ID.String()))
		return nil
	}
	delete(r.clients, info.ID)
	r.mu.Unlock()

	if state == nil {
		return nil
	}

	if state.lockRefreshCancel != nil {
		r.log.Debug("stopping lock refresh goroutine",
			slog.String("instanceId", info.ID.String()))
		state.lockRefreshCancel()
	}

	if state.lock != nil {
		r.log.Debug("releasing redis lock",
			slog.String("instanceId", info.ID.String()),
			slog.String("lockMode", state.lockMode))

		if err := state.lock.Release(context.Background()); err != nil {
			r.log.Warn("failed to release redis lock during disconnect",
				slog.String("instanceId", info.ID.String()),
				slog.String("error", err.Error()))
		}

		r.activeLocksMu.Lock()
		delete(r.activeLocks, info.ID)
		r.activeLocksMu.Unlock()
	}

	if state.client != nil {
		state.client.EnableAutoReconnect = false
		state.client.Disconnect()

		if state.client.Store != nil && state.client.Store.ID != nil {
			deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := state.client.Store.Delete(deleteCtx); err != nil {
				r.log.Error("delete store device during disconnect",
					slog.String("instanceId", info.ID.String()),
					slog.String("error", err.Error()))
			} else {
				r.log.Info("device store deleted during disconnect",
					slog.String("instanceId", info.ID.String()))
			}
		}
	}

	r.log.Info("complete disconnect finished",
		slog.String("instanceId", info.ID.String()))

	return nil
}

func (r *ClientRegistry) GetQRCode(ctx context.Context, info InstanceInfo) (string, error) {
	_, qrChan, err := r.startPairingSession(ctx, info)
	if err != nil {
		return "", err
	}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case item, ok := <-qrChan:
			if !ok {
				return "", errors.New("qr channel closed")
			}
			if item.Event == whatsmeow.QRChannelEventCode {
				return item.Code, nil
			} else if item.Event == whatsmeow.QRChannelEventError {
				return "", item.Error
			} else if item.Event != "success" {
				r.log.Info("qr channel event", slog.String("instanceId", info.ID.String()), slog.String("event", item.Event))
			}
		}
	}
}

func (r *ClientRegistry) PairPhone(ctx context.Context, info InstanceInfo, phone string) (string, error) {
	client, _, err := r.EnsureClient(ctx, info)
	if err != nil {
		return "", err
	}
	if client.Store != nil && client.Store.ID != nil {
		return "", ErrInstanceAlreadyPaired
	}

	if !client.IsConnected() {
		r.log.Info("connecting to websocket before generating pairing code",
			slog.String("instanceId", info.ID.String()))

		if connErr := client.Connect(); connErr != nil {
			r.log.Error("failed to connect before pair phone",
				slog.String("instanceId", info.ID.String()),
				slog.String("error", connErr.Error()))
			return "", fmt.Errorf("failed to connect websocket: %w", connErr)
		}

		time.Sleep(500 * time.Millisecond)
	}

	pairCtx := context.Background()

	code, err := client.PairPhone(pairCtx, phone, true, whatsmeow.PairClientSafari, "Safari (macOS)")

	if err != nil {
		return "", fmt.Errorf("failed to generate pairing code: %w", err)
	}

	r.log.Info("pairing code generated successfully",
		slog.String("instanceId", info.ID.String()),
		slog.String("phone", phone))

	return code, nil
}

func (r *ClientRegistry) startPairingSession(ctx context.Context, info InstanceInfo) (*whatsmeow.Client, <-chan whatsmeow.QRChannelItem, error) {
	client, _, err := r.EnsureClient(ctx, info)
	if err != nil {
		return nil, nil, err
	}
	if client.Store != nil && client.Store.ID != nil {
		return nil, nil, ErrInstanceAlreadyPaired
	}

	pairCtx, cancel := context.WithCancel(context.Background())
	qrChan, err := client.GetQRChannel(pairCtx)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	r.registerPairingSession(info.ID, cancel)
	go func() {
		if !client.IsConnected() {
			if err := client.Connect(); err != nil {
				r.log.Error("connect for qr", slog.String("instanceId", info.ID.String()), slog.String("error", err.Error()))
			}
		} else {
			r.log.Debug("websocket already connected for qr pairing", slog.String("instanceId", info.ID.String()))
		}
	}()
	return client, qrChan, nil
}

func (r *ClientRegistry) HasStoreDevice(ctx context.Context, storeJID string) (bool, error) {
	if storeJID == "" {
		return false, nil
	}
	jid, err := types.ParseJID(storeJID)
	if err != nil {
		return false, fmt.Errorf("parse store jid: %w", err)
	}
	if r.container == nil {
		return false, errors.New("whatsmeow container not initialised")
	}

	var device *store.Device
	err = r.withDatabaseRetry(ctx, func() error {
		var getErr error
		device, getErr = r.container.GetDevice(ctx, jid)
		return getErr
	})
	if err != nil {
		return false, fmt.Errorf("lookup store device: %w", err)
	}
	return device != nil && device.ID != nil, nil
}

func (r *ClientRegistry) ConnectExistingClients(ctx context.Context) (connected, skipped int, err error) {
	if r.repo == nil {
		return 0, 0, fmt.Errorf("repository not configured")
	}

	links, err := r.repo.ListInstancesWithStoreJID(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("list instances: %w", err)
	}

	if len(links) == 0 {
		r.log.Info("no paired instances to reconnect")
		return 0, 0, nil
	}

	r.log.Info("starting reconnection of paired instances", slog.Int("total", len(links)))

	for _, link := range links {
		key := fmt.Sprintf("funnelchat:instance:%s", link.ID.String())
		lock, acquired, err := r.lockManager.Acquire(ctx, key, 30)

		if err != nil {
			errStr := err.Error()
			isConnectionError := strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "connection reset") ||
				strings.Contains(errStr, "i/o timeout") ||
				strings.Contains(errStr, "EOF") ||
				strings.Contains(errStr, "broken pipe") ||
				strings.Contains(errStr, "network is unreachable") ||
				strings.Contains(errStr, "no route to host") ||
				strings.Contains(errStr, "context deadline exceeded")

			if isConnectionError {
				r.log.Warn("Redis unavailable during startup - using local mode",
					slog.String("instanceId", link.ID.String()),
					slog.String("error", err.Error()))
				lock = &noOpLock{}
				acquired = true
			} else {
				r.log.Debug("skipping instance (lock error)",
					slog.String("instanceId", link.ID.String()),
					slog.String("error", err.Error()))
				skipped++
				continue
			}
		}

		if !acquired {
			r.log.Debug("skipping instance (lock held by another replica)",
				slog.String("instanceId", link.ID.String()))
			skipped++
			continue
		}

		info := InstanceInfo{
			ID:       link.ID,
			StoreJID: &link.StoreJID,
		}

		client, created, connectErr := r.EnsureClientWithLock(ctx, info, lock)
		if connectErr != nil {
			r.log.Error("ensure client failed",
				slog.String("instanceId", link.ID.String()),
				slog.String("error", connectErr.Error()))
			if lock != nil {
				lock.Release(context.Background())
			}
			skipped++
			continue
		}

		if client.IsConnected() {
			r.log.Debug("instance already connected",
				slog.String("instanceId", link.ID.String()))
			connected++
			continue
		}

		if err := client.Connect(); err != nil {
			r.log.Warn("failed to connect instance",
				slog.String("instanceId", link.ID.String()),
				slog.String("error", err.Error()),
				slog.Bool("wasCreated", created))

			r.mu.Lock()
			if state, exists := r.clients[link.ID]; exists {
				if state.lockRefreshCancel != nil {
					state.lockRefreshCancel()
				}
				if state.lock != nil {
					state.lock.Release(context.Background())
				}
				delete(r.clients, link.ID)
				r.activeLocksMu.Lock()
				delete(r.activeLocks, link.ID)
				r.activeLocksMu.Unlock()
			}
			r.mu.Unlock()

			skipped++
			continue
		}

		r.log.Info("connected instance",
			slog.String("instanceId", link.ID.String()),
			slog.Bool("wasCreated", created))
		connected++

	}

	r.log.Info("reconnection complete",
		slog.Int("connected", connected),
		slog.Int("skipped", skipped),
		slog.Int("total", len(links)))

	return connected, skipped, nil
}

func (r *ClientRegistry) DisconnectAll(ctx context.Context) (disconnected int) {
	r.mu.RLock()
	clientIDs := make([]uuid.UUID, 0, len(r.clients))
	for id := range r.clients {
		clientIDs = append(clientIDs, id)
	}
	r.mu.RUnlock()

	if len(clientIDs) == 0 {
		r.log.Info("no active clients to disconnect")
		return 0
	}

	r.log.Info("starting graceful disconnect of all clients", slog.Int("total", len(clientIDs)))

	var wg sync.WaitGroup
	var disconnectedCount int
	var countMu sync.Mutex

	for _, id := range clientIDs {
		r.mu.RLock()
		state, ok := r.clients[id]
		r.mu.RUnlock()

		if !ok || state == nil || state.client == nil {
			continue
		}

		if !state.client.IsConnected() {
			r.log.Debug("client already disconnected", slog.String("instanceId", id.String()))
			continue
		}

		wg.Add(1)
		go func(clientID uuid.UUID, client *whatsmeow.Client) {
			defer wg.Done()

			disconnectCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			done := make(chan struct{})

			go func() {
				client.Disconnect()
				close(done)
			}()

			select {
			case <-done:
				r.log.Info("disconnected client", slog.String("instanceId", clientID.String()))
				countMu.Lock()
				disconnectedCount++
				countMu.Unlock()
			case <-disconnectCtx.Done():
				r.log.Warn("client disconnect timeout - continuing shutdown",
					slog.String("instanceId", clientID.String()))
			}
		}(id, state.client)
	}

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		r.log.Info("graceful disconnect complete", slog.Int("disconnected", disconnectedCount))
	case <-ctx.Done():
		r.log.Warn("disconnect all timeout - some clients may not have disconnected",
			slog.Int("disconnected", disconnectedCount),
			slog.Int("total", len(clientIDs)))
	}

	return disconnectedCount
}

func (r *ClientRegistry) SetSplitBrainMetrics(splitBrainCounter func()) {
	r.splitBrainMu.Lock()
	defer r.splitBrainMu.Unlock()
	r.splitBrainCounter = splitBrainCounter
}

func (r *ClientRegistry) StartSplitBrainDetection() {
	r.splitBrainMu.Lock()
	defer r.splitBrainMu.Unlock()

	if r.splitBrainRunning {
		r.log.Warn("split-brain detection already running")
		return
	}

	r.splitBrainTicker = time.NewTicker(60 * time.Second)
	r.splitBrainStop = make(chan struct{})
	r.splitBrainRunning = true

	r.log.Info("started split-brain detection worker", slog.Duration("interval", 60*time.Second))

	go func() {
		for {
			select {
			case <-r.splitBrainTicker.C:
				r.detectSplitBrain()
			case <-r.splitBrainStop:
				r.log.Info("split-brain detection worker stopped")
				return
			}
		}
	}()
}

func (r *ClientRegistry) StopSplitBrainDetection() {
	r.splitBrainMu.Lock()
	defer r.splitBrainMu.Unlock()

	if !r.splitBrainRunning {
		return
	}

	r.splitBrainRunning = false
	close(r.splitBrainStop)
	if r.splitBrainTicker != nil {
		r.splitBrainTicker.Stop()
	}
}

func (r *ClientRegistry) detectSplitBrain() {
	r.mu.RLock()
	clientIDs := make([]uuid.UUID, 0, len(r.clients))
	for id := range r.clients {
		clientIDs = append(clientIDs, id)
	}
	r.mu.RUnlock()

	if len(clientIDs) == 0 {
		return
	}

	r.mu.RLock()
	redisLockCount := 0
	localLockCount := 0
	invalidLockCount := 0

	for id, state := range r.clients {
		if state != nil {
			if state.lockMode == "redis" {
				r.activeLocksMu.RLock()
				lock, hasLock := r.activeLocks[id]
				r.activeLocksMu.RUnlock()

				if hasLock && lock != nil && lock.GetValue() != "" {
					redisLockCount++
				} else {
					r.log.Warn("client marked as redis mode but lock has no token",
						slog.String("instanceId", id.String()),
						slog.String("reportedMode", state.lockMode))

					invalidLockCount++
					localLockCount++
				}
			} else if state.lockMode == "local" {
				localLockCount++
			}
		}
	}
	r.mu.RUnlock()

	r.log.Debug("running split-brain detection",
		slog.Int("clients", len(clientIDs)),
		slog.Int("redisLocks", redisLockCount),
		slog.Int("localLocks", localLockCount),
		slog.Int("invalidLocks", invalidLockCount))

	if redisLockCount == 0 && localLockCount > 0 {
		r.log.Debug("skipping split-brain detection - all locks in local mode",
			slog.Int("localLocks", localLockCount),
			slog.String("reason", "Redis unavailable for all instances, single-replica fallback mode"))
		return
	}

	if redisLockCount > 0 {
		r.log.Debug("proceeding with split-brain detection - have Redis locks",
			slog.Int("redisLocks", redisLockCount),
			slog.Int("localLocks", localLockCount))
	}

	splitBrainDetected := 0
	for _, instanceID := range clientIDs {
		r.activeLocksMu.RLock()
		expectedLock, haveLock := r.activeLocks[instanceID]
		r.activeLocksMu.RUnlock()

		if !haveLock {
			r.log.Debug("client exists but no lock tracked in activeLocks",
				slog.String("instanceId", instanceID.String()))
			continue
		}

		lockKey := fmt.Sprintf("funnelchat:instance:%s", instanceID.String())
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		var stillOwned bool
		var err error

		type lockOwnershipChecker interface {
			CheckLockOwnership(ctx context.Context, key string, expectedLock locks.Lock) (bool, error)
		}

		if checker, ok := r.lockManager.(lockOwnershipChecker); ok {
			stillOwned, err = checker.CheckLockOwnership(ctx, lockKey, expectedLock)
		} else {
			err = expectedLock.Refresh(ctx, 30)
			stillOwned = (err == nil)
		}
		cancel()

		if err != nil {
			r.log.Debug("error verifying lock ownership during split-brain detection",
				slog.String("instanceId", instanceID.String()),
				slog.String("error", err.Error()))

			stillOwned = false
		}

		if !stillOwned {
			r.log.Error("SPLIT-BRAIN DETECTED: Client exists but lock ownership lost",
				slog.String("instanceId", instanceID.String()),
				slog.String("workerID", r.workerID))

			splitBrainDetected++

			if r.splitBrainCounter != nil {
				r.splitBrainCounter()
			}

			r.mu.Lock()
			r.activeLocksMu.Lock()

			state, exists := r.clients[instanceID]
			delete(r.clients, instanceID)
			delete(r.activeLocks, instanceID)

			r.activeLocksMu.Unlock()
			r.mu.Unlock()

			r.log.Warn("removed split-brain client from registry",
				slog.String("instanceId", instanceID.String()))

			if exists && state != nil {
				if state.lockRefreshCancel != nil {
					state.lockRefreshCancel()
				}

				if state.lock != nil {
					if releaseErr := state.lock.Release(context.Background()); releaseErr != nil {
						r.log.Warn("failed to release lock after split-brain detection",
							slog.String("instanceId", instanceID.String()),
							slog.String("error", releaseErr.Error()))
					}
				}

				if state.client != nil {
					state.client.Disconnect()
					r.log.Warn("forcefully disconnected split-brain client",
						slog.String("instanceId", instanceID.String()))
				}
			}
		}
	}

	if splitBrainDetected > 0 {
		r.log.Error("split-brain detection completed",
			slog.Int("split_brain_clients", splitBrainDetected),
			slog.Int("total_checked", len(clientIDs)))
	}
}

func (r *ClientRegistry) ReleaseAllLocks() int {
	if r == nil || r.lockManager == nil {
		return 0
	}

	r.activeLocksMu.Lock()
	locks := make([]locks.Lock, 0, len(r.activeLocks))
	instanceIDs := make([]uuid.UUID, 0, len(r.activeLocks))

	for id, lock := range r.activeLocks {
		locks = append(locks, lock)
		instanceIDs = append(instanceIDs, id)
	}

	for id := range r.activeLocks {
		delete(r.activeLocks, id)
	}
	r.activeLocksMu.Unlock()

	if len(locks) == 0 {
		r.log.Info("no active redis locks to release")
		return 0
	}

	r.log.Info("releasing all redis locks", slog.Int("count", len(locks)))

	released := 0
	for i, lock := range locks {
		if err := lock.Release(context.Background()); err != nil {
			r.log.Warn("failed to release redis lock during shutdown",
				slog.String("instanceId", instanceIDs[i].String()),
				slog.String("error", err.Error()))
		} else {
			released++
		}
	}

	r.log.Info("redis locks released",
		slog.Int("released", released),
		slog.Int("failed", len(locks)-released))

	return released
}

func (r *ClientRegistry) Close() error {
	if r == nil {
		return nil
	}

	r.StopSplitBrainDetection()

	if r.container != nil {
		return r.container.Close()
	}
	return nil
}
