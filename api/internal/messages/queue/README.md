# WhatsApp Message Queue System

Sistema de fila de mensagens para WhatsApp API usando River Queue v0.26.0, garantindo ordem FIFO estrita e alta performance.

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Garantia FIFO](#garantia-fifo)
- [Componentes](#componentes)
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Uso](#uso)
- [API](#api)
- [Testes](#testes)
- [Monitoramento](#monitoramento)

## 🎯 Visão Geral

Este sistema implementa uma fila de mensagens robusta e escalável para o envio de mensagens WhatsApp, garantindo:

- ✅ **Ordem FIFO Estrita**: Mensagens enviadas na ordem exata em que foram recebidas
- ✅ **Isolamento por Instância**: Uma instância não bloqueia outra
- ✅ **Delays Ilimitados**: Suporte para DelayMessage sem limites de tempo
- ✅ **DelayTyping**: Indicador "digitando..." antes do envio
- ✅ **Non-blocking HTTP**: API responde imediatamente após enfileirar
- ✅ **Alta Performance**: Processamento paralelo entre instâncias
- ✅ **Persistência**: Jobs armazenados no PostgreSQL (ACID)
- ✅ **Observabilidade**: Métricas, logs e views para monitoramento

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTP API Handler                         │
│                  (Recebe POST /send-text)                   │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     Coordinator                              │
│  - Gerencia todo o sistema de fila                          │
│  - Coordena Manager, Enqueue e Worker                       │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼             ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Manager    │ │   Enqueue    │ │    Worker    │
│              │ │   Service    │ │              │
│ - River      │ │              │ │ - Processa   │
│   Client     │ │ - FIFO       │ │   mensagens  │
│ - Queues     │ │   Ordering   │ │ - whatsmeow  │
│ - Migrations │ │ - Sequence   │ │ - DelayTyping│
└──────────────┘ └──────────────┘ └──────────────┘
        │                │                │
        └────────────────┼────────────────┘
                         ▼
                  ┌────────────┐
                  │ PostgreSQL │
                  │ (River DB) │
                  └────────────┘
```

### Fluxo de uma Mensagem

```
1. Cliente → POST /send-text
   ↓
2. Handler → Coordinator.EnqueueMessage()
   ↓
3. EnqueueService:
   - Gera zaapID único
   - Obtém último scheduled_at
   - Calcula novo scheduled_at (último + delay + jitter)
   - Obtém próximo sequence number (atomic)
   - Insere job com priority = -sequence
   ↓
4. Handler → Retorna zaapID imediatamente (non-blocking)
   ↓
5. River Worker (background):
   - Aguarda scheduled_at
   - Verifica instância conectada
   - Aplica DelayTyping (composing)
   - Envia mensagem via whatsmeow
   - Marca como completed
```

## 🔒 Garantia FIFO

### Estratégia de 5 Camadas

A ordem FIFO é garantida através de 5 mecanismos complementares:

#### 1️⃣ **Isolamento de Fila**
- Cada instância tem sua própria fila: `instance-{uuid}`
- Formato: `instance-550e8400-e29b-41d4-a716-446655440000`
- Evita interferência entre instâncias

#### 2️⃣ **MaxWorkers=1**
- Configuração **obrigatória** por fila
- Processa apenas 1 mensagem por vez
- Garante sequencialidade

#### 3️⃣ **Números de Sequência**
- Função SQL atômica: `get_next_message_sequence()`
- Contador monotônico por instância
- Previne race conditions

#### 4️⃣ **Encadeamento de Tempo**
- Cada mensagem agendada **após** a anterior
- Formula: `scheduled_at = MAX(last_scheduled_at, NOW) + delay + jitter`
- Cria cadeia de dependências

#### 5️⃣ **Prioridade por Sequência**
- Priority = -sequence_number
- River processa prioridades menores primeiro
- -1, -2, -3, ... garante ordem exata

### Exemplo Prático

```
Mensagem 1: sequence=1, priority=-1, scheduled_at=10:00:00
Mensagem 2: sequence=2, priority=-2, scheduled_at=10:00:03
Mensagem 3: sequence=3, priority=-3, scheduled_at=10:00:06

River processa: -1 → -2 → -3 (ordem FIFO garantida)
```

## 🧩 Componentes

### 1. Coordinator (`coordinator.go`)

**Responsabilidade**: Orquestrar todo o sistema

**Métodos principais**:
```go
// Criar coordenador
coordinator, err := NewCoordinator(ctx, &CoordinatorConfig{
    Pool:           dbPool,
    ClientRegistry: whatsappRegistry,
    Logger:         logger,
})

// Adicionar fila de instância
coordinator.AddInstanceQueue(ctx, instanceID)

// Enfileirar mensagem
zaapID, err := coordinator.EnqueueMessage(ctx, instanceID, SendMessageArgs{
    Phone: "5511999999999",
    MessageType: MessageTypeText,
    TextContent: &TextMessage{Message: "Olá!"},
    DelayMessage: 5000, // 5 segundos
    DelayTyping: 2000,  // 2 segundos
})

// Listar jobs da fila
jobs, err := coordinator.ListQueueJobs(ctx, instanceID, 50, 0)

// Cancelar job
err := coordinator.CancelJob(ctx, zaapID)

// Shutdown gracioso
coordinator.Stop(ctx)
```

### 2. RiverQueueManager (`river_client.go`)

**Responsabilidade**: Gerenciar River client e filas

**Características**:
- Executa migrations automaticamente
- Gerencia lifecycle do River client
- Adiciona/remove filas dinamicamente
- Configuração MaxWorkers=1 por fila

### 3. EnqueueService (`enqueue.go`)

**Responsabilidade**: Enfileirar mensagens com ordem FIFO

**Características**:
- Geração de zaapID único
- Cálculo de scheduled_at com encadeamento
- Obtenção atômica de sequence number
- Suporte transacional

### 4. SendMessageWorker (`worker.go`)

**Responsabilidade**: Processar envio de mensagens

**Características**:
- Implementa interface `river.Worker`
- Valida instância conectada
- Aplica DelayTyping
- Envia via whatsmeow
- Retry automático em falhas

### 5. Models (`models.go`)

**Tipos de Mensagem**:
- ✅ Text (implementado)
- 🚧 Image (TODO)
- 🚧 Audio (TODO)
- 🚧 Video (TODO)
- 🚧 Document (TODO)
- 🚧 Location (TODO)
- 🚧 Contact (TODO)
- 🚧 Interactive (TODO)

## 📦 Instalação

### Pré-requisitos

- Go 1.24+
- PostgreSQL 13+
- River Queue v0.26.0

### Dependências

```bash
go get github.com/riverqueue/river@v0.26.0
go get github.com/riverqueue/river/riverdriver/riverpgxv5@v0.26.0
go get github.com/riverqueue/river/rivermigrate@v0.26.0
```

### Migrations

As migrations são executadas automaticamente ao criar o `RiverQueueManager`:

```sql
-- River core tables
river_job
river_leader
river_migration

-- Custom tables
message_sequences

-- Functions
get_next_message_sequence(instance_id UUID)

-- Views
v_queue_stats_by_instance
v_recent_failed_jobs
v_active_queues
v_message_sequences_summary
```

## ⚙️ Configuração

### Variáveis de Ambiente

```bash
# Habilitar sistema de fila
MESSAGE_QUEUE_ENABLED=true

# Workers por fila (DEVE SER 1 para FIFO)
MESSAGE_QUEUE_WORKER_MAX_WORKERS=1

# Intervalo de polling
MESSAGE_QUEUE_POLL_INTERVAL=100ms

# Tentativas máximas de retry
MESSAGE_QUEUE_MAX_JOB_ATTEMPTS=3

# Timeout por job
MESSAGE_QUEUE_JOB_TIMEOUT=5m

# Retenção de jobs completados
MESSAGE_QUEUE_RETENTION_PERIOD=24h

# Jitter aleatório (para evitar colisões)
MESSAGE_QUEUE_MIN_JITTER=1s
MESSAGE_QUEUE_MAX_JITTER=3s
```

### Configuração em Go

```go
import "github.com/your-org/zedaapi/api/internal/config"

cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Acessar configurações
enabled := cfg.MessageQueue.Enabled
maxWorkers := cfg.MessageQueue.WorkerMaxWorkers
```

## 🚀 Uso

### Setup Básico

```go
package main

import (
    "context"
    "log/slog"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/Setup-Automatizado/zedaapi/api/internal/messages/queue"
)

func main() {
    ctx := context.Background()

    // 1. Criar pool de conexões
    pool, err := pgxpool.New(ctx, "postgres://...")
    if err != nil {
        panic(err)
    }
    defer pool.Close()

    // 2. Criar coordenador (inicia River automaticamente)
    coordinator, err := queue.NewCoordinator(ctx, &queue.CoordinatorConfig{
        Pool:           pool,
        ClientRegistry: whatsappClientRegistry,
        Logger:         slog.Default(),
    })
    if err != nil {
        panic(err)
    }
    defer coordinator.Stop(ctx)

    // 3. Adicionar fila para instância
    instanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
    if err := coordinator.AddInstanceQueue(ctx, instanceID); err != nil {
        panic(err)
    }

    // 4. Enfileirar mensagem
    zaapID, err := coordinator.EnqueueMessage(ctx, instanceID, queue.SendMessageArgs{
        Phone:        "5511999999999",
        MessageType:  queue.MessageTypeText,
        TextContent:  &queue.TextMessage{Message: "Olá, mundo!"},
        DelayMessage: 0,
        DelayTyping:  2000, // 2 segundos de "digitando..."
    })
    if err != nil {
        panic(err)
    }

    log.Printf("Mensagem enfileirada: %s", zaapID)
}
```

### Envio com Delay

```go
// Agendar mensagem para 1 hora no futuro
zaapID, err := coordinator.EnqueueMessage(ctx, instanceID, queue.SendMessageArgs{
    Phone:        "5511999999999",
    MessageType:  queue.MessageTypeText,
    TextContent:  &queue.TextMessage{Message: "Lembrete!"},
    DelayMessage: 3600000, // 1 hora em ms
    DelayTyping:  1000,
})
```

### Consultar Status

```go
// Obter informações do job
job, err := coordinator.GetJobByZaapID(ctx, zaapID)
if err != nil {
    // Job não encontrado
}

fmt.Printf("Status: %s\n", job.Status)
fmt.Printf("Sequência: %d\n", job.SequenceNumber)
fmt.Printf("Agendado para: %s\n", job.ScheduledFor)
```

### Listar Fila

```go
// Listar primeiros 50 jobs da fila
response, err := coordinator.ListQueueJobs(ctx, instanceID, 50, 0)

fmt.Printf("Total de jobs: %d\n", response.Total)
for _, job := range response.Jobs {
    fmt.Printf("- %s: %s (%s)\n", job.ZaapID, job.Phone, job.Status)
}
```

### Cancelar Mensagem

```go
// Cancelar mensagem pendente
err := coordinator.CancelJob(ctx, zaapID)
if err != nil {
    // Job não encontrado ou já processado
}
```

## 📊 API

### Coordinator

```go
type Coordinator interface {
    // Lifecycle
    AddInstanceQueue(ctx context.Context, instanceID uuid.UUID) error
    RemoveInstanceQueue(ctx context.Context, instanceID uuid.UUID) error
    Stop(ctx context.Context) error
    IsStarted() bool

    // Enqueue
    EnqueueMessage(ctx context.Context, instanceID uuid.UUID, args SendMessageArgs) (zaapID string, err error)

    // Query
    GetQueueStats(ctx context.Context, instanceID uuid.UUID) (*QueueStats, error)
    GetQueuePosition(ctx context.Context, zaapID string) (int, error)
    GetJobByZaapID(ctx context.Context, zaapID string) (*QueueJobInfo, error)
    ListQueueJobs(ctx context.Context, instanceID uuid.UUID, limit, offset int) (*QueueListResponse, error)
    ListActiveQueues() []*InstanceQueue

    // Management
    CancelJob(ctx context.Context, zaapID string) error
}
```

### SendMessageArgs

```go
type SendMessageArgs struct {
    // Identificação
    ZaapID     string    // Gerado automaticamente
    InstanceID uuid.UUID // UUID da instância

    // Destinatário
    Phone string // Formato: 5511999999999 ou 5511999999999@s.whatsapp.net

    // Tipo e conteúdo
    MessageType MessageType // text, image, audio, etc.
    TextContent *TextMessage
    ImageContent *MediaMessage
    // ... outros tipos

    // Delays
    DelayMessage int64 // Delay antes de agendar (ms)
    DelayTyping  int64 // Indicador "digitando..." (ms)

    // FIFO (preenchido automaticamente)
    SequenceNumber int64
    ScheduledFor   time.Time

    // Metadata
    EnqueuedAt time.Time
    Metadata   map[string]interface{}
}
```

### JobStatus

```go
const (
    JobStatusAvailable = "available" // Pronto para processar
    JobStatusScheduled = "scheduled" // Aguardando scheduled_at
    JobStatusRunning   = "running"   // Em processamento
    JobStatusCompleted = "completed" // Sucesso
    JobStatusCancelled = "cancelled" // Cancelado manualmente
    JobStatusDiscarded = "discarded" // Falhou permanentemente
    JobStatusRetryable = "retryable" // Falhou, vai retentar
)
```

## 🧪 Testes

### Teste de FIFO

```go
func TestFIFOOrdering(t *testing.T) {
    // 1. Enfileirar 10 mensagens rapidamente
    var zaapIDs []string
    for i := 0; i < 10; i++ {
        zaapID, err := coordinator.EnqueueMessage(ctx, instanceID, SendMessageArgs{
            Phone: fmt.Sprintf("551199999%04d", i),
            MessageType: MessageTypeText,
            TextContent: &TextMessage{Message: fmt.Sprintf("Mensagem %d", i)},
        })
        require.NoError(t, err)
        zaapIDs = append(zaapIDs, zaapID)
    }

    // 2. Aguardar processamento
    time.Sleep(5 * time.Second)

    // 3. Verificar ordem de envio
    for i, zaapID := range zaapIDs {
        job, err := coordinator.GetJobByZaapID(ctx, zaapID)
        require.NoError(t, err)
        assert.Equal(t, JobStatusCompleted, job.Status)
        assert.Equal(t, int64(i+1), job.SequenceNumber)
    }
}
```

## 📈 Monitoramento

### Views SQL

```sql
-- Status das filas por instância
SELECT * FROM v_queue_stats_by_instance;

-- Jobs falhados recentemente
SELECT * FROM v_recent_failed_jobs;

-- Filas ativas
SELECT * FROM v_active_queues;

-- Sequências de mensagens
SELECT * FROM v_message_sequences_summary;
```

### Métricas

```go
// Obter estatísticas da fila
stats, err := coordinator.GetQueueStats(ctx, instanceID)
fmt.Printf("Jobs disponíveis: %d\n", stats.AvailableJobs)
fmt.Printf("Jobs rodando: %d\n", stats.RunningJobs)
fmt.Printf("Jobs completados: %d\n", stats.CompletedJobs)
fmt.Printf("Jobs falhados: %d\n", stats.FailedJobs)
```

### Logs Estruturados

```json
{
  "level": "info",
  "msg": "message enqueued",
  "zaap_id": "a1b2c3d4e5f6...",
  "instance_id": "550e8400-...",
  "phone": "5511999999999",
  "message_type": "text",
  "sequence": 42,
  "scheduled_at": "2024-01-01T10:00:05Z",
  "job_id": 12345
}
```

## 🔧 Troubleshooting

### Mensagens fora de ordem

**Causa**: MaxWorkers > 1
**Solução**: Garantir `MESSAGE_QUEUE_WORKER_MAX_WORKERS=1`

### Jobs presos em "running"

**Causa**: Worker travou ou crashou
**Solução**: River Rescuer marca jobs como "retryable" após timeout

### Mensagens não enviadas

**Causa**: Instância desconectada
**Solução**: Worker retorna `JobSnooze` e tenta novamente em 30s

### Alto uso de CPU

**Causa**: Poll interval muito curto
**Solução**: Aumentar `MESSAGE_QUEUE_POLL_INTERVAL`

## 📚 Referências

- [River Queue v0.26.0](https://github.com/riverqueue/river/releases/tag/v0.26.0)
- [River Documentation](https://riverqueue.com/docs)
- [whatsmeow](https://github.com/tulir/whatsmeow)
- [PostgreSQL ACID](https://www.postgresql.org/docs/current/tutorial-transactions.html)

## 🎯 Roadmap

### Sprint 2 (Próximos Passos)
- [ ] Implementar HTTP handlers (POST /send-text, GET /queue, DELETE /queue/:id)
- [ ] Adicionar suporte para mensagens de imagem
- [ ] Adicionar suporte para mensagens de áudio/vídeo
- [ ] Adicionar suporte para documentos
- [ ] Implementar testes unitários completos
- [ ] Adicionar métricas Prometheus

### Sprint 3
- [ ] Suporte para mensagens interativas (botões, listas)
- [ ] Suporte para localização
- [ ] Suporte para contatos
- [ ] Dashboard de monitoramento
- [ ] Rate limiting por instância

## 📄 Licença

Este código faz parte do projeto WhatsApp API e segue a mesma licença do projeto principal.
