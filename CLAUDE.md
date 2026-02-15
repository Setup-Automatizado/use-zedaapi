# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Production WhatsApp API built on the whatsmeow library. Single Go module (`go.mau.fi/whatsmeow`) containing both the WhatsApp client library (root) and the HTTP API server (`api/`). Targets Z-API webhook compatibility.

## Build & Development Commands

```bash
# Build
go build -v ./...

# Run server
go run api/cmd/server/main.go        # loads api/.env automatically

# Tests
go test -v ./...                      # all tests
go test ./api/internal/nats/... -v    # single package
go test ./... -run TestName -v        # single test
go test -cover ./...                  # coverage
go test -race ./...                   # race detector

# Linting (pre-commit hooks)
pre-commit run --all-files
goimports -local go.mau.fi/whatsmeow -w <file>
go vet ./...

# Code generation
go generate ./...
```

**Go version**: 1.24+ (toolchain 1.25.6). Tabs for indentation (`.editorconfig`).

**Import grouping** (enforced by pre-commit): stdlib, third-party, then `go.mau.fi/whatsmeow/...` local imports.

## Architecture

### Entry Point

`api/cmd/server/main.go` (~1300 lines) wires everything. Startup order: config > logger/sentry/prometheus > databases > redis > NATS (optional) > event orchestrator > dispatch coordinator > media coordinator > message queue > client registry > HTTP server.

### Dual-Mode: PostgreSQL vs NATS

The system supports two backends controlled by `NATS_ENABLED` env var:

```
NATS_ENABLED=false  →  PostgreSQL outbox + LISTEN/NOTIFY + polling
NATS_ENABLED=true   →  NATS JetStream pub/sub
```

Both paths share identical interfaces, swapped at startup via adapters in `main.go`:

| Interface | PostgreSQL impl | NATS impl |
|-----------|----------------|-----------|
| `queue.QueueCoordinator` | `queue.Coordinator` | `queue.NATSCoordinator` |
| `capture.EventWriter` | `capture.TransactionalWriter` | `capture.NATSEventWriter` |
| `dispatch.DispatchCoordinator` | `dispatch.Coordinator` | `dispatch.NATSDispatchCoordinator` |
| `media.MediaCoordinatorProvider` | `media.MediaCoordinator` | `media.NATSMediaCoordinator` |

### Event Pipeline (5 stages)

```
Capture → Buffer → Write → Dispatch → Deliver
```

1. **Capture** (`events/capture/handler.go`): whatsmeow callbacks → `types.InternalEvent`
2. **Buffer** (`events/capture/buffer.go`): per-instance in-memory channel
3. **Write**: `TransactionalWriter` (PG outbox) or `NATSEventWriter` (JetStream)
4. **Dispatch**: workers poll outbox or NATS consumers → resolve webhook URL → Z-API transform
5. **Deliver** (`events/transport/`): HTTP POST to partner webhooks with retry

### Message Queue (FIFO)

Strict per-instance ordering: `MaxAckPending=1`, one worker per instance. Messages published to `messages.{instance_id}`. WhatsApp requires this to avoid out-of-order delivery.

### NATS JetStream Streams

| Stream | Subjects | Retention | Purpose |
|--------|----------|-----------|---------|
| `MESSAGE_QUEUE` | `messages.>` | WorkQueue | Per-instance message sending |
| `WHATSAPP_EVENTS` | `events.>` | Limits | Event capture + webhook dispatch |
| `MEDIA_PROCESSING` | `media.tasks.>`, `media.done.>` | Limits | Async media download/upload |
| `DLQ` | `dlq.>` | Limits (30d) | Dead letter queue |

### Adapter Pattern

`main.go` uses adapter structs to bridge package boundaries without circular imports. Each package defines narrow interfaces for its dependencies (e.g., `queue.ClientRegistry`, `dispatch.WebhookResolver`, `media.ClientProvider`), and `main.go` adapts broader types to satisfy them.

### Webhook URL Routing

Webhook delivery URLs are resolved per event type from `webhook_configs` table. The routing logic in `dispatch/nats_worker.go:resolveWebhookURL` and `capture/writer.go:resolveWebhookURL` must stay in sync. Key rules:
- API-sent messages (`from_api=true`): use `ReceivedDeliveryURL`, fallback to `ReceivedURL`
- `NotifySentByMe` enabled: `ReceivedDeliveryURL` first
- Receipts: `MessageStatusURL`
- Presence: `ChatPresenceURL`
- Unknown event types: drop (return empty string)

## Key Packages

| Package | Purpose |
|---------|---------|
| `api/internal/config` | Environment-driven config, all vars parsed in `Load()` |
| `api/internal/nats` | NATS JetStream client, stream/consumer configs, health checks |
| `api/internal/events/capture` | Event handlers, buffer, writers (PG + NATS) |
| `api/internal/events/dispatch` | Webhook delivery coordinators (PG + NATS) |
| `api/internal/events/media` | Media download/upload, fast path, NATS workers |
| `api/internal/events/nats` | Event envelope types + publisher |
| `api/internal/messages/queue` | Message queue coordinators (PG + NATS) |
| `api/internal/whatsmeow` | ClientRegistry, WhatsApp connection lifecycle |
| `api/internal/locks` | Redis distributed locking + circuit breaker |
| `api/internal/instances` | Instance CRUD repository |
| `api/internal/http/handlers` | Chi HTTP handlers |
| `api/internal/observability` | Prometheus metrics struct |

## Database

**Two PostgreSQL databases**: application DB (`POSTGRES_DSN`) and whatsmeow store (`WAMEOW_POSTGRES_DSN`). Auto-created on startup via `database.EnsureMultipleDatabases`.

**Migrations** in `api/migrations/` (000001-000006), applied automatically via `migrations.Apply()`. Uses raw SQL with pgx, no ORM.

**Key tables**: `instances`, `webhook_configs` (per-instance webhook URLs), `event_outbox` (outbox pattern with sequence_number), `message_queue`, `media_metadata`.

## Configuration

Env vars loaded from `api/.env` (see `api/.env.example` for all options). Critical vars:

```env
POSTGRES_DSN=postgres://user:pass@localhost:5432/funnelchat_api?sslmode=disable
WAMEOW_POSTGRES_DSN=postgres://user:pass@localhost:5432/funnelchat_store?sslmode=disable
REDIS_ADDR=localhost:6379
NATS_ENABLED=true
NATS_URL=nats://localhost:4222
S3_ENDPOINT=http://localhost:9000
CLIENT_AUTH_TOKEN=<min 16 chars, required>
```

Custom duration parsing supports `d` (days) and `w` (weeks): `S3_URL_EXPIRATION=6d`.

## Testing

NATS tests use an embedded NATS server (`nats-server/v2/test`), no external dependencies needed. Test helpers in `api/internal/nats/client_test.go`: `startEmbeddedNATS(t)`, `testConfig(srv)`, `testLogger()`, `testMetrics(t)`.

Table-driven tests colocated with source files. Conventional `*_test.go` naming.

## Critical Invariants

- **Event delivery must never block on media processing**. Fast path tries 5s, then queues async task and delivers event with `media_pending=true`.
- **Message queue FIFO**: `batch_size=1`, `workers_per_instance=1`, `MaxAckPending=1`. Violating this causes WhatsApp message reordering.
- **CLIENT_AUTH_TOKEN**: global, required, >=16 chars. Server aborts if missing. Never log or expose.
- **JetStream MsgID deduplication**: all publishers use stable MsgID keys (not timestamps) to enable safe retry.
- **Webhook URL routing parity**: PG writer and NATS dispatch worker must use identical routing logic.

## Conventions

- **Logging**: `log/slog` with structured fields (`instance_id`, `component`, `event_id`). Never `fmt.Println`.
- **Errors**: wrap with context `fmt.Errorf("operation: %w", err)`. NATS errors include stream/subject.
- **Metrics**: Prometheus counters/histograms for every NATS and event operation.
- **Commits**: Conventional Commits (`fix:`, `feat:`, `chore:`), lowercase, imperative.
