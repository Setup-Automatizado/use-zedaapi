# Repository Guidelines

## Project Structure & Module Organization
Core client logic sits at the module root alongside feature-specific files such as `message.go`, `presence.go`, and `upload.go`. Shared protocol definitions are under `proto/`, transport plumbing in `socket/`, and storage abstractions in `store/` (with `store/sqlstore` providing SQL-backed persistence). Cryptographic helpers and logging utilities live in `util/`, while device state fixtures reside in `appstate/`. Tests belong next to the code they cover; see `client_test.go` for layout and package conventions.

## Build, Test, and Development Commands
Use `go build -v ./...` for a full compile check across supported Go 1.24–1.25 toolchains. Run `go test -v ./...` before every push; narrow investigations with `go test ./... -run TestName`. Generate internal scaffolding via `go generate ./...` whenever touching code referenced by `internals_generate.go`. Format and tidy imports with `goimports -local go.mau.fi/whatsmeow -w <file>`; `pre-commit run --all-files` mirrors the CI workflow.

## Coding Style & Naming Conventions
Rely on `gofmt`/`goimports`; tabs are the default indent per `.editorconfig`, with two-space YAML overrides. Match existing zero-values, option structs, and helper names when extending features. Keep exported identifiers descriptive (`Client`, `DangerousInternals`), and prefix experimental or risky APIs with clear warnings. Never edit generated files in `proto/` or `internals.go` by hand—update source definitions and regenerate instead.

## Testing Guidelines
Write table-driven Go tests in files named `*_test.go` and keep them in the same package as the code under test. Assert observable behaviour rather than private state, and cover both happy paths and WhatsApp edge cases (retries, presence updates, media downloads). Aim to keep `go test -cover ./...` stable; add regression tests whenever fixing bugs surfaced in production logs.

## Observability & Logging Standards (MANDATORY FOR CODE REVIEW)

All PRs MUST pass these observability checks before approval. Violations will result in automatic rejection.

### 1. Structured Logging - MANDATORY

**✅ REQUIRED Pattern**:
```go
// HTTP Handlers - retrieve logger from context
logger := internallogging.ContextLogger(r.Context(), nil)
logger.Info("processing instance operation",
    slog.String("instance_id", instanceID),
    slog.String("operation", "connect"))

// Add attributes to context for downstream propagation
ctx = internallogging.WithAttrs(ctx,
    slog.String("instance_id", instanceID),
    slog.String("partner_id", partnerID))

// Error logging with structured fields
logger.Error("operation failed",
    slog.String("error", err.Error()),
    slog.Int("retry_count", retries))
```

**❌ FORBIDDEN Patterns**:
```go
fmt.Println("processing request")        // REJECTED - use slog
log.Printf("instance %s failed", id)     // REJECTED - use slog
logger.Info("processing ⚠️ critical")    // REJECTED - no emoji
logger.Info("token: " + authToken)       // REJECTED - PII leak
```

**Code Review Checklist - Logging**:
- [ ] ALL logs use `slog` with structured fields (no `fmt.Println`, `log.Print`)
- [ ] Context propagated through function calls (`context.Context` parameter)
- [ ] `ContextLogger` used to retrieve logger from context
- [ ] `WithAttrs` adds domain fields (`instance_id`, `partner_id`, `request_id`)
- [ ] Error logs include `slog.String("error", err.Error())`
- [ ] No emoji characters in log messages
- [ ] No sensitive data (credentials, tokens, PII) in logs
- [ ] Log levels appropriate: DEBUG (diagnostic), INFO (operational), WARN (degraded), ERROR (failures)

### 2. Context Propagation - MANDATORY

**✅ REQUIRED Pattern**:
```go
// Middleware injects logger into context
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqLogger := logger.With(
            slog.String("request_id", uuid.New().String()),
            slog.String("host", r.Host))
        ctx := internallogging.WithLogger(r.Context(), reqLogger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Handler retrieves and enriches context
func (h *Handler) ProcessInstance(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    instanceID := chi.URLParam(r, "id")

    // Add instance_id to context
    ctx = internallogging.WithAttrs(ctx, slog.String("instance_id", instanceID))

    // Pass enriched context to service
    if err := h.service.Execute(ctx, instanceID); err != nil {
        logger := internallogging.ContextLogger(ctx, nil)
        logger.Error("execution failed", slog.String("error", err.Error()))
    }
}
```

**Code Review Checklist - Context**:
- [ ] ALL functions accept `context.Context` as first parameter
- [ ] Middleware injects logger via `WithLogger`
- [ ] Handlers retrieve logger via `ContextLogger`
- [ ] Domain attributes added via `WithAttrs` (instance_id, partner_id)
- [ ] Context passed to ALL downstream calls (services, repositories)
- [ ] Background goroutines receive context copy via `WithLogger`

### 3. Prometheus Metrics - MANDATORY

**✅ REQUIRED Pattern**:
```go
// Update metrics where events occur (not in separate functions)
func (r *Registry) AcquireLock(ctx context.Context, instanceID string) error {
    start := time.Now()

    err := r.lockManager.Acquire(ctx, instanceID)

    // Update metrics immediately after operation
    result := "success"
    if err != nil {
        result = "failure"
    }
    metrics.LockAcquisitions.WithLabelValues(instanceID, result).Inc()
    metrics.LockAcquisitionDuration.WithLabelValues(instanceID).Observe(time.Since(start).Seconds())

    return err
}
```

**Existing Metrics** (MUST update when relevant):
- `HTTPRequestsTotal`: HTTP request counts (method, path, status)
- `HTTPRequestDuration`: HTTP request latency (method, path)
- `WebhookQueueSize`: Webhook outbox size
- `WebhookDeliveryAttempts`: Webhook delivery results (success, failure, retry)
- `LockAcquisitions`: Lock acquisition results (instance_id, success/failure)
- `LockReacquisitionAttempts`: Lock renewal attempts (instance_id)
- `LockReacquisitionDuration`: Lock renewal latency (instance_id)
- `HealthChecks`: Health check results (component, healthy/unhealthy)

**Code Review Checklist - Metrics**:
- [ ] Metrics updated at event occurrence point (not delayed)
- [ ] Labels include relevant context (component, status, instance_id)
- [ ] Avoid high-cardinality labels (no user_id, request_id)
- [ ] Histogram/Summary for durations (not counters)
- [ ] Counter names end with `_total` suffix
- [ ] New metrics documented in `api/internal/observability/metrics.go`

### 4. Sentry Error Tracking - MANDATORY

**✅ REQUIRED Pattern**:
```go
// Capture critical errors with context
if err := criticalOperation(); err != nil {
    logger := internallogging.ContextLogger(ctx, nil)
    logger.Error("critical operation failed", slog.String("error", err.Error()))

    // Capture in Sentry with tags and context
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetTag("component", "registry")
        scope.SetTag("instance_id", instanceID)
        scope.SetTag("severity", "critical")
        scope.SetContext("operation", map[string]interface{}{
            "type":         "lock_acquisition",
            "retry_count":  retries,
            "duration_ms":  duration.Milliseconds(),
        })
        sentry.CaptureException(err)
    })

    return err
}
```

**❌ DO NOT Capture**:
- Expected errors (validation failures, 400/404 responses)
- Errors already logged as warnings
- High-frequency errors (use sampling)

**Code Review Checklist - Sentry**:
- [ ] Only critical errors captured (not validation/expected errors)
- [ ] `WithScope` used to add context (never global tags)
- [ ] Tags include: `component`, `instance_id`, `severity`
- [ ] Context includes operation details (type, attempts, duration)
- [ ] No sensitive data in Sentry captures (credentials, tokens, PII)
- [ ] Sampling configured for high-frequency errors

### 5. Health Checks - MANDATORY

**✅ REQUIRED Pattern**:
```go
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := internallogging.ContextLogger(ctx, nil)

    start := time.Now()
    err := h.db.PingContext(ctx)
    duration := time.Since(start)

    status := "healthy"
    if err != nil {
        status = "unhealthy"

        // Log with structured fields
        logger.Error("database health check failed",
            slog.String("error", err.Error()),
            slog.Duration("duration", duration))

        // Update metrics
        metrics.HealthChecks.WithLabelValues("database", "unhealthy").Inc()

        // Capture in Sentry
        sentry.WithScope(func(scope *sentry.Scope) {
            scope.SetTag("component", "healthcheck")
            scope.SetContext("database", map[string]interface{}{
                "duration_ms": duration.Milliseconds(),
            })
            sentry.CaptureException(err)
        })

        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }

    metrics.HealthChecks.WithLabelValues("database", "healthy").Inc()
    logger.Debug("health check passed", slog.Duration("duration", duration))
    w.WriteHeader(http.StatusOK)
}
```

**Code Review Checklist - Health Checks**:
- [ ] Log duration and status for all checks
- [ ] Update `HealthChecks` metric with component and status labels
- [ ] Capture failures in Sentry with context
- [ ] Return 200 for healthy, 503 for unhealthy
- [ ] `/health`: Liveness probe (basic service health)
- [ ] `/ready`: Readiness probe (dependencies available)

### 6. Complete Code Review Checklist

Before approving PR, verify ALL items:

**Logging**:
- [ ] ALL logs use `slog` with structured fields
- [ ] No `fmt.Println`, `log.Print`, or plain text logging
- [ ] Context propagated through all function calls
- [ ] `instance_id` added to context via `WithAttrs`
- [ ] Errors logged with `slog.String("error", err.Error())`
- [ ] No emoji characters in log messages
- [ ] No sensitive data in logs

**Context**:
- [ ] ALL functions accept `context.Context` as first parameter
- [ ] Middleware injects logger via `WithLogger`
- [ ] Handlers use `ContextLogger` to retrieve logger
- [ ] `WithAttrs` used for domain fields
- [ ] Context passed to downstream calls

**Metrics**:
- [ ] Metrics updated where events occur
- [ ] Labels include relevant context
- [ ] No high-cardinality labels
- [ ] Histogram/Summary for durations

**Sentry**:
- [ ] Only critical errors captured
- [ ] `WithScope` with proper tags/context
- [ ] No sensitive data captured

**Health Checks**:
- [ ] Log duration and status
- [ ] Update metrics
- [ ] Capture failures in Sentry

**General**:
- [ ] Tests pass: `go test ./...`
- [ ] Pre-commit checks pass: `pre-commit run --all-files`
- [ ] OpenAPI docs updated (if API changes)
- [ ] See `PLAN.md` for complete observability standards

## Commit & Pull Request Guidelines
Follow the existing history by using lowercase conventional prefixes (`fix:`, `feat:`, `chore:`) and imperative phrasing (`fix race in prekey upload`). Squash intermediate WIP commits before opening a PR. PRs should explain protocol impacts, reference related issues, and include screenshots or logs when altering client-visible flows. Confirm that `go test ./...` and `pre-commit run --all-files` pass locally before requesting review.

## Security & Configuration Tips
Do not commit real WhatsApp credentials or device states; keep personal `appstate` files out of version control. Sanitise sample payloads added to `store/` or `proto/` and note any redactions in comments. Review dependency bumps for upstream security advisories, and flag breaking protocol changes early so integrators can adapt.
