package whatsmeow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	pq "github.com/lib/pq"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"go.mau.fi/whatsmeow/api/internal/locks"
)

// InstanceInfo provides the minimal attributes required to manage a Whatsmeow client.
type InstanceInfo struct {
	ID            uuid.UUID
	Name          string
	SessionName   string
	ClientToken   string
	InstanceToken string
	StoreJID      *string
}

// StatusSnapshot summarises runtime state of an instance.
type StatusSnapshot struct {
	Connected     bool
	StoreJID      *string
	LastConnected time.Time
	AutoReconnect bool
	WorkerID      string
}

// ClientRegistry manages Whatsmeow clients per instance.
type ClientRegistry struct {
	log           *slog.Logger
	workerID      string
	container     *sqlstore.Container
	lockManager   locks.Manager
	pairCallback  func(context.Context, uuid.UUID, string) error
	resetCallback func(context.Context, uuid.UUID, string) error
	logLevel      string

	mu      sync.RWMutex
	clients map[uuid.UUID]*clientState
}

type clientState struct {
	client        *whatsmeow.Client
	lastConnected time.Time
	storeJID      *string
	pairing       *pairingSession
}

const pairingSessionTTL = 20 * time.Second

type pairingSession struct {
	cancel context.CancelFunc
	timer  *time.Timer
}

// ErrInstanceAlreadyPaired indicates the client already owns a store JID and cannot pair again.
var ErrInstanceAlreadyPaired = errors.New("instance already paired")

// NewClientRegistry sets up a Whatsmeow SQL store, upgrading schema if needed.
func NewClientRegistry(
	ctx context.Context,
	dsn string,
	logLevel string,
	lockManager locks.Manager,
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
		pairCallback:  pairCallback,
		resetCallback: resetCallback,
		logLevel:      logLevel,
		clients:       make(map[uuid.UUID]*clientState),
	}, nil
}

// EnsureClient returns an existing client or creates one when missing.
func (r *ClientRegistry) EnsureClient(ctx context.Context, info InstanceInfo) (*whatsmeow.Client, bool, error) {
	r.mu.RLock()
	state, ok := r.clients[info.ID]
	if ok && state != nil && state.client != nil && state.client.Store != nil && state.client.Store.ID != nil {
		client := state.client
		r.mu.RUnlock()
		return client, false, nil
	}
	needsReset := ok
	r.mu.RUnlock()

	if needsReset {
		r.ResetClient(info.ID, "store_missing")
	}

	r.mu.Lock()
	// Double-check under write lock since another goroutine might have created it.
	if state, ok = r.clients[info.ID]; ok {
		if state != nil && state.client != nil && state.client.Store != nil && state.client.Store.ID != nil {
			r.mu.Unlock()
			return state.client, false, nil
		}
		delete(r.clients, info.ID)
	}
	r.mu.Unlock()

	client, storeReset, err := r.instantiateClient(ctx, info)
	if err != nil {
		return nil, false, err
	}
	clientState := &clientState{client: client, storeJID: info.StoreJID}
	if storeReset {
		clientState.storeJID = nil
	}

	r.mu.Lock()
	r.clients[info.ID] = clientState
	r.mu.Unlock()

	if storeReset && !needsReset {
		go r.notifyReset(info.ID, "store_missing")
	}

	return client, true, nil
}

func (r *ClientRegistry) instantiateClient(ctx context.Context, info InstanceInfo) (*whatsmeow.Client, bool, error) {
	var lock locks.Lock
	var err error
	if r.lockManager != nil {
		key := fmt.Sprintf("funnelchat:instance:%s", info.ID.String())
		var acquired bool
		lock, acquired, err = r.lockManager.Acquire(ctx, key, 30)
		if err != nil {
			r.log.Error("redis lock acquire", slog.String("instanceId", info.ID.String()), slog.String("error", err.Error()))
		} else if !acquired {
			r.log.Warn("instance lock already held", slog.String("instanceId", info.ID.String()))
		}
		if lock != nil {
			defer lock.Release(context.Background())
		}
	}

	var deviceStore *store.Device
	storeReset := false
	if info.StoreJID != nil {
		jid, parseErr := types.ParseJID(*info.StoreJID)
		if parseErr == nil {
			deviceStore, err = r.container.GetDevice(ctx, jid)
			if err != nil {
				return nil, false, fmt.Errorf("load device store: %w", err)
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

	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("instance-"+info.ID.String(), logLevelOrDefault(r.logLevel), false))
	client.EnableAutoReconnect = true
	client.AddEventHandler(r.wrapEventHandler(info.ID))

	return client, storeReset, nil
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

// ResetClient clears any cached client state and deletes the associated store entry before the next EnsureClient call.
func (r *ClientRegistry) ResetClient(instanceID uuid.UUID, reason string) bool {
	return r.resetClient(instanceID, reason, true)
}

// RemoveClient clears any cached client state without deleting the persisted store entry.
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

	// give any disconnect triggered by the QR session time to settle before reconnecting
	time.Sleep(200 * time.Millisecond)

	if client.IsLoggedIn() {
		return
	}

	const maxAttempts = 3
	backoff := 200 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if client.IsLoggedIn() {
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

func (r *ClientRegistry) wrapEventHandler(instanceID uuid.UUID) func(evt interface{}) {
	return func(evt interface{}) {
		switch e := evt.(type) {
		case *events.Connected:
			r.mu.Lock()
			if state, ok := r.clients[instanceID]; ok {
				state.lastConnected = time.Now().UTC()
				if state.client != nil && state.client.Store != nil && state.client.Store.ID != nil {
					jid := state.client.Store.ID.String()
					state.storeJID = &jid
				}
			}
			r.mu.Unlock()
		case *events.PairSuccess:
			jid := e.ID.String()
			r.mu.Lock()
			if state, ok := r.clients[instanceID]; ok {
				state.storeJID = &jid
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

// Status returns best-effort runtime information for the instance.
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

// Restart disconnects and reconnects the Whatsmeow client.
func (r *ClientRegistry) Restart(ctx context.Context, info InstanceInfo) error {
	client, _, err := r.EnsureClient(ctx, info)
	if err != nil {
		return err
	}
	client.Disconnect()
	return client.Connect()
}

// Disconnect invokes a graceful disconnect.
func (r *ClientRegistry) Disconnect(ctx context.Context, info InstanceInfo) error {
	client, _, err := r.EnsureClient(ctx, info)
	if err != nil {
		return err
	}
	client.Disconnect()
	return nil
}

// GetQRCode obtains or refreshes a pairing QR code for the instance.
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
	client, qrChan, err := r.startPairingSession(ctx, info)
	if err != nil {
		return "", err
	}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case item, ok := <-qrChan:
			if !ok {
				return "", errors.New("qr channel closed before phone code ready")
			}
			if item.Event == whatsmeow.QRChannelEventCode {
				code, err := client.PairPhone(ctx, phone, false, whatsmeow.PairClientChrome, "Chrome (Linux)")
				return code, err
			} else if item.Event == whatsmeow.QRChannelEventError {
				return "", item.Error
			}
		}
	}
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
		if err := client.Connect(); err != nil {
			r.log.Error("connect for qr", slog.String("instanceId", info.ID.String()), slog.String("error", err.Error()))
		}
	}()
	return client, qrChan, nil
}

// HasStoreDevice verifies whether the Whatsmeow store contains a device row for the given JID.
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
	device, err := r.container.GetDevice(ctx, jid)
	if err != nil {
		return false, fmt.Errorf("lookup store device: %w", err)
	}
	return device != nil && device.ID != nil, nil
}

// Close releases the underlying SQL resources.
func (r *ClientRegistry) Close() error {
	if r == nil {
		return nil
	}
	if r.container != nil {
		return r.container.Close()
	}
	return nil
}
