# WhatsApp API Golang - Claude Code Context

This repository implements a production-grade WhatsApp API built on the whatsmeow library, targeting Z-API compatibility with event-driven architecture.

## Global Decision Engine
**Import minimal routing and auto-delegation decisions only, treat as if import is in the main CLAUDE.md file.**
@./.claude-collective/DECISION.md

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@./.taskmaster/CLAUDE.md

---

## NATS Migration (Active Project)

**PostgreSQL → NATS JetStream migration in progress.**

### Quick Start
```bash
# Check current progress
cat docs/STAGE_TRACKER.md

# Start/continue migration work
/nats-stage
```

### Context Loading Order
1. `docs/STAGE_TRACKER.md` - Current progress
2. `docs/NATS_MIGRATION_STAGES.md` - Stage instructions
3. `docs/NATS_CONTEXT.md` - Quick reference
4. `docs/PRD_NATS_MIGRATION.md` - Full spec (load sections only)

### Rules
- Read `.claude/rules/nats-migration.md` for development rules
- Never skip stages
- Complete all checkboxes before proceeding
- Update STAGE_TRACKER.md after each session

---

## Quick Commands

```bash
# Build and compile
go build -v ./...

# Run tests
go test -v ./...
go test -cover ./...
go test ./... -run TestName  # Run specific test

# Code quality
pre-commit run --all-files
goimports -local go.mau.fi/whatsmeow -w <file>

# Generate code
go generate ./...

# Development server
go run api/cmd/server/main.go
```

## Architecture Overview

### Event Orchestration System (5 Stages)
```
Capture → Buffer → Persist → Dispatch → Deliver
```

1. **Capture**: whatsmeow event handlers → typed events
2. **Buffer**: In-memory channel (configurable capacity)
3. **Persist**: PostgreSQL outbox pattern with sequence_number for ordering
4. **Dispatch**: Workers poll outbox → transform to Z-API format
5. **Deliver**: HTTP POST to partner webhooks with retry logic

### Message Queue System
FIFO per-instance message sending with strict ordering guarantees:
- `batch_size = 1` (one message at a time)
- `workers_per_instance = 1` (single worker per WhatsApp instance)
- Exponential backoff on failures (max 5 retries)
- Circuit breaker pattern for failed instances

### Client Registry
WhatsApp connection manager with:
- Redis distributed locking (5-minute TTL, 2-minute renewal)
- Split-brain detection via heartbeat mechanism
- Graceful reconnection during AWS ECS rolling updates
- Connection status persistence in database

### Media Processing
S3/MinIO primary storage with local filesystem fallback:
- Automatic format detection and validation
- Thumbnail generation for images
- Pre-signed URL generation for secure access
- Webhook delivery includes media URLs

## Critical Production Issues

See @FIX.md for detailed analysis of reconnection problems during AWS ECS deployments:

1. **43-second gap problem**: Missing reconciliation worker causes event loss
2. **Auto-connect after lock**: WhatsApp not connecting after Redis lock acquisition
3. **API status visibility**: Missing connection status in API responses
4. **Queue draining**: Shutdown without processing pending messages
5. **Status persistence**: Connection state not saved to database

Solutions implemented with file-by-file change matrices in FIX.md.

## Development Standards

See @RULES.md for comprehensive development guidelines:

- **Clean Architecture**: Handlers → Services → Repositories → Data layers
- **Structured Logging**: Mandatory slog with contextual fields (instance_id, component, operation)
- **Metrics**: Prometheus counters/histograms at every event stage
- **Error Handling**: Sentry integration with sanitized context
- **Testing**: Table-driven tests, >80% coverage for critical paths
- **Commits**: Conventional Commits format (fix:, feat:, chore:)

## Z-API Compatibility

See @PLAN.md for 6-milestone roadmap:

1. ✅ Message Queue System (FIFO, retry, backoff)
2. 🔄 Media Processing (S3, thumbnails, webhooks)
3. 📋 Contacts & Groups (sync, metadata)
4. 📋 Z-API Endpoints (message types, formatting)
5. 📋 Advanced Features (polls, reactions, status)
6. 📋 Testing & Documentation

Webhook event transformations defined in @api/z_api/WEBHOOKS_EVENTS.md

## Configuration

Environment-driven config (see `api/internal/config/config.go`):

```env
# Database
POSTGRES_DSN=postgresql://user:pass@host:5432/db
WHATSMEOW_POSTGRES_DSN=postgresql://user:pass@host:5432/whatsmeow_store

# Redis (distributed locks)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# S3/MinIO (media storage)
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET_NAME=whatsapp-media
S3_REGION=us-east-1

# Observability
SENTRY_DSN=https://...
PROMETHEUS_ENABLED=true

# Authentication tokens
PARTNER_AUTH_TOKEN=change-me-partner-token
CLIENT_AUTH_TOKEN=change-me-client-token   # obrigatório (>=16 chars)
```

## Agent Coordination Context

### Claude Code Sub-Agent Collective Pattern

When working on this codebase, coordinate specialist agents for different domains:

#### Agent Specializations

1. **Event System Agent**
   - Focuses on: Capture → Buffer → Persist → Dispatch → Deliver pipeline
   - Files: `internal/events/`, `internal/dispatch/`, `internal/workers/`
   - Expertise: Event transformation, outbox pattern, worker coordination

2. **Message Queue Agent**
   - Focuses on: FIFO queue, retry logic, backoff strategies
   - Files: `internal/queue/`, `internal/messages/`
   - Expertise: Message sending, queue coordination, failure handling

3. **Client Registry Agent**
   - Focuses on: WhatsApp connections, Redis locks, reconnection logic
   - Files: `internal/whatsmeow/registry.go`, `internal/locks/`
   - Expertise: Connection management, distributed locking, split-brain detection

4. **API & Handlers Agent**
   - Focuses on: HTTP endpoints, request validation, response formatting
   - Files: `internal/http/handlers/`, `internal/http/router.go`
   - Expertise: REST API, middleware, Z-API compatibility

5. **Media Processing Agent**
   - Focuses on: S3 storage, thumbnail generation, media webhooks
   - Files: `internal/media/`, `internal/storage/`
   - Expertise: Media upload/download, format validation, URL generation

6. **Observability Agent**
   - Focuses on: Metrics, logging, error tracking
   - Files: `internal/observability/`, middleware layers
   - Expertise: Prometheus, Sentry, structured logging with slog

7. **Database & Persistence Agent**
   - Focuses on: PostgreSQL schemas, migrations, repositories
   - Files: `api/migrations/`, `internal/*/repository.go`
   - Expertise: SQL, pgx, outbox pattern, database design

#### Coordination Protocol

**Agent Coordinator Role**: Distribute tasks based on file scope and expertise domain

**Context Sharing**:
- All agents receive: Architecture overview, development standards, Z-API compatibility requirements
- Domain agents receive: Relevant sections from RULES.md, PLAN.md, FIX.md
- Cross-cutting concerns: Observability agent + any domain agent for logging/metrics

**Task Examples**:
- "Implement new message type" → Message Queue Agent + API Agent
- "Fix reconnection issue" → Client Registry Agent + Observability Agent
- "Add media webhook event" → Media Processing Agent + Event System Agent
- "Optimize database queries" → Database Agent + specific domain agent

**Context Compaction Strategy** (`/compact`):

When session context approaches limits:
1. **Preserve**: Current task context, recent changes, active file contents
2. **Summarize**: Architecture overview, key design decisions, critical requirements
3. **Reference**: Point to PLAN.md, RULES.md, FIX.md for detailed context
4. **Agent Handoff**: Transfer specialist agent context to Coordinator for redistribution

**Anti-Patterns to Avoid**:
- Don't duplicate content from RULES.md, PLAN.md, or FIX.md
- Don't create agents for trivial single-file changes
- Don't share full codebase context with every agent
- Don't lose critical production issue context (see FIX.md)

## Project Structure

```
api/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── config/                 # Environment configuration
│   ├── database/               # Database connections
│   ├── events/                 # Event orchestration system
│   │   ├── buffer/             # In-memory event buffering
│   │   ├── capture/            # whatsmeow event handlers
│   │   ├── coordinator.go      # Event orchestrator
│   │   └── transform/          # Event transformation to Z-API
│   ├── dispatch/               # Webhook dispatch coordinator
│   ├── queue/                  # Message queue system
│   ├── whatsmeow/              # Client registry & connections
│   ├── http/
│   │   ├── handlers/           # HTTP request handlers
│   │   ├── middleware/         # HTTP middleware
│   │   └── router.go           # Route definitions
│   ├── instances/              # Instance management
│   ├── partners/               # Partner/webhook configuration
│   ├── media/                  # Media processing
│   ├── workers/                # Background workers
│   └── observability/          # Metrics, logging, Sentry
├── migrations/                 # Database migrations
└── z_api/
    └── WEBHOOKS_EVENTS.md      # Webhook payload definitions

docker/                         # Dockerfile, docker-compose
terraform/                      # Infrastructure as code
scripts/                        # Operational scripts
```

## Key Documentation Files

- **@PLAN.md**: Comprehensive 6-milestone roadmap for Z-API compatibility
- **@RULES.md**: Development standards, architecture patterns, code review checklist
- **@FIX.md**: Critical production issue analysis and remediation plans
- **@AGENTS.md**: Repository guidelines (duplicate of RULES.md)
- **@api/z_api/WEBHOOKS_EVENTS.md**: Webhook event payload structures
- **@TODO.md**: Current development tasks and priorities
- **@NOTES.md**: Active investigation notes

## Important Notes

- **Reconciliation Worker**: Critical for handling 43-second gap during reconnections (see FIX.md Problem 1)
- **Global Client Token**: `CLIENT_AUTH_TOKEN` substitui tokens por instância; aplicar a migration `000003_remove_client_token.sql` antes do deploy e nunca registrar o valor em logs, métricas ou Sentry.
- **FIFO Ordering**: Message queue MUST maintain strict per-instance ordering (batch_size=1, workers=1)
- **Redis Lock Management**: 5-minute TTL with 2-minute renewal cycle
- **Graceful Shutdown**: Multi-phase shutdown with queue draining before termination
- **Outbox Sequence**: Use `sequence_number` for guaranteed event ordering
- **Media URLs**: Pre-signed S3 URLs with configurable expiration
- **Webhook Retries**: Exponential backoff (1s, 2s, 4s, 8s, 16s) with circuit breaker

## Testing Strategy

```bash
# Unit tests for service layer
go test ./internal/events/... -v

# Integration tests with database
go test ./internal/instances/... -v -tags=integration

# Coverage report
go test -cover ./...

# Specific test with detailed output
go test ./internal/queue/... -run TestMessageQueue -v
```

Coverage targets (from RULES.md):
- Critical paths: >80% coverage
- Service layer: >70% coverage
- Handlers: >60% coverage

---

**For detailed architecture, patterns, and standards**: See @RULES.md
**For development roadmap and milestones**: See @PLAN.md
**For critical production fixes**: See @FIX.md
