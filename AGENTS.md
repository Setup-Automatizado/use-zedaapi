# Repository Guidelines

## Project Structure & Module Organization
Core client logic sits at the module root beside feature files such as `message.go`, `presence.go`, `upload.go`, and `client.go`. Shared protocol definitions live under `proto/`, socket plumbing under `socket/`, and persistence abstractions in `store/` (SQL adapters in `store/sqlstore`). Device state fixtures live in `appstate/`, crypto and logging helpers in `util/`, while API surface docs and compatibility work stay in `api/` (e.g., `api/z_api/WEBHOOKS_EVENTS.md`). Operational assets are under `docker/`, `terraform/`, and `scripts/`. Tests belong next to their targets (see `client_test.go`), and generated artifacts (`proto/*`, `internals.go`) must only be updated via `go generate ./...`.

## Build, Test, and Development Commands
- `go build -v ./...` — compile every package against the supported Go 1.24–1.25 toolchains.
- `go test -v ./...` — execute all tests; scope investigations with `go test ./... -run TestName`.
- `go test -cover ./...` — confirm coverage stability for regressions.
- `go generate ./...` — regenerate internals whenever touching generator inputs like `internals_generate.go`.
- `pre-commit run --all-files` — run the same lint, fmt, and static analysis gates as CI.
- `goimports -local go.mau.fi/whatsmeow -w <file>` — enforce canonical formatting and import grouping.

## Coding Style & Naming Conventions
Always run `gofmt`/`goimports`; tabs are enforced via `.editorconfig`. Exported identifiers stay descriptive (`Client`, `DangerousInternals`) and experimental or risky APIs require clear naming notes. Extend existing option structs and zero values instead of inventing new patterns. Logging must use structured `slog` via `logging.ContextLogger` and `logging.WithAttrs`; `fmt.Println`/`log.Printf` are forbidden. Keep secrets, tokens, and PII out of logs and comments.

## Testing Guidelines
Write table-driven tests in `*_test.go` files colocated with the code under test. Cover observable behaviour, especially WhatsApp-specific edge cases (presence updates, retries, media downloads). Prefer reusable fixtures from `appstate/` over ad-hoc payloads. Maintain coverage with `go test -cover ./...` and add regression cases for any production bug fix.

## Commit & Pull Request Guidelines
Follow lowercase Conventional Commit prefixes (`fix:`, `feat:`, `chore:`) with imperative phrasing (e.g., `fix: handle expired prekeys`). Squash WIP commits before requesting review, reference related issues, and attach logs or screenshots for client-facing changes. Confirm `go test ./...` and `pre-commit run --all-files` pass locally, and note protocol or observability impacts in the PR body.

## Observability & Security Essentials
Propagate `context.Context` everywhere, enriching it with `logging.WithAttrs` (instance IDs, components) before calling downstream services. Update Prometheus metrics at the point of every event, choose labels carefully, and capture only critical failures in Sentry using `sentry.WithScope` with sanitized tags. Never commit real WhatsApp credentials or personal `appstate` snapshots; document redactions in sample payloads.

## JID ↔ Phone Normalization
- Sempre derive o MSISDN para chats e participantes assim que o evento é capturado. Utilize `eventctx.LIDResolver` para popular `metadata["chat_pn"]`, `metadata["sender_pn"]`, `metadata["call_from_pn"]`, `metadata["call_creator_pn"]` e os arrays específicos de grupo (`join_participants_pn`, `leave_participants_pn`, etc.).
- Na camada Z-API, consuma esses campos para preencher `phone`, `participantPhone`, `PresenceChatCallback.participant`, etc., mantendo os campos `chatLid`/`participantLid` originais para retrocompatibilidade.
- Qualquer evento novo que exponha `@lid` deve oferecer o PN correspondente, preservando a regra “prefira phone, mas nunca remova o LID”.
- Eventos de call/presence também devem carregar essas chaves auxiliares para que webhooks privados recebam MSISDNs consistentes.

## Poll Events
As enquetes ainda expõem apenas o `pollMessageId` nas votações. Existe um TODO ativo para descriptografar as opções usando `msgsecret.DecryptPollVote` assim que o pipeline persistir os segredos necessários; mantenha essa limitação documentada nos PRs.

## Authentication Tokens
- `CLIENT_AUTH_TOKEN` é obrigatório, único e carregado via ambiente; o servidor aborta se estiver vazio ou com menos de 16 caracteres. Todos os handlers/serviços devem validar o cabeçalho recebido comparando com `config.Client.AuthToken`.
- Não persista nem retorne tokens de cliente no banco ou em payloads JSON. A coluna `instances.client_token` foi removida; reutilizar migrations antigas é proibido.
- Webhooks devem continuar enviando o header `Client-Token` com o valor do ambiente. Nunca gere tokens por instância nem exponha o segredo em logs/metrics.

## Reference Docs & Matrices
- `RULES.md` / `rules.md` — comprehensive architecture, coding standards, observability checklists, and reviewer matrices.
- `PLAN.md` — roadmap for Z-API compatibility, queue architecture, and milestone sequencing.
- `FIX.md` — incident analysis, remediation plans, file-by-file change matrices, and deployment playbooks.
- `api/z_api/WEBHOOKS_EVENTS.md` — canonical structure for webhook payloads tracked during compatibility work.
- `NOTES.md` — scratchpad for active investigations; treat as working context.
Consult these alongside this guide to ensure every change aligns with project-wide principles, domain rules, and operational expectations.
