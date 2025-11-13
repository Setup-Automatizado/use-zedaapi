# Plano de implementacao para compatibilidade Z-API e fila de mensagens

## Panorama do estado atual
- API atual cobre cadastros de instancias, configuracao de webhooks e operacoes basicas de parceiro, mas ainda nao expoe camada HTTP para envio de mensagens ou dominios como contatos, grupos e comunidades.
- Os eventos recebidos do WhatsApp passam pelo orquestrador (`api/internal/events`) que grava no Postgres (`event_outbox`) e despacha via workers dedicados (`dispatch`, `media`). Essa fundacao garante ordenacao por instancia, reprocessamento e observabilidade robusta.
- O `ClientRegistry` (`api/internal/whatsmeow`) gerencia conexoes WhatsApp, integra locks Redis e ja notifica coordenadores existentes na conexao (`dispatch`, `media`). Ainda nao ha coordenador de fila de envio.
- Observabilidade possui infraestrutura pronta (Prometheus, Sentry, logging estruturado) e precisa ser estendida para novas features.
- Postman collection (`api/z_api/Nova API Collection.postman_collection.json`) lista todo o contrato alvo; grande parte dos endpoints nao existe na API.

## Objetivos principais
- Garantir compatibilidade funcional com a collection Z-API nas areas de instancias, mensagens, contatos, chats, grupos, comunidades, fila de mensagens, webhooks e parceiros.
- Implementar fila de envio por instancia com comportamento identico ao Z-API (delay humano randomico 1-3s, `delayMessage` opcional, isolamento por instancia, tolerancia a desconexoes).
- Manter os padroes de observabilidade obrigatorios (logs, contexto, metricas, Sentry, health) em todas as novas rotas, workers e integracoes.
- Preservar alta disponibilidade e escalabilidade horizontal (workers idempotentes, locks adequados, armazenar estado em Postgres).

## Requisitos chave e restricoes
- Cada funcao exposta via HTTP deve aceitar `context.Context` e propagar logger com `logging.ContextLogger`/`logging.WithAttrs`.
- Nao usar `fmt.Printf` ou logs sem `slog`. Nenhum emoji ou dado sensivel em logs ou Sentry.
- Atualizar `observability.Metrics` com novos contadores/histogramas e acionar metricas no ponto do evento.
- Erros criticos capturados no Sentry devem usar `WithScope` com tags `component`, `instance_id`, `severity` e contexto adicional.
- Nao editar arquivos gerados (`proto/`, `internals.go`). Ajustes exigem regeneracao.
- Priorizar ASCII em arquivos de texto e codigo (sem acentos).
- `go test ./...` e `pre-commit run --all-files` devem passar antes de concluir ciclos de implementacao.

## Arquitetura proposta da fila de envio

### Modelagem de dados
- Criar tabela `message_queue` (Postgres) com colunas:
  - `id BIGSERIAL`, `instance_id UUID`, `message_id UUID` (tracking interno), `type` (enum textual), `payload JSONB`, `status` (`pending`, `processing`, `sent`, `failed`, `canceled`), `attempts`, `max_attempts`, `scheduled_at TIMESTAMP`, `next_attempt_at TIMESTAMP`, `delivered_at`, `error`, `metadata JSONB`, `created_at`, `updated_at`.
  - `sequence_number BIGINT` gerado via funcao similar a `get_next_event_sequence` para manter ordenacao por instancia.
  - Campos para midia (por ex. `media_id`, `media_status`, `upload_retry_count`) amarrados ao pipeline de upload.
- Opcional: tabela `message_queue_attempts` historizando tentativas (timestamp, erro, duracao, contexto de Sentry) para auditoria e debugging.
- Funcoes auxiliares SQL:
  - `get_next_message_sequence(instance_id UUID)` (plpgsql) para atribuir sequencial per-instancia.
  - Indexes: `idx_message_queue_instance_status_next` (instance_id, status, next_attempt_at, sequence_number), `idx_message_queue_status` para consultas globais, `idx_message_queue_message_id` para recuperacao por ID exposto ao cliente.

### Camadas de codigo
- `api/internal/messages/repository.go`
  - Operacoes CRUD no Postgres (enqueue, poll, atualizar status, cancelar, listar fila, limpar, obter contagem).
  - Consultas com `FOR UPDATE SKIP LOCKED` para evitar conflitos quando houver varios workers.
- `api/internal/messages/model.go`
  - Estruturas dominio (`QueuedMessage`, `SendParams`, `Attachment`, `DelayPolicy`, etc.).
- `api/internal/messages/serializer`
  - Conversao das cargas HTTP (text, image base64/url, audio, documento, contato, poll, event, etc.) para um formato interno padronizado.
  - Validacao e normalizacao (JIDs, DDD, remover caracteres nao numericos).
- `api/internal/messages/service.go`
  - Autenticacao de instancia/tokens usando `instances.Repository`.
  - Resolver store JID, verificar status da assinatura, armazenar `delayMessage`, calcular `scheduled_at`.
  - Enfileirar comando de envio sem bloquear requisicao HTTP.
  - Mapear `messageId` custom fornecido pelo cliente (Z-API exige capacidade de enviar ID proprio; armazenar no payload para correlacao com recibos).
- `api/internal/messages/handler.go`
  - Endpoints `POST /send-*`, `POST /forward-message`, `POST /send-reaction`, etc., agrupados por dominio.
  - Utilizar sub-rotas `chi` reusando validacao de instancia e `Client-Token` header.
  - Respostas padronizadas com status `Queue` (similar Z-API: `{"status":"QUEUED","queueId":...,"messageId":...}`).
- `api/internal/messages/coordinator.go`
  - Estrutura similar a `events/dispatch.Coordinator`, registrando worker por instancia ao conectar.
  - Gatilhos no `ClientRegistry` (`wrapEventHandler` para `events.Connected`/`Disconnected`) para `RegisterInstance` e `UnregisterInstance`.
  - `Start`/`Stop` com `sync.WaitGroup`, `context.Context`, `ShutdownGracePeriod`.
- `api/internal/messages/worker.go`
  - Loop `poll -> send -> mark`.
  - `Poll`: busca ate `cfg.Messages.BatchSize` itens `pending`/`retrying` com `next_attempt_at <= now`.
  - Respeitar `delayMessage`: armazenar em `metadata` e, apos enviar uma mensagem, aplicar `sleep` randomico 1-3s + `delayMessage` se presente antes de prosseguir.
  - Checar se cliente esta conectado (`ClientRegistry.IsConnected(instanceID)`); se nao, reagendar `next_attempt_at = now + backoff`.
  - Para midia, integrar com `mediaCoordinator` (upload antes do envio, fallback S3/local).
  - Atualizar metricas: backlog gauge, counter de sucesso/falha, histogram de duracao, counter por tipo.
  - Em caso de erro:
    - Atualizar `status=retrying`, `attempts++`, `next_attempt_at = now + backoff`.
    - Registrar `last_error`, capturar Sentry para erros nao esperados.
  - Ao exceder `max_attempts`, mover para `failed` e opcionalmente inserir em `message_queue_dlq` (planejar se necessario).
- `api/internal/messages/sender.go`
  - Encapsular interacao com `whatsmeow.Client` (text, media, reaction, poll).
  - Converter `phone` em `types.JID`.
  - Realizar uploads (base64 vs url): reusar utils `upload.go` ou criar `MediaUploadService`.
  - Para `delayMessage` per message, reusar a informacao para `sleep`.
  - Garantir que `SendMessage` receba `context` com timeout configuravel (`cfg.Messages.SendTimeout`).

### Observabilidade dedicada
- Novas metricas em `observability.Metrics`:
  - `MessagesQueued *prometheus.CounterVec` (instance_id, type).
  - `MessageQueueBacklog *prometheus.GaugeVec` (instance_id, priority).
  - `MessageSendAttempts *prometheus.CounterVec` (instance_id, type, status).
  - `MessageSendDuration *prometheus.HistogramVec` (instance_id, type).
  - `MessageQueueLatency *prometheus.HistogramVec` (tempo entre enqueue e envio).
  - `MessageQueueFailures *prometheus.CounterVec` (instance_id, error_class).
- Logar inicio/fim de cada envio com `instance_id`, `queue_id`, `message_type`, `target`, `duration`.
- Capturar erros criticos no Sentry com tags `component=message_queue_worker`, `message_type`, `instance_id`, `severity`.
- Atualizar `/ready` para incluir checagem opcional da fila (ping leve no Postgres e validacao de conexao com Redis se a fila usar locks).

### API de fila (compatibilidade Z-API)
- `GET /queue`: listar mensagens pendentes com filtros (`status`, `limit`, `offset`, `type`).
- `DELETE /queue`: limpar fila pendente (opcao `all=true` ou `olderThan`).
- `DELETE /queue/{messageId}`: remover mensagem especifica (caso exista antes do envio).
- Resposta deve incluir contadores e `queueLength`.
- Validar `Client-Token` e `instanceToken` em todos.
- Integrar com metricas (e.g., counter de `queue_purged`).

## Cobertura de endpoints por dominio

### Instancias (dados do Postman)
- Ja implementados: `/status`, `/qr-code`, `/qr-code/image`, `/phone-code`, `/restart`, `/disconnect`, webhooks update.
- Pendentes:
  - Atualizacoes de perfil: `/profile-name`, `/profile-picture`, `/profile-description`, `/update-name`, `/me`.
  - Configuracoes: `/update-call-reject-auto`, `/update-call-reject-message`, `/update-auto-read-message`, `/device`, `/update-notify-sent-by-me`, `/update-every-webhooks`.
  - Endpoint `/update-webhook-*` ja existem, validar nomes/contratos.
  - Necessario mapear adjacency no `ClientRegistry` (ex: metodos para editar perfil via `client.SetProfilePicture`, `client.SetProfileName`).
- Planejar rotas de health: `GET /device`, `GET /status` ja prontos, mas garantir shape da resposta de acordo com Z-API.

### Mensagens
- `POST /send-text`, `/send-image`, `/send-sticker`, `/send-gif`, `/send-audio`, `/send-video`, `/send-ptv`, `/send-document/mp4`, `/send-link`, `/send-location`, `/send-product`, `/send-catalog`, `/send-contact`, `/send-contacts`, `/send-button-actions`, `/send-button-list`, `/send-option-list`, `/send-button-otp`, `/send-button-pix`, `/send-poll`, `/send-poll-vote`, `/send-order`, `/order-status-update`, `/order-payment-update`, `/send-newsletter-admin-invite`, `/send-event`, `/send-edit-event`, `/send-event-response`, `/forward-message`, `/send-reaction`, `/send-remove-reaction`, `/pin-message`, `/read-message`, `/messages` (fetch single message info).
- Planejar implementacao incremental (text -> midia -> interativos -> commerce -> eventos).
- Cada mensagem deve gerar `queueId`, respeitar `delayMessage`, permitir `messageId` custom, e validar se destino e grupo/wide (e.g., `isGroup`).

### Contatos e chats
- `GET /contacts`, `GET /contacts/{PHONE}`, `GET /phone-exists/{PHONE}`, `GET /chats`, `GET /chats/{PHONE}`.
- Mapear funcoes do `whatsmeow.Client` (`GetContacts`, `GetChats`, `IsOnWhatsApp`).
- Adicionar caches e metricas (latencia, hits).

### Grupos
- Cria/atualiza: `/create-group`, `/update-group-name`, `/update-group-photo`, `/update-group-description`, `/update-group-settings`.
- Gerenciar participacao: `/add-participant`, `/remove-participant`, `/add-admin`, `/remove-admin`, `/leave-group`.
- Metadata: `GET /group-metadata/{GROUP_PHONE}`, `GET /group-invitation-metadata`.
- Usar chamadas `client.CreateGroup`, `client.SetGroupSubject`, `client.SetGroupDescription`, etc.
- Garantir logs com `group_jid`, `participants`.

### Comunidades
- `/communities`, `/communities-metadata/{COMMUNITY_ID}`, `/communities` (GET/POST).
- Verificar suporte do whatsmeow (newsletter / communities) e planejar wrappers.

### Produtos e catalogos
- `/products`, `/catalogs`, `/catalogs/config`, `/catalogs/collection`, `/business/*`, `/tags`.
- Avaliar dependencias de API WhatsApp Business (WA Business Catalog). Planejar wrappers no `Client` ou modulos dedicados (pode exigir endpoints externos).
- Priorizar depois que fila/mensagens estiverem estaveis (risco alto).

### Parceiros
- Ja implementado: `POST /instances/integrator/on-demand`, `POST /subscription`, `POST /cancel`, `DELETE /instances`, `GET /instances`.
- Verificar se body/response bate com Postman; ajustar se necessario.

### Webhooks
- Validar se updates de presence, connected, message-status ja batem com collection.
- Documentar (PLAN) necessidade de homologar payload transformado (Z-API schema).

## Roadmap sugerido por marcos

### Marco 0 - Descoberta e preparacao (1-2 sprints)
- [ ] Catalogar gaps entre OpenAPI atual e Postman, produzir matriz rota->status.
- [ ] Validar requisitos de integracao (S3, Redis, Postgres, whatsmeow versao).
- [ ] Definir nomenclatura de metricas e formato de payload da fila.
- [ ] Escrever especificacao tecnica detalhada da tabela `message_queue`.
- [ ] Criar testes de fumaca para instancias existentes (baseline).

### Marco 1 - Fundacao da fila e infraestrutura (1 sprint)
- [ ] Criar migracoes SQL para `message_queue` e funcoes auxiliares.
- [ ] Implementar repositorio `messages.Repository` com testes unitarios usando `pgxmock` ou banco em memoria.
- [ ] Adicionar configuracoes em `config.Config` (`Messages` section: `BatchSize`, `PollInterval`, `SendTimeout`, `MaxAttempts`, `DefaultDelayMin/Max`).
- [ ] Estender `observability.Metrics` com novos contadores/gauges e expor em `/metrics`.
- [ ] Implementar `Coordinator`/`Worker` com send stub (sem chamar whatsmeow ainda) e cobertura de testes.
- [ ] Integrar `Coordinator` ao `ClientRegistry` (register/unregister).
- [ ] Adicionar endpoints de filas (`GET /queue`, `DELETE /queue`, `DELETE /queue/{id}`) retornando dados fakes para validar pipeline.

### Marco 2 - Envio de texto e controle basico (1-2 sprints)
- [ ] Implementar `POST /send-text` com serializer, validacao, enfileiramento.
- [ ] Implementar sender para texto puro (sem midia) com `whatsmeow.Client.SendMessage`.
- [ ] Garantir `delayMessage` e random delay 1-3s aplicados.
- [ ] Atualizar logs e metricas (duracao, tentativas).
- [ ] Adicionar testes de integracao (usar `whatsmeow.DangerousInternals` mock ou `testcontainers`).
- [ ] Atualizar OpenAPI (`docs/openapi.yaml`) e Postman docs se necessario.
- [ ] Validar `read-message` e `messages` (consultas) usando dados do whatsmeow store.

### Marco 3 - Midia e anexos (2-3 sprints)
- [ ] Extender pipeline para suportar uploads (imagem, audio, video, documento, sticker, gif).
- [ ] Reusar `mediaCoordinator` ou criar helpers de upload (S3/local).
- [ ] Implementar endpoints `send-image`, `send-audio`, `send-video`, `send-document/mp4`, `send-sticker`, `send-gif`, `send-ptv`.
- [ ] Suportar envio via URL ou base64; armazenar metadados (sha256, mime).
- [ ] Incluir metricas de upload (tamanho, duracao, erros) aproveitando estrutura existente.
- [ ] Escrever testes com fixtures base64 e mocks de upload.

### Marco 4 - Mensagens interativas e commerce (2-3 sprints)
- [ ] Implementar `send-button-actions`, `send-button-list`, `send-option-list`, `send-button-otp`, `send-button-pix`, `send-link`, `send-location`, `send-product`, `send-catalog`.
- [ ] Adicionar validadores especificos (por exemplo, limite de botoes, campos obrigatorios).
- [ ] Integrar com APIs de product/catalog se disponiveis; caso contrario, planejar stub/resposta padrao.
- [ ] Implementar `/send-poll`, `/send-poll-vote`, `/send-order`, `/order-status-update`, `/order-payment-update`, `/send-event*`, `send-newsletter-admin-invite`.
- [ ] Garantir serializacao de payloads seguindo Z-API (ver exemplos no Postman).
- [ ] Atualizar docs e adicionar testes.

### Marco 5 - Contatos, chats, grupos, comunidades (2-3 sprints)
- [ ] Implementar endpoints de contatos/chats usando dados do store (com caching e metricas).
- [ ] Implementar funcoes de grupos (criar, listar, editar, participantes, metadata).
- [ ] Implementar comunidades (avaliar limite do whatsmeow; pode exigir features beta).
- [ ] Adicionar metricas e logs para cada dominio (`component=groups_service`, etc.).
- [ ] Documentar quotas/limites (ex: max participantes).

### Marco 6 - Finalizacao e hardening (1 sprint)
- [ ] Revisar todos os endpoints contra Postman (paridade de caminho, metodo, campos).
- [ ] Auditar observabilidade (log checklist, metricas, Sentry, health).
- [ ] Adicionar testes end-to-end rodando fluxo completo (enfileirar -> enviar -> webhook simulado).
- [ ] Atualizar README/Docs com instrucoes de uso.
- [ ] Validar performance (carga de fila, throughput) e ajustar configuracoes.
- [ ] Preparar scripts de monitoramento (alertas Prometheus, dashboards).

## Checklists transversais

- **Logs**: toda funcao HTTP/worker precisa logar inicio/fim e erros com campos padrao (`instance_id`, `queue_id`, `message_type`, `target`). Utilizar `DEBUG` para payloads sensiveis em dev, nunca logar base64 ou tokens.
- **Contexto**: garantir `ctx` como primeiro parametro e usar `logging.WithAttrs` para enriquecer com `instance_id`, `component`.
- **Metricas**: adicionar docs em `api/internal/observability/metrics.go` ao introduzir metricas novas e registrar no registrador global.
- **Sentry**: capturar apenas erros inesperados. Evitar flood (usar `sentry.WithScope` e, se necessario, amostragem manual).
- **Health**: estender `HealthHandler` para validar conectividade Postgres (fila), Redis (locks), S3 (midia). Retornar 503 se qualquer dependencia critica estiver indisponivel.
- **Configuracao**: atualizar `api/docker-compose.dev.yml`, variaveis `.env` e `docs` com novos parametros (`MESSAGES_*`).
- **Testes**:
  - Unitarios para serializadores, repositorios, workers (usando context com timeout).
  - Integracao com banco (pode usar `testcontainers-go`).
  - Contratos HTTP (talvez `go-testdeep` ou `httptest`).
  - Manter cobertura `go test -cover ./...`.
- **Documentacao**: alinhar `docs/openapi.yaml` e possivelmente gerar doc HTML. Sincronizar com Postman se alteracoes ocorrerm.
- **DevEx**: incluir scripts `make`/`mage` para migrar banco, rodar workers, seed de dados.

## Riscos e mitigacoes
- **Complexidade do contrato Z-API**: grande numero de rotas. Mitigar priorizando por uso e entregando em marcos; manter rastreador de progresso.
- **Suporte do whatsmeow a recursos de commerce/comunidades**: pode estar parcial. Documentar limitacoes e considerar usar APIs Business oficiais ou stubs ate que suporte esteja disponivel.
- **Carga de fila e escalabilidade**: risco de acumulo se instancia ficar offline. Mitigar com metricas de backlog, politica de expiracao e alertas.
- **Sincronizacao de tokens**: garantir validacao consistente (Client-Token + instance token). Revisar funcoes existentes para evitar divergencias.
- **Uploads grandes**: manter limites (`MediaMaxFileSize`), streaming para S3, e fallback local se S3 falhar. Implementar limpeza de arquivos temporarios.
- **Conflitos de locks Redis**: fila deve respeitar locks ja usados pelo `ClientRegistry`. Evitar deadlocks usando design sem locks extras ou reusando `lockManager`.
- **Observabilidade incompleta**: checklists obrigatorios devem ser auditados antes de merges. Considerar linter custom (ou revisao manual).

## Dependencias externas e perguntas em aberto
- Confirmar se endpoints de commerce exigem credenciais adicionais (p. ex. Facebook Graph).
- Validar exigencia de registro S3 (ou se base64 direto basta) para envios.
- Especificar comportamento de `delayMessage` quando valor < 1s ou > limite (definir maximo).
- Definir se `DELETE /queue` remove tambem itens `processing` e qual status retornar.
- Decidir estrategia para correlacao webhook -> mensagem enfileirada (usar `messageId` custom do cliente e/ou `queueId`).
- Verificar necessidade de DLQ e reprocessamento para fila de envio (talvez replicar padrao de `event_dlq`).
- Confirmar tolerancia a tempo fora (quantos dias uma mensagem pode ficar na fila antes de expirar).

## Proximos passos imediatos
- Validar este plano com stakeholders e priorizar marcos.
- Preparar issues/tarefas no tracker com estimativas.
- Criar branch de preparacao para migracoes da fila com testes iniciais.
- Configurar dashboards Prometheus/Grafana placeholder para novas metricas (`messages_queue_*`).
- Agendar kickoff tecnico para alinhar contratos de payload detalhados (ex. estrutura de botoes, polls).
