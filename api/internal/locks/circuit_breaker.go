package locks

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type CircuitState int32

const (
	StateClosed   CircuitState = 0
	StateOpen     CircuitState = 1
	StateHalfOpen CircuitState = 2
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

type CircuitBreakerConfig struct {
	FailureThreshold    int
	OpenDuration        time.Duration
	HalfOpenMaxAttempts int
	HealthCheckInterval time.Duration
}

func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:    3,
		OpenDuration:        30 * time.Second,
		HalfOpenMaxAttempts: 2,
		HealthCheckInterval: 10 * time.Second,
	}
}

type CircuitBreakerManager struct {
	underlying            Manager
	config                CircuitBreakerConfig
	state                 atomic.Int32
	consecutiveFailures   atomic.Int32
	halfOpenAttempts      atomic.Int32
	lastFailureTime       atomic.Int64
	mu                    sync.RWMutex
	healthCheckTicker     *time.Ticker
	stopHealthCheck       chan struct{}
	isHealthChecking      bool
	onStateChange         func(old, new CircuitState)
	lockSuccessCounter    func()
	lockFailureCounter    func()
	circuitStateGauge     func(float64)
	lockReacquireAttempt  func(instanceID, result string)
	lockReacquireFallback func(instanceID, circuitState string)
}

type CircuitBreakerMetricsCallbacks struct {
	LockSuccess       func()
	LockFailure       func()
	CircuitState      func(float64)
	ReacquireAttempt  func(instanceID, result string)
	ReacquireFallback func(instanceID, circuitState string)
}

func NewCircuitBreakerManager(underlying Manager, config CircuitBreakerConfig) *CircuitBreakerManager {
	cbm := &CircuitBreakerManager{
		underlying:      underlying,
		config:          config,
		stopHealthCheck: make(chan struct{}),
	}
	cbm.state.Store(int32(StateClosed))

	cbm.startHealthCheck()

	return cbm
}

func (cbm *CircuitBreakerManager) Acquire(ctx context.Context, key string, ttlSeconds int) (Lock, bool, error) {
	currentState := CircuitState(cbm.state.Load())

	switch currentState {
	case StateClosed:
		return cbm.tryAcquire(ctx, key, ttlSeconds)

	case StateOpen:
		if cbm.shouldAttemptRecovery() {
			cbm.transitionTo(StateHalfOpen)
			return cbm.tryAcquire(ctx, key, ttlSeconds)
		}
		return cbm.fallbackLock(), true, nil

	case StateHalfOpen:
		lock, acquired, err := cbm.tryAcquire(ctx, key, ttlSeconds)
		if err == nil {
			attempts := cbm.halfOpenAttempts.Add(1)
			if attempts >= int32(cbm.config.HalfOpenMaxAttempts) {
				cbm.transitionTo(StateClosed)
				cbm.consecutiveFailures.Store(0)
				cbm.halfOpenAttempts.Store(0)
			}
		} else {
			cbm.recordFailure()
			cbm.transitionTo(StateOpen)
			return cbm.fallbackLock(), true, nil
		}
		return lock, acquired, err

	default:
		return cbm.fallbackLock(), true, errors.New("circuit breaker in unknown state")
	}
}

func (cbm *CircuitBreakerManager) tryAcquire(ctx context.Context, key string, ttlSeconds int) (Lock, bool, error) {
	lock, acquired, err := cbm.underlying.Acquire(ctx, key, ttlSeconds)

	if err != nil {
		cbm.recordFailure()
		if cbm.lockFailureCounter != nil {
			cbm.lockFailureCounter()
		}

		failures := cbm.consecutiveFailures.Load()
		if failures >= int32(cbm.config.FailureThreshold) {
			cbm.transitionTo(StateOpen)
			return cbm.fallbackLock(), true, nil
		}

		return nil, false, err
	}

	cbm.consecutiveFailures.Store(0)
	if cbm.lockSuccessCounter != nil {
		cbm.lockSuccessCounter()
	}
	return lock, acquired, nil
}

func (cbm *CircuitBreakerManager) recordFailure() {
	cbm.consecutiveFailures.Add(1)
	cbm.lastFailureTime.Store(time.Now().Unix())
}

func (cbm *CircuitBreakerManager) shouldAttemptRecovery() bool {
	lastFailure := cbm.lastFailureTime.Load()
	if lastFailure == 0 {
		return true
	}
	elapsed := time.Since(time.Unix(lastFailure, 0))
	return elapsed >= cbm.config.OpenDuration
}

func (cbm *CircuitBreakerManager) transitionTo(newState CircuitState) {
	oldState := CircuitState(cbm.state.Swap(int32(newState)))
	if oldState != newState {
		if cbm.onStateChange != nil {
			cbm.onStateChange(oldState, newState)
		}
		if cbm.circuitStateGauge != nil {
			cbm.circuitStateGauge(float64(newState))
		}
		if newState == StateHalfOpen {
			cbm.halfOpenAttempts.Store(0)
		}
	}
}

func (cbm *CircuitBreakerManager) fallbackLock() Lock {
	return &noOpLock{}
}

func (cbm *CircuitBreakerManager) GetState() CircuitState {
	return CircuitState(cbm.state.Load())
}

func (cbm *CircuitBreakerManager) CheckLockOwnership(ctx context.Context, key string, expectedLock Lock) (bool, error) {
	if expectedLock == nil {
		return false, nil
	}

	ourToken := expectedLock.GetValue()
	if ourToken == "" {
		return false, nil
	}

	err := expectedLock.Refresh(ctx, 30)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (cbm *CircuitBreakerManager) OnStateChange(callback func(old, new CircuitState)) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	cbm.onStateChange = callback
}

func (cbm *CircuitBreakerManager) SetMetrics(callbacks CircuitBreakerMetricsCallbacks) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	cbm.lockSuccessCounter = callbacks.LockSuccess
	cbm.lockFailureCounter = callbacks.LockFailure
	cbm.circuitStateGauge = callbacks.CircuitState
	cbm.lockReacquireAttempt = callbacks.ReacquireAttempt
	cbm.lockReacquireFallback = callbacks.ReacquireFallback

	if callbacks.CircuitState != nil {
		callbacks.CircuitState(float64(cbm.state.Load()))
	}
}

func (cbm *CircuitBreakerManager) RecordLockReacquire(instanceID, result string) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	if cbm.lockReacquireAttempt != nil {
		cbm.lockReacquireAttempt(instanceID, result)
	}
}

func (cbm *CircuitBreakerManager) RecordLockReacquireFallback(instanceID string, state CircuitState) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	if cbm.lockReacquireFallback != nil {
		cbm.lockReacquireFallback(instanceID, state.String())
	}
}

func (cbm *CircuitBreakerManager) startHealthCheck() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if cbm.isHealthChecking {
		return
	}

	cbm.healthCheckTicker = time.NewTicker(cbm.config.HealthCheckInterval)
	cbm.isHealthChecking = true

	go func() {
		for {
			select {
			case <-cbm.healthCheckTicker.C:
				cbm.performHealthCheck()
			case <-cbm.stopHealthCheck:
				return
			}
		}
	}()
}

func (cbm *CircuitBreakerManager) performHealthCheck() {
	currentState := cbm.GetState()

	if currentState != StateOpen {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testKey := "circuit_breaker:health:test"
	lock, acquired, err := cbm.underlying.Acquire(ctx, testKey, 5)

	if err == nil && acquired && lock != nil {
		_ = lock.Release(context.Background())
		if cbm.shouldAttemptRecovery() {
			cbm.transitionTo(StateHalfOpen)
		}
	}
}

func (cbm *CircuitBreakerManager) StopHealthCheck() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if !cbm.isHealthChecking {
		return
	}

	cbm.isHealthChecking = false
	close(cbm.stopHealthCheck)
	if cbm.healthCheckTicker != nil {
		cbm.healthCheckTicker.Stop()
	}
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
