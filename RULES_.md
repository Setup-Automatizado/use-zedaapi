# Development Rules

## Authentication Tokens
- `CLIENT_AUTH_TOKEN` is a mandatory, single value loaded from the environment. The server refuses to start when it is blank or shorter than 16 characters (`config.Load`).
- Never persist or expose per-instance client tokens. All code paths must accept the shared header, compare against the configured env value, and avoid including it in API payloads, logs, or database tables.
- `PARTNER_AUTH_TOKEN` remains the only token for partner-prefixed routes. Keep it secret-managed alongside `CLIENT_AUTH_TOKEN` (compose, Terraform, Secrets Manager).

## Schema & Migrations
- Fresh installs rely on `api/migrations/000001_init.sql` with no `client_token` column. Existing databases migrate via `000003_remove_client_token.sql` (must be applied before deploying the refactored binaries).
- Whenever a migration manipulates authentication state, document the rollback behaviour and verify goose down scripts.
- Run `go test ./...` after altering migrations or repositories to confirm repository structs stay aligned with schema changes.

## Service & Registry Layers
- `instances.Service` must receive the global token through dependency injection (`NewService`). Any helper mapping instances to `whatsmeow.InstanceInfo` must exclude token fields.
- Whenever you add new handlers or providers that need to authenticate instance calls, reuse `instances.Service.tokensMatch` instead of reimplementing header checks.
- Webhook dispatch (`events/capture`) must populate the `Client-Token` header using `cfg.Client.AuthToken`. Do not introduce alternate secret sources.

## Configuration & Infrastructure
- Compose, Terraform modules, and examples must declare both `PARTNER_AUTH_TOKEN` and `CLIENT_AUTH_TOKEN` within `secret_env_mapping` to keep runtime environments consistent.
- Update secrets when rotating tokens; the binary requires a restart to load new env values.
- Never hardcode production credentials. Use `additional_secret_values` only for local templates and mark real values as `CHANGE_ME`.

## Observability & Security
- Logs must not emit token values. When you need to confirm presence of a token in a request, log boolean flags (e.g., `has_client_token=true`).
- Prometheus metrics or tracing spans should not include secret values; use identifiers such as instance IDs or request hashes instead.

## Testing & Verification
- Regression tests that previously depended on per-instance tokens must be updated to use a shared header value from fixtures or environment overrides.
- When adding new integration tests, seed `CLIENT_AUTH_TOKEN` via `t.Setenv` to mirror production behaviour.

# Regras de Desenvolvimento e Arquitetura - WhatsApp API

**Versão Completa e Definitiva**

Este documento é a fonte única de verdade para o desenvolvimento do projeto. Todos os Pull Requests (PRs) e contribuições **devem** seguir estas regras para serem aprovados.

---

## Índice

1. [Visão Geral e Princípios](#1-visão-geral-e-princípios)
2. [Arquitetura Completa](#2-arquitetura-completa)
3. [Estrutura do Projeto](#3-estrutura-do-projeto)
4. [Padrões de Código](#4-padrões-de-código)
5. [Regras de Negócio](#5-regras-de-negócio)
6. [Observabilidade (MANDATÓRIO)](#6-observabilidade-mandatório)
7. [Database e Migrations](#7-database-e-migrations)
8. [Worker Patterns](#8-worker-patterns)
9. [HTTP API Patterns](#9-http-api-patterns)
10. [Testing Strategy](#10-testing-strategy)
11. [Configuration Management](#11-configuration-management)
12. [Security Best Practices](#12-security-best-practices)
13. [Performance Requirements](#13-performance-requirements)
14. [Z-API Compatibility](#14-z-api-compatibility)
15. [Deployment & Operations](#15-deployment--operations)
16. [Code Review Checklist](#16-code-review-checklist)
17. [Commits e Pull Requests](#17-commits-e-pull-requests)

---

## 1. Visão Geral e Princípios

### 1.1. Objetivo do Projeto

Desenvolver uma API WhatsApp robusta, escalável e observável que:
- Mantenha **compatibilidade funcional com Z-API**
- Implemente **fila de mensagens confiável** com ordenação por instância
- Garanta **alta disponibilidade** através de workers distribuídos
- Forneça **observabilidade completa** (logs, métricas, tracing)

### 1.2. Princípios Fundamentais

**Arquitetura**:
- **Clean Architecture**: Separação clara de responsabilidades em camadas
- **Event-Driven**: Sistema orientado a eventos com buffers e workers
- **Idempotência**: Operações podem ser reexecutadas sem efeitos colaterais
- **Horizontal Scalability**: Múltiplos workers processando em paralelo

**Qualidade de Código**:
- **Testabilidade**: Interfaces e injeção de dependências
- **Observabilidade**: Logs estruturados, métricas em todos os pontos críticos
- **Error Handling**: Erros sempre tratados e propagados com contexto
- **Performance**: Otimização baseada em métricas reais

**Operacional**:
- **Graceful Shutdown**: Desligamento ordenado sem perda de dados
- **Circuit Breakers**: Proteção contra falhas em cascata
- **Retry Logic**: Tentativas com backoff exponencial
- **Lock Management**: Coordenação distribuída via Redis

---

## 2. Arquitetura Completa

### 2.1. Clean Architecture - Camadas

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  /internal/http/handlers - HTTP Endpoints                    │
│  Responsabilidades:                                          │
│  - Parse e validação de requests                             │
│  - Autenticação (Client-Token + instance token)              │
│  - Serialização de responses                                 │
│  - NÃO DEVE conter lógica de negócio                         │
└─────────────────────────────────────────────────────────────┘
                           ↓ calls
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  /internal/{domain}/service.go                               │
│  Responsabilidades:                                          │
│  - Orquestração de lógica de negócio                         │
│  - Validações complexas de domínio                           │
│  - Coordenação entre múltiplos repositórios                  │
│  - Enforcement de regras de negócio                          │
└─────────────────────────────────────────────────────────────┘
                           ↓ calls
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                          │
│  /internal/{domain}/repository.go                            │
│  Responsabilidades:                                          │
│  - Queries SQL (pgx/sqlx)                                    │
│  - Mapeamento ORM → Domain Models                            │
│  - Transaction management                                    │
│  - NÃO DEVE conter lógica de negócio                         │
└─────────────────────────────────────────────────────────────┘
                           ↓ persists to
┌─────────────────────────────────────────────────────────────┐
│                     Data Layer                               │
│  PostgreSQL (main + whatsmeow_store)                         │
│  Redis (distributed locks)                                   │
│  S3/MinIO (media storage)                                    │
└─────────────────────────────────────────────────────────────┘
```

**Fluxo de Dados**:
```
HTTP Request → Handler → Service → Repository → Database
            ↓ logs       ↓ metrics ↓ context   ↓ transaction
         Response ← DTO ← Domain ← Entity ← SQL Result
```

**Regras de Dependência**:
- Handlers **SÓ** chamam Services (nunca Repositories diretamente)
- Services **PODEM** chamar múltiplos Repositories
- Repositories **NÃO** chamam outros Repositories
- Camadas internas **NÃO** conhecem camadas externas
- Domain Models são definidos na camada Service/Repository

### 2.2. Event Orchestration System

O sistema de eventos é o coração da aplicação, responsável por capturar, processar e entregar eventos WhatsApp de forma confiável e ordenada.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    WhatsApp Events (via whatsmeow)                   │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  STAGE 1: CAPTURE (/internal/events/capture)                        │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │ EventHandler (per instance)                              │       │
│  │ - Receives raw WhatsApp events                           │       │
│  │ - Enriches with metadata (instance_id, timestamp)        │       │
│  │ - Routes to appropriate buffer                           │       │
│  └──────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  STAGE 2: BUFFER (/internal/events/capture)                         │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │ EventBuffer (per instance)                               │       │
│  │ - In-memory buffer (size: EVENT_BUFFER_SIZE)             │       │
│  │ - Batching (EVENT_BATCH_SIZE events)                     │       │
│  │ - Auto-flush (EVENT_POLL_INTERVAL)                       │       │
│  │ - Backpressure handling                                  │       │
│  └──────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  STAGE 3: PERSIST (/internal/events/persistence)                    │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │ TransactionalWriter                                      │       │
│  │ - Assigns sequence_number per instance                   │       │
│  │ - Stores in event_outbox (PostgreSQL)                    │       │
│  │ - Extracts media metadata → media_outbox                 │       │
│  │ - Transaction guarantees                                 │       │
│  └──────────────────────────────────────────────────────────┘       │
│                                                                       │
│  Database Tables:                                                    │
│  - event_outbox: Main queue (status: pending/processing/sent/failed) │
│  - event_dlq: Dead Letter Queue for permanent failures               │
│  - media_outbox: Media processing queue                              │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  STAGE 4: DISPATCH (/internal/events/dispatch)                      │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │ DispatchCoordinator                                      │       │
│  │ - Manages InstanceWorkers (one per connected instance)   │       │
│  │ - Registers/Unregisters on connect/disconnect events     │       │
│  │ - Monitors worker health                                 │       │
│  └──────────────────────────────────────────────────────────┘       │
│                       ↓ spawns                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │ InstanceWorker (per instance)                            │       │
│  │ - Poll loop: SELECT FOR UPDATE SKIP LOCKED               │       │
│  │ - Batch processing (EVENT_BATCH_SIZE)                    │       │
│  │ - Respects sequence_number ordering                      │       │
│  │ - Retry logic with exponential backoff                   │       │
│  └──────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  STAGE 5: DELIVER (/internal/events/transport)                      │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │ TransportRegistry                                        │       │
│  │ - HTTP transport (webhooks)                              │       │
│  │ - Retry with circuit breaker                             │       │
│  │ - Timeout handling (WEBHOOK_TIMEOUT)                     │       │
│  └──────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                    Webhook Endpoints (Customer)                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Garantias do Sistema**:
1. **Ordenação por Instância**: sequence_number assegura ordem FIFO per instance
2. **At-Least-Once Delivery**: Retry até sucesso ou DLQ
3. **Idempotência**: event_id UUID previne duplicação
4. **Observabilidade Total**: Logs e métricas em cada stage
5. **Graceful Degradation**: Circuit breakers e fallbacks

### 2.3. Media Processing System

Sistema paralelo para download e upload de mídias WhatsApp.

```
┌─────────────────────────────────────────────────────────────────────┐
│  WhatsApp Media Event → media_outbox (during event persistence)     │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  MediaCoordinator (/internal/events/media)                          │
│  - Spawns workers based on MEDIA_WORKER_CONCURRENCY                 │
│  - Distributes work across workers                                  │
│  - Monitors download/upload progress                                │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  MediaWorker (concurrent pool)                                      │
│  STEP 1: Download from WhatsApp                                     │
│    → whatsmeow.Client.Download()                                    │
│    → Timeout: MEDIA_DOWNLOAD_TIMEOUT (5m)                           │
│                                                                      │
│  STEP 2: Upload to S3/MinIO                                         │
│    → Primary: S3Uploader                                            │
│    → Fallback: LocalMediaStorage                                    │
│    → Timeout: MEDIA_UPLOAD_TIMEOUT (10m)                            │
│                                                                      │
│  STEP 3: Generate URL                                               │
│    → S3: Presigned URL (S3_URL_EXPIRATION)                          │
│    → Local: Signed local URL (MEDIA_LOCAL_URL_EXPIRY)               │
│                                                                      │
│  STEP 4: Update event_outbox                                        │
│    → Replace directPath with uploaded URL                           │
│    → Update media_status to 'uploaded'                              │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  Storage Backends                                                    │
│  - S3/MinIO (primary)                                                │
│  - Local filesystem (fallback + dev)                                 │
└─────────────────────────────────────────────────────────────────────┘
```

**Cleanup Pipeline**:
```
MediaReaper (periodic job)
  → Scans old media records
  → Uploads local → S3 (if not already there)
  → Deletes local files > LOCAL_MEDIA_RETENTION
  → Deletes S3 objects > S3_MEDIA_RETENTION
  → Uses distributed lock (Redis) to prevent concurrent cleanup
```

### 2.4. Message Queue System (Planned)

Sistema de fila de envio de mensagens com compatibilidade Z-API.

```
┌─────────────────────────────────────────────────────────────────────┐
│  POST /send-text, /send-image, etc. (HTTP API)                      │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  MessageService                                                      │
│  - Validates request (phone format, tokens, subscription)            │
│  - Serializes message payload                                        │
│  - Calculates scheduled_at (delay + delayMessage)                    │
│  - Enqueues to message_queue                                         │
│  - Returns: {"status":"QUEUED","queueId":"...","messageId":"..."}    │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  message_queue (PostgreSQL)                                          │
│  - sequence_number per instance (ordering)                           │
│  - status: pending/processing/sent/failed/canceled                   │
│  - scheduled_at, next_attempt_at (retry control)                     │
│  - delayMessage, attempts, max_attempts                              │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  MessageCoordinator                                                  │
│  - One MessageWorker per connected instance                          │
│  - Registers on instance connect events                              │
│  - Unregisters on disconnect                                         │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  MessageWorker (per instance)                                        │
│  Poll Loop:                                                          │
│    1. SELECT pending messages WHERE next_attempt_at <= NOW()         │
│       ORDER BY sequence_number FOR UPDATE SKIP LOCKED                │
│    2. Check instance connected status                                │
│    3. Send via whatsmeow.Client.SendMessage()                        │
│    4. Apply delay: random(1-3s) + delayMessage                       │
│    5. Update status (sent/retrying/failed)                           │
│    6. Retry with exponential backoff if needed                       │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  WhatsApp Business API (via whatsmeow)                               │
└─────────────────────────────────────────────────────────────────────┘
```

**Características**:
- **Ordenação FIFO** por instância
- **Delay Humanizado**: 1-3s random + delayMessage opcional
- **Isolamento**: Falha de uma instância não afeta outras
- **Retry Inteligente**: Backoff exponencial, max_attempts
- **Tolerância a Desconexões**: Reagenda automaticamente

### 2.5. ClientRegistry Architecture

Gerenciador central de conexões WhatsApp.

```
ClientRegistry
├── Container (whatsmeow.Container)
│   └── SQLStore (Postgres whatsmeow_store DB)
│
├── Client Management
│   ├── clients: map[uuid.UUID]*whatsmeow.Client
│   ├── EnsureClient(instanceInfo) → creates/returns client
│   ├── DisconnectAll() → graceful shutdown
│   └── ReleaseAllLocks() → Redis cleanup
│
├── Lock Management
│   ├── lockManager (CircuitBreaker + RedisManager)
│   ├── AcquireLock() → distributed coordination
│   ├── ReacquireLock() → periodic renewal (30s)
│   └── FallbackLock() → in-memory when Redis fails
│
├── Event Integration
│   ├── EventOrchestrator.RegisterInstance() on connect
│   ├── EventOrchestrator.UnregisterInstance() on disconnect
│   ├── DispatchCoordinator.RegisterInstance()
│   └── MediaCoordinator.RegisterInstance()
│
├── Split-Brain Detection
│   ├── Periodic scan (every 60s)
│   ├── Detects: clients with empty lock tokens
│   └── Action: force disconnect invalid clients
│
└── Callbacks
    ├── PairCallback → updates instances.store_jid
    └── ResetCallback → clears instances.store_jid on logout
```

**Lock States**:
- **Memory Mode**: No Redis lock (single instance mode)
- **Redis Mode**: Distributed lock with token
- **Fallback Mode**: In-memory lock when circuit breaker open

**Worker Sessions & Ownership**
- Cada worker registra batidas de coração na tabela `worker_sessions` (Postgres) usando `WORKER_HEARTBEAT_INTERVAL` e expirações controladas por `WORKER_HEARTBEAT_EXPIRY`.
- A posse de cada instância é definida por hashing determinístico (rendezvous) sobre os workers ativos; apenas o worker eleito cria/gera conexões.
- O registrador agenda um rebalanceamento periódico (`WORKER_REBALANCE_INTERVAL`) que encerra clientes atendidos pelo worker errado e libera o lock para o novo dono.
- `desired_worker_id` em `instances` reflete o dono esperado, enquanto `worker_id` guarda o dono atual.

**Redis Lock Configurável**
- `REDIS_LOCK_KEY_PREFIX` define o namespace das chaves (ex.: `prod`, `staging`).
- `REDIS_LOCK_TTL` e `REDIS_LOCK_REFRESH_INTERVAL` controlam o tempo de vida do lock e a cadência de renovação (default 30s/15s). Sempre mantenha `refresh < ttl` para evitar perda acidental de lock.
- Healthchecks e circuit breaker continuam reportando `local`/`redis` mode, mas agora o fallback só é utilizado pelo worker eleito, evitando split-brain durante falhas no Redis.

### 2.6. Configuration Management

```
config.Config (loaded from env vars)
├── AppEnv (development/staging/production)
├── HTTP (addr, timeouts, base URL)
├── Postgres (DSN, max conns)
├── WhatsmeowStore (DSN, log level)
	├── Redis (addr, credentials, TLS)
	├── RedisLock (key prefix, ttl, refresh interval)
	├── S3 (endpoint, bucket, credentials, presigned URLs)
├── Media (storage path, retention, cleanup)
├── Sentry (DSN, environment, release)
	├── Workers (dispatcher count, media count)
	├── WorkerRegistry (heartbeat interval, expiry, rebalance)
├── Prometheus (namespace)
├── Partner (auth token)
└── Events (buffer, batch, intervals, retries, circuit breaker)
```

Todas configurações via **variáveis de ambiente** (12-factor app).

---

## 3. Estrutura do Projeto

### 3.1. Diretórios Principais

```
api/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point, initialization
│
├── internal/                          # Private application code
│   ├── config/                        # Configuration loading
│   │   └── config.go
│   │
│   ├── database/                      # Database connections
│   │   └── postgres.go
│   │
│   ├── http/                          # HTTP layer
│   │   ├── router.go                  # Route registration
│   │   ├── server.go                  # HTTP server
│   │   ├── handlers/                  # Request handlers
│   │   │   ├── instances.go
│   │   │   ├── partners.go
│   │   │   ├── health.go
│   │   │   ├── media.go
│   │   │   └── response.go            # Standard responses
│   │   └── middleware/                # HTTP middleware
│   │       ├── auth.go
│   │       ├── logging.go
│   │       └── metrics.go
│   │
│   ├── instances/                     # Instance domain
│   │   ├── model.go                   # Domain models
│   │   ├── repository.go              # Data access
│   │   └── service.go                 # Business logic
│   │
│   ├── events/                        # Event orchestration
│   │   ├── orchestrator.go            # Main coordinator
│   │   ├── integration.go             # Integration helpers
│   │   ├── capture/                   # Event capture
│   │   │   ├── handler.go
│   │   │   ├── buffer.go
│   │   │   ├── router.go
│   │   │   └── writer.go
│   │   ├── dispatch/                  # Event dispatch
│   │   │   ├── coordinator.go
│   │   │   └── worker.go
│   │   ├── media/                     # Media processing
│   │   │   ├── coordinator.go
│   │   │   ├── worker.go
│   │   │   ├── uploader.go
│   │   │   ├── storage.go
│   │   │   └── reaper.go
│   │   ├── persistence/               # Data layer
│   │   │   ├── outbox.go
│   │   │   ├── dlq.go
│   │   │   └── media.go
│   │   ├── transport/                 # Delivery transports
│   │   │   ├── registry.go
│   │   │   └── http/
│   │   ├── transform/                 # Payload transformation
│   │   ├── types/                     # Event types
│   │   └── eventctx/                  # Context helpers
│   │
│   ├── whatsmeow/                     # WhatsApp client
│   │   ├── registry.go                # ClientRegistry
│   │   └── split_brain.go             # Split-brain detection
│   │
│   ├── locks/                         # Distributed locking
│   │   ├── manager.go
│   │   ├── circuit_breaker.go
│   │   └── redis.go
│   │
│   ├── observability/                 # Monitoring
│   │   ├── metrics.go                 # Prometheus metrics
│   │   └── async.go                   # Async metric updates
│   │
│   ├── logging/                       # Structured logging
│   │   ├── logging.go
│   │   └── context.go                 # Context propagation
│   │
│   ├── redis/                         # Redis client
│   │   └── client.go
│   │
│   └── sentry/                        # Error tracking
│       └── init.go
│
├── migrations/                        # SQL migrations
│   ├── 000001_init.sql
│   └── ...
│
├── docs/                              # API documentation
│   ├── openapi.yaml                   # OpenAPI spec (source of truth)
│   ├── openapi.json
│   └── http.go                        # Docs server
│
├── z_api/                             # Z-API reference
│   └── Nova API Collection.postman_collection.json
│
├── docker-compose.dev.yml             # Dev environment
├── .env.example                       # Example configuration
├── DEVELOPMENT_RULES.md               # This file
└── PLAN.md                            # Implementation roadmap
```

### 3.2. Convenções de Nomenclatura

**Pacotes**:
- Sempre minúsculas, sem underscores
- Nome descritivo do domínio: `instances`, `events`, `locks`
- Evite pacotes genéricos como `utils`, `helpers`

**Arquivos**:
- `model.go`: Domain models e DTOs
- `repository.go`: Data access layer
- `service.go`: Business logic orchestration
- `handler.go`: HTTP handlers
- `*_test.go`: Tests collocated with code

**Estruturas e Interfaces**:
- PascalCase para exportados: `ClientRegistry`, `EventHandler`
- camelCase para não-exportados: `workerPool`, `eventBuffer`
- Interface names: sem sufixo "Interface" (Go convention)
  - Correto: `Repository`, `Uploader`
  - Incorreto: `RepositoryInterface`, `IUploader`

**Funções**:
- Construtores: `NewXxx()` retorna `*Xxx`
- Factory: `CreateXxx()` quando lógica complexa
- Getters: omitir "Get" prefix (`Status()` não `GetStatus()`)
  - Exceção: quando ambiguidade (`GetByID` é claro)

---

## 4. Padrões de Código

### 4.1. Context Propagation (OBRIGATÓRIO)

**REGRA**: `context.Context` deve ser o **primeiro parâmetro** de toda função que:
- Faz I/O (DB, HTTP, Redis)
- Chama outra função que aceita context
- Pode ser cancelada ou ter timeout

**Pattern Completo**:

```go
// main.go - Criação do contexto raiz
func main() {
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT, syscall.SIGTERM,
    )
    defer cancel()

    logger := logging.New(cfg.Log.Level)

    // Injeta logger no contexto raiz
    ctx = logging.WithLogger(ctx, logger)

    // Passa para inicialização
    if err := run(ctx, cfg); err != nil {
        logger.Error("startup failed", slog.String("error", err.Error()))
        os.Exit(1)
    }
}

// Middleware - Enriquece contexto por request
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Cria logger com request_id
            reqLogger := logger.With(
                slog.String("request_id", uuid.New().String()),
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
                slog.String("remote_addr", r.RemoteAddr),
            )

            // Injeta no contexto
            ctx := logging.WithLogger(r.Context(), reqLogger)

            // Passa request com novo contexto
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Handler - Adiciona atributos de domínio
func (h *InstanceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    instanceID := chi.URLParam(r, "instanceId")

    // Enriquece contexto com instance_id
    ctx = logging.WithAttrs(ctx,
        slog.String("instance_id", instanceID),
        slog.String("operation", "get_status"),
    )

    // Recupera logger enriquecido
    logger := logging.ContextLogger(ctx, h.log)
    logger.Info("processing status request")

    // Passa contexto enriquecido para service
    status, err := h.service.GetStatus(ctx, instanceID, clientToken, instanceToken)
    if err != nil {
        logger.Error("status request failed",
            slog.String("error", err.Error()))
        respondError(w, http.StatusInternalServerError, "failed to get status")
        return
    }

    logger.Info("status request completed",
        slog.Bool("connected", status.Connected))
    respondJSON(w, http.StatusOK, status)
}

// Service - Propaga contexto
func (s *Service) GetStatus(
    ctx context.Context,
    instanceID uuid.UUID,
    clientToken, instanceToken string,
) (*Status, error) {
    logger := logging.ContextLogger(ctx, s.log)

    // Valida tokens
    inst, err := s.repo.GetByID(ctx, instanceID)
    if err != nil {
        return nil, fmt.Errorf("get instance: %w", err)
    }

    if !s.tokensMatch(inst, clientToken, instanceToken) {
        logger.Warn("unauthorized access attempt",
            slog.String("provided_client_token", clientToken[:8]+"..."))
        return nil, ErrUnauthorized
    }

    // Consulta registry (passa contexto)
    snapshot := s.registry.Status(ctx, toInstanceInfo(*inst))

    return &Status{
        Connected:     snapshot.Connected,
        StoreJID:      snapshot.StoreJID,
        LastConnected: snapshot.LastConnected,
    }, nil
}

// Repository - Context em queries
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Instance, error) {
    query := `SELECT id, name, client_token, ... FROM instances WHERE id = $1`

    // pgx usa context para timeout/cancelamento
    row := r.pool.QueryRow(ctx, query, id)

    var inst Instance
    if err := row.Scan(&inst.ID, &inst.Name, ...); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrInstanceNotFound
        }
        return nil, fmt.Errorf("query instance: %w", err)
    }

    return &inst, nil
}

// Background goroutine - Cria novo contexto
func (s *Service) processAsync(parentCtx context.Context, data Data) {
    // Cria contexto filho com timeout
    ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
    defer cancel()

    // Propaga logger do parent
    logger := logging.ContextLogger(parentCtx, s.log)
    ctx = logging.WithLogger(ctx, logger)

    // Adiciona contexto específico da goroutine
    ctx = logging.WithAttrs(ctx,
        slog.String("task", "async_processing"),
        slog.String("data_id", data.ID.String()),
    )

    go func() {
        if err := s.process(ctx, data); err != nil {
            logger := logging.ContextLogger(ctx, s.log)
            logger.Error("async processing failed",
                slog.String("error", err.Error()))
        }
    }()
}
```

**Checklist Context Propagation**:
- [ ] `context.Context` é **primeiro parâmetro** em todas as funções
- [ ] Middleware injeta logger via `logging.WithLogger`
- [ ] Handlers adicionam atributos de domínio via `logging.WithAttrs`
- [ ] Services recuperam logger via `logging.ContextLogger(ctx, fallback)`
- [ ] Repositories usam `ctx` em todas as queries DB/Redis
- [ ] Background goroutines criam contexto filho com `WithTimeout`
- [ ] Nunca passar `context.Background()` exceto em `main()` e testes

#### 4.1.1. Padrão whatsmeow Client API (CRÍTICO)

**IMPORTANTE**: A partir de 2025, a biblioteca whatsmeow foi atualizada para exigir `context.Context` como **primeiro parâmetro** em todas as chamadas de API do cliente WhatsApp.

**Padrão Obrigatório para Adapter Functions**:

```go
// ✅ CORRETO: Função que recebe ctx - passa ele através
func (a *whatsAppClientAdapter) GetSubscribedNewsletters(ctx context.Context) ([]*types.NewsletterMetadata, error) {
    if a.client == nil {
        return nil, ErrClientNotConnected
    }
    return a.client.GetSubscribedNewsletters(ctx)  // Passa ctx recebido
}

func (a *whatsAppClientAdapter) CreateNewsletter(ctx context.Context, params whatsmeowclient.CreateNewsletterParams) (*types.NewsletterMetadata, error) {
    if a.client == nil {
        return nil, ErrClientNotConnected
    }
    return a.client.CreateNewsletter(ctx, params)  // Passa ctx recebido
}

// ✅ CORRETO: Função sem ctx parameter - usa context.Background()
func (a *whatsAppClientAdapter) FollowNewsletter(id types.JID) error {
    if a.client == nil {
        return ErrClientNotConnected
    }
    return a.client.FollowNewsletter(context.Background(), id)
}

func (a *whatsAppClientAdapter) GetGroupInfo(jid types.JID) (*types.GroupInfo, error) {
    if a.client == nil {
        return nil, ErrClientNotConnected
    }
    return a.client.GetGroupInfo(context.Background(), jid)
}

// ❌ ERRADO: Não passar context
func (a *whatsAppClientAdapter) GetGroupInfo(jid types.JID) (*types.GroupInfo, error) {
    return a.client.GetGroupInfo(jid)  // ERRO: falta context.Context
}

// ❌ ERRADO: Descartar ctx recebido
func (a *whatsAppClientAdapter) CreateNewsletter(ctx context.Context, params whatsmeowclient.CreateNewsletterParams) (*types.NewsletterMetadata, error) {
    _ = ctx  // NÃO FAZER ISSO
    return a.client.CreateNewsletter(params)  // ERRO: falta context
}
```

**Funções whatsmeow Comumente Usadas que Exigem Context**:

Grupos:
- `client.GetGroupInfo(ctx, jid)`
- `client.GetJoinedGroups(ctx)`
- `client.CreateGroup(ctx, req)`
- `client.SetGroupName(ctx, jid, name)`
- `client.SetGroupPhoto(ctx, jid, avatar)`
- `client.UpdateGroupParticipants(ctx, jid, changes, action)`
- `client.LeaveGroup(ctx, jid)`
- `client.GetGroupInviteLink(ctx, jid, reset)`
- `client.JoinGroupWithLink(ctx, code)`
- `client.SetGroupAnnounce(ctx, jid, announce)`
- `client.SetGroupLocked(ctx, jid, locked)`
- `client.SetGroupDescription(ctx, jid, description)`

Comunidades:
- `client.GetSubGroups(ctx, community)`
- `client.GetLinkedGroupsParticipants(ctx, community)`
- `client.LinkGroup(ctx, parent, child)`
- `client.UnlinkGroup(ctx, parent, child)`

Newsletters:
- `client.GetSubscribedNewsletters(ctx)`
- `client.CreateNewsletter(ctx, params)`
- `client.FollowNewsletter(ctx, id)`
- `client.UnfollowNewsletter(ctx, id)`
- `client.NewsletterToggleMute(ctx, id, mute)`
- `client.GetNewsletterInfo(ctx, id)`

Mensagens e Presença:
- `client.SendMessage(ctx, to, message)`
- `client.SendChatPresence(ctx, jid, state, media)`
- `client.SendPresence(ctx, presence)`

Contatos e Perfil:
- `client.GetProfilePictureInfo(ctx, jid, params)`

**Regra de Decisão**:
1. Se a função adapter **recebe** `ctx context.Context` como parâmetro → **SEMPRE passe ele através** para a API whatsmeow
2. Se a função adapter **NÃO recebe** ctx → Use `context.Background()` na chamada whatsmeow
3. **NUNCA** descarte ctx recebido com `_ = ctx`
4. **NUNCA** omita o parâmetro context em chamadas whatsmeow

**Histórico de Mudança**:
- **Antes de 2025**: whatsmeow APIs não exigiam context
- **Após merge upstream 2025**: Todas as APIs whatsmeow foram atualizadas para exigir `context.Context` como primeiro parâmetro
- **Impacto**: 42 funções foram atualizadas em 8 arquivos durante a integração upstream

**Arquivos Adapter Afetados**:
- `api/internal/groups/client_provider.go`
- `api/internal/communities/client_provider.go`
- `api/internal/newsletters/client_provider.go`
- `api/internal/messages/queue/*.go`
- `api/internal/whatsmeow/contact_metadata.go`
- `api/internal/whatsmeow/registry.go`

### 4.2. Error Handling (OBRIGATÓRIO)

**REGRAS**:
1. **Nunca ignore erros** (`_`)
2. **Sempre adicione contexto** ao propagar erros
3. **Wrap errors** das camadas inferiores
4. **Log antes de retornar** erros críticos
5. **Nunca use `panic`** para erros esperados

**Pattern**:

```go
// Repository - Erro base
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Instance, error) {
    row := r.pool.QueryRow(ctx, query, id)

    var inst Instance
    if err := row.Scan(&inst.ID, ...); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrInstanceNotFound // Erro de domínio
        }
        // Wrap com contexto
        return nil, fmt.Errorf("query instance %s: %w", id, err)
    }

    return &inst, nil
}

// Service - Adiciona contexto de negócio
func (s *Service) Restart(ctx context.Context, id uuid.UUID) error {
    logger := logging.ContextLogger(ctx, s.log)

    inst, err := s.repo.GetByID(ctx, id)
    if err != nil {
        // Distingue erros de domínio
        if errors.Is(err, ErrInstanceNotFound) {
            logger.Warn("restart failed: instance not found")
            return err // Propaga erro de domínio
        }
        // Wrap erro inesperado
        return fmt.Errorf("restart instance: %w", err)
    }

    if !inst.SubscriptionActive {
        logger.Warn("restart blocked: subscription inactive")
        return ErrInstanceInactive
    }

    if err := s.registry.Restart(ctx, id); err != nil {
        logger.Error("registry restart failed",
            slog.String("error", err.Error()))

        // Captura em Sentry para erros críticos
        sentry.WithScope(func(scope *sentry.Scope) {
            scope.SetTag("component", "instance_service")
            scope.SetTag("instance_id", id.String())
            scope.SetTag("operation", "restart")
            sentry.CaptureException(err)
        })

        return fmt.Errorf("restart client: %w", err)
    }

    logger.Info("instance restarted successfully")
    return nil
}

// Handler - Traduz erros para HTTP
func (h *InstanceHandler) Restart(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    instanceID, clientToken, instanceToken := h.parseParams(r)

    if err := h.service.Restart(ctx, instanceID, clientToken, instanceToken); err != nil {
        h.handleServiceError(ctx, w, err)
        return
    }

    respondJSON(w, http.StatusOK, map[string]bool{"restarted": true})
}

func (h *InstanceHandler) handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
    logger := logging.ContextLogger(ctx, h.log)

    switch {
    case errors.Is(err, ErrUnauthorized):
        logger.Warn("unauthorized request")
        respondError(w, http.StatusUnauthorized, "unauthorized")

    case errors.Is(err, ErrInstanceNotFound):
        logger.Warn("instance not found")
        respondError(w, http.StatusNotFound, "instance not found")

    case errors.Is(err, ErrInstanceInactive):
        logger.Warn("subscription inactive")
        respondError(w, http.StatusForbidden, "subscription inactive")

    default:
        logger.Error("internal error", slog.String("error", err.Error()))
        respondError(w, http.StatusInternalServerError, "internal server error")
    }
}
```

**Definição de Erros de Domínio**:

```go
// errors.go - Erros de domínio no pacote do service
var (
    ErrInstanceNotFound      = errors.New("instance not found")
    ErrUnauthorized          = errors.New("unauthorized")
    ErrInstanceInactive      = errors.New("subscription inactive")
    ErrInstanceAlreadyPaired = errors.New("instance already paired")
    ErrInvalidWebhookURL     = errors.New("webhook url must use https")
)

// Erros com contexto
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s - %s", e.Field, e.Message)
}
```

**Checklist Error Handling**:
- [ ] Nenhum erro ignorado com `_`
- [ ] Todos os erros wrapeados com `fmt.Errorf("context: %w", err)`
- [ ] Erros de domínio definidos como `var Err... = errors.New(...)`
- [ ] Services usam `errors.Is()` para distinguir erros
- [ ] Handlers traduzem erros para status HTTP apropriados
- [ ] Erros críticos capturados no Sentry com contexto
- [ ] Logs antes de retornar erros

### 4.3. Adapter Pattern

Usado para compatibilizar interfaces entre camadas e pacotes.

**Caso 1: Repository → ClientRegistry**

```go
// ClientRegistry espera:
type StoreJIDRepository interface {
    ListInstancesWithStoreJID(ctx context.Context) ([]StoreLink, error)
}

// Mas instances.Repository tem método diferente
// Criamos adapter:

type repositoryAdapter struct {
    repo *instances.Repository
}

func (a *repositoryAdapter) ListInstancesWithStoreJID(ctx context.Context) ([]whatsmeow.StoreLink, error) {
    links, err := a.repo.ListInstancesWithStoreJID(ctx)
    if err != nil {
        return nil, err
    }

    // Converte entre tipos
    result := make([]whatsmeow.StoreLink, len(links))
    for i, link := range links {
        result[i] = whatsmeow.StoreLink{
            ID:       link.ID,
            StoreJID: link.StoreJID,
        }
    }
    return result, nil
}

// Usage em main.go:
repoAdapter := &repositoryAdapter{repo: instancesRepo}
registry := whatsmeow.NewClientRegistry(ctx, cfg, lockManager, repoAdapter, ...)
```

**Caso 2: Webhook Resolver**

```go
// EventOrchestrator espera:
type WebhookResolver interface {
    Resolve(ctx context.Context, instanceID uuid.UUID) (*ResolvedWebhookConfig, error)
}

// Adapter combina instances.Repository + webhook config:

type webhookResolverAdapter struct {
    repo *instances.Repository
}

func (a *webhookResolverAdapter) Resolve(ctx context.Context, id uuid.UUID) (*capture.ResolvedWebhookConfig, error) {
    inst, err := a.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    webhook, err := a.repo.GetWebhookConfig(ctx, id)
    if err != nil {
        return nil, err
    }

    deref := func(ptr *string) string {
        if ptr == nil {
            return ""
        }
        return *ptr
    }

    return &capture.ResolvedWebhookConfig{
        DeliveryURL:         deref(webhook.DeliveryURL),
        ReceivedURL:         deref(webhook.ReceivedURL),
        MessageStatusURL:    deref(webhook.MessageStatusURL),
        StoreJID:            inst.StoreJID,
        ClientToken:         inst.ClientToken,
    }, nil
}
```

**Quando Usar Adapters**:
- Compatibilizar interfaces entre pacotes diferentes
- Evitar dependências circulares
- Manter contratos limpos nas camadas internas
- Facilitar testes com mocks

### 4.4. Graceful Shutdown

**OBRIGATÓRIO**: Todos os workers, coordinators e servidores devem suportar shutdown gracioso.

**Pattern Completo**:

```go
// main.go - Orchestration
func main() {
    // Signal context - cancela em SIGINT/SIGTERM
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT, syscall.SIGTERM,
    )
    defer cancel()

    // ... inicialização ...

    // Emergency timeout: força exit após 45s
    emergencyTimeout := time.AfterFunc(45*time.Second, func() {
        logger.Error("EMERGENCY TIMEOUT: Forcing exit after 45 seconds")
        os.Exit(1)
    })
    defer emergencyTimeout.Stop()

    // 1. HTTP Server shutdown
    if err := server.Run(ctx); err != nil {
        logger.Error("http server stopped", slog.String("error", err.Error()))
    }

    logger.Info("starting graceful shutdown sequence")

    // 2. Release Redis locks
    logger.Info("releasing redis locks")
    released := registry.ReleaseAllLocks()
    logger.Info("redis locks released", slog.Int("count", released))

    // 3. Disconnect WhatsApp clients
    logger.Info("disconnecting whatsapp clients")
    disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
    disconnected := registry.DisconnectAll(disconnectCtx)
    disconnectCancel()
    logger.Info("clients disconnected", slog.Int("count", disconnected))

    // 4. Close ClientRegistry
    logger.Info("closing client registry")
    registryCloseDone := make(chan error, 1)
    go func() {
        registryCloseDone <- registry.Close()
    }()

    select {
    case err := <-registryCloseDone:
        if err != nil {
            logger.Error("registry close failed", slog.String("error", err.Error()))
        }
    case <-time.After(5 * time.Second):
        logger.Warn("registry close timeout after 5 seconds")
    }

    // 5. Stop dispatch coordinator
    logger.Info("stopping dispatch coordinator")
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    dispatchCoordinator.Stop(shutdownCtx)
    shutdownCancel()
    logger.Info("dispatch coordinator stopped")

    // 6. Stop event orchestrator
    logger.Info("stopping event orchestrator")
    shutdownCtx2, shutdownCancel2 := context.WithTimeout(context.Background(), 10*time.Second)
    eventOrchestrator.Stop(shutdownCtx2)
    shutdownCancel2()
    logger.Info("event orchestrator stopped")

    // 7. Flush Sentry events
    if cfg.Sentry.DSN != "" {
        logger.Info("flushing sentry events")
        sentry.Flush(5 * time.Second)
    }

    logger.Info("shutdown complete")
}

// Worker - Implementação padrão
type InstanceWorker struct {
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
    instanceID   uuid.UUID
    pollInterval time.Duration
    log          *slog.Logger
}

func (w *InstanceWorker) Start() {
    w.ctx, w.cancel = context.WithCancel(context.Background())

    w.wg.Add(1)
    go func() {
        defer w.wg.Done()
        w.run()
    }()
}

func (w *InstanceWorker) run() {
    logger := w.log.With(
        slog.String("instance_id", w.instanceID.String()),
        slog.String("component", "instance_worker"),
    )

    logger.Info("worker started")

    ticker := time.NewTicker(w.pollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-w.ctx.Done():
            logger.Info("worker shutting down gracefully")
            return

        case <-ticker.C:
            // Process batch
            if err := w.processBatch(w.ctx); err != nil {
                logger.Error("batch processing failed",
                    slog.String("error", err.Error()))
            }
        }
    }
}

func (w *InstanceWorker) Stop() {
    w.log.Info("stopping worker",
        slog.String("instance_id", w.instanceID.String()))

    w.cancel() // Signal goroutine to stop
    w.wg.Wait() // Wait for completion

    w.log.Info("worker stopped")
}

// Coordinator - Manages multiple workers
type Coordinator struct {
    ctx     context.Context
    cancel  context.CancelFunc
    workers map[uuid.UUID]*InstanceWorker
    mu      sync.RWMutex
    wg      sync.WaitGroup
}

func (c *Coordinator) Start(ctx context.Context) error {
    c.ctx, c.cancel = context.WithCancel(ctx)
    c.workers = make(map[uuid.UUID]*InstanceWorker)

    logger.Info("coordinator started")
    return nil
}

func (c *Coordinator) RegisterInstance(instanceID uuid.UUID) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if _, exists := c.workers[instanceID]; exists {
        return // Already registered
    }

    worker := NewInstanceWorker(c.ctx, instanceID, c.config)
    worker.Start()

    c.workers[instanceID] = worker

    logger.Info("worker registered",
        slog.String("instance_id", instanceID.String()))
}

func (c *Coordinator) UnregisterInstance(instanceID uuid.UUID) {
    c.mu.Lock()
    worker, exists := c.workers[instanceID]
    if exists {
        delete(c.workers, instanceID)
    }
    c.mu.Unlock()

    if exists {
        worker.Stop()
        logger.Info("worker unregistered",
            slog.String("instance_id", instanceID.String()))
    }
}

func (c *Coordinator) Stop(ctx context.Context) {
    logger.Info("stopping coordinator")

    c.cancel() // Cancel coordinator context

    // Stop all workers concurrently
    c.mu.Lock()
    workersCopy := make(map[uuid.UUID]*InstanceWorker)
    for id, w := range c.workers {
        workersCopy[id] = w
    }
    c.workers = nil
    c.mu.Unlock()

    // Stop with timeout
    done := make(chan struct{})
    go func() {
        for id, worker := range workersCopy {
            worker.Stop()
            logger.Debug("worker stopped",
                slog.String("instance_id", id.String()))
        }
        close(done)
    }()

    select {
    case <-done:
        logger.Info("all workers stopped")
    case <-ctx.Done():
        logger.Warn("coordinator stop timeout, some workers may still be running")
    }
}
```

**Checklist Graceful Shutdown**:
- [ ] `context.Context` usado para cancelamento
- [ ] `sync.WaitGroup` para esperar goroutines
- [ ] Ticker/Timer com `defer ticker.Stop()`
- [ ] Workers respondem a `ctx.Done()`
- [ ] Coordinators param workers antes de retornar
- [ ] Timeout em operações de shutdown (30s típico)
- [ ] Logs de inicio e fim de shutdown
- [ ] Emergency timeout global (45s)

---

## 5. Regras de Negócio e Domínio

### 5.1. Instance Lifecycle

**Estados de uma Instância**:

```
  [Created]
      ↓ EnsureClient()
  [Initialized]
      ↓ GetQRCode() ou GetPhoneCode()
  [Pairing]
      ↓ User scans QR / enters code
  [Paired] (store_jid populated)
      ↓ Auto-connect
  [Connected]
      ↓ User disconnect / network issue
  [Disconnected]
      ↓ Restart()
  [Connected]
      ↓ Subscription expires
  [Inactive]
```

**Regras de Transição**:

| Estado Atual | Ação Permitida | Próximo Estado | Validações |
|--------------|----------------|----------------|------------|
| Created | EnsureClient | Initialized | subscription_active = true |
| Initialized | GetQRCode/PhoneCode | Pairing | Não já paired |
| Pairing | [user action] | Paired | store_jid recebido |
| Paired | Auto-connect | Connected | Redis lock acquired |
| Connected | Disconnect | Disconnected | Release Redis lock |
| Disconnected | Restart | Connected | subscription_active = true |
| Any | [subscription expires] | Inactive | subscription_active = false |

**Business Rules**:

1. **Subscription Enforcement**:
   ```go
   func (s *Service) validateSubscription(inst *Instance) error {
       if !inst.SubscriptionActive {
           return ErrInstanceInactive
       }
       if inst.CanceledAt != nil && time.Now().After(*inst.CanceledAt) {
           return ErrInstanceInactive
       }
       return nil
   }
   ```

2. **Pairing Prevention**:
   ```go
   func (s *Service) GetQRCode(ctx context.Context, id uuid.UUID) (string, error) {
       inst, err := s.repo.GetByID(ctx, id)
       if err != nil {
           return "", err
       }

       if inst.StoreJID != nil && *inst.StoreJID != "" {
           return "", ErrInstanceAlreadyPaired
       }

       // ... rest of logic
   }
   ```

3. **Auto-Connect on Pairing**:
   - Quando `PairCallback` é invocado (store_jid populated)
   - ClientRegistry automaticamente chama `client.Connect()`
   - DispatchCoordinator.RegisterInstance() é chamado
   - EventOrchestrator.RegisterInstance() é chamado

### 5.2. Authentication & Authorization

**Modelo de Autenticação Dupla**:

Todas as rotas de instância requerem **dois tokens**:

1. **Client-Token** (Header): Autenticação a nível de cliente
2. **Instance Token** (Path): Autorização a nível de instância

```
GET /instances/{instanceId}/token/{instanceToken}/status
Headers:
  Client-Token: <client_token_from_instance_record>
```

**Validação Pattern**:

```go
func (s *Service) tokensMatch(
    inst *Instance,
    clientToken, instanceToken string,
) bool {
    return inst.ClientToken == clientToken &&
           inst.InstanceToken == instanceToken
}

func (s *Service) requireAuth(
    ctx context.Context,
    id uuid.UUID,
    clientToken, instanceToken string,
) (*Instance, error) {
    inst, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if !s.tokensMatch(inst, clientToken, instanceToken) {
        logger := logging.ContextLogger(ctx, s.log)
        logger.Warn("authentication failed",
            slog.String("instance_id", id.String()))
        return nil, ErrUnauthorized
    }

    return inst, nil
}
```

**Token Generation**:
```go
// Na criação da instância
inst := Instance{
    ID:            uuid.New(),
    ClientToken:   uuid.NewString(), // Gerado uma única vez
    InstanceToken: uuid.NewString(), // Gerado uma única vez
    // ...
}
```

**Partner API Authentication**:

Partner endpoints usam token global:

```go
// Middleware
func PartnerAuth(partnerToken string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")

            if authHeader != "Bearer "+partnerToken {
                respondError(w, http.StatusUnauthorized, "invalid partner token")
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 5.3. Webhook Configuration Rules

**HTTPS Enforcement**:

```go
func validateWebhookURL(url string) error {
    if url == "" {
        return nil // Empty is valid (webhook disabled)
    }

    parsed, err := url.Parse(url)
    if err != nil {
        return fmt.Errorf("invalid url: %w", err)
    }

    if parsed.Scheme != "https" {
        return ErrInvalidWebhookURL
    }

    return nil
}
```

**Webhook Types**:

| Webhook | Evento | Obrigatório | Default |
|---------|--------|-------------|---------|
| delivery_url | message.delivery | Não | null |
| received_url | message.received | Não | null |
| received_delivery_url | message.received + delivery | Não | null |
| message_status_url | message.status | Não | null |
| disconnected_url | instance.disconnected | Não | null |
| connected_url | instance.connected | Não | null |
| chat_presence_url | chat.presence | Não | null |

**notify_sent_by_me Flag**:
- Quando `true`: inclui mensagens enviadas pelo próprio usuário
- Quando `false`: filtra mensagens own no EventRouter

**Atualização Pattern**:

```go
func (s *Service) UpdateWebhookDelivery(
    ctx context.Context,
    id uuid.UUID,
    clientToken, instanceToken, url string,
) error {
    inst, err := s.requireAuth(ctx, id, clientToken, instanceToken)
    if err != nil {
        return err
    }

    if err := validateWebhookURL(url); err != nil {
        return err
    }

    return s.repo.UpdateWebhookDeliveryURL(ctx, id, &url)
}
```

### 5.4. Media Processing Rules

**Size Limits**:

```go
const (
    MaxFileSize   = 100 * 1024 * 1024 // 100MB
    ChunkSize     = 5 * 1024 * 1024   // 5MB for streaming
)

func validateMediaSize(size int64) error {
    if size > MaxFileSize {
        return fmt.Errorf("file too large: %d bytes (max %d)", size, MaxFileSize)
    }
    return nil
}
```

**Storage Strategy**:

1. **Primary**: S3/MinIO
   - Presigned URLs (default)
   - Public URLs (if S3_USE_PRESIGNED_URLS=false)
   - Expiration: S3_URL_EXPIRATION (30d default)

2. **Fallback**: Local filesystem
   - Activated when S3 upload fails
   - Signed URLs com HMAC
   - Expiration: MEDIA_LOCAL_URL_EXPIRY (720h default)

**Retention Policy**:

```go
// MediaReaper - Cleanup Job
func (r *MediaReaper) Run(ctx context.Context) {
    // 1. Upload local files to S3 (if not already there)
    r.uploadLocalToS3(ctx)

    // 2. Delete local files older than LOCAL_MEDIA_RETENTION
    r.cleanupLocal(ctx, time.Now().Add(-cfg.LocalRetention))

    // 3. Delete S3 objects older than S3_MEDIA_RETENTION
    r.cleanupS3(ctx, time.Now().Add(-cfg.S3Retention))
}
```

**Media Types Supported**:

- **Image**: image/jpeg, image/png, image/webp
- **Video**: video/mp4, video/3gpp
- **Audio**: audio/ogg, audio/mpeg, audio/aac
- **Document**: application/pdf, application/msword, etc.
- **Sticker**: image/webp (animated)

### 5.5. Event Ordering Guarantees

**Per-Instance Ordering**:

Events for a single instance are processed in the order received from WhatsApp.

```sql
-- Sequence function assegura ordenação
CREATE OR REPLACE FUNCTION get_next_event_sequence(p_instance_id UUID)
RETURNS BIGINT AS $$
DECLARE
    next_seq BIGINT;
BEGIN
    SELECT COALESCE(MAX(sequence_number), 0) + 1
    INTO next_seq
    FROM event_outbox
    WHERE instance_id = p_instance_id;

    RETURN next_seq;
END;
$$ LANGUAGE plpgsql;
```

**Worker Processing**:

```go
// InstanceWorker garante ordenação
func (w *InstanceWorker) poll(ctx context.Context) ([]*OutboxEvent, error) {
    query := `
        SELECT id, instance_id, event_type, payload, sequence_number, attempts
        FROM event_outbox
        WHERE instance_id = $1
          AND status = 'pending'
          AND next_attempt_at <= NOW()
        ORDER BY sequence_number ASC  -- CRITICAL: ordem FIFO
        LIMIT $2
        FOR UPDATE SKIP LOCKED
    `

    return w.repo.Query(ctx, query, w.instanceID, w.batchSize)
}
```

**Garantia**: Se evento A foi gerado antes de evento B na mesma instância, A será processado antes de B.

### 5.6. Retry and Backoff Strategy

**Exponential Backoff**:

```go
var DefaultRetryDelays = []time.Duration{
    0 * time.Second,      // Attempt 1: immediate
    10 * time.Second,     // Attempt 2: 10s
    30 * time.Second,     // Attempt 3: 30s
    2 * time.Minute,      // Attempt 4: 2m
    5 * time.Minute,      // Attempt 5: 5m
    15 * time.Minute,     // Attempt 6: 15m
}

func calculateNextAttemptAt(attempt int) time.Time {
    if attempt >= len(DefaultRetryDelays) {
        // Max attempts reached → DLQ
        return time.Time{}
    }

    delay := DefaultRetryDelays[attempt]
    return time.Now().Add(delay)
}
```

**Retry Decision Logic**:

```go
func (w *InstanceWorker) handleDeliveryError(
    ctx context.Context,
    event *OutboxEvent,
    err error,
) error {
    event.Attempts++
    event.LastError = err.Error()

    if event.Attempts >= event.MaxAttempts {
        // Move to DLQ
        logger.Error("max attempts reached, moving to DLQ",
            slog.String("event_id", event.ID.String()),
            slog.Int("attempts", event.Attempts))

        if err := w.dlqRepo.Insert(ctx, event, "max_attempts_exceeded"); err != nil {
            return fmt.Errorf("dlq insert: %w", err)
        }

        if err := w.outboxRepo.Delete(ctx, event.ID); err != nil {
            return fmt.Errorf("outbox delete: %w", err)
        }

        metrics.DLQEventsTotal.WithLabelValues(
            event.InstanceID.String(),
            event.EventType,
            "max_attempts",
        ).Inc()

        return nil
    }

    // Schedule retry
    event.NextAttemptAt = calculateNextAttemptAt(event.Attempts)
    event.Status = "retrying"

    return w.outboxRepo.Update(ctx, event)
}
```

**Circuit Breaker Integration**:

Se webhook endpoint está falhando consistentemente:
1. Circuit breaker abre
2. Eventos são mantidos em `pending` (não tentados)
3. After cooldown period, circuit vai para `half-open`
4. Tenta algumas requests de teste
5. Se sucesso → `closed`, retoma processamento normal
6. Se falha → `open`, aguarda novo cooldown

### 5.7. Lock Management Rules

**Distributed Lock Pattern**:

```go
// Acquire lock before connecting client
func (r *ClientRegistry) EnsureClient(
    ctx context.Context,
    info InstanceInfo,
) (*whatsmeow.Client, error) {
    // Try acquire Redis lock
    lock, err := r.lockManager.AcquireLock(ctx, info.ID.String(), 60*time.Second)
    if err != nil {
        if errors.Is(err, locks.ErrLockHeld) {
            return nil, fmt.Errorf("instance locked by another process")
        }
        // Circuit breaker open → use fallback
        lock = r.lockManager.FallbackLock(info.ID.String())
    }

    // Store lock with client
    r.mu.Lock()
    r.locks[info.ID] = lock
    r.mu.Unlock()

    // Create client...
}
```

**Lock Renewal**:

```go
// Background goroutine renews lock every 30s
func (r *ClientRegistry) renewLocks(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            r.mu.RLock()
            locksCopy := make(map[uuid.UUID]locks.Lock)
            for id, lock := range r.locks {
                locksCopy[id] = lock
            }
            r.mu.RUnlock()

            for id, lock := range locksCopy {
                if err := lock.Renew(ctx, 60*time.Second); err != nil {
                    logger.Error("lock renewal failed",
                        slog.String("instance_id", id.String()),
                        slog.String("error", err.Error()))

                    // Disconnect client if lock lost
                    r.handleLockLoss(ctx, id)
                }
            }
        }
    }
}
```

**Split-Brain Prevention**:

```go
// Periodic check: ensure all connected clients have valid locks
func (r *ClientRegistry) detectSplitBrain(ctx context.Context) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for id, client := range r.clients {
        if !client.IsConnected() {
            continue
        }

        lock, hasLock := r.locks[id]

        // Invalid state: connected but no lock
        if !hasLock || lock.Token() == "" {
            logger.Error("split-brain detected: connected without valid lock",
                slog.String("instance_id", id.String()))

            metrics.SplitBrainDetected.Inc()

            // Force disconnect
            client.Disconnect()
            delete(r.clients, id)
            delete(r.locks, id)
        }
    }
}
```

### 5.8. Message Queue Rules (Planned)

**Delay Policies**:

```go
type SendMessageParams struct {
    Phone         string                 `json:"phone"`
    Message       string                 `json:"message"`
    DelayMessage  int                    `json:"delayMessage,omitempty"` // milliseconds
    MessageID     string                 `json:"messageId,omitempty"`    // custom ID
}

// Cálculo do scheduled_at
func calculateScheduledAt(delayMessage int) time.Time {
    // Base delay: random 1-3 segundos (humanização)
    baseDelay := time.Duration(rand.Intn(2000)+1000) * time.Millisecond

    // Add custom delay if specified
    customDelay := time.Duration(delayMessage) * time.Millisecond

    return time.Now().Add(baseDelay + customDelay)
}
```

**Queue Management**:

```go
// Enqueue
func (s *MessageService) SendText(
    ctx context.Context,
    instanceID uuid.UUID,
    params SendMessageParams,
) (*QueueResponse, error) {
    // Validate instance + tokens
    inst, err := s.validateInstance(ctx, instanceID, clientToken, instanceToken)
    if err != nil {
        return nil, err
    }

    // Check subscription
    if !inst.SubscriptionActive {
        return nil, ErrInstanceInactive
    }

    // Normalize phone para enviar @s.whatsapp.net, @g.us, @newsletter, @lid, @broadcast e @hosted!
    jid, err := normalizeJID(params.Phone)
    if err != nil {
        return nil, ErrInvalidPhoneNumber
    }

    // Create queue entry
    queueEntry := &QueuedMessage{
        ID:            uuid.New(),
        InstanceID:    instanceID,
        MessageID:     params.MessageID, // client-provided or generated
        Type:          "text",
        Payload:       marshalPayload(params),
        Status:        "pending",
        ScheduledAt:   calculateScheduledAt(params.DelayMessage),
        NextAttemptAt: calculateScheduledAt(params.DelayMessage),
        Attempts:      0,
        MaxAttempts:   3,
    }

    if err := s.queueRepo.Enqueue(ctx, queueEntry); err != nil {
        return nil, err
    }

    metrics.MessagesQueued.WithLabelValues(
        instanceID.String(),
        "text",
    ).Inc()

    return &QueueResponse{
        Status:    "QUEUED",
        QueueID:   queueEntry.ID.String(),
        MessageID: queueEntry.MessageID,
    }, nil
}

// Dequeue and Send
func (w *MessageWorker) processBatch(ctx context.Context) error {
    messages, err := w.queueRepo.Poll(ctx, w.instanceID, w.batchSize)
    if err != nil {
        return err
    }

    for _, msg := range messages {
        // Check if instance still connected
        if !w.registry.IsConnected(w.instanceID) {
            // Reschedule for later
            msg.NextAttemptAt = time.Now().Add(30 * time.Second)
            w.queueRepo.Update(ctx, msg)
            continue
        }

        // Send via whatsmeow
        if err := w.sender.Send(ctx, msg); err != nil {
            w.handleSendError(ctx, msg, err)
            continue
        }

        // Mark as sent
        msg.Status = "sent"
        msg.DeliveredAt = time.Now()
        w.queueRepo.Update(ctx, msg)

        metrics.MessageSendAttempts.WithLabelValues(
            w.instanceID.String(),
            msg.Type,
            "success",
        ).Inc()

        // Apply delay before next message
        applyDelay(msg.DelayMessage)
    }

    return nil
}

func applyDelay(customDelay int) {
    // Random 1-3s
    baseDelay := time.Duration(rand.Intn(2000)+1000) * time.Millisecond

    // Add custom delay
    totalDelay := baseDelay + time.Duration(customDelay)*time.Millisecond

    time.Sleep(totalDelay)
}
```

---

## 6. Observabilidade (MANDATÓRIO)

**ESTA É A SEÇÃO MAIS CRÍTICA. Nenhum PR será aprovado sem aderir estritamente a estas regras.**

### 6.1. Structured Logging (slog) - OBRIGATÓRIO

**✅ Padrão Exigido**:

```go
// Em Handlers HTTP - recupere o logger do contexto
logger := logging.ContextLogger(r.Context(), nil)
logger.Info("processando operação da instância",
    slog.String("instance_id", instanceID),
    slog.String("operation", "connect"))

// Adicione atributos ao contexto para propagação
ctx = logging.WithAttrs(ctx,
    slog.String("instance_id", instanceID),
    slog.String("operation", "restart"))

// Log de erro com campos estruturados
logger.Error("operação falhou",
    slog.String("error", err.Error()),
    slog.Int("retry_count", retries))
```

**❌ Padrões Proibidos**:

```go
fmt.Println("processando request")        // REJEITADO - use slog
log.Printf("instância %s falhou", id)     // REJEITADO - use slog
logger.Info("processando ⚠️ crítico")    // REJEITADO - sem emojis
logger.Info("token: " + authToken)       // REJEITADO - vazamento de PII
```

**Log Levels**:

| Level | Uso | Exemplo |
|-------|-----|---------|
| **DEBUG** | Diagnóstico detalhado, payloads (dev only) | "raw event payload", "query params", "internal state" |
| **INFO** | Operational events normais | "worker started", "message sent", "instance connected" |
| **WARN** | Degradação ou situação anormal recuperável | "redis timeout (using fallback)", "retry attempt 3/6" |
| **ERROR** | Falhas que exigem atenção | "database connection lost", "webhook delivery failed" |

**Checklist de Revisão de Código - Logging**:
- [ ] **TODOS** os logs usam `slog` com campos estruturados (sem `fmt.Println`, `log.Print`)
- [ ] O contexto é propagado através das chamadas de função (`context.Context` como parâmetro)
- [ ] `ContextLogger` é usado para recuperar o logger do contexto
- [ ] `WithAttrs` adiciona campos de domínio (`instance_id`, `request_id`)
- [ ] Logs de erro incluem `slog.String("error", err.Error())`
- [ ] Sem caracteres de emoji nas mensagens de log
- [ ] Sem dados sensíveis (credenciais, tokens, PII) nos logs
- [ ] Log level apropriado (DEBUG para diagnóstico, INFO para operacional, WARN para degradação, ERROR para falhas)

### 6.2. Propagação de Contexto - OBRIGATÓRIO

Já documentado em detalhes na **Seção 4.1**. Checklist resumido:

**Checklist de Revisão de Código - Contexto**:
- [ ] **TODAS** as funções aceitam `context.Context` como primeiro parâmetro
- [ ] Middleware injeta o logger via `WithLogger`
- [ ] Handlers usam `ContextLogger` para recuperar o logger
- [ ] Atributos de domínio são adicionados via `WithAttrs`
- [ ] O contexto é passado para **TODAS** as chamadas downstream (serviços, repositórios)
- [ ] Background goroutines recebem contexto filho via `WithLogger`

### 6.3. Métricas Prometheus - OBRIGATÓRIO

**60+ Métricas Definidas** (ver `internal/observability/metrics.go`):

#### HTTP & Health Metrics

```go
// HTTPRequests - Contador de requests HTTP
metrics.HTTPRequests.WithLabelValues(method, path, strconv.Itoa(status)).Inc()

// HTTPDuration - Histograma de latência HTTP
metrics.HTTPDuration.WithLabelValues(method, path, strconv.Itoa(status)).Observe(duration.Seconds())

// HealthChecks - Status de health checks
metrics.HealthChecks.WithLabelValues(component, status).Inc()
// component: "database", "redis", "s3"
// status: "healthy", "unhealthy"
```

#### Lock & Circuit Breaker Metrics

```go
// LockAcquisitions - Aquisições de lock Redis
metrics.LockAcquisitions.WithLabelValues(status).Inc()
// status: "success", "failure"

// LockReacquisitionAttempts - Tentativas de renovação de lock
metrics.LockReacquisitionAttempts.WithLabelValues(instanceID, result).Inc()
// result: "success", "failure", "fallback"

// CircuitBreakerState - Estado do circuit breaker global
metrics.CircuitBreakerState.Set(float64(state))
// 0=CLOSED, 1=OPEN, 2=HALF_OPEN

// CircuitBreakerStatePerInstance - Estado por instância
metrics.CircuitBreakerStatePerInstance.WithLabelValues(instanceID).Set(float64(state))

// SplitBrainDetected - Detecção de split-brain
metrics.SplitBrainDetected.Inc()

// SplitBrainInvalidLocks - Locks inválidos detectados
metrics.SplitBrainInvalidLocks.WithLabelValues(instanceID).Inc()
```

#### Event System Metrics

```go
// EventsCaptured - Eventos capturados do WhatsApp
metrics.EventsCaptured.WithLabelValues(instanceID, eventType, sourceLib).Inc()
// eventType: "message.received", "message.delivery", etc.
// sourceLib: "whatsmeow"

// EventsBuffered - Gauge de eventos no buffer
metrics.EventsBuffered.Set(float64(bufferSize))

// EventsInserted - Eventos inseridos no outbox
metrics.EventsInserted.WithLabelValues(instanceID, eventType, status).Inc()
// status: "success", "failure"

// EventsProcessed - Eventos processados por workers
metrics.EventsProcessed.WithLabelValues(instanceID, eventType, status).Inc()
// status: "success", "retrying", "failed"

// EventProcessingDuration - Duração do processamento
metrics.EventProcessingDuration.WithLabelValues(instanceID, eventType).Observe(duration.Seconds())

// EventRetries - Tentativas de retry
metrics.EventRetries.WithLabelValues(instanceID, eventType, attemptStr).Inc()

// EventsFailed - Eventos falhados permanentemente
metrics.EventsFailed.WithLabelValues(instanceID, eventType, reason).Inc()

// EventsDelivered - Eventos entregues com sucesso
metrics.EventsDelivered.WithLabelValues(instanceID, eventType, transport).Inc()

// EventDeliveryDuration - Latência de entrega completa
metrics.EventDeliveryDuration.WithLabelValues(instanceID, eventType, transport).Observe(duration.Seconds())

// EventSequenceGaps - Gaps de sequência detectados
metrics.EventSequenceGaps.WithLabelValues(instanceID).Set(float64(gapCount))

// EventOutboxBacklog - Backlog no outbox
metrics.EventOutboxBacklog.WithLabelValues(instanceID).Set(float64(pending Count))
```

#### DLQ Metrics

```go
// DLQEventsTotal - Eventos movidos para DLQ
metrics.DLQEventsTotal.WithLabelValues(instanceID, eventType, failureReason).Inc()

// DLQReprocessAttempts - Tentativas de reprocessamento
metrics.DLQReprocessAttempts.WithLabelValues(instanceID, eventID).Inc()

// DLQReprocessSuccess - Reprocessamentos bem-sucedidos
metrics.DLQReprocessSuccess.WithLabelValues(instanceID, eventType).Inc()

// DLQBacklog - Total de eventos na DLQ
metrics.DLQBacklog.Set(float64(dlqCount))
```

#### Media Processing Metrics

```go
// MediaDownloadsTotal - Downloads do WhatsApp
metrics.MediaDownloadsTotal.WithLabelValues(instanceID, mediaType, status).Inc()
// mediaType: "image", "video", "audio", "document"
// status: "success", "failure"

// MediaDownloadDuration - Duração de downloads
metrics.MediaDownloadDuration.WithLabelValues(instanceID, mediaType).Observe(duration.Seconds())

// MediaDownloadSize - Tamanho de arquivos baixados
metrics.MediaDownloadSize.WithLabelValues(instanceID, mediaType).Observe(float64(bytes))

// MediaUploadsTotal - Uploads para S3
metrics.MediaUploadsTotal.WithLabelValues(instanceID, mediaType, status).Inc()

// MediaUploadDuration - Duração de uploads S3
metrics.MediaUploadDuration.WithLabelValues(instanceID, mediaType).Observe(duration.Seconds())

// MediaUploadSizeBytes - Bytes enviados para S3
metrics.MediaUploadSizeBytes.WithLabelValues(mediaType).Add(float64(bytes))

// MediaFailures - Falhas no processamento de mídia
metrics.MediaFailures.WithLabelValues(instanceID, mediaType, stage).Inc()
// stage: "download", "upload", "validation"

// MediaBacklog - Fila de mídia pendente
metrics.MediaBacklog.Set(float64(pendingCount))

// MediaFallbackAttempts - Tentativas de fallback
metrics.MediaFallbackAttempts.WithLabelValues(instanceID, mediaType, fallbackType).Inc()
// fallbackType: "s3", "local"

// MediaFallbackSuccess/Failure
metrics.MediaFallbackSuccess.WithLabelValues(instanceID, mediaType, storageType).Inc()
metrics.MediaFallbackFailure.WithLabelValues(instanceID, mediaType, errorType).Inc()

// MediaCleanup* - Métricas de limpeza
metrics.MediaCleanupRuns.WithLabelValues(result).Inc()
// result: "success", "partial", "error", "empty"

metrics.MediaCleanupDeletedBytes.WithLabelValues(storageType).Add(float64(bytes))
// storageType: "s3", "local"

metrics.MediaCleanupDuration.Observe(duration.Seconds())
```

#### Transport Metrics

```go
// TransportDeliveries - Entregas via transport
metrics.TransportDeliveries.WithLabelValues(instanceID, transportType, status).Inc()
// transportType: "http"
// status: "success", "failure", "timeout"

// TransportDuration - Latência de entrega
metrics.TransportDuration.WithLabelValues(instanceID, transportType).Observe(duration.Seconds())

// TransportErrors - Erros de transporte
metrics.TransportErrors.WithLabelValues(instanceID, transportType, errorType).Inc()
// errorType: "network", "timeout", "http_4xx", "http_5xx"

// TransportRetries - Tentativas de retry
metrics.TransportRetries.WithLabelValues(instanceID, transportType, attemptStr).Inc()
```

#### Worker Metrics

```go
// WorkersActive - Workers ativos
metrics.WorkersActive.WithLabelValues(workerType).Set(float64(count))
// workerType: "dispatch", "media", "message_queue"

// WorkerTaskDuration - Duração de tarefas
metrics.WorkerTaskDuration.WithLabelValues(workerType, taskType).Observe(duration.Seconds())

// WorkerErrors - Erros de workers
metrics.WorkerErrors.WithLabelValues(workerType, errorType).Inc()
```

**Pattern de Uso**:

```go
func (w *InstanceWorker) processBatch(ctx context.Context) error {
    start := time.Now()

    events, err := w.poll(ctx)
    if err != nil {
        metrics.WorkerErrors.WithLabelValues("dispatch", "poll_error").Inc()
        return err
    }

    for _, event := range events {
        eventStart := time.Now()

        if err := w.deliver(ctx, event); err != nil {
            metrics.EventsProcessed.WithLabelValues(
                event.InstanceID.String(),
                event.EventType,
                "failure",
            ).Inc()
            continue
        }

        // Sucesso
        metrics.EventsProcessed.WithLabelValues(
            event.InstanceID.String(),
            event.EventType,
            "success",
        ).Inc()

        metrics.EventProcessingDuration.WithLabelValues(
            event.InstanceID.String(),
            event.EventType,
        ).Observe(time.Since(eventStart).Seconds())
    }

    metrics.WorkerTaskDuration.WithLabelValues("dispatch", "batch").Observe(time.Since(start).Seconds())

    return nil
}
```

**Checklist de Revisão de Código - Métricas**:
- [ ] Métricas são atualizadas no ponto de ocorrência do evento (não em funções separadas)
- [ ] Labels incluem contexto relevante (componente, status, `instance_id`)
- [ ] Evitar labels de alta cardinalidade (sem `user_id`, `request_id`)
- [ ] Usar Histogram/Summary para durações (não counters)
- [ ] Nomes de Counter terminam com o sufixo `_total`
- [ ] Gauges para valores instantâneos (backlog, workers ativos)
- [ ] Histograms para distribuições (latências, tamanhos)

### 6.4. Rastreamento de Erros com Sentry - OBRIGATÓRIO

**✅ Padrão Exigido**:

```go
// Capture erros críticos com contexto
if err := criticalOperation(); err != nil {
    logger := logging.ContextLogger(ctx, nil)
    logger.Error("operação crítica falhou", slog.String("error", err.Error()))

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

**❌ NÃO Capturar**:
- Erros esperados (validation failures, 400/404 responses)
- Erros já logados como warnings
- Erros de alta frequência (use sampling)
- Erros de validação de input do usuário

**Severity Levels**:

| Severity | Critério | Exemplo |
|----------|----------|---------|
| **critical** | Sistema indisponível, perda de dados | Database down, Redis split-brain |
| **error** | Falha de operação importante | Webhook delivery failed, media upload failed |
| **warning** | Degradação ou fallback ativado | Circuit breaker opened, using fallback lock |
| **info** | Eventos incomuns mas não problemáticos | DLQ reprocess success, manual intervention |

**Checklist de Revisão de Código - Sentry**:
- [ ] Apenas erros críticos são capturados (não erros de validação ou 4xx)
- [ ] `WithScope` é usado para adicionar contexto (nunca tags globais)
- [ ] Tags incluem: `component`, `instance_id`, `severity`
- [ ] Context includes operation details (type, attempts, duration)
- [ ] Sem dados sensíveis nas capturas do Sentry (credentials, tokens, PII)
- [ ] Sampling configurado para erros de alta frequência

### 6.5. Health Checks - OBRIGATÓRIO

**Endpoints**:

- **`/health`**: Liveness probe - verifica se o serviço está rodando
- **`/ready`**: Readiness probe - verifica dependências

**Implementação Completa**:

```go
type HealthHandler struct {
    pool        *pgxpool.Pool
    lockManager locks.Manager
    setMetrics  func(component, status string)
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
    // Liveness: apenas verifica se o serviço responde
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := logging.ContextLogger(ctx, nil)

    checks := []struct{
        name string
        fn   func(context.Context) error
    }{
        {"database", h.checkDatabase},
        {"redis", h.checkRedis},
    }

    allHealthy := true

    for _, check := range checks {
        start := time.Now()
        err := check.fn(ctx)
        duration := time.Since(start)

        status := "healthy"
        if err != nil {
            status = "unhealthy"
            allHealthy = false

            logger.Error("health check failed",
                slog.String("component", check.name),
                slog.String("error", err.Error()),
                slog.Duration("duration", duration))

            // Capture no Sentry
            sentry.WithScope(func(scope *sentry.Scope) {
                scope.SetTag("component", "healthcheck")
                scope.SetTag("check", check.name)
                scope.SetContext("check", map[string]interface{}{
                    "duration_ms": duration.Milliseconds(),
                })
                sentry.CaptureException(err)
            })
        } else {
            logger.Debug("health check passed",
                slog.String("component", check.name),
                slog.Duration("duration", duration))
        }

        // Update metrics
        if h.setMetrics != nil {
            h.setMetrics(check.name, status)
        }
    }

    if allHealthy {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Ready"))
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Not Ready"))
    }
}

func (h *HealthHandler) checkDatabase(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    return h.pool.Ping(ctx)
}

func (h *HealthHandler) checkRedis(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Try to acquire a test lock
    lock, err := h.lockManager.AcquireLock(ctx, "healthcheck", 10*time.Second)
    if err != nil {
        return err
    }
    defer lock.Release(ctx)

    return nil
}
```

**Checklist de Revisão de Código - Health Checks**:
- [ ] Log duration and status for all checks
- [ ] Update `HealthChecks` metric with component and status labels
- [ ] Capture failures in Sentry with context
- [ ] Return 200 for healthy, 503 for unhealthy
- [ ] `/health`: Basic liveness (service up)
- [ ] `/ready`: Dependencies available (database, redis, s3)
- [ ] Timeouts configurados (5s máximo por check)

### 6.6. Complete Code Review Checklist

**Antes de aprovar PR, verificar TODOS os itens**:

#### Logging ✅
- [ ] ALL logs use `slog` with structured fields
- [ ] No `fmt.Println`, `log.Print`, or plain text logging
- [ ] Context propagated through all function calls
- [ ] `instance_id` added to context via `WithAttrs`
- [ ] Errors logged with `slog.String("error", err.Error())`
- [ ] No emoji characters in log messages
- [ ] No sensitive data in logs (credentials, tokens, PII)
- [ ] Log levels appropriate: DEBUG (diagnostic), INFO (operational), WARN (degraded), ERROR (failures)

#### Context Propagation ✅
- [ ] ALL functions accept `context.Context` as first parameter
- [ ] Middleware injects logger via `WithLogger`
- [ ] Handlers use `ContextLogger` to retrieve logger
- [ ] `WithAttrs` used for domain fields (instance_id, operation)
- [ ] Context passed to downstream calls (services, repositories)
- [ ] Background goroutines receive context copy via `WithLogger`

#### Metrics ✅
- [ ] Metrics updated where events occur (not delayed)
- [ ] Labels include relevant context (component, status, instance_id)
- [ ] No high-cardinality labels (no user_id, request_id)
- [ ] Histogram/Summary for durations (not counters)
- [ ] Counter names end with `_total` suffix
- [ ] Gauges for instantaneous values
- [ ] Histograms for distributions

#### Sentry ✅
- [ ] Only critical errors captured (not validation/expected errors)
- [ ] `WithScope` with proper tags/context
- [ ] Tags include: component, instance_id, severity
- [ ] Context includes operation details (type, attempts, duration)
- [ ] No sensitive data captured (credentials, tokens, PII)
- [ ] Sampling configured for high-frequency errors

#### Health Checks ✅
- [ ] Log duration and status for all checks
- [ ] Update metrics with component and status
- [ ] Capture failures in Sentry with context
- [ ] Correct status codes (200/503)
- [ ] Timeouts configured (5s max)

#### Dispatch System (se aplicável) ✅
- [ ] Coordinator logs worker lifecycle events
- [ ] Workers include instance_id in all logs
- [ ] Processor tracks delivery metrics immediately
- [ ] Failed deliveries captured in Sentry with context
- [ ] Graceful shutdown with timeout tracking
- [ ] No blocking operations in worker poll loop
- [ ] Context cancellation respected in all goroutines

#### General ✅
- [ ] Tests pass: `go test ./...`
- [ ] Pre-commit checks pass: `pre-commit run --all-files`
- [ ] OpenAPI docs updated (if API changes)
- [ ] Migrations created (if schema changes)
- [ ] No `panic` for expected errors
- [ ] Error wrapping with context (`fmt.Errorf("context: %w", err)`)

---

## 7. Database e Migrations

### 7.1. Migration System (Goose)

#### Migration File Conventions

**Naming Pattern**: `NNNNNN_description.sql`

```bash
# Examples:
migrations/000001_init.sql              # Initial schema
migrations/000002_add_user_roles.sql    # Add user roles table
migrations/000003_event_system_v2.sql   # Event system v2
```

**File Structure (goose format)**:
```sql
-- +goose Up
-- +goose StatementBegin

-- Your DDL statements here
CREATE TABLE IF NOT EXISTS your_table (...);
CREATE INDEX IF NOT EXISTS idx_name ON your_table(...);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback statements (reverse order)
DROP INDEX IF EXISTS idx_name;
DROP TABLE IF EXISTS your_table;

-- +goose StatementEnd
```

#### Migration Rules

**✅ ALWAYS**:
- Include `IF NOT EXISTS` for CREATE statements
- Include `IF EXISTS` for DROP statements
- Write both `Up` and `Down` migrations
- Test rollback before committing
- Add comments explaining WHY, not WHAT
- Use `StatementBegin/StatementEnd` for multi-statement blocks
- Create indexes AFTER data insertion (not before)
- Use `CONSTRAINT` names for easier debugging

**❌ NEVER**:
- Change existing migrations (create new one instead)
- Drop columns with data (add tombstone flag instead)
- Use `CASCADE` without careful consideration
- Mix DDL and DML in same migration (separate concerns)
- Create indexes on small tables (<1000 rows) unless needed
- Use `TEXT` for enum-like fields (use `VARCHAR(N)` with CHECK constraint)

#### Creating New Migration

```bash
# Generate migration file
goose -dir migrations create add_message_queue sql

# Apply migration
goose -dir migrations postgres "user=postgres dbname=mydb sslmode=disable" up

# Rollback last migration
goose -dir migrations postgres "user=postgres dbname=mydb sslmode=disable" down

# Check status
goose -dir migrations postgres "user=postgres dbname=mydb sslmode=disable" status
```

### 7.2. Schema Design Principles

#### Core Tables

**instances** (main entity):
```sql
CREATE TABLE IF NOT EXISTS instances (
    id UUID PRIMARY KEY,
    name TEXT,
    session_name TEXT,
    client_token TEXT NOT NULL,
    instance_token TEXT NOT NULL,

    -- WhatsApp state
    store_jid TEXT,                         -- WhatsApp JID
    is_device BOOLEAN NOT NULL DEFAULT FALSE,
    business_device BOOLEAN NOT NULL DEFAULT FALSE,

    -- Subscription and features
    subscription_active BOOLEAN NOT NULL DEFAULT FALSE,
    call_reject_auto BOOLEAN NOT NULL DEFAULT FALSE,
    call_reject_message TEXT,
    auto_read_message BOOLEAN NOT NULL DEFAULT FALSE,

    -- Lifecycle
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ALWAYS create unique indexes for authentication tokens
CREATE UNIQUE INDEX IF NOT EXISTS idx_instances_client_token ON instances(client_token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_instances_instance_token ON instances(instance_token);
```

**event_outbox** (event queue with ordering guarantees):
```sql
CREATE TABLE IF NOT EXISTS event_outbox (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    source_lib VARCHAR(20) NOT NULL DEFAULT 'whatsmeow',

    -- Event data
    payload JSONB NOT NULL,
    metadata JSONB,

    -- CRITICAL: Ordering control
    sequence_number BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Processing status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 6,
    next_attempt_at TIMESTAMP,

    -- Media tracking
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    media_processed BOOLEAN NOT NULL DEFAULT FALSE,

    -- Delivery tracking
    delivered_at TIMESTAMP,
    transport_type VARCHAR(20) NOT NULL DEFAULT 'webhook',
    last_error TEXT,

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- CRITICAL: Enforce per-instance sequence uniqueness
    CONSTRAINT event_outbox_sequence_unique UNIQUE (instance_id, sequence_number),
    CONSTRAINT event_outbox_status_check CHECK (status IN ('pending', 'processing', 'retrying', 'delivered', 'failed'))
);
```

**event_dlq** (Dead Letter Queue for failed events):
```sql
CREATE TABLE IF NOT EXISTS event_dlq (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL,
    event_id UUID NOT NULL,

    -- Preserve original data for debugging
    original_payload JSONB NOT NULL,
    original_sequence_number BIGINT NOT NULL,

    -- Failure context
    failure_reason TEXT NOT NULL,
    last_error TEXT NOT NULL,
    total_attempts INT NOT NULL,
    attempt_history JSONB NOT NULL DEFAULT '[]'::JSONB,

    -- Reprocessing control
    reprocess_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reprocessed_at TIMESTAMP,

    moved_to_dlq_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT event_dlq_event_unique UNIQUE (event_id)
);
```

### 7.3. Index Strategy

#### Performance-Critical Indexes

**event_outbox**: Optimized for worker polling
```sql
-- PRIMARY: Worker poll query (most important)
CREATE INDEX IF NOT EXISTS idx_outbox_instance_pending
    ON event_outbox(instance_id, status, next_attempt_at, sequence_number)
    WHERE status IN ('pending', 'retrying');

-- SECONDARY: Media processing
CREATE INDEX IF NOT EXISTS idx_outbox_media_pending
    ON event_outbox(instance_id, id)
    WHERE has_media = TRUE AND media_processed = FALSE;

-- MONITORING: Recent events per instance
CREATE INDEX IF NOT EXISTS idx_outbox_instance_recent
    ON event_outbox(instance_id, created_at DESC);

-- ALERTING: Failed events requiring attention
CREATE INDEX IF NOT EXISTS idx_outbox_failed
    ON event_outbox(instance_id, created_at DESC)
    WHERE status = 'failed';
```

**event_dlq**: Optimized for monitoring and reprocessing
```sql
-- Reprocessing queue
CREATE INDEX IF NOT EXISTS idx_dlq_reprocess_pending
    ON event_dlq(reprocess_status, moved_to_dlq_at)
    WHERE reprocess_status = 'pending';

-- Failure analysis
CREATE INDEX IF NOT EXISTS idx_dlq_failure_reason
    ON event_dlq(failure_reason, moved_to_dlq_at DESC);

-- Instance monitoring
CREATE INDEX IF NOT EXISTS idx_dlq_instance
    ON event_dlq(instance_id, moved_to_dlq_at DESC);
```

**media_metadata**: Optimized for parallel processing
```sql
-- Worker poll query
CREATE INDEX IF NOT EXISTS idx_media_pending
    ON media_metadata(download_status, next_retry_at, created_at)
    WHERE download_status IN ('pending', 'failed')
    AND download_attempts < max_retries;

-- Failed media requiring manual intervention
CREATE INDEX IF NOT EXISTS idx_media_failed
    ON media_metadata(instance_id, created_at DESC)
    WHERE download_status = 'failed' AND download_attempts >= max_retries;
```

#### Partial Indexes (WHERE clause)

**Use partial indexes to reduce size and improve query performance**:

```sql
-- ✅ GOOD: Only index pending/retrying events (80%+ reduction)
CREATE INDEX idx_outbox_pending ON event_outbox(instance_id, status)
    WHERE status IN ('pending', 'retrying');

-- ❌ BAD: Indexes all rows including delivered (99% of data)
CREATE INDEX idx_outbox_all ON event_outbox(instance_id, status);
```

### 7.4. Sequence Generation (Critical for Ordering)

#### Atomic Sequence Function

**Purpose**: Guarantee ordered event processing per instance

```sql
CREATE OR REPLACE FUNCTION get_next_event_sequence(p_instance_id UUID)
RETURNS BIGINT AS $$
DECLARE
    v_sequence BIGINT;
BEGIN
    -- Atomic INSERT ... ON CONFLICT for concurrency safety
    INSERT INTO instance_event_sequence (
        instance_id,
        current_sequence,
        last_event_at,
        total_events,
        updated_at,
        created_at
    )
    VALUES (
        p_instance_id,
        1,
        NOW(),
        1,
        NOW(),
        NOW()
    )
    ON CONFLICT (instance_id)
    DO UPDATE SET
        current_sequence = instance_event_sequence.current_sequence + 1,
        last_event_at = NOW(),
        total_events = instance_event_sequence.total_events + 1,
        updated_at = NOW()
    RETURNING current_sequence INTO v_sequence;

    RETURN v_sequence;
END;
$$ LANGUAGE plpgsql;
```

#### Usage in Application Code

```go
// internal/events/capture/transactional_writer.go
func (w *TransactionalWriter) Write(ctx context.Context, ev *Event) error {
    logger := logging.ContextLogger(ctx, w.log)

    // Get next sequence number atomically
    var sequenceNumber int64
    err := w.db.QueryRowContext(ctx,
        "SELECT get_next_event_sequence($1)",
        ev.InstanceID,
    ).Scan(&sequenceNumber)
    if err != nil {
        logger.Error("failed to get sequence number",
            slog.String("instance_id", ev.InstanceID.String()),
            slog.Any("error", err))
        return fmt.Errorf("get_next_event_sequence: %w", err)
    }

    logger.Debug("sequence generated",
        slog.Int64("sequence_number", sequenceNumber),
        slog.String("event_type", ev.EventType))

    // Insert event with sequence_number
    _, err = w.db.ExecContext(ctx, `
        INSERT INTO event_outbox (
            instance_id, event_id, event_type, source_lib,
            payload, metadata, sequence_number,
            status, attempts, next_attempt_at,
            has_media, transport_type, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()
        )`,
        ev.InstanceID, ev.EventID, ev.EventType, ev.SourceLib,
        ev.Payload, ev.Metadata, sequenceNumber,
        "pending", 0, time.Now(),
        ev.HasMedia, "webhook",
    )

    if err != nil {
        return fmt.Errorf("insert event_outbox: %w", err)
    }

    // Update metrics IMMEDIATELY after success
    metrics.EventsPersisted.WithLabelValues(
        ev.InstanceID.String(),
        ev.EventType,
        "success",
    ).Inc()

    return nil
}
```

#### Monitoring Sequence Gaps

```sql
-- Function to detect missing sequences (data loss detection)
CREATE OR REPLACE FUNCTION check_sequence_gaps(p_instance_id UUID)
RETURNS TABLE(missing_sequence BIGINT) AS $$
BEGIN
    RETURN QUERY
    WITH expected_sequences AS (
        SELECT generate_series(
            1::BIGINT,
            (SELECT current_sequence FROM instance_event_sequence WHERE instance_id = p_instance_id)
        ) AS seq
    ),
    existing_sequences AS (
        SELECT sequence_number
        FROM event_outbox
        WHERE instance_id = p_instance_id
    )
    SELECT e.seq
    FROM expected_sequences e
    LEFT JOIN existing_sequences x ON e.seq = x.sequence_number
    WHERE x.sequence_number IS NULL
    ORDER BY e.seq;
END;
$$ LANGUAGE plpgsql;
```

**Alert on gaps**:
```bash
# Monitoring script (run every 5 minutes)
psql -c "SELECT instance_id, check_sequence_gaps(instance_id)
         FROM instances
         WHERE canceled_at IS NULL" \
| grep -v "0 rows" && alert_ops_team
```

### 7.5. Transaction Boundaries

#### Repository Pattern

**✅ CORRECT: Transaction in service layer**
```go
// internal/instances/service.go
func (s *Service) CreateInstanceAndSetup(ctx context.Context, req CreateRequest) (*Instance, error) {
    logger := logging.ContextLogger(ctx, s.log)

    // Start transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // Safe to call even after commit

    // Create instance
    inst, err := s.repo.CreateWithTx(ctx, tx, req)
    if err != nil {
        logger.Error("failed to create instance", slog.Any("error", err))
        return nil, fmt.Errorf("create instance: %w", err)
    }

    // Initialize webhook config (same transaction)
    err = s.webhookRepo.CreateDefaultWithTx(ctx, tx, inst.ID)
    if err != nil {
        logger.Error("failed to create webhook config",
            slog.String("instance_id", inst.ID.String()),
            slog.Any("error", err))
        return nil, fmt.Errorf("create webhook config: %w", err)
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        logger.Error("failed to commit transaction", slog.Any("error", err))
        return nil, fmt.Errorf("commit transaction: %w", err)
    }

    logger.Info("instance created successfully",
        slog.String("instance_id", inst.ID.String()),
        slog.String("name", inst.Name))

    return inst, nil
}
```

**❌ INCORRECT: Transaction spans repository calls without control**
```go
// ❌ DON'T DO THIS
func (s *Service) CreateInstance(ctx context.Context, req CreateRequest) error {
    // No transaction control - each repo call is separate transaction
    inst, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }

    // If this fails, instance already created (data inconsistency)
    err = s.webhookRepo.CreateDefault(ctx, inst.ID)
    if err != nil {
        return err // Instance orphaned without webhook config
    }

    return nil
}
```

#### Transaction Isolation Levels

**Default**: `READ COMMITTED` (PostgreSQL default, sufficient for most cases)

**Use `SERIALIZABLE` for critical operations**:
```go
// Example: Lock acquisition with serializable isolation
func (s *LockService) AcquireLock(ctx context.Context, instanceID uuid.UUID) error {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    if err != nil {
        return fmt.Errorf("begin serializable transaction: %w", err)
    }
    defer tx.Rollback()

    // Check if lock exists
    var exists bool
    err = tx.QueryRowContext(ctx,
        "SELECT EXISTS(SELECT 1 FROM locks WHERE instance_id = $1)",
        instanceID,
    ).Scan(&exists)
    if err != nil {
        return fmt.Errorf("check lock existence: %w", err)
    }

    if exists {
        return ErrLockAlreadyAcquired
    }

    // Acquire lock
    _, err = tx.ExecContext(ctx,
        "INSERT INTO locks (instance_id, acquired_at, expires_at) VALUES ($1, NOW(), NOW() + INTERVAL '5 minutes')",
        instanceID,
    )
    if err != nil {
        return fmt.Errorf("acquire lock: %w", err)
    }

    return tx.Commit()
}
```

### 7.6. Connection Pool Configuration

#### Environment Variables (.env)

```bash
# PostgreSQL connection string
DATABASE_URL=postgres://user:password@localhost:5432/whatsapp_api?sslmode=disable

# Connection pool settings (pgx defaults shown)
DB_MAX_OPEN_CONNS=25           # Maximum open connections
DB_MAX_IDLE_CONNS=25           # Maximum idle connections
DB_CONN_MAX_LIFETIME=5m        # Connection max lifetime
DB_CONN_MAX_IDLE_TIME=5m       # Idle connection max lifetime

# Recommended for production:
# DB_MAX_OPEN_CONNS=50          # Higher for more concurrent requests
# DB_MAX_IDLE_CONNS=10          # Lower to free resources
# DB_CONN_MAX_LIFETIME=30m      # Rotate connections periodically
# DB_CONN_MAX_IDLE_TIME=10m     # Close idle connections faster
```

#### Connection Pool Initialization

```go
// internal/config/database.go
func InitializeDatabase(cfg *Config) (*sql.DB, error) {
    db, err := sql.Open("pgx", cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(cfg.DBMaxOpenConns)         // Default: 25
    db.SetMaxIdleConns(cfg.DBMaxIdleConns)         // Default: 25
    db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)   // Default: 5m
    db.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)   // Default: 5m

    // Verify connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return db, nil
}
```

#### Monitoring Connection Pool

```go
// Expose metrics for observability
func (s *Server) recordDatabaseMetrics() {
    stats := s.db.Stats()

    metrics.DBConnectionsOpen.Set(float64(stats.OpenConnections))
    metrics.DBConnectionsInUse.Set(float64(stats.InUse))
    metrics.DBConnectionsIdle.Set(float64(stats.Idle))
    metrics.DBConnectionWaitCount.Set(float64(stats.WaitCount))
    metrics.DBConnectionWaitDuration.Set(stats.WaitDuration.Seconds())
}
```

### 7.7. PostgreSQL-Specific Patterns

#### JSONB for Flexible Schema

**Use JSONB for semi-structured data**:
```sql
-- event_outbox.payload stores Z-API compatible event data
CREATE TABLE event_outbox (
    payload JSONB NOT NULL,
    metadata JSONB,
    transport_config JSONB
);

-- Query JSONB fields with indexes
CREATE INDEX idx_payload_event_type ON event_outbox ((payload->>'event'));
CREATE INDEX idx_metadata_priority ON event_outbox ((metadata->>'priority'));
```

**Application code**:
```go
// Unmarshal JSONB into Go struct
var event struct {
    Event   string `json:"event"`
    Phone   string `json:"phone"`
    Message struct {
        Body string `json:"body"`
    } `json:"message"`
}

err := json.Unmarshal(row.Payload, &event)
```

#### UUID Generation

**Use `gen_random_uuid()` (PostgreSQL 13+)**:
```sql
-- Faster than uuid-ossp extension
CREATE TABLE event_outbox (
    event_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid()
);
```

**Application code**:
```go
import "github.com/google/uuid"

eventID := uuid.New() // Generate UUID in application
```

#### Timestamps with Timezone

**ALWAYS use `TIMESTAMPTZ`** (not `TIMESTAMP`):
```sql
-- ✅ GOOD: Stores timezone information
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- ❌ BAD: No timezone (ambiguous)
created_at TIMESTAMP NOT NULL DEFAULT NOW()
```

#### Array Types

**Use for multi-valued fields**:
```sql
-- Example: Store multiple webhook URLs
CREATE TABLE webhook_configs (
    instance_id UUID PRIMARY KEY,
    delivery_urls TEXT[] NOT NULL DEFAULT '{}'
);

-- Query
SELECT * FROM webhook_configs WHERE 'https://example.com' = ANY(delivery_urls);
```

### 7.8. Database Checklist for Code Review ✅

#### Migrations ✅
- [ ] Migration follows naming convention `NNNNNN_description.sql`
- [ ] Both `Up` and `Down` migrations implemented
- [ ] Rollback tested successfully
- [ ] Uses `IF NOT EXISTS` / `IF EXISTS` appropriately
- [ ] Constraints have meaningful names
- [ ] Comments explain WHY (not WHAT)
- [ ] No mixing of DDL and DML
- [ ] Indexes created with performance consideration

#### Schema ✅
- [ ] Primary keys defined (UUID or BIGSERIAL)
- [ ] Foreign keys with appropriate `ON DELETE` action
- [ ] NOT NULL constraints for required fields
- [ ] DEFAULT values for optional fields
- [ ] CHECK constraints for enums and validation
- [ ] UNIQUE constraints where needed
- [ ] Timestamps use TIMESTAMPTZ (not TIMESTAMP)

#### Indexes ✅
- [ ] Partial indexes used for filtered queries
- [ ] Composite indexes match query patterns
- [ ] Index column order optimized (high selectivity first)
- [ ] No redundant indexes
- [ ] Indexes on foreign keys
- [ ] JSONB fields indexed with expression indexes

#### Transactions ✅
- [ ] Transaction boundaries in service layer (not repository)
- [ ] `defer tx.Rollback()` after `BeginTx`
- [ ] Explicit `tx.Commit()` before returning
- [ ] Isolation level appropriate for operation
- [ ] No long-running transactions (< 1 second target)
- [ ] Context cancellation respected

#### Connection Pool ✅
- [ ] Max open connections configured
- [ ] Max idle connections configured
- [ ] Connection max lifetime configured
- [ ] Connection max idle time configured
- [ ] Pool metrics exposed for monitoring

#### Query Performance ✅
- [ ] `EXPLAIN ANALYZE` used for complex queries
- [ ] N+1 queries avoided (use JOINs or eager loading)
- [ ] Pagination implemented for large result sets
- [ ] Index usage verified with `EXPLAIN`
- [ ] Query timeout configured in context

---

## 8. Worker Patterns

### 8.1. Worker Architecture Overview

**Three-Layer Worker System**:

```
┌─────────────────────────────────────────────────────────────────────┐
│  COORDINATOR (Global)                                                │
│  - Manages worker lifecycle (start/stop/register/unregister)        │
│  - Spawns one InstanceWorker per connected instance                 │
│  - Handles graceful shutdown with WaitGroup                          │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  INSTANCE WORKER (Per Instance)                                     │
│  - Poll-process loop for single instance                            │
│  - Fetches pending events ordered by sequence_number                │
│  - Processes events via EventProcessor                              │
│  - Respects context cancellation                                    │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│  EVENT PROCESSOR (Per Event)                                        │
│  - Transform event (internal → Z-API format)                        │
│  - Deliver webhook with retry logic                                 │
│  - Handle success/failure/skipped events                            │
│  - Move to DLQ after max attempts                                   │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.2. Coordinator Pattern

**Purpose**: Manage multiple instance workers with lifecycle control

```go
// internal/events/dispatch/coordinator.go
type Coordinator struct {
    cfg               *config.Config
    pool              *pgxpool.Pool
    outboxRepo        persistence.OutboxRepository
    dlqRepo           persistence.DLQRepository
    transportRegistry *transport.Registry
    lookup            InstanceLookup
    metrics           *observability.Metrics

    mu       sync.RWMutex                      // Protects workers map
    workers  map[uuid.UUID]*InstanceWorker     // Per-instance workers
    stopChan chan struct{}                     // Broadcast shutdown
    wg       sync.WaitGroup                    // Wait for worker completion
    running  bool                              // State flag
}

func (c *Coordinator) Start(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.running {
        return fmt.Errorf("coordinator already running")
    }

    c.running = true

    logger := logging.ContextLogger(ctx, nil)
    logger.Info("dispatch coordinator started",
        slog.Int("workers", len(c.workers)))

    return nil
}

func (c *Coordinator) RegisterInstance(ctx context.Context, instanceID uuid.UUID) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if !c.running {
        return fmt.Errorf("coordinator not running")
    }

    // Idempotent: skip if already registered
    if _, exists := c.workers[instanceID]; exists {
        return nil
    }

    logger := logging.ContextLogger(ctx, nil)
    logger.Info("registering dispatch worker for instance",
        slog.String("instance_id", instanceID.String()))

    // Create worker (not started yet)
    worker := NewInstanceWorker(
        instanceID,
        c.cfg,
        c.outboxRepo,
        c.dlqRepo,
        c.transportRegistry,
        c.lookup,
        c.metrics,
    )

    c.workers[instanceID] = worker

    // Start worker in background goroutine
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()

        // Create cancellable context
        workerCtx, cancel := context.WithCancel(context.Background())
        defer cancel()

        // Listen for shutdown signal
        go func() {
            <-c.stopChan
            cancel() // Cancel worker context on shutdown
        }()

        // Start worker (blocks until context cancelled)
        if err := worker.Start(workerCtx); err != nil {
            logger.Error("worker failed",
                slog.String("instance_id", instanceID.String()),
                slog.Any("error", err))
        }
    }()

    c.metrics.WorkersActive.Inc()

    return nil
}

func (c *Coordinator) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    worker, exists := c.workers[instanceID]
    if !exists {
        return nil
    }

    logger := logging.ContextLogger(ctx, nil)
    logger.Info("unregistering dispatch worker",
        slog.String("instance_id", instanceID.String()))

    // Stop worker (graceful)
    if err := worker.Stop(ctx); err != nil {
        logger.Warn("failed to stop worker gracefully",
            slog.String("instance_id", instanceID.String()),
            slog.Any("error", err))
    }

    delete(c.workers, instanceID)
    c.metrics.WorkersActive.Dec()

    return nil
}

func (c *Coordinator) Stop(ctx context.Context) error {
    c.mu.Lock()
    if !c.running {
        c.mu.Unlock()
        return nil
    }
    c.running = false
    c.mu.Unlock()

    logger := logging.ContextLogger(ctx, nil)
    logger.Info("stopping dispatch coordinator",
        slog.Int("active_workers", len(c.workers)))

    // Broadcast shutdown to all workers
    close(c.stopChan)

    // Wait for all workers to finish (with timeout)
    done := make(chan struct{})
    go func() {
        c.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        logger.Info("all workers stopped gracefully")
        return nil
    case <-ctx.Done():
        logger.Warn("coordinator shutdown timeout, forcing exit")
        return ctx.Err()
    }
}
```

**Key Patterns**:
- **Map Protection**: `sync.RWMutex` protects `workers` map from concurrent access
- **Graceful Shutdown**: `stopChan` broadcasts shutdown, `WaitGroup` waits for completion
- **Context Cancellation**: Each worker gets cancellable context
- **Idempotency**: `RegisterInstance` is idempotent (safe to call multiple times)
- **Metrics**: Track active workers with Prometheus gauge

### 8.3. Instance Worker Pattern (Poll-Process Loop)

**Purpose**: Process events for single instance in ordered sequence

```go
// internal/events/dispatch/worker.go
type InstanceWorker struct {
    instanceID uuid.UUID
    processor  *EventProcessor
    outboxRepo persistence.OutboxRepository
    metrics    *observability.Metrics

    stopChan chan struct{}
    running  bool
}

func (w *InstanceWorker) Start(ctx context.Context) error {
    w.running = true
    defer func() { w.running = false }()

    logger := logging.WithAttrs(ctx,
        slog.String("instance_id", w.instanceID.String()),
        slog.String("component", "instance_worker"))

    logger.Info("instance worker started")

    ticker := time.NewTicker(500 * time.Millisecond) // Poll interval
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            logger.Info("instance worker stopping (context cancelled)")
            return ctx.Err()

        case <-w.stopChan:
            logger.Info("instance worker stopping (stop signal)")
            return nil

        case <-ticker.C:
            // Poll for pending events (non-blocking)
            if err := w.pollAndProcess(ctx); err != nil {
                // Log but continue (transient errors shouldn't kill worker)
                logger.Error("poll error",
                    slog.Any("error", err))
            }
        }
    }
}

func (w *InstanceWorker) pollAndProcess(ctx context.Context) error {
    logger := logging.ContextLogger(ctx, nil)

    // Fetch pending events ORDERED BY sequence_number (critical)
    events, err := w.outboxRepo.FetchPendingEvents(ctx, w.instanceID, 10)
    if err != nil {
        logger.Error("failed to fetch pending events",
            slog.Any("error", err))
        return fmt.Errorf("fetch pending events: %w", err)
    }

    if len(events) == 0 {
        return nil // No work to do
    }

    logger.Debug("fetched pending events",
        slog.Int("count", len(events)))

    // Process events sequentially (maintain order)
    for _, event := range events {
        // Check context cancellation before processing
        if ctx.Err() != nil {
            return ctx.Err()
        }

        // Enrich context with event details
        eventCtx := logging.WithAttrs(ctx,
            slog.String("event_id", event.EventID.String()),
            slog.String("event_type", event.EventType),
            slog.Int64("sequence_number", event.SequenceNumber))

        // Process event (delegate to EventProcessor)
        if err := w.processor.Process(eventCtx, event); err != nil {
            logger.Error("failed to process event",
                slog.String("event_id", event.EventID.String()),
                slog.Any("error", err))
            // Continue with next event (don't fail entire batch)
        }
    }

    return nil
}

func (w *InstanceWorker) Stop(ctx context.Context) error {
    if !w.running {
        return nil
    }

    logger := logging.WithAttrs(ctx,
        slog.String("instance_id", w.instanceID.String()))

    logger.Info("stopping instance worker")

    close(w.stopChan)

    // Wait for current processing to finish (with timeout)
    select {
    case <-time.After(5 * time.Second):
        logger.Warn("worker stop timeout")
        return fmt.Errorf("stop timeout")
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

**Key Patterns**:
- **Poll Interval**: 500ms ticker for balance between latency and database load
- **Ordered Processing**: ALWAYS fetch events `ORDER BY sequence_number`
- **Context Cancellation**: Check `ctx.Err()` before processing each event
- **Error Isolation**: Single event failure doesn't stop worker
- **Batch Size**: Fetch 10 events per poll (configurable based on load)
- **Non-Blocking**: Never block on I/O in main loop

### 8.4. Event Processor Pattern

**Purpose**: Transform and deliver single event

```go
// internal/events/dispatch/processor.go
type EventProcessor struct {
    instanceID        uuid.UUID
    cfg               *config.Config
    outboxRepo        persistence.OutboxRepository
    dlqRepo           persistence.DLQRepository
    transportRegistry *transport.Registry
    lookup            InstanceLookup
    metrics           *observability.Metrics
}

func (p *EventProcessor) Process(ctx context.Context, event *persistence.OutboxEvent) error {
    start := time.Now()
    logger := logging.ContextLogger(ctx, nil)

    logger.Debug("processing event",
        slog.String("event_id", event.EventID.String()),
        slog.String("event_type", event.EventType),
        slog.Int("attempt", event.Attempts+1))

    // Update metrics IMMEDIATELY
    p.metrics.EventsProcessed.WithLabelValues(
        p.instanceID.String(),
        event.EventType,
        "started",
    ).Inc()

    // Update status to 'processing'
    if err := p.outboxRepo.UpdateEventStatus(ctx,
        event.EventID,
        persistence.EventStatusProcessing,
        event.Attempts,
        event.NextAttemptAt,
        event.LastError,
    ); err != nil {
        logger.Error("failed to update status to processing",
            slog.Any("error", err))
        return fmt.Errorf("update status: %w", err)
    }

    // STEP 1: Transform event (internal → Z-API format)
    webhookPayload, err := p.transformEvent(ctx, event)
    if err != nil {
        var skippedErr *skippedEventError
        if errors.As(err, &skippedErr) {
            return p.handleSkippedEvent(ctx, event, skippedErr.Reason)
        }
        return p.handleTransformError(ctx, event, err)
    }

    // STEP 2: Deliver webhook
    deliveryResult, err := p.deliverWebhook(ctx, event, webhookPayload)
    if err != nil {
        return p.handleDeliveryError(ctx, event, err)
    }

    // STEP 3: Handle result
    if deliveryResult.Success {
        return p.handleSuccess(ctx, event, start)
    }

    if deliveryResult.Retryable {
        return p.handleRetryableError(ctx, event, deliveryResult)
    }

    return p.handlePermanentError(ctx, event, deliveryResult)
}

func (p *EventProcessor) transformEvent(ctx context.Context, event *persistence.OutboxEvent) ([]byte, error) {
    logger := logging.ContextLogger(ctx, nil)

    // Decode internal event format
    var encoded string
    if err := json.Unmarshal(event.Payload, &encoded); err != nil {
        return nil, fmt.Errorf("decode payload string: %w", err)
    }

    internalEvent, err := encoding.DecodeInternalEvent(encoded)
    if err != nil {
        return nil, fmt.Errorf("decode internal event: %w", err)
    }

    // Skip history_sync events (not relevant for webhooks)
    if internalEvent.EventType == "history_sync" {
        logger.Warn("skipping history sync event",
            slog.String("event_id", event.EventID.String()))
        return nil, &skippedEventError{Reason: "history_sync_temporarily_ignored"}
    }

    // Inject media_url if available
    if event.MediaURL != nil && *event.MediaURL != "" {
        if internalEvent.Metadata == nil {
            internalEvent.Metadata = make(map[string]string)
        }
        internalEvent.Metadata["media_url"] = *event.MediaURL
    }

    // Transform to Z-API format
    zapiEvent, err := transformzapi.Transform(ctx, internalEvent)
    if err != nil {
        return nil, fmt.Errorf("transform to zapi: %w", err)
    }

    // Marshal to JSON
    payload, err := json.Marshal(zapiEvent)
    if err != nil {
        return nil, fmt.Errorf("marshal zapi event: %w", err)
    }

    return payload, nil
}

func (p *EventProcessor) handleSuccess(ctx context.Context, event *persistence.OutboxEvent, start time.Time) error {
    logger := logging.ContextLogger(ctx, nil)

    duration := time.Since(start)

    // Update event status to 'delivered'
    if err := p.outboxRepo.MarkEventDelivered(ctx, event.EventID); err != nil {
        logger.Error("failed to mark event delivered",
            slog.Any("error", err))
        return fmt.Errorf("mark delivered: %w", err)
    }

    // Update metrics IMMEDIATELY
    p.metrics.EventsProcessed.WithLabelValues(
        p.instanceID.String(),
        event.EventType,
        "success",
    ).Inc()

    p.metrics.EventProcessingDuration.WithLabelValues(
        p.instanceID.String(),
        event.EventType,
    ).Observe(duration.Seconds())

    logger.Info("event delivered successfully",
        slog.String("event_id", event.EventID.String()),
        slog.Duration("duration", duration),
        slog.Int("attempts", event.Attempts+1))

    return nil
}

func (p *EventProcessor) handleRetryableError(ctx context.Context, event *persistence.OutboxEvent, result *DeliveryResult) error {
    logger := logging.ContextLogger(ctx, nil)

    // Calculate next attempt with exponential backoff
    nextAttempt := time.Now().Add(calculateBackoff(event.Attempts + 1))

    if err := p.outboxRepo.UpdateEventStatus(ctx,
        event.EventID,
        persistence.EventStatusRetrying,
        event.Attempts+1,
        &nextAttempt,
        &result.ErrorMessage,
    ); err != nil {
        logger.Error("failed to update retry status",
            slog.Any("error", err))
        return fmt.Errorf("update retry status: %w", err)
    }

    // Update metrics
    p.metrics.EventsProcessed.WithLabelValues(
        p.instanceID.String(),
        event.EventType,
        "retrying",
    ).Inc()

    logger.Warn("event delivery failed, will retry",
        slog.String("event_id", event.EventID.String()),
        slog.Int("attempt", event.Attempts+1),
        slog.Time("next_attempt", nextAttempt),
        slog.String("error", result.ErrorMessage))

    return nil
}

func (p *EventProcessor) handlePermanentError(ctx context.Context, event *persistence.OutboxEvent, result *DeliveryResult) error {
    logger := logging.ContextLogger(ctx, nil)

    // Move to DLQ after max attempts
    if event.Attempts+1 >= event.MaxAttempts {
        return p.moveToDLQ(ctx, event, result.ErrorMessage)
    }

    // Update to failed status
    if err := p.outboxRepo.UpdateEventStatus(ctx,
        event.EventID,
        persistence.EventStatusFailed,
        event.Attempts+1,
        nil,
        &result.ErrorMessage,
    ); err != nil {
        logger.Error("failed to update failed status",
            slog.Any("error", err))
        return fmt.Errorf("update failed status: %w", err)
    }

    // Update metrics
    p.metrics.EventsProcessed.WithLabelValues(
        p.instanceID.String(),
        event.EventType,
        "failed",
    ).Inc()

    // Capture in Sentry for critical failures
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetTag("component", "event_processor")
        scope.SetTag("instance_id", p.instanceID.String())
        scope.SetTag("event_type", event.EventType)
        scope.SetContext("event", map[string]interface{}{
            "event_id":        event.EventID.String(),
            "sequence_number": event.SequenceNumber,
            "attempts":        event.Attempts + 1,
        })
        sentry.CaptureMessage(fmt.Sprintf("Event delivery failed: %s", result.ErrorMessage))
    })

    logger.Error("event delivery failed permanently",
        slog.String("event_id", event.EventID.String()),
        slog.Int("attempts", event.Attempts+1),
        slog.String("error", result.ErrorMessage))

    return nil
}

// calculateBackoff returns exponential backoff with jitter
func calculateBackoff(attempt int) time.Duration {
    // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s
    base := time.Second * time.Duration(1<<uint(attempt-1))

    // Cap at 1 minute
    if base > time.Minute {
        base = time.Minute
    }

    // Add 20% jitter to prevent thundering herd
    jitter := time.Duration(rand.Float64() * 0.2 * float64(base))

    return base + jitter
}
```

**Key Patterns**:
- **Status Tracking**: Update status at each stage (processing → retrying → delivered/failed)
- **Exponential Backoff**: `1s → 2s → 4s → 8s → 16s → 32s` with jitter
- **DLQ Movement**: Events move to DLQ after max attempts (default: 6)
- **Metrics First**: Update metrics IMMEDIATELY after status changes
- **Sentry Capture**: Critical failures captured with full context
- **Error Classification**: Distinguish retryable vs permanent errors

### 8.5. Worker Lifecycle Best Practices

#### Graceful Shutdown Sequence

```go
// cmd/server/main.go
func main() {
    // ... initialization ...

    // Start coordinator
    if err := coordinator.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    logger.Info("shutdown signal received")

    // Create shutdown context with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // STEP 1: Stop HTTP server (no new requests)
    if err := server.Shutdown(shutdownCtx); err != nil {
        logger.Error("http shutdown failed", slog.Any("error", err))
    }

    // STEP 2: Stop coordinator (finish processing current events)
    if err := coordinator.Stop(shutdownCtx); err != nil {
        logger.Warn("coordinator shutdown timeout", slog.Any("error", err))
    }

    // STEP 3: Close database connections
    dbPool.Close()

    logger.Info("shutdown complete")
}
```

#### Worker Registration on Instance Connect

```go
// internal/instances/service.go
func (s *Service) ConnectInstance(ctx context.Context, instanceID uuid.UUID) error {
    logger := logging.ContextLogger(ctx, s.log)

    // Update instance status
    if err := s.repo.UpdateConnectionStatus(ctx, instanceID, true); err != nil {
        return fmt.Errorf("update connection status: %w", err)
    }

    // Register dispatch worker for event processing
    if err := s.coordinator.RegisterInstance(ctx, instanceID); err != nil {
        logger.Error("failed to register dispatch worker",
            slog.String("instance_id", instanceID.String()),
            slog.Any("error", err))
        return fmt.Errorf("register dispatch worker: %w", err)
    }

    logger.Info("instance connected and worker registered",
        slog.String("instance_id", instanceID.String()))

    return nil
}
```

#### Worker Unregistration on Instance Disconnect

```go
func (s *Service) DisconnectInstance(ctx context.Context, instanceID uuid.UUID) error {
    logger := logging.ContextLogger(ctx, s.log)

    // Unregister dispatch worker first
    if err := s.coordinator.UnregisterInstance(ctx, instanceID); err != nil {
        logger.Warn("failed to unregister dispatch worker",
            slog.String("instance_id", instanceID.String()),
            slog.Any("error", err))
        // Continue with disconnect even if unregister fails
    }

    // Update instance status
    if err := s.repo.UpdateConnectionStatus(ctx, instanceID, false); err != nil {
        return fmt.Errorf("update connection status: %w", err)
    }

    logger.Info("instance disconnected and worker unregistered",
        slog.String("instance_id", instanceID.String()))

    return nil
}
```

### 8.6. Worker Checklist for Code Review ✅

#### Coordinator ✅
- [ ] Map protection with `sync.RWMutex`
- [ ] Graceful shutdown with `WaitGroup`
- [ ] Context cancellation propagated to workers
- [ ] Metrics track active workers (`WorkersActive` gauge)
- [ ] Idempotent register/unregister operations
- [ ] Shutdown timeout configured (30 seconds recommended)

#### Instance Worker ✅
- [ ] Poll interval configured (500ms recommended)
- [ ] Context cancellation checked before processing
- [ ] Events fetched `ORDER BY sequence_number`
- [ ] Batch size reasonable (10 events recommended)
- [ ] Error isolation (single event failure doesn't kill worker)
- [ ] Non-blocking poll loop
- [ ] Stop channel for graceful shutdown

#### Event Processor ✅
- [ ] Status updated at each stage (processing → retrying → delivered/failed)
- [ ] Metrics updated IMMEDIATELY after status changes
- [ ] Exponential backoff with jitter
- [ ] DLQ movement after max attempts
- [ ] Sentry capture for critical failures
- [ ] Error classification (retryable vs permanent)
- [ ] Context propagation through processing pipeline
- [ ] Logging includes event_id, event_type, sequence_number

#### General ✅
- [ ] No goroutine leaks (all goroutines cleaned up on shutdown)
- [ ] No blocking operations in main loop
- [ ] Context cancellation respected everywhere
- [ ] Resource cleanup with `defer`
- [ ] Tests cover graceful shutdown scenarios
- [ ] Metrics exposed for monitoring

---

## 9. HTTP API Patterns

### 9.1. Router Structure (chi)

```go
// internal/http/router.go
func NewRouter(deps RouterDeps) http.Handler {
    r := chi.NewRouter()

    // Global middleware (order matters)
    r.Use(chiMiddleware.RequestID)           // Generate request ID
    r.Use(chiMiddleware.RealIP)              // Extract real IP
    r.Use(chiMiddleware.Recoverer)           // Recover from panics
    r.Use(chiMiddleware.Timeout(60 * time.Second))
    r.Use(ourMiddleware.RequestLogger(deps.Logger))
    r.Use(ourMiddleware.PrometheusMiddleware(deps.Metrics))
    r.Use(deps.SentryHandler.Handle)         // Sentry error tracking

    // Health endpoints (no auth)
    r.Get("/health", deps.HealthHandler.Health)
    r.Get("/ready", deps.HealthHandler.Ready)

    // Metrics endpoint
    r.Method(http.MethodGet, "/metrics", promhttp.Handler())

    // Debug profiler (development only)
    r.Mount("/debug", chiMiddleware.Profiler())

    // Documentation
    r.Route("/docs", func(dr chi.Router) {
        dr.Get("/", docs.UIHandler())
        dr.Get("/openapi.yaml", docs.YAMLHandler(deps.DocsConfig))
        dr.Get("/openapi.json", docs.JSONHandler(deps.DocsConfig))
    })

    // Instance API (authenticated)
    if deps.InstanceHandler != nil {
        deps.InstanceHandler.Register(r)
    }

    // Partner API (requires partner token)
    if deps.PartnerHandler != nil {
        r.Group(func(pr chi.Router) {
            pr.Use(ourMiddleware.PartnerAuth(deps.PartnerToken))
            deps.PartnerHandler.Register(pr)
        })
    }

    return r
}
```

### 9.2. Handler Pattern

```go
// internal/http/handlers/instances.go
type InstanceHandler struct {
    service *instances.Service  // Business logic layer
    log     *slog.Logger        // Structured logger
}

func (h *InstanceHandler) Register(r chi.Router) {
    // Z-API compatible routes
    r.Route("/instances/{instanceId}/token/{token}", func(r chi.Router) {
        r.Get("/status", h.getStatus)
        r.Get("/qr-code", h.getQRCode)
        r.Post("/disconnect", h.disconnect)
        r.Put("/update-webhook-delivery", h.updateWebhookDelivery)
    })
}

func (h *InstanceHandler) getStatus(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Parse and validate instance ID
    instanceID, ok := h.parseInstanceID(w, r)
    if !ok {
        return // Error already sent
    }

    // Enrich context with instance_id for logging
    ctx = logging.WithAttrs(ctx,
        slog.String("instance_id", instanceID.String()),
        slog.String("operation", "get_status"))

    // Extract authentication tokens
    instanceToken := chi.URLParam(r, "token")
    clientToken := r.Header.Get("Client-Token")

    // Delegate to service layer
    status, err := h.service.GetStatus(ctx, instanceID, clientToken, instanceToken)
    if err != nil {
        h.handleServiceError(ctx, w, err)
        return
    }

    // Success response
    respondJSON(w, http.StatusOK, status)
}

// Helper: Parse and validate UUID from path
func (h *InstanceHandler) parseInstanceID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
    idStr := chi.URLParam(r, "instanceId")
    if idStr == "" {
        respondError(w, http.StatusBadRequest, "missing instance ID")
        return uuid.Nil, false
    }

    id, err := uuid.Parse(idStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "invalid instance ID format")
        return uuid.Nil, false
    }

    return id, true
}

// Helper: Map service errors to HTTP status codes
func (h *InstanceHandler) handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
    logger := logging.ContextLogger(ctx, h.log)

    switch {
    case errors.Is(err, instances.ErrNotFound):
        respondError(w, http.StatusNotFound, "instance not found")
    case errors.Is(err, instances.ErrUnauthorized):
        respondError(w, http.StatusUnauthorized, "invalid credentials")
    case errors.Is(err, instances.ErrInvalidState):
        respondError(w, http.StatusConflict, "invalid instance state")
    default:
        logger.Error("internal error", slog.Any("error", err))
        respondError(w, http.StatusInternalServerError, "internal server error")
    }
}
```

### 9.3. Response Helpers

```go
// internal/http/handlers/response.go
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}
```

### 9.4. HTTP Checklist ✅

- [ ] Routes follow Z-API conventions
- [ ] Context enriched with request metadata
- [ ] Authentication tokens extracted from headers/path
- [ ] Input validation before service calls
- [ ] Service errors mapped to HTTP status codes
- [ ] Structured JSON responses
- [ ] No business logic in handlers
- [ ] Middleware applied in correct order
- [ ] Timeout configured (60s default)
- [ ] Metrics middleware active

---

## 10. Testing, Security, Performance & Deployment

### 10.1. Testing Strategy

**Coverage Requirements**:
- Unit tests: ≥80% coverage
- Integration tests: Critical paths
- E2E tests: Main workflows

**Test Structure**:
```go
// internal/instances/service_test.go
func TestService_GetStatus(t *testing.T) {
    tests := []struct {
        name        string
        instanceID  uuid.UUID
        setupMock   func(*mocks.Repository)
        wantStatus  *Status
        wantErr     error
    }{
        {
            name: "success",
            instanceID: uuid.New(),
            setupMock: func(m *mocks.Repository) {
                m.On("GetByID", mock.Anything, mock.Anything).
                    Return(&Instance{Status: "connected"}, nil)
            },
            wantStatus: &Status{Connected: true},
            wantErr: nil,
        },
        {
            name: "not found",
            instanceID: uuid.New(),
            setupMock: func(m *mocks.Repository) {
                m.On("GetByID", mock.Anything, mock.Anything).
                    Return(nil, instances.ErrNotFound)
            },
            wantErr: instances.ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := new(mocks.Repository)
            tt.setupMock(repo)

            svc := instances.NewService(repo, nil, nil)

            status, err := svc.GetStatus(context.Background(), tt.instanceID, "client", "instance")

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
            require.Equal(t, tt.wantStatus.Connected, status.Connected)
        })
    }
}
```

### 10.2. Security Best Practices

**Authentication**:
- Dual token validation (Client-Token + instance token)
- Tokens stored as hashed values
- Partner API requires separate token

**PII Protection**:
- NEVER log tokens, passwords, phone numbers
- Sanitize sensitive data in Sentry
- Use `[REDACTED]` in logs for sensitive fields

**Input Validation**:
- Validate UUID format before parsing
- Sanitize webhook URLs
- Enforce max payload size (10MB)

### 10.3. Performance Requirements

**Latency Targets**:
- API responses: p95 <200ms
- Event persistence: <50ms
- Webhook delivery: <5s (including retries)

**Resource Limits**:
- Max 100 concurrent instances per node
- Event batch size: 10 events
- Database connections: 25-50

### 10.4. Configuration Management

**Environment Variables** (.env):
```bash
# Server
HTTP_PORT=8080
HTTP_TIMEOUT=60s

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/whatsapp_api
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_LOCK_TTL=5m

# S3 Storage
S3_BUCKET=whatsapp-media
S3_REGION=us-east-1
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx

# Observability
SENTRY_DSN=https://xxx@sentry.io/xxx
METRICS_PORT=9090

# Partners
PARTNER_TOKEN=secret_token_here
```

### 10.5. Z-API Compatibility

**Target**: 200+ endpoints

**Endpoint Mapping**:
```
GET  /instances/{id}/token/{token}/status
GET  /instances/{id}/token/{token}/qr-code
POST /instances/{id}/token/{token}/disconnect
PUT  /instances/{id}/token/{token}/update-webhook-delivery
...
```

**Response Format**:
```json
{
  "connected": true,
  "session": "active",
  "phone": "+5511999999999"
}
```

### 10.6. Deployment & Operations

**Docker Compose**:
```yaml
version: '3.8'
services:
  api:
    image: whatsapp-api:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/whatsapp_api
      REDIS_URL: redis://redis:6379/0
    depends_on:
      - db
      - redis

  db:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
```

**Monitoring**:
- Prometheus metrics on :9090/metrics
- Health checks: /health (liveness), /ready (readiness)
- Grafana dashboards for visualization
- Sentry for error tracking
- Alert on: high error rate (>1%), DLQ growth, worker failures

**Scaling**:
- Horizontal: Run multiple API nodes behind load balancer
- Vertical: Increase DB connections, worker concurrency
- Database: Read replicas for reporting queries

---

## 11. Final Checklist (Comprehensive)

### Code Quality ✅
- [ ] No `panic` for expected errors
- [ ] Error wrapping with context
- [ ] All tests pass
- [ ] Coverage ≥80%
- [ ] Linter clean

### Architecture ✅
- [ ] Clean Architecture layers respected
- [ ] No business logic in handlers
- [ ] Repository pattern for data access
- [ ] Service layer for business logic

### Observability ✅  (MANDATORY)
- [ ] Context propagation everywhere
- [ ] Structured logging (slog)
- [ ] Metrics at event points
- [ ] Sentry for critical errors
- [ ] Health checks implemented

### Database ✅
- [ ] Migrations with Up/Down
- [ ] Transactions in service layer
- [ ] Indexes for performance
- [ ] Sequence ordering guaranteed

### Workers ✅
- [ ] Graceful shutdown
- [ ] Context cancellation
- [ ] Ordered event processing
- [ ] Exponential backoff retry
- [ ] DLQ for failed events

### HTTP API ✅
- [ ] Z-API compatible routes
- [ ] Input validation
- [ ] Authentication enforced
- [ ] Error responses consistent

### Security ✅
- [ ] No PII in logs
- [ ] Tokens validated
- [ ] Input sanitized
- [ ] HTTPS enforced

### Performance ✅
- [ ] Latency targets met
- [ ] Resource limits configured
- [ ] Connection pooling optimized

---

**END OF DEVELOPMENT RULES**

**Última Atualização**: 2025-10-12
**Versão**: 2.0.0

Para dúvidas sobre estas regras, consulte AGENTS.md para padrões de agente ou PLAN.md para roadmap técnico.
