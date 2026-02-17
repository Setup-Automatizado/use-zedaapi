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
	"github.com/redis/go-redis/v9"

	"go.mau.fi/libsignal/keys/prekey"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	internalevents "go.mau.fi/whatsmeow/api/internal/events"
	"go.mau.fi/whatsmeow/api/internal/events/dispatch"
	eventctx "go.mau.fi/whatsmeow/api/internal/events/eventctx"
	"go.mau.fi/whatsmeow/api/internal/events/media"
	"go.mau.fi/whatsmeow/api/internal/locks"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type StoreLink struct {
	ID       uuid.UUID
	StoreJID string
}

type ConnectionState struct {
	Connected       bool
	Status          string
	LastConnectedAt *time.Time
	WorkerID        *string
	DesiredWorkerID *string
}

type InstanceRepository interface {
	ListInstancesWithStoreJID(ctx context.Context) ([]StoreLink, error)
	UpdateConnectionStatus(ctx context.Context, id uuid.UUID, connected bool, status string, workerID *string, desiredWorkerID *string) error
	GetConnectionState(ctx context.Context, id uuid.UUID) (*ConnectionState, error)
	GetCallRejectConfig(ctx context.Context, id uuid.UUID) (bool, *string, error)
}

type InstanceInfo struct {
	ID            uuid.UUID
	Name          string
	SessionName   string
	InstanceToken string
	StoreJID      *string
}

type StatusSnapshot struct {
	Connected        bool
	ConnectionStatus string
	StoreJID         *string
	LastConnected    *time.Time
	AutoReconnect    bool
	WorkerID         string
}

type ContactMetadataConfig struct {
	CacheCapacity   int
	NameTTL         time.Duration
	PhotoTTL        time.Duration
	ErrorTTL        time.Duration
	PrefetchWorkers int
	FetchQueueSize  int
}

func (c ContactMetadataConfig) withDefaults() ContactMetadataConfig {
	if c.CacheCapacity <= 0 {
		c.CacheCapacity = defaultCacheCapacity
	}
	if c.NameTTL <= 0 {
		c.NameTTL = defaultNameTTL
	}
	if c.PhotoTTL <= 0 {
		c.PhotoTTL = defaultPhotoTTL
	}
	if c.ErrorTTL <= 0 {
		c.ErrorTTL = defaultErrorTTL
	}
	if c.PrefetchWorkers <= 0 {
		c.PrefetchWorkers = 4
	}
	if c.FetchQueueSize <= 0 {
		c.FetchQueueSize = 1024
	}
	return c
}

type AutoConnectConfig struct {
	Enabled        bool
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func (c AutoConnectConfig) withDefaults() AutoConnectConfig {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 3
	}
	if c.InitialBackoff <= 0 {
		c.InitialBackoff = time.Second
	}
	if c.MaxBackoff <= 0 {
		c.MaxBackoff = 10 * time.Second
	}
	if c.MaxBackoff < c.InitialBackoff {
		c.MaxBackoff = c.InitialBackoff
	}
	return c
}

type LockConfig struct {
	KeyPrefix       string
	TTL             time.Duration
	RefreshInterval time.Duration
}

func (c LockConfig) withDefaults() LockConfig {
	if c.KeyPrefix == "" {
		c.KeyPrefix = "zedaapi"
	}
	if c.TTL <= 0 {
		c.TTL = 30 * time.Second
	}
	if c.RefreshInterval <= 0 || c.RefreshInterval >= c.TTL {
		c.RefreshInterval = c.TTL / 2
	}
	return c
}

type WorkerDirectory interface {
	AssignedOwner(id uuid.UUID) string
}

var ErrNotAssignedOwner = errors.New("instance not assigned to this worker")

// QueuePauser allows pausing/resuming message queue during proxy swaps.
// Defined locally to avoid circular dependency with the proxy package.
type QueuePauser interface {
	PauseInstance(ctx context.Context, instanceID uuid.UUID, reason string) error
	ResumeInstance(ctx context.Context, instanceID uuid.UUID) error
}

// QueueInstanceRemover allows removing message queue workers for an instance.
// Defined locally to avoid circular dependency with the queue package.
type QueueInstanceRemover interface {
	RemoveInstance(ctx context.Context, instanceID uuid.UUID) error
}

type ClientRegistry struct {
	log                  *slog.Logger
	workerID             string
	hostname             string
	container            *sqlstore.Container
	lockManager          locks.Manager
	repo                 InstanceRepository
	pairCallback         func(context.Context, uuid.UUID, string) error
	resetCallback        func(context.Context, uuid.UUID, string) error
	logLevel             string
	mu                   sync.RWMutex
	clients              map[uuid.UUID]*clientState
	creationLocks        sync.Map
	activeLocks          map[uuid.UUID]locks.Lock
	activeLocksMu        sync.RWMutex
	splitBrainTicker     *time.Ticker
	splitBrainStop       chan struct{}
	splitBrainRunning    bool
	splitBrainMu         sync.Mutex
	metrics              ClientRegistryMetrics
	obsMetrics           *observability.Metrics
	eventIntegration     *internalevents.IntegrationHelper
	dispatchCoordinator  dispatch.DispatchCoordinator
	mediaCoordinator     media.MediaCoordinatorProvider
	contactMetadataCfg   ContactMetadataConfig
	contactMetadataRedis *redis.Client
	eventHandlerTimeout  time.Duration
	autoConnectCfg       AutoConnectConfig
	lockTTL              time.Duration
	lockRefreshInterval  time.Duration
	lockKeyPrefix        string
	workerDirectory      WorkerDirectory
	rebalanceTicker      *time.Ticker
	rebalanceStop        chan struct{}

	reconcileMu       sync.Mutex
	reconcileTicker   *time.Ticker
	reconcileCancel   context.CancelFunc
	reconcileDone     chan struct{}
	reconcileRunning  bool
	reconcileInterval time.Duration

	connectOverride        func(*whatsmeow.Client) error
	isConnectedOverride    func(*whatsmeow.Client) bool
	ensureWithLockOverride func(context.Context, InstanceInfo, locks.Lock) (*whatsmeow.Client, bool, error)

	pairingCodeCache *pairingCache
	proxyRepo        ProxyRepository
	queuePauser      QueuePauser
	queueCoordinator QueueInstanceRemover
}

type clientState struct {
	client             *whatsmeow.Client
	lastConnected      time.Time
	connectionStatus   string
	storeJID           *string
	pairing            *pairingSession
	wasNewInstance     bool
	createdAt          time.Time
	primarySession     bool
	lock               locks.Lock
	lockRefreshCancel  context.CancelFunc
	lockMode           string
	contactCache       *contactMetadataCache
	lidResolver        eventctx.LIDResolver
	pollDecrypter      eventctx.PollDecrypter
	autoConnectStarted bool
	proxyURL           string
}

type ClientRegistryMetrics struct {
	SplitBrainDetected    func()
	SplitBrainInvalidLock func(instanceID string)
}

type lockReacquireMetrics interface {
	RecordLockReacquire(instanceID, result string)
	RecordLockReacquireFallback(instanceID string, state locks.CircuitState)
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
	eventIntegration *internalevents.IntegrationHelper,
	dispatchCoordinator dispatch.DispatchCoordinator,
	mediaCoordinator media.MediaCoordinatorProvider,
	eventHandlerTimeout time.Duration,
	contactMetadataCfg ContactMetadataConfig,
	contactMetadataRedis *redis.Client,
	autoConnectCfg AutoConnectConfig,
	lockCfg LockConfig,
	obsMetrics *observability.Metrics,
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
	contactCfg := contactMetadataCfg.withDefaults()
	autoCfg := autoConnectCfg.withDefaults()
	lockConfiguration := lockCfg.withDefaults()
	workerIdentifier := fmt.Sprintf("%s-%s", hostname, uuid.NewString())
	return &ClientRegistry{
		log:                  logger,
		workerID:             workerIdentifier,
		hostname:             hostname,
		container:            container,
		lockManager:          lockManager,
		repo:                 repo,
		pairCallback:         pairCallback,
		resetCallback:        resetCallback,
		logLevel:             logLevel,
		clients:              make(map[uuid.UUID]*clientState),
		activeLocks:          make(map[uuid.UUID]locks.Lock),
		eventIntegration:     eventIntegration,
		dispatchCoordinator:  dispatchCoordinator,
		mediaCoordinator:     mediaCoordinator,
		contactMetadataCfg:   contactCfg,
		contactMetadataRedis: contactMetadataRedis,
		eventHandlerTimeout:  eventHandlerTimeout,
		autoConnectCfg:       autoCfg,
		lockTTL:              lockConfiguration.TTL,
		lockRefreshInterval:  lockConfiguration.RefreshInterval,
		lockKeyPrefix:        lockConfiguration.KeyPrefix,
		obsMetrics:           obsMetrics,
		pairingCodeCache:     newPairingCache(),
	}, nil
}

func (r *ClientRegistry) WorkerID() string {
	return r.workerID
}

func (r *ClientRegistry) WorkerHostname() string {
	return r.hostname
}

// SetQueueCoordinator sets the optional queue coordinator for removing message queue
// workers during instance reset. Defined as a setter to avoid circular dependencies.
func (r *ClientRegistry) SetQueueCoordinator(q QueueInstanceRemover) {
	r.queueCoordinator = q
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

	if !ok && !r.canHandleInstance(info.ID) {
		return nil, false, ErrNotAssignedOwner
	}

	if !ok && !r.canHandleInstance(info.ID) {
		return nil, false, ErrNotAssignedOwner
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

	appliedProxy := r.resolveProxyURL(ctx, info.ID)

	clientState := &clientState{
		client:         client,
		storeJID:       info.StoreJID,
		wasNewInstance: wasNew,
		createdAt:      time.Now().UTC(),
		lock:           lock,
		lockMode:       lockMode,
		proxyURL:       appliedProxy,
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
		if lockMode != "redis" && !r.canHandleInstance(info.ID) {
			if releaseErr := lock.Release(context.Background()); releaseErr != nil {
				r.log.Warn("failed to release lock for non-owner",
					slog.String("instanceId", info.ID.String()),
					slog.String("error", releaseErr.Error()))
			}
			return nil, false, ErrNotAssignedOwner
		}

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

	if r.eventIntegration != nil {
		registerCtx := ctx
		if registerCtx == nil {
			registerCtx = context.Background()
		}
		registerCtx, cancel := context.WithTimeout(registerCtx, 5*time.Second)
		err := r.eventIntegration.EnsureRegistered(registerCtx, info.ID)
		cancel()
		if err != nil {
			r.log.Error("failed to pre-register instance with event system",
				slog.String("instanceId", info.ID.String()),
				slog.String("error", err.Error()))
		}
	}

	r.maybeStartAutoConnect(info.ID)

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

	appliedProxy := r.resolveProxyURL(ctx, info.ID)

	clientState := &clientState{
		client:         client,
		storeJID:       info.StoreJID,
		wasNewInstance: wasNew,
		createdAt:      time.Now().UTC(),
		lock:           lock,
		lockMode:       lockMode,
		proxyURL:       appliedProxy,
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
		if lockMode != "redis" && !r.canHandleInstance(info.ID) {
			if releaseErr := lock.Release(context.Background()); releaseErr != nil {
				r.log.Warn("failed to release lock for non-owner",
					slog.String("instanceId", info.ID.String()),
					slog.String("error", releaseErr.Error()))
			}
			return nil, false, ErrNotAssignedOwner
		}

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

	if r.eventIntegration != nil {
		registerCtx := ctx
		if registerCtx == nil {
			registerCtx = context.Background()
		}
		registerCtx, cancel := context.WithTimeout(registerCtx, 5*time.Second)
		err := r.eventIntegration.EnsureRegistered(registerCtx, info.ID)
		cancel()
		if err != nil {
			r.log.Error("failed to pre-register instance with event system",
				slog.String("instanceId", info.ID.String()),
				slog.String("error", err.Error()))
		}
	}

	r.maybeStartAutoConnect(info.ID)

	return client, true, nil
}

func (r *ClientRegistry) maybeStartAutoConnect(instanceID uuid.UUID) {
	if !r.autoConnectCfg.Enabled {
		return
	}

	r.mu.Lock()
	state, ok := r.clients[instanceID]
	if !ok || state == nil || state.client == nil {
		r.mu.Unlock()
		return
	}
	if state.autoConnectStarted {
		r.mu.Unlock()
		return
	}

	paired := state.storeJID != nil && *state.storeJID != ""
	if !state.wasNewInstance && !paired {
		r.mu.Unlock()
		return
	}

	state.autoConnectStarted = true
	client := state.client
	wasNew := state.wasNewInstance
	r.mu.Unlock()

	go r.autoConnectInstance(instanceID, client, wasNew)
}

func (r *ClientRegistry) isClientConnected(client *whatsmeow.Client) bool {
	if r.isConnectedOverride != nil {
		return r.isConnectedOverride(client)
	}
	return client.IsConnected()
}

func (r *ClientRegistry) connectClient(client *whatsmeow.Client) error {
	if r.connectOverride != nil {
		return r.connectOverride(client)
	}
	return client.Connect()
}

func (r *ClientRegistry) autoConnectInstance(instanceID uuid.UUID, client *whatsmeow.Client, wasNew bool) {
	if client == nil {
		return
	}

	r.mu.RLock()
	state, ok := r.clients[instanceID]
	if !ok || state == nil || state.client != client {
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()

	cfg := r.autoConnectCfg.withDefaults()
	logger := r.log.With(
		slog.String("component", "client_registry"),
		slog.String("worker", "auto_connect"),
		slog.String("instanceId", instanceID.String()),
		slog.Bool("new_instance", wasNew),
	)

	time.Sleep(100 * time.Millisecond)

	backoff := cfg.InitialBackoff
	maxBackoff := cfg.MaxBackoff
	if maxBackoff < backoff {
		maxBackoff = backoff
	}

	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		r.mu.RLock()
		currentState, ok := r.clients[instanceID]
		r.mu.RUnlock()
		if !ok || currentState == nil || currentState.client != client {
			logger.Debug("auto-connect aborted - client state removed",
				slog.Int("attempt", attempt))
			return
		}

		if r.isClientConnected(client) {
			logger.Debug("auto-connect skipped, already connected",
				slog.Int("attempt", attempt))
			r.persistConnectionStatus(instanceID, true, "connected", nil)
			return
		}

		r.persistConnectionStatus(instanceID, false, "connecting", nil)

		if err := r.connectClient(client); err != nil {
			logger.Warn("auto-connect attempt failed",
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()))
			r.persistConnectionStatus(instanceID, false, "connect_failed", nil)

			if attempt == maxAttempts {
				break
			}

			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * 2)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		logger.Info("auto-connect successful", slog.Int("attempt", attempt))
		r.persistConnectionStatus(instanceID, true, "connected", nil)
		return
	}

	logger.Error("auto-connect exhausted attempts",
		slog.Int("attempts", maxAttempts))
}

func (r *ClientRegistry) startLockRefresh(instanceID uuid.UUID, lock locks.Lock, lockMode string) context.CancelFunc {
	if lock == nil {
		r.log.Debug("no lock to refresh (lockManager not configured)", slog.String("instanceId", instanceID.String()))
		return func() {}
	}

	workerCtx := observability.AsyncContext(observability.AsyncContextOptions{
		Logger:     r.log,
		Component:  "client_registry",
		Worker:     "lock_refresh",
		InstanceID: instanceID.String(),
		Extra: []slog.Attr{
			slog.String("lock_mode", lockMode),
		},
	})
	baseLogger := r.log.With(
		slog.String("component", "client_registry"),
		slog.String("worker", "lock_refresh"),
		slog.String("instance_id", instanceID.String()),
		slog.String("lock_mode", lockMode),
	)

	var metricsReporter lockReacquireMetrics
	if reporter, ok := r.lockManager.(lockReacquireMetrics); ok {
		metricsReporter = reporter
	}

	ctx, cancel := context.WithCancel(workerCtx)

	go func() {
		ticker := time.NewTicker(r.lockRefreshIntervalDuration())
		defer ticker.Stop()

		baseLogger.Debug("started lock refresh goroutine")

		for {
			select {
			case <-ctx.Done():
				baseLogger.Debug("stopping lock refresh goroutine")
				return

			case <-ticker.C:
				r.mu.RLock()
				currentState, stateExists := r.clients[instanceID]
				currentLockMode := ""
				if stateExists && currentState != nil {
					currentLockMode = currentState.lockMode
				}
				r.mu.RUnlock()

				iterationLogger := baseLogger.With(slog.String("current_lock_mode", currentLockMode))

				if currentLockMode == "local" && stateExists {
					type circuitStateChecker interface {
						GetState() locks.CircuitState
					}

					if checker, ok := r.lockManager.(circuitStateChecker); ok {
						circuitState := checker.GetState()

						iterationLogger.Debug("checking circuit breaker state for lock reacquisition",
							slog.Int("circuit_state", int(circuitState)))

						if circuitState == locks.StateClosed {
							iterationLogger.Info("Redis recovered - attempting to reacquire real lock")

							lockKey := r.lockKey(instanceID)

							var newLock locks.Lock
							var acquired bool
							var acquireErr error

							for attempt := 1; attempt <= 3; attempt++ {
								attemptLogger := iterationLogger.With(slog.Int("attempt", attempt))
								acquireCtx, acquireCancel := context.WithTimeout(workerCtx, 3*time.Second)
								newLock, acquired, acquireErr = r.lockManager.Acquire(acquireCtx, lockKey, r.lockTTLSeconds())
								acquireCancel()

								if acquireErr == nil && acquired && newLock != nil {
									lockToken := newLock.GetValue()

									if lockToken == "" {
										attemptLogger.Debug("lock reacquisition returned fallback lock - circuit breaker not ready",
											slog.String("reason", "empty_token"))
										if metricsReporter != nil {
											metricsReporter.RecordLockReacquire(instanceID.String(), observability.LockReacquireResultFallback)
											metricsReporter.RecordLockReacquireFallback(instanceID.String(), circuitState)
										}
										if attempt < 3 {
											backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
											attemptLogger.Debug("waiting before retry", slog.String("backoff", backoff.String()))
											time.Sleep(backoff)
										}
										continue
									}

									attemptLogger.Info("successfully reacquired Redis lock after recovery",
										slog.String("lock_token", func() string {
											if len(lockToken) > 8 {
												return lockToken[:8] + "..."
											}
											return lockToken
										}()))
									if metricsReporter != nil {
										metricsReporter.RecordLockReacquire(instanceID.String(), observability.LockReacquireResultSuccess)
									}

									r.mu.Lock()
									r.activeLocksMu.Lock()
									if state, exists := r.clients[instanceID]; exists && state != nil {
										if state.lock != nil {
											_ = state.lock.Release(context.Background())
										}

										state.lock = newLock
										state.lockMode = "redis"
										r.activeLocks[instanceID] = newLock
									}
									r.activeLocksMu.Unlock()
									r.mu.Unlock()

									lock = newLock
									currentLockMode = "redis"
									iterationLogger = baseLogger.With(slog.String("current_lock_mode", currentLockMode))
									break
								}

								if metricsReporter != nil {
									metricsReporter.RecordLockReacquire(instanceID.String(), observability.LockReacquireResultFailure)
								}

								if attempt < 3 {
									backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
									attemptLogger.Debug("lock reacquisition failed, retrying",
										slog.String("backoff", backoff.String()),
										slog.String("error", func() string {
											if acquireErr != nil {
												return acquireErr.Error()
											}
											if !acquired {
												return "lock not acquired"
											}
											return "unknown error"
										}()))
									time.Sleep(backoff)
								}
							}

							if acquireErr != nil || !acquired || newLock == nil {
								reportErr := acquireErr
								if reportErr == nil {
									if !acquired {
										reportErr = errors.New("lock reacquisition unsuccessful: lock held by another replica")
									} else {
										reportErr = errors.New("lock reacquisition unsuccessful: nil lock returned")
									}
								}

								iterationLogger.Warn("failed to reacquire Redis lock after recovery (all attempts)",
									slog.String("error", reportErr.Error()))
								observability.CaptureWorkerException(workerCtx, "client_registry", "lock_reacquire", instanceID.String(), reportErr)
							}
						}
					}
				}

				currentToken := lock.GetValue()

				if currentToken == "" {
					iterationLogger.Debug("skipping lock refresh - have noOpLock (fallback mode)")
					continue
				}

				refreshCtx, refreshCancel := context.WithTimeout(workerCtx, 2*time.Second)
				err := lock.Refresh(refreshCtx, r.lockTTLSeconds())
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
						iterationLogger.Warn("lock refresh failed due to Redis connection issue - switching to local mode",
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
					var circuitState locks.CircuitState
					if checker, ok := r.lockManager.(circuitStateChecker); ok {
						circuitState = checker.GetState()
						circuitOpen = (circuitState != locks.StateClosed)
					}

					if circuitOpen {
						iterationLogger.Debug("lock refresh failed but circuit breaker open - continuing",
							slog.String("error", err.Error()))
						r.mu.Lock()
						if state, exists := r.clients[instanceID]; exists && state != nil {
							state.lockMode = "local"
						}
						r.mu.Unlock()
					} else {
						iterationLogger.Error("CRITICAL: lock refresh failed with circuit breaker CLOSED - SPLIT-BRAIN DETECTED",
							slog.String("error", err.Error()))
						observability.CaptureWorkerException(workerCtx, "client_registry", "lock_refresh", instanceID.String(), err)

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
									iterationLogger.Warn("failed to release lock after lock loss detection",
										slog.String("error", releaseErr.Error()))
								}
							}

							if state.client != nil {
								state.client.Disconnect()
								iterationLogger.Warn("forcefully disconnected client due to lock loss")
							}
						}

						return
					}
				} else {
					iterationLogger.Debug("lock refreshed successfully")

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
		key := r.lockKey(info.ID)
		var acquired bool
		lock, acquired, err = r.lockManager.Acquire(ctx, key, r.lockTTLSeconds())
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

	// Apply per-instance proxy configuration before Connect()
	proxyURL := r.applyProxyFromConfig(ctx, info.ID, client)
	_ = proxyURL // tracked in clientState after return

	client.AddEventHandler(r.wrapEventHandler(info.ID))

	return client, storeReset, lock, nil
}

func logLevelOrDefault(level string) string {
	if level == "" {
		return "INFO"
	}
	return level
}

func (r *ClientRegistry) getCachedPairingCode(instanceID uuid.UUID, codeType PairingCodeType) (*CachedPairingCode, bool) {
	if r.pairingCodeCache == nil {
		return nil, false
	}
	return r.pairingCodeCache.Get(instanceID, codeType)
}

func (r *ClientRegistry) getCachedPhoneCode(instanceID uuid.UUID, phone string) (*CachedPairingCode, bool) {
	if r.pairingCodeCache == nil {
		return nil, false
	}
	return r.pairingCodeCache.GetForPhone(instanceID, phone)
}

func (r *ClientRegistry) setCachedPairingCode(instanceID uuid.UUID, code *CachedPairingCode) {
	if r.pairingCodeCache == nil {
		return
	}
	r.pairingCodeCache.Set(instanceID, code)
	r.log.Debug("cached pairing code",
		slog.String("instanceId", instanceID.String()),
		slog.String("type", string(code.Type)),
		slog.Duration("ttl", code.RemainingTTL()))
}

// invalidatePairingCache remove codigo do cache
func (r *ClientRegistry) invalidatePairingCache(instanceID uuid.UUID, reason string) {
	if r.pairingCodeCache == nil {
		return
	}
	r.pairingCodeCache.Invalidate(instanceID)
	r.log.Debug("invalidated pairing cache",
		slog.String("instanceId", instanceID.String()),
		slog.String("reason", reason))
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
	r.invalidatePairingCache(instanceID, reason)

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

	// Flush event buffers before unregistering workers
	if r.eventIntegration != nil {
		flushCtx, flushCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := r.eventIntegration.OnInstanceDisconnect(flushCtx, instanceID); err != nil {
			r.log.Warn("failed to flush instance buffer during reset",
				slog.String("instanceId", instanceID.String()),
				slog.String("reason", reason),
				slog.String("error", err.Error()))
		}
		flushCancel()
	}

	// Unregister NATS dispatch worker
	if r.dispatchCoordinator != nil {
		unregCtx, unregCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := r.dispatchCoordinator.UnregisterInstance(unregCtx, instanceID); err != nil {
			r.log.Warn("failed to unregister instance from dispatch coordinator during reset",
				slog.String("instanceId", instanceID.String()),
				slog.String("reason", reason),
				slog.String("error", err.Error()))
		}
		unregCancel()
	}

	// Unregister NATS media worker
	if r.mediaCoordinator != nil {
		if err := r.mediaCoordinator.UnregisterInstance(instanceID); err != nil {
			r.log.Warn("failed to unregister instance from media coordinator during reset",
				slog.String("instanceId", instanceID.String()),
				slog.String("reason", reason),
				slog.String("error", err.Error()))
		}
	}

	// Unregister NATS message queue worker (created lazily on first message, may not exist)
	if r.queueCoordinator != nil {
		queueCtx, queueCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := r.queueCoordinator.RemoveInstance(queueCtx, instanceID); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				r.log.Warn("failed to remove instance from message queue coordinator during reset",
					slog.String("instanceId", instanceID.String()),
					slog.String("reason", reason),
					slog.String("error", err.Error()))
			}
		}
		queueCancel()
	}

	// Unregister event handler and buffer
	if r.eventIntegration != nil {
		removeCtx, removeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := r.eventIntegration.OnInstanceRemove(removeCtx, instanceID); err != nil {
			r.log.Warn("failed to unregister instance from event system during reset",
				slog.String("instanceId", instanceID.String()),
				slog.String("reason", reason),
				slog.String("error", err.Error()))
		}
		removeCancel()
	}

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

	r.persistConnectionStatus(instanceID, false, reason, nil)

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

	_, _, encErr := client.DangerousInternals().EncryptMessageForDevice(ctx, []byte{0x00}, primaryLID, bundle, nil, nil)
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
		var (
			provider  eventctx.ContactMetadataProvider
			resolver  eventctx.LIDResolver
			decrypter eventctx.PollDecrypter
		)

		r.mu.Lock()
		if state, ok := r.clients[instanceID]; ok && state != nil && state.client != nil {
			if state.contactCache == nil {
				state.contactCache = newContactMetadataCache(
					state.client,
					r.log.With(
						slog.String("component", "contact_metadata_cache"),
						slog.String("instanceId", instanceID.String()),
					),
					instanceID,
					r.contactMetadataCfg,
					r.contactMetadataRedis,
				)
			}
			provider = state.contactCache
			if state.lidResolver == nil {
				state.lidResolver = newStoreLIDResolver(
					state.client,
					r.log.With(
						slog.String("component", "lid_resolver"),
						slog.String("instanceId", instanceID.String()),
					),
				)
			}
			resolver = state.lidResolver
			if state.pollDecrypter == nil {
				state.pollDecrypter = newPollDecrypter(
					state.client,
					r.log.With(
						slog.String("component", "poll_decrypter"),
						slog.String("instanceId", instanceID.String()),
					),
				)
			}
			decrypter = state.pollDecrypter
		}
		r.mu.Unlock()

		ctx := context.Background()
		if provider != nil {
			ctx = eventctx.WithContactProvider(ctx, provider)
		}
		if resolver != nil {
			ctx = eventctx.WithLIDResolver(ctx, resolver)
		}
		if decrypter != nil {
			ctx = eventctx.WithPollDecrypter(ctx, decrypter)
		}

		if r.eventIntegration != nil {
			timeout := r.eventHandlerTimeout
			if timeout <= 0 {
				timeout = 60 * time.Second
			}
			eventCtx, cancel := context.WithTimeout(ctx, timeout)
			r.eventIntegration.WrapEventHandler(eventCtx, instanceID, evt)
			cancel()
		}

		switch e := evt.(type) {
		case *events.Connected:
			if r.eventIntegration != nil {
				if err := r.eventIntegration.OnInstanceConnect(ctx, instanceID); err != nil {
					r.log.Error("failed to register instance with event system",
						slog.String("instanceId", instanceID.String()),
						slog.String("error", err.Error()))
				}
			}

			if r.dispatchCoordinator != nil {
				if err := r.dispatchCoordinator.RegisterInstance(ctx, instanceID); err != nil {
					r.log.Error("failed to register instance with dispatch coordinator",
						slog.String("instanceId", instanceID.String()),
						slog.String("error", err.Error()))
				}
			}

			if r.mediaCoordinator != nil {
				r.mu.RLock()
				state, ok := r.clients[instanceID]
				var client *whatsmeow.Client
				if ok && state != nil {
					client = state.client
				}
				r.mu.RUnlock()

				if client != nil {
					if err := r.mediaCoordinator.RegisterInstance(instanceID, client); err != nil {
						r.log.Error("failed to register instance with media coordinator",
							slog.String("instanceId", instanceID.String()),
							slog.String("error", err.Error()))
					}
				}
			}

			r.mu.Lock()
			var (
				client       *whatsmeow.Client
				sessionReady bool
			)
			if state, ok := r.clients[instanceID]; ok {
				state.connectionStatus = "connected"
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
			r.announcePresenceAvailable(instanceID, "connected")
			r.persistConnectionStatus(instanceID, true, "connected", nil)

		case *events.PushNameSetting:
			r.announcePresenceAvailable(instanceID, "push_name_setting")

		case *events.AppStateSyncComplete:
			r.announcePresenceAvailable(instanceID, "app_state_sync_complete")

		case *events.KeepAliveRestored:
			r.announcePresenceAvailable(instanceID, "keep_alive_restored")

		case *events.StreamReplaced:
			r.announcePresenceAvailable(instanceID, "stream_replaced")

		case *events.PairSuccess:
			r.invalidatePairingCache(instanceID, "pair_success")

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
					callbackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := r.pairCallback(callbackCtx, instanceID, jid); err != nil {
						r.log.Error("pair callback",
							slog.String("instanceId", instanceID.String()),
							slog.String("error", err.Error()))
					}
				}()
			}

		case *events.Disconnected:
			r.mu.Lock()
			if state, ok := r.clients[instanceID]; ok {
				state.connectionStatus = "disconnected"
			}
			r.mu.Unlock()
			r.announcePresenceUnavailable(instanceID, "disconnected")
			if r.eventIntegration != nil {
				if err := r.eventIntegration.OnInstanceDisconnect(ctx, instanceID); err != nil {
					r.log.Warn("failed to flush instance buffer",
						slog.String("instanceId", instanceID.String()),
						slog.String("error", err.Error()))
				}
			}

			r.log.Warn("instance disconnected", slog.String("instanceId", instanceID.String()))
			r.persistConnectionStatus(instanceID, false, "disconnected", nil)

		case *events.LoggedOut:
			r.invalidatePairingCache(instanceID, "logged_out")

			r.mu.Lock()
			if state, ok := r.clients[instanceID]; ok {
				state.connectionStatus = "logged_out"
			}
			r.mu.Unlock()
			r.announcePresenceUnavailable(instanceID, "logged_out")

			reason := "logged_out"
			if desc := e.Reason.String(); desc != "" {
				reason = "logged_out:" + desc
			}
			r.log.Warn("instance logged out",
				slog.String("instanceId", instanceID.String()),
				slog.String("reason", reason))
			emptyWorker := ""
			r.persistConnectionStatus(instanceID, false, reason, &emptyWorker)
			r.ResetClient(instanceID, reason)

		case *events.PairError:
			if e.Error != nil &&
				(errors.Is(e.Error, whatsmeow.ErrPairInvalidDeviceIdentityHMAC) ||
					errors.Is(e.Error, whatsmeow.ErrPairInvalidDeviceSignature)) {
				reason := "pair_error"
				r.log.Warn("pairing error reset",
					slog.String("instanceId", instanceID.String()),
					slog.String("error", e.Error.Error()))
				r.ResetClient(instanceID, reason)
			}

		case *events.QR:
			// Capture fresh QR codes from whatsmeow events and update cache
			if len(e.Codes) > 0 {
				now := time.Now()
				cached := &CachedPairingCode{
					Code:      e.Codes[0],
					Type:      PairingCodeTypeQR,
					CreatedAt: now,
					ExpiresAt: now.Add(pairingCodeTTL),
					SessionID: uuid.NewString(),
				}
				r.setCachedPairingCode(instanceID, cached)

				r.log.Info("QR code cache updated from event",
					slog.String("instanceId", instanceID.String()),
					slog.Int("codes_received", len(e.Codes)))

				// If multiple codes received, start monitoring for renewals
				if len(e.Codes) > 1 {
					go r.monitorQRCodesFromEvent(instanceID, cached.SessionID, e.Codes[1:])
				}
			}

		case *events.CallOffer:
			// Handle automatic call rejection if configured
			r.mu.RLock()
			state, ok := r.clients[instanceID]
			var client *whatsmeow.Client
			if ok && state != nil {
				client = state.client
			}
			r.mu.RUnlock()

			if client != nil && r.repo != nil {
				handler := newCallRejectHandler(
					client,
					instanceID,
					r.repo,
					r.log,
				)
				go handler.HandleCallOffer(ctx, e)
			}

		case *events.CallOfferNotice:
			// Handle automatic call rejection for group calls if configured
			r.mu.RLock()
			state, ok := r.clients[instanceID]
			var client *whatsmeow.Client
			if ok && state != nil {
				client = state.client
			}
			r.mu.RUnlock()

			if client != nil && r.repo != nil {
				handler := newCallRejectHandler(
					client,
					instanceID,
					r.repo,
					r.log,
				)
				go handler.HandleCallOfferNotice(ctx, e)
			}
		}
	}
}

func (r *ClientRegistry) announcePresenceAvailable(instanceID uuid.UUID, reason string) {
	r.announcePresence(instanceID, types.PresenceAvailable, reason)
}

func (r *ClientRegistry) announcePresenceUnavailable(instanceID uuid.UUID, reason string) {
	r.announcePresence(instanceID, types.PresenceUnavailable, reason)
}

func (r *ClientRegistry) announcePresence(instanceID uuid.UUID, presence types.Presence, reason string) {
	client := r.getClient(instanceID)
	if client == nil {
		return
	}

	go func() {
		if err := client.SendPresence(context.Background(), presence); err != nil {
			r.log.Warn("failed to send presence update",
				slog.String("instanceId", instanceID.String()),
				slog.String("reason", reason),
				slog.String("presence", string(presence)),
				slog.String("error", err.Error()))
			return
		}
		r.log.Debug("presence state updated",
			slog.String("instanceId", instanceID.String()),
			slog.String("reason", reason),
			slog.String("presence", string(presence)))
	}()
}

func (r *ClientRegistry) getClient(instanceID uuid.UUID) *whatsmeow.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if state, ok := r.clients[instanceID]; ok && state != nil {
		return state.client
	}
	return nil
}

// GetClient returns a WhatsApp client for the given instance ID.
// This is a public wrapper around getClient for use by other packages.
// Returns the client and a boolean indicating if it was found.
func (r *ClientRegistry) GetClient(instanceID uuid.UUID) (*whatsmeow.Client, bool) {
	client := r.getClient(instanceID)
	return client, client != nil
}

// IsConnected checks if an instance has an active WhatsApp connection.
// Returns true if the client exists and is logged in (authenticated).
func (r *ClientRegistry) IsConnected(instanceID uuid.UUID) bool {
	r.mu.RLock()
	state, ok := r.clients[instanceID]
	r.mu.RUnlock()

	if !ok || state.client == nil {
		return false
	}

	return state.client.IsLoggedIn()
}

func (r *ClientRegistry) Status(info InstanceInfo) StatusSnapshot {
	r.mu.RLock()
	state, ok := r.clients[info.ID]
	r.mu.RUnlock()
	if !ok {
		snapshot := StatusSnapshot{
			Connected:        false,
			ConnectionStatus: "disconnected",
			StoreJID:         info.StoreJID,
			AutoReconnect:    false,
			WorkerID:         r.workerID,
		}

		if repoState := r.loadRepoConnectionState(info.ID); repoState != nil {
			snapshot.Connected = repoState.Connected
			if repoState.Status != "" {
				snapshot.ConnectionStatus = repoState.Status
			}
			if repoState.LastConnectedAt != nil {
				snapshot.LastConnected = repoState.LastConnectedAt
			}
			if repoState.WorkerID != nil && *repoState.WorkerID != "" {
				snapshot.WorkerID = *repoState.WorkerID
			}
		}

		return snapshot
	}

	var storeJID *string
	if state.storeJID != nil {
		storeJID = state.storeJID
	} else if state.client != nil && state.client.Store != nil && state.client.Store.ID != nil {
		jid := state.client.Store.ID.String()
		storeJID = &jid
	}

	connected := state.client != nil && state.client.IsLoggedIn()
	connectionStatus := state.connectionStatus
	if connectionStatus == "" {
		if connected {
			connectionStatus = "connected"
		} else if state.wasNewInstance {
			connectionStatus = "initializing"
		} else {
			connectionStatus = "disconnected"
		}
	}

	var lastConnected *time.Time
	if !state.lastConnected.IsZero() {
		lc := state.lastConnected
		lastConnected = &lc
	} else if repoState := r.loadRepoConnectionState(info.ID); repoState != nil && repoState.LastConnectedAt != nil {
		lastConnected = repoState.LastConnectedAt
		if connectionStatus == "" && repoState.Status != "" {
			connectionStatus = repoState.Status
		}
	}

	return StatusSnapshot{
		Connected:        connected,
		ConnectionStatus: connectionStatus,
		StoreJID:         storeJID,
		LastConnected:    lastConnected,
		AutoReconnect:    state.client != nil && state.client.EnableAutoReconnect,
		WorkerID:         r.workerID,
	}
}

func (r *ClientRegistry) loadRepoConnectionState(instanceID uuid.UUID) *ConnectionState {
	if r.repo == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	state, err := r.repo.GetConnectionState(ctx, instanceID)
	if err != nil {
		r.log.Debug("failed to load connection state from repository",
			slog.String("instanceId", instanceID.String()),
			slog.String("error", err.Error()))
		return nil
	}

	return state
}

func (r *ClientRegistry) persistConnectionStatus(instanceID uuid.UUID, connected bool, status string, workerIDOverride *string) {
	if r.repo == nil {
		return
	}

	var workerPtr *string
	if workerIDOverride != nil {
		if *workerIDOverride != "" {
			worker := *workerIDOverride
			workerPtr = &worker
		}
	} else if r.workerID != "" {
		worker := r.workerID
		workerPtr = &worker
	}

	desiredPtr := r.desiredWorkerPointer(instanceID)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.repo.UpdateConnectionStatus(ctx, instanceID, connected, status, workerPtr, desiredPtr); err != nil {
			r.log.Warn("failed to persist connection status",
				slog.String("instanceId", instanceID.String()),
				slog.String("status", status),
				slog.String("error", err.Error()))
		} else {
			r.log.Debug("persisted connection status",
				slog.String("instanceId", instanceID.String()),
				slog.String("status", status),
				slog.Bool("connected", connected))
		}
	}()
}

func (r *ClientRegistry) desiredWorkerPointer(instanceID uuid.UUID) *string {
	if r.workerDirectory == nil {
		if r.workerID == "" {
			return nil
		}
		owner := r.workerID
		return &owner
	}
	owner := r.workerDirectory.AssignedOwner(instanceID)
	if owner == "" {
		return nil
	}
	ownerCopy := owner
	return &ownerCopy
}

func (r *ClientRegistry) canHandleInstance(instanceID uuid.UUID) bool {
	if r.workerDirectory == nil || r.workerID == "" {
		return true
	}
	owner := r.workerDirectory.AssignedOwner(instanceID)
	return owner == "" || owner == r.workerID
}

func (r *ClientRegistry) lockKey(instanceID uuid.UUID) string {
	prefix := r.lockKeyPrefix
	if prefix == "" {
		prefix = "zedaapi"
	}
	return fmt.Sprintf("%s:instance:%s", prefix, instanceID.String())
}

func (r *ClientRegistry) lockTTLSeconds() int {
	ttl := r.lockTTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	seconds := int(ttl / time.Second)
	if seconds <= 0 {
		seconds = 30
	}
	return seconds
}

func (r *ClientRegistry) lockRefreshIntervalDuration() time.Duration {
	interval := r.lockRefreshInterval
	if interval <= 0 || interval >= r.lockTTL {
		interval = r.lockTTL / 2
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	return interval
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

	r.invalidatePairingCache(info.ID, "disconnect")

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

	if state.contactCache != nil {
		state.contactCache.Close()
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
	r.mu.RLock()
	state, exists := r.clients[info.ID]
	if exists && state != nil && state.client != nil {
		if state.client.Store != nil && state.client.Store.ID != nil {
			r.mu.RUnlock()
			return "", ErrInstanceAlreadyPaired
		}
	}
	r.mu.RUnlock()

	if cached, ok := r.getCachedPairingCode(info.ID, PairingCodeTypeQR); ok {
		r.log.Debug("returning cached QR code",
			slog.String("instanceId", info.ID.String()),
			slog.Duration("remaining_ttl", cached.RemainingTTL()))
		return cached.Code, nil
	}

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
				now := time.Now()
				sessionID := uuid.NewString()
				cached := &CachedPairingCode{
					Code:      item.Code,
					Type:      PairingCodeTypeQR,
					CreatedAt: now,
					ExpiresAt: now.Add(pairingCodeTTL),
					SessionID: sessionID,
				}
				r.setCachedPairingCode(info.ID, cached)

				go r.monitorQRCodeRenewals(info.ID, qrChan, sessionID)

				return item.Code, nil
			} else if item.Event == whatsmeow.QRChannelEventError {
				return "", item.Error
			} else if item.Event == "success" {
				r.invalidatePairingCache(info.ID, "paired_during_qr_generation")
				return "", ErrInstanceAlreadyPaired
			}
		}
	}
}

func (r *ClientRegistry) monitorQRCodeRenewals(instanceID uuid.UUID, qrChan <-chan whatsmeow.QRChannelItem, sessionID string) {
	for item := range qrChan {
		switch item.Event {
		case whatsmeow.QRChannelEventCode:
			now := time.Now()
			cached := &CachedPairingCode{
				Code:      item.Code,
				Type:      PairingCodeTypeQR,
				CreatedAt: now,
				ExpiresAt: now.Add(pairingCodeTTL),
				SessionID: sessionID,
			}
			r.setCachedPairingCode(instanceID, cached)
			r.log.Debug("QR code renewed",
				slog.String("instanceId", instanceID.String()),
				slog.String("sessionId", sessionID))

		case "success":
			r.invalidatePairingCache(instanceID, "pair_success")
			r.log.Info("pairing successful via QR code",
				slog.String("instanceId", instanceID.String()))
			return

		case "timeout":
			r.invalidatePairingCache(instanceID, "session_timeout")
			r.log.Debug("QR session timed out",
				slog.String("instanceId", instanceID.String()))
			return

		default:
			if item.Error != nil {
				r.invalidatePairingCache(instanceID, "error")
				r.log.Warn("QR channel error",
					slog.String("instanceId", instanceID.String()),
					slog.String("event", item.Event),
					slog.String("error", item.Error.Error()))
				return
			}
		}
	}
	r.invalidatePairingCache(instanceID, "channel_closed")
}

// monitorQRCodesFromEvent handles QR code rotation from *events.QR payloads.
// Each code is cached with pairingCodeTTL (20s) and rotated sequentially.
func (r *ClientRegistry) monitorQRCodesFromEvent(instanceID uuid.UUID, sessionID string, remainingCodes []string) {
	const qrCodeInterval = 20 * time.Second

	for i, code := range remainingCodes {
		// Check if session is still valid before sleeping
		if cached, ok := r.getCachedPairingCode(instanceID, PairingCodeTypeQR); !ok || cached.SessionID != sessionID {
			r.log.Debug("QR event monitor stopped - session changed",
				slog.String("instanceId", instanceID.String()),
				slog.String("sessionId", sessionID))
			return
		}

		// Check if instance is already paired
		r.mu.RLock()
		state, exists := r.clients[instanceID]
		isPaired := exists && state != nil && state.client != nil &&
			state.client.Store != nil && state.client.Store.ID != nil
		r.mu.RUnlock()

		if isPaired {
			r.log.Debug("QR event monitor stopped - instance paired",
				slog.String("instanceId", instanceID.String()))
			return
		}

		time.Sleep(qrCodeInterval)

		// Update cache with next QR code
		now := time.Now()
		cached := &CachedPairingCode{
			Code:      code,
			Type:      PairingCodeTypeQR,
			CreatedAt: now,
			ExpiresAt: now.Add(pairingCodeTTL),
			SessionID: sessionID,
		}
		r.setCachedPairingCode(instanceID, cached)
		r.log.Debug("QR code rotated from event",
			slog.String("instanceId", instanceID.String()),
			slog.Int("code_index", i+1),
			slog.Int("remaining", len(remainingCodes)-i-1))
	}

	r.log.Debug("QR event codes exhausted",
		slog.String("instanceId", instanceID.String()),
		slog.String("sessionId", sessionID))
}

func (r *ClientRegistry) PairPhone(ctx context.Context, info InstanceInfo, phone string) (string, error) {
	client, _, err := r.EnsureClient(ctx, info)
	if err != nil {
		return "", err
	}
	if client.Store != nil && client.Store.ID != nil {
		return "", ErrInstanceAlreadyPaired
	}

	if cached, ok := r.getCachedPhoneCode(info.ID, phone); ok {
		r.log.Debug("returning cached phone code",
			slog.String("instanceId", info.ID.String()),
			slog.String("phone", phone),
			slog.Duration("remaining_ttl", cached.RemainingTTL()))
		return cached.Code, nil
	}

	r.invalidatePairingCache(info.ID, "phone_changed")

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

	pairCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	code, err := client.PairPhone(pairCtx, phone, true, whatsmeow.PairClientSafari, "Safari (macOS)")
	if err != nil {
		return "", fmt.Errorf("failed to generate pairing code: %w", err)
	}

	now := time.Now()
	cached := &CachedPairingCode{
		Code:      code,
		Type:      PairingCodeTypePhone,
		Phone:     phone,
		CreatedAt: now,
		ExpiresAt: now.Add(pairingCodeTTL),
		SessionID: uuid.NewString(),
	}
	r.setCachedPairingCode(info.ID, cached)

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
		if !r.canHandleInstance(link.ID) {
			r.log.Debug("skipping instance (assigned to another worker)",
				slog.String("instanceId", link.ID.String()))
			skipped++
			continue
		}

		key := r.lockKey(link.ID)
		lock, acquired, err := r.lockManager.Acquire(ctx, key, r.lockTTLSeconds())

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

func (r *ClientRegistry) SetMetrics(metrics ClientRegistryMetrics) {
	r.splitBrainMu.Lock()
	defer r.splitBrainMu.Unlock()
	r.metrics = metrics
}

func (r *ClientRegistry) SetSplitBrainMetrics(splitBrainCounter func()) {
	r.SetMetrics(ClientRegistryMetrics{SplitBrainDetected: splitBrainCounter})
}

func (r *ClientRegistry) startOwnershipRebalancer(interval time.Duration) {
	r.stopOwnershipRebalancer()
	r.rebalanceTicker = time.NewTicker(interval)
	r.rebalanceStop = make(chan struct{})
	go r.runOwnershipRebalancer()
}

func (r *ClientRegistry) stopOwnershipRebalancer() {
	if r.rebalanceStop != nil {
		close(r.rebalanceStop)
		r.rebalanceStop = nil
	}
	if r.rebalanceTicker != nil {
		r.rebalanceTicker.Stop()
		r.rebalanceTicker = nil
	}
}

func (r *ClientRegistry) runOwnershipRebalancer() {
	for {
		select {
		case <-r.rebalanceStop:
			return
		case <-r.rebalanceTicker.C:
			r.performOwnershipRebalance()
		}
	}
}

func (r *ClientRegistry) performOwnershipRebalance() {
	if r.workerDirectory == nil {
		return
	}
	r.mu.RLock()
	ids := make([]uuid.UUID, 0, len(r.clients))
	for id := range r.clients {
		ids = append(ids, id)
	}
	r.mu.RUnlock()

	for _, id := range ids {
		owner := r.workerDirectory.AssignedOwner(id)
		if owner == "" || owner == r.workerID {
			continue
		}
		r.log.Info("handing off instance to new worker",
			slog.String("instanceId", id.String()),
			slog.String("currentWorker", r.workerID),
			slog.String("assignedWorker", owner))
		r.resetClient(id, "ownership_rebalance", false)
	}
}

func (r *ClientRegistry) AttachWorkerDirectory(dir WorkerDirectory, rebalanceInterval time.Duration) {
	r.workerDirectory = dir
	if dir == nil {
		r.stopOwnershipRebalancer()
		return
	}
	if rebalanceInterval <= 0 {
		rebalanceInterval = 30 * time.Second
	}
	r.startOwnershipRebalancer(rebalanceInterval)
}

func (r *ClientRegistry) StartReconciliationWorker(ctx context.Context, interval time.Duration) {
	if r.repo == nil {
		r.log.Warn("reconciliation worker disabled - repository not configured")
		return
	}

	if interval <= 0 {
		interval = 10 * time.Second
	}

	if ctx == nil {
		ctx = context.Background()
	}

	r.reconcileMu.Lock()
	if r.reconcileRunning {
		r.reconcileMu.Unlock()
		r.log.Warn("reconciliation worker already running")
		return
	}

	workerLogger := r.log.With(
		slog.String("component", "client_registry"),
		slog.String("worker", "reconciliation"),
	)
	workerCtx := logging.WithLogger(ctx, workerLogger)
	workerCtx, cancel := context.WithCancel(workerCtx)

	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	r.reconcileCancel = cancel
	r.reconcileDone = done
	r.reconcileTicker = ticker
	r.reconcileRunning = true
	r.reconcileInterval = interval
	r.reconcileMu.Unlock()

	workerLogger.Info("started reconciliation worker",
		slog.Duration("interval", interval))

	go r.runReconciliationWorker(workerCtx, ticker, done, workerLogger)
}

func (r *ClientRegistry) StopReconciliationWorker() {
	r.reconcileMu.Lock()
	if !r.reconcileRunning {
		r.reconcileMu.Unlock()
		return
	}

	cancel := r.reconcileCancel
	done := r.reconcileDone
	ticker := r.reconcileTicker

	r.reconcileCancel = nil
	r.reconcileDone = nil
	r.reconcileTicker = nil
	r.reconcileRunning = false
	r.reconcileInterval = 0
	r.reconcileMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if ticker != nil {
		ticker.Stop()
	}

	if done != nil {
		select {
		case <-done:
			r.log.Info("reconciliation worker stopped")
		case <-time.After(5 * time.Second):
			r.log.Warn("reconciliation worker stop timeout")
		}
	}
}

func (r *ClientRegistry) runReconciliationWorker(ctx context.Context, ticker *time.Ticker, done chan struct{}, logger *slog.Logger) {
	defer close(done)
	defer func() {
		if rec := recover(); rec != nil {
			err := fmt.Errorf("reconciliation worker panic: %v", rec)
			logger.Error("reconciliation worker panic", slog.Any("panic", rec))
			observability.CaptureWorkerException(ctx, "client_registry", "reconciliation", "", err)
		}
	}()
	defer ticker.Stop()

	r.reconcileOrphanedInstances(ctx, logger)

	for {
		select {
		case <-ctx.Done():
			logger.Info("reconciliation worker stopped by context")
			return
		case <-ticker.C:
			r.reconcileOrphanedInstances(ctx, logger)
		}
	}
}

func (r *ClientRegistry) reconcileOrphanedInstances(ctx context.Context, logger *slog.Logger) {
	if r.repo == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if logger == nil {
		logger = logging.ContextLogger(ctx, r.log).With(
			slog.String("component", "client_registry"),
			slog.String("worker", "reconciliation"),
		)
	}

	listCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	links, err := r.repo.ListInstancesWithStoreJID(listCtx)
	cancel()
	if err != nil {
		logger.Error("reconciliation: list instances failed",
			slog.String("error", err.Error()))
		r.recordReconciliationAttempt("error")
		return
	}

	r.mu.RLock()
	orphaned := make([]StoreLink, 0, len(links))
	for _, link := range links {
		if _, exists := r.clients[link.ID]; exists {
			continue
		}
		orphaned = append(orphaned, link)
	}
	r.mu.RUnlock()

	if r.obsMetrics != nil && r.obsMetrics.OrphanedInstances != nil {
		r.obsMetrics.OrphanedInstances.Set(float64(len(orphaned)))
	}

	if len(orphaned) == 0 {
		logger.Debug("reconciliation: no orphaned instances found",
			slog.Int("paired_instances", len(links)))
		return
	}

	logger.Debug("reconciliation: recovering orphaned instances",
		slog.Int("count", len(orphaned)))

	for _, link := range orphaned {
		if ctx.Err() != nil {
			logger.Debug("reconciliation: context cancelled, stopping recovery")
			return
		}

		start := time.Now()
		instanceLogger := logger.With(slog.String("instanceId", link.ID.String()))

		if !r.canHandleInstance(link.ID) {
			logger.Debug("reconciliation: instance assigned to another worker",
				slog.String("instanceId", link.ID.String()))
			r.recordReconciliationAttempt("skipped")
			continue
		}

		key := r.lockKey(link.ID)

		var (
			lock     locks.Lock
			acquired bool
			lockErr  error
		)

		if r.lockManager != nil {
			lockCtx, lockCancel := context.WithTimeout(ctx, 5*time.Second)
			lock, acquired, lockErr = r.lockManager.Acquire(lockCtx, key, r.lockTTLSeconds())
			lockCancel()
		} else {
			lock = &noOpLock{}
			acquired = true
		}

		if lockErr != nil {
			if isRedisConnectivityError(lockErr) {
				instanceLogger.Warn("reconciliation: redis unavailable, using in-memory lock")
				lock = &noOpLock{}
				acquired = true
			} else {
				instanceLogger.Debug("reconciliation: lock acquire error",
					slog.String("error", lockErr.Error()))
				r.recordReconciliationAttempt("error")
				continue
			}
		}

		if !acquired {
			instanceLogger.Debug("reconciliation: lock held by another replica")
			r.recordReconciliationAttempt("skipped")
			continue
		}

		info := InstanceInfo{
			ID:       link.ID,
			StoreJID: &link.StoreJID,
		}

		var (
			client    *whatsmeow.Client
			created   bool
			ensureErr error
		)
		if r.ensureWithLockOverride != nil {
			client, created, ensureErr = r.ensureWithLockOverride(ctx, info, lock)
		} else {
			client, created, ensureErr = r.EnsureClientWithLock(ctx, info, lock)
		}
		if ensureErr != nil {
			instanceLogger.Error("reconciliation: ensure client failed",
				slog.String("error", ensureErr.Error()))
			r.recordReconciliationAttempt("failure")
			if lock != nil {
				if releaseErr := lock.Release(context.Background()); releaseErr != nil {
					instanceLogger.Warn("reconciliation: failed to release lock after ensure failure",
						slog.String("error", releaseErr.Error()))
				}
			}
			continue
		}

		if !r.isClientConnected(client) {
			if connectErr := r.connectClient(client); connectErr != nil {
				instanceLogger.Error("reconciliation: connect failed",
					slog.String("error", connectErr.Error()))
				r.recordReconciliationAttempt("failure")
				continue
			}
		}

		r.recordReconciliationAttempt("success")
		if r.obsMetrics != nil && r.obsMetrics.ReconciliationDuration != nil {
			r.obsMetrics.ReconciliationDuration.Observe(time.Since(start).Seconds())
		}

		instanceLogger.Info("reconciliation: instance recovered",
			slog.Duration("duration", time.Since(start)),
			slog.Bool("was_created", created))
	}

	if r.obsMetrics != nil && r.obsMetrics.OrphanedInstances != nil {
		r.obsMetrics.OrphanedInstances.Set(0)
	}
}

func (r *ClientRegistry) recordReconciliationAttempt(result string) {
	if result == "" {
		result = "unknown"
	}
	if r.obsMetrics == nil || r.obsMetrics.ReconciliationAttempts == nil {
		return
	}
	r.obsMetrics.ReconciliationAttempts.WithLabelValues(result).Inc()
}

func isRedisConnectivityError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "context deadline exceeded")
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
					if r.metrics.SplitBrainInvalidLock != nil {
						r.metrics.SplitBrainInvalidLock(id.String())
					}
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

		lockKey := r.lockKey(instanceID)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		var stillOwned bool
		var err error

		type lockOwnershipChecker interface {
			CheckLockOwnership(ctx context.Context, key string, expectedLock locks.Lock) (bool, error)
		}

		if checker, ok := r.lockManager.(lockOwnershipChecker); ok {
			stillOwned, err = checker.CheckLockOwnership(ctx, lockKey, expectedLock)
		} else {
			err = expectedLock.Refresh(ctx, r.lockTTLSeconds())
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

			if r.metrics.SplitBrainDetected != nil {
				r.metrics.SplitBrainDetected()
			}

			observability.CaptureWorkerException(
				observability.AsyncContext(observability.AsyncContextOptions{
					Logger:     r.log,
					Component:  "client_registry",
					Worker:     "split_brain_detection",
					InstanceID: instanceID.String(),
				}),
				"client_registry",
				"split_brain_detection",
				instanceID.String(),
				fmt.Errorf("split-brain detected: lock ownership lost for instance %s", instanceID.String()),
			)

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

	r.StopReconciliationWorker()
	r.StopSplitBrainDetection()
	r.stopOwnershipRebalancer()

	r.mu.RLock()
	for _, state := range r.clients {
		if state != nil && state.contactCache != nil {
			state.contactCache.Close()
		}
	}
	r.mu.RUnlock()

	if r.container != nil {
		return r.container.Close()
	}
	return nil
}
