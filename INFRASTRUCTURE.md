# Infrastructure & CI/CD Documentation (English)

> Last updated: 2026-02-13

## Table of Contents

1. [Environments](#1-environments)
2. [AWS Infrastructure](#2-aws-infrastructure)
3. [Docker Build Architecture](#3-docker-build-architecture)
4. [CI/CD Workflows](#4-cicd-workflows)
5. [Semantic Versioning](#5-semantic-versioning)
6. [Version Display](#6-version-display)
7. [Deployment Flow](#7-deployment-flow)
8. [Troubleshooting](#8-troubleshooting)
9. [Quick Reference Commands](#9-quick-reference-commands)

---

## 1. Environments

This project operates with two deployment environments. The `develop` branch is used exclusively for development and does **not** trigger any deployments.

| Environment | Branch     | Purpose                   | Deployment Trigger        |
|-------------|------------|---------------------------|---------------------------|
| homolog     | `homolog`  | Pre-production testing    | Push to `homolog` branch  |
| production  | `production` | Live / stable release   | Push to `production` branch |

---

## 2. AWS Infrastructure

### ECS Clusters

| Cluster Name                    | Environment |
|---------------------------------|-------------|
| `production-whatsmeow-cluster`  | production  |
| `homolog-whatsmeow-cluster`     | homolog     |

### ECS Services (4 total)

| Service Name                     | Component | Environment |
|----------------------------------|-----------|-------------|
| `production-whatsmeow-service`   | API       | production  |
| `production-manager-service`     | Manager   | production  |
| `homolog-whatsmeow-service`      | API       | homolog     |
| `homolog-manager-service`        | Manager   | homolog     |

### ECR Repositories

| Repository Name         | Component |
|--------------------------|-----------|
| `whatsapp-api`           | API (Go backend)     |
| `manager-whatsapp-api`   | Manager (Next.js frontend) |

### Application Load Balancers (ALBs)

There are 4 ALBs in total -- one per service per environment. Each ALB routes traffic to its corresponding ECS service via target groups.

### Image Tagging Strategy

| Environment | Runner Image Tag                              | Migration Image Tag              |
|-------------|-----------------------------------------------|----------------------------------|
| production  | `:latest` (default) or `:v{version}` (when semantic release publishes) | `:latest-migrate` or `:v{version}-migrate` |
| homolog     | `:homolog`                                    | `:homolog-migrate`               |

---

## 3. Docker Build Architecture

### API (`docker/Dockerfile`)

The API Dockerfile contains 5 stages. The **default target** is `production` (Stage 3).

| Stage | Name          | Base Image          | Purpose                                                        |
|-------|---------------|---------------------|----------------------------------------------------------------|
| 1     | `deps`        | `golang:1.25-alpine` | Download and verify Go module dependencies                    |
| 2     | `builder`     | `golang:1.25-alpine` | Compile Go binary (linux/amd64 only, CGO enabled for go-fitz/MuPDF) |
| 3     | `production`  | `alpine:3.21`       | Runtime image with ffmpeg, curl, ca-certificates, non-root user |
| 4     | `development` | `golang:1.25-alpine` | Development image with hot-reload tooling                     |
| 5     | `testing`     | `golang:1.25` (Debian) | Test image with race detector (requires glibc)              |

**Build arguments:**

| Argument     | Description                        |
|--------------|------------------------------------|
| `VERSION`    | Semantic version string            |
| `COMMIT`     | Git commit SHA                     |
| `BUILD_TIME` | UTC build timestamp                |

**Version injection:** The version is compiled into the binary via ldflags:

```
-X go.mau.fi/whatsmeow/api/internal/version.version=${VERSION}
-X go.mau.fi/whatsmeow/api/internal/version.gitCommit=${COMMIT}
-X go.mau.fi/whatsmeow/api/internal/version.buildTime=${BUILD_TIME}
```

The `VERSION` file from the repository root is also copied to `/app/VERSION` inside the container as a runtime fallback.

### Manager (`manager-whatsapp-api-golang/Dockerfile`)

The Manager Dockerfile contains 4 stages. **CRITICAL: The last stage is `migration`, not `runner`.**

| Stage | Name        | Base Image           | Purpose                                              |
|-------|-------------|----------------------|------------------------------------------------------|
| 1     | `deps`      | `oven/bun:1-alpine`  | Install npm dependencies with Bun, generate Prisma client |
| 2     | `builder`   | `oven/bun:1-alpine`  | Build Next.js application                            |
| 3     | `runner`    | `oven/bun:1-alpine`  | Production runner (standalone Next.js with Bun runtime) |
| 4     | `migration` | `oven/bun:1-alpine`  | One-shot Prisma migration runner                     |

**Build arguments:**

| Argument                        | Description                                  |
|---------------------------------|----------------------------------------------|
| `VERSION`                       | Semantic version string                      |
| `COMMIT`                        | Git commit SHA                               |
| `BUILD_TIME`                    | UTC build timestamp                          |
| `NEXT_PUBLIC_APP_URL`           | Public URL of the Manager application        |
| `NEXT_PUBLIC_WHATSAPP_API_URL`  | Public URL of the WhatsApp API               |
| `NEXT_PUBLIC_API_BASE_URL`      | Base URL for API calls                       |

> **WARNING:** The Manager Dockerfile's last stage is `migration`, NOT `runner`. Without specifying `--target runner`, Docker will build the `migration` stage by default. The migration stage only runs `bun prisma migrate deploy` and exits immediately. This will cause ECS containers to enter a crash loop. Always build the runner image with `--target runner`.

---

## 4. CI/CD Workflows

### ci.yml -- Continuous Integration

**Trigger:** Push to all branches EXCEPT `production`, `homolog`, and `main`. All pull requests.

**Concurrency:** Cancels in-progress runs for the same branch.

| Job            | Description                                                   |
|----------------|---------------------------------------------------------------|
| `lint`         | Code formatting (gofmt, goimports), go vet, pre-commit hooks |
| `build`        | Compile Go library and API binary (Go 1.24 + 1.25 matrix)    |
| `test`         | Run tests with race detector, coverage upload (PostgreSQL + Redis services) |
| `security`     | Gosec security scanner                                        |
| `docker-build` | Docker build test (no push)                                   |

This workflow does **not** deploy anything.

### cd.yml -- Release & Deploy

**Trigger:** Push to `production` or `homolog`.

**Concurrency:** Does not cancel in-progress runs (sequential deployments).

| Job              | Depends On                         | Description                                                    |
|------------------|------------------------------------|----------------------------------------------------------------|
| `release`        | --                                 | Semantic Release: analyzes commits, creates version tags, updates VERSION file and CHANGELOG |
| `docker-api`     | `release`                          | Build and push API image to ECR                                |
| `docker-manager` | `release`                          | Build and push Manager runner image AND migration image to ECR |
| `deploy`         | `release`, `docker-api`, `docker-manager` | Update ECS services with force-new-deployment            |
| `health-check`   | `deploy`                           | Post-deploy health verification (up to 5 retries, 10s apart)  |
| `notify`         | all jobs                           | Generate deployment summary in GitHub Actions                  |

### manager-deploy.yml -- Manager-Specific Deploy

**Trigger:** Push to `production` or `homolog` when files in `manager-whatsapp-api-golang/` change. Also supports `workflow_dispatch` with environment selection.

| Step                   | Description                                                |
|------------------------|------------------------------------------------------------|
| Build and Test         | Install deps, Prisma generate, type check, lint, build     |
| Build runner image     | Docker build with `--target runner`, push to ECR           |
| Build migration image  | Docker build with `--target migration`, push to ECR        |
| Run database migrations| Run ECS one-shot Fargate task with the migration image     |
| Deploy ECS service     | Force new deployment, wait for stabilization               |

---

## 5. Semantic Versioning

Configuration is defined in `.releaserc.json`.

### Branch Configuration

| Branch       | Channel    | Release Type       |
|--------------|------------|---------------------|
| `production` | default    | Stable releases     |
| `homolog`    | `homolog`  | Pre-releases (`x.y.z-homolog.n`) |

### Commit Types and Release Impact

| Commit Type | Triggers Release | Version Bump |
|-------------|------------------|--------------|
| `feat`      | Yes              | minor        |
| `fix`       | Yes              | patch        |
| `perf`      | Yes              | patch        |
| `refactor`  | Yes              | patch        |
| `revert`    | Yes              | patch        |
| Breaking change (any type with `BREAKING CHANGE`) | Yes | major |
| `docs`      | No               | --           |
| `style`     | No               | --           |
| `test`      | No               | --           |
| `chore`     | No               | --           |
| `build`     | No               | --           |
| `ci`        | No               | --           |

### Plugins (execution order)

1. `@semantic-release/commit-analyzer` -- Determines release type from conventional commits
2. `@semantic-release/release-notes-generator` -- Generates release notes
3. `@semantic-release/changelog` -- Updates `CHANGELOG.md`
4. `@semantic-release/exec` -- Writes version to `VERSION` file, updates `manager-whatsapp-api-golang/package.json`
5. `@semantic-release/git` -- Commits `CHANGELOG.md`, `VERSION`, and `package.json` with `[skip ci]`
6. `@semantic-release/github` -- Creates GitHub release, adds comments to related issues/PRs

### Tag Format

```
v${version}
```

Examples: `v2.0.0`, `v2.1.0-homolog.3`

---

## 6. Version Display

### API `/health` Endpoint

The `/health` endpoint returns a JSON response that includes the `version` field. The version is resolved from `api/internal/version/version.go` with the following priority:

1. **ldflags** (compile-time) -- Injected during Docker build via `-X` flags
2. **VERSION file** (runtime fallback) -- Read from `/app/VERSION` or relative paths
3. **"unknown"** -- Default when neither source is available

### API OpenAPI Documentation (`/docs`)

The Swagger UI at `/docs` dynamically injects the version into the OpenAPI spec. The file `api/docs/http.go` calls `version.String()` to populate the `info.version` field in the spec before serving it.

### Manager Sidebar

The Manager frontend displays the API version in its sidebar via the component `components/layout/api-version.tsx`. This component calls the API `/health` endpoint every 60 seconds and renders the version string. When the sidebar is collapsed, it shows only the major version (`vX.Y.Z`); when expanded, it shows the full version with the `API v` prefix.

---

## 7. Deployment Flow

Step-by-step deployment process:

```
1. Developer merges PR into `production` or `homolog`
       |
2. cd.yml workflow triggers
       |
3. Semantic Release analyzes commits
   - If release-worthy commits exist: creates tag, updates VERSION, CHANGELOG
   - If not: proceeds without new version
       |
4. Docker images built in parallel:
   - API image: docker/Dockerfile (default target = production)
   - Manager runner image: manager-whatsapp-api-golang/Dockerfile --target runner
   - Manager migration image: manager-whatsapp-api-golang/Dockerfile --target migration
       |
5. Version injected via build-args (VERSION, COMMIT, BUILD_TIME)
       |
6. Images pushed to ECR with appropriate tags:
   - production: :latest (+ :v{version} if new release)
   - homolog: :homolog
   - migration: :{tag}-migrate
       |
7. ECS services updated with --force-new-deployment
       |
8. Wait for services to stabilize (aws ecs wait services-stable)
       |
9. Health check verifies API responds on /health (5 retries, 10s interval)
       |
10. Deployment summary generated in GitHub Actions
```

---

## 8. Troubleshooting

### Manager ECS Containers Crashing (Immediate Exit)

**Cause:** The Manager Docker image was built without `--target runner`. Since the `migration` stage is the last stage in the Dockerfile, Docker defaults to building it. The migration container executes `bun prisma migrate deploy` and exits immediately, causing ECS to enter a crash loop.

**Solution:** Ensure all Docker build commands for the Manager runner include `--target runner`:

```bash
docker build --target runner -t manager-whatsapp-api:latest ./manager-whatsapp-api-golang
```

### API Showing Wrong Version

**Possible causes:**

1. The `VERSION` file in the repository has not been updated by semantic release.
2. The `VERSION` build-arg was not passed during Docker build.
3. The ldflags were not applied correctly during compilation.

**Diagnosis:**

```bash
# Check the VERSION file in the repo
cat VERSION

# Check the running container's version
curl http://<alb-dns>/health | jq '.version'
```

### CI Failing on Lint

**Solution:** Run formatting tools locally before pushing:

```bash
go mod tidy
goimports -local go.mau.fi/whatsmeow -w .
gofmt -w .
```

### Docker Build Timeout on arm64

**Cause:** This project only builds for `linux/amd64`. If arm64 appears in the build, check the `platforms` parameter in the workflow or Docker build command.

**Solution:** Ensure `platforms: linux/amd64` is set in all build steps. Remove any multi-platform configuration.

### ECS Service Not Stabilizing

**Possible causes:**

1. Container health check failing (check `/health` for API, `/api/health` for Manager).
2. Missing environment variables in task definition.
3. Database connectivity issues.

**Diagnosis:**

```bash
# Check service events
aws ecs describe-services --cluster <cluster> --services <service> \
  --query 'services[0].events[:10]'

# Check stopped task reason
aws ecs describe-tasks --cluster <cluster> --tasks <task-arn> \
  --query 'tasks[0].{status:lastStatus,reason:stoppedReason,container:containers[0].reason}'
```

---

## 9. Quick Reference Commands

```bash
# ------------------------------------------------
# ECS Service Status
# ------------------------------------------------

# Check services in a cluster
aws ecs describe-services \
  --cluster production-whatsmeow-cluster \
  --services production-whatsmeow-service production-manager-service \
  --query 'services[*].{name:serviceName,status:status,desired:desiredCount,running:runningCount,pending:pendingCount}'

# List running tasks
aws ecs list-tasks --cluster production-whatsmeow-cluster

# Describe a specific task
aws ecs describe-tasks \
  --cluster production-whatsmeow-cluster \
  --tasks <task-arn>

# ------------------------------------------------
# Force Redeployment
# ------------------------------------------------

# Force redeploy API
aws ecs update-service \
  --cluster production-whatsmeow-cluster \
  --service production-whatsmeow-service \
  --force-new-deployment

# Force redeploy Manager
aws ecs update-service \
  --cluster production-whatsmeow-cluster \
  --service production-manager-service \
  --force-new-deployment

# ------------------------------------------------
# ECR Images
# ------------------------------------------------

# List recent API images
aws ecr describe-images \
  --repository-name whatsapp-api \
  --query 'sort_by(imageDetails,&imagePushedAt)[-5:].[imageTags,imagePushedAt]'

# List recent Manager images
aws ecr describe-images \
  --repository-name manager-whatsapp-api \
  --query 'sort_by(imageDetails,&imagePushedAt)[-5:].[imageTags,imagePushedAt]'

# ------------------------------------------------
# Logs
# ------------------------------------------------

# View ECS logs (replace <stream> with actual log stream name)
aws logs get-log-events \
  --log-group-name /ecs/production-whatsmeow \
  --log-stream-name <stream>

# List recent log streams
aws logs describe-log-streams \
  --log-group-name /ecs/production-whatsmeow \
  --order-by LastEventTime \
  --descending \
  --limit 5

# ------------------------------------------------
# Health Check
# ------------------------------------------------

# Check API health
curl -s http://<alb-dns>/health | jq .

# Check Manager health
curl -s http://<manager-alb-dns>/api/health
```

---
---

# Documentacao de Infraestrutura & CI/CD (Portugues)

> Ultima atualizacao: 2026-02-13

## Indice

1. [Ambientes](#1-ambientes)
2. [Infraestrutura AWS](#2-infraestrutura-aws)
3. [Arquitetura de Build Docker](#3-arquitetura-de-build-docker)
4. [Workflows CI/CD](#4-workflows-cicd)
5. [Versionamento Semantico](#5-versionamento-semantico)
6. [Exibicao de Versao](#6-exibicao-de-versao)
7. [Fluxo de Deploy](#7-fluxo-de-deploy)
8. [Resolucao de Problemas](#8-resolucao-de-problemas)
9. [Comandos de Referencia Rapida](#9-comandos-de-referencia-rapida)

---

## 1. Ambientes

Este projeto opera com dois ambientes de deploy. A branch `develop` e usada exclusivamente para desenvolvimento e **nao** dispara nenhum deploy.

| Ambiente    | Branch       | Finalidade                 | Gatilho de Deploy             |
|-------------|--------------|----------------------------|-------------------------------|
| homolog     | `homolog`    | Testes de pre-producao     | Push na branch `homolog`      |
| production  | `production` | Producao / release estavel | Push na branch `production`   |

---

## 2. Infraestrutura AWS

### Clusters ECS

| Nome do Cluster                 | Ambiente    |
|---------------------------------|-------------|
| `production-whatsmeow-cluster`  | production  |
| `homolog-whatsmeow-cluster`     | homolog     |

### Servicos ECS (4 no total)

| Nome do Servico                  | Componente | Ambiente    |
|----------------------------------|------------|-------------|
| `production-whatsmeow-service`   | API        | production  |
| `production-manager-service`     | Manager    | production  |
| `homolog-whatsmeow-service`      | API        | homolog     |
| `homolog-manager-service`        | Manager    | homolog     |

### Repositorios ECR

| Nome do Repositorio    | Componente                    |
|------------------------|-------------------------------|
| `whatsapp-api`         | API (backend Go)              |
| `manager-whatsapp-api` | Manager (frontend Next.js)    |

### Application Load Balancers (ALBs)

Existem 4 ALBs no total -- um por servico por ambiente. Cada ALB roteia o trafego para seu servico ECS correspondente via target groups.

### Estrategia de Tags das Imagens

| Ambiente    | Tag da Imagem Runner                          | Tag da Imagem de Migracao    |
|-------------|-----------------------------------------------|------------------------------|
| production  | `:latest` (padrao) ou `:v{version}` (quando semantic release publica) | `:latest-migrate` ou `:v{version}-migrate` |
| homolog     | `:homolog`                                    | `:homolog-migrate`           |

---

## 3. Arquitetura de Build Docker

### API (`docker/Dockerfile`)

O Dockerfile da API contem 5 estagios. O **target padrao** e `production` (Estagio 3).

| Estagio | Nome          | Imagem Base           | Finalidade                                                      |
|---------|---------------|-----------------------|-----------------------------------------------------------------|
| 1       | `deps`        | `golang:1.25-alpine`  | Download e verificacao de dependencias Go                       |
| 2       | `builder`     | `golang:1.25-alpine`  | Compilacao do binario Go (linux/amd64, CGO habilitado para go-fitz/MuPDF) |
| 3       | `production`  | `alpine:3.21`         | Imagem de runtime com ffmpeg, curl, ca-certificates, usuario nao-root |
| 4       | `development` | `golang:1.25-alpine`  | Imagem de desenvolvimento com ferramentas de hot-reload         |
| 5       | `testing`     | `golang:1.25` (Debian) | Imagem de teste com race detector (requer glibc)               |

**Argumentos de build:**

| Argumento    | Descricao                          |
|--------------|------------------------------------|
| `VERSION`    | String de versao semantica         |
| `COMMIT`     | SHA do commit Git                  |
| `BUILD_TIME` | Timestamp UTC do build             |

**Injecao de versao:** A versao e compilada no binario via ldflags:

```
-X go.mau.fi/whatsmeow/api/internal/version.version=${VERSION}
-X go.mau.fi/whatsmeow/api/internal/version.gitCommit=${COMMIT}
-X go.mau.fi/whatsmeow/api/internal/version.buildTime=${BUILD_TIME}
```

O arquivo `VERSION` da raiz do repositorio tambem e copiado para `/app/VERSION` dentro do container como fallback em runtime.

### Manager (`manager-whatsapp-api-golang/Dockerfile`)

O Dockerfile do Manager contem 4 estagios. **CRITICO: O ultimo estagio e `migration`, nao `runner`.**

| Estagio | Nome        | Imagem Base          | Finalidade                                                |
|---------|-------------|----------------------|-----------------------------------------------------------|
| 1       | `deps`      | `oven/bun:1-alpine`  | Instalacao de dependencias npm com Bun, geracao do Prisma client |
| 2       | `builder`   | `oven/bun:1-alpine`  | Build da aplicacao Next.js                                |
| 3       | `runner`    | `oven/bun:1-alpine`  | Runner de producao (Next.js standalone com runtime Bun)   |
| 4       | `migration` | `oven/bun:1-alpine`  | Runner de migracao one-shot com Prisma                    |

**Argumentos de build:**

| Argumento                       | Descricao                                    |
|---------------------------------|----------------------------------------------|
| `VERSION`                       | String de versao semantica                   |
| `COMMIT`                        | SHA do commit Git                            |
| `BUILD_TIME`                    | Timestamp UTC do build                       |
| `NEXT_PUBLIC_APP_URL`           | URL publica da aplicacao Manager             |
| `NEXT_PUBLIC_WHATSAPP_API_URL`  | URL publica da API WhatsApp                  |
| `NEXT_PUBLIC_API_BASE_URL`      | URL base para chamadas de API                |

> **AVISO:** O ultimo estagio do Dockerfile do Manager e `migration`, NAO `runner`. Sem especificar `--target runner`, o Docker construira o estagio `migration` por padrao. O container de migracao apenas executa `bun prisma migrate deploy` e encerra imediatamente. Isso fara com que os containers ECS entrem em loop de crash. Sempre construa a imagem runner com `--target runner`.

---

## 4. Workflows CI/CD

### ci.yml -- Integracao Continua

**Gatilho:** Push em todas as branches EXCETO `production`, `homolog` e `main`. Todos os pull requests.

**Concorrencia:** Cancela execucoes em andamento para a mesma branch.

| Job            | Descricao                                                      |
|----------------|----------------------------------------------------------------|
| `lint`         | Formatacao de codigo (gofmt, goimports), go vet, pre-commit hooks |
| `build`        | Compilacao da biblioteca Go e binario da API (matriz Go 1.24 + 1.25) |
| `test`         | Execucao de testes com race detector, upload de cobertura (servicos PostgreSQL + Redis) |
| `security`     | Scanner de seguranca Gosec                                     |
| `docker-build` | Teste de build Docker (sem push)                               |

Este workflow **nao** realiza nenhum deploy.

### cd.yml -- Release & Deploy

**Gatilho:** Push em `production` ou `homolog`.

**Concorrencia:** Nao cancela execucoes em andamento (deploys sequenciais).

| Job              | Depende De                         | Descricao                                                      |
|------------------|------------------------------------|----------------------------------------------------------------|
| `release`        | --                                 | Semantic Release: analisa commits, cria tags de versao, atualiza arquivo VERSION e CHANGELOG |
| `docker-api`     | `release`                          | Build e push da imagem da API para ECR                         |
| `docker-manager` | `release`                          | Build e push da imagem runner e imagem de migracao do Manager para ECR |
| `deploy`         | `release`, `docker-api`, `docker-manager` | Atualizacao dos servicos ECS com force-new-deployment    |
| `health-check`   | `deploy`                           | Verificacao de saude pos-deploy (ate 5 tentativas, 10s de intervalo) |
| `notify`         | todos os jobs                      | Geracao de resumo do deploy no GitHub Actions                  |

### manager-deploy.yml -- Deploy Especifico do Manager

**Gatilho:** Push em `production` ou `homolog` quando arquivos em `manager-whatsapp-api-golang/` sao alterados. Tambem suporta `workflow_dispatch` com selecao de ambiente.

| Etapa                   | Descricao                                                  |
|-------------------------|-------------------------------------------------------------|
| Build e Teste           | Instalacao de deps, Prisma generate, type check, lint, build |
| Build da imagem runner  | Docker build com `--target runner`, push para ECR           |
| Build da imagem migrate | Docker build com `--target migration`, push para ECR        |
| Execucao de migracoes   | Execucao de task one-shot Fargate no ECS com imagem de migracao |
| Deploy do servico ECS   | Force new deployment, aguarda estabilizacao                 |

---

## 5. Versionamento Semantico

A configuracao esta definida em `.releaserc.json`.

### Configuracao de Branches

| Branch       | Canal      | Tipo de Release      |
|--------------|------------|----------------------|
| `production` | padrao     | Releases estaveis    |
| `homolog`    | `homolog`  | Pre-releases (`x.y.z-homolog.n`) |

### Tipos de Commit e Impacto na Release

| Tipo de Commit | Gera Release | Bump de Versao |
|----------------|--------------|----------------|
| `feat`         | Sim          | minor          |
| `fix`          | Sim          | patch          |
| `perf`         | Sim          | patch          |
| `refactor`     | Sim          | patch          |
| `revert`       | Sim          | patch          |
| Breaking change (qualquer tipo com `BREAKING CHANGE`) | Sim | major |
| `docs`         | Nao          | --             |
| `style`        | Nao          | --             |
| `test`         | Nao          | --             |
| `chore`        | Nao          | --             |
| `build`        | Nao          | --             |
| `ci`           | Nao          | --             |

### Plugins (ordem de execucao)

1. `@semantic-release/commit-analyzer` -- Determina o tipo de release a partir dos commits convencionais
2. `@semantic-release/release-notes-generator` -- Gera notas de release
3. `@semantic-release/changelog` -- Atualiza o `CHANGELOG.md`
4. `@semantic-release/exec` -- Escreve a versao no arquivo `VERSION`, atualiza `manager-whatsapp-api-golang/package.json`
5. `@semantic-release/git` -- Faz commit de `CHANGELOG.md`, `VERSION` e `package.json` com `[skip ci]`
6. `@semantic-release/github` -- Cria release no GitHub, adiciona comentarios em issues/PRs relacionados

### Formato de Tag

```
v${version}
```

Exemplos: `v2.0.0`, `v2.1.0-homolog.3`

---

## 6. Exibicao de Versao

### Endpoint `/health` da API

O endpoint `/health` retorna uma resposta JSON que inclui o campo `version`. A versao e resolvida a partir de `api/internal/version/version.go` com a seguinte prioridade:

1. **ldflags** (tempo de compilacao) -- Injetada durante o build Docker via flags `-X`
2. **Arquivo VERSION** (fallback em runtime) -- Lido de `/app/VERSION` ou caminhos relativos
3. **"unknown"** -- Padrao quando nenhuma fonte esta disponivel

### Documentacao OpenAPI da API (`/docs`)

A interface Swagger UI em `/docs` injeta a versao dinamicamente na spec OpenAPI. O arquivo `api/docs/http.go` chama `version.String()` para popular o campo `info.version` na spec antes de servi-la.

### Barra Lateral do Manager

O frontend do Manager exibe a versao da API na barra lateral atraves do componente `components/layout/api-version.tsx`. Este componente chama o endpoint `/health` da API a cada 60 segundos e renderiza a string de versao. Quando a barra lateral esta recolhida, mostra apenas a versao principal (`vX.Y.Z`); quando expandida, mostra a versao completa com o prefixo `API v`.

---

## 7. Fluxo de Deploy

Processo de deploy passo a passo:

```
1. Desenvolvedor faz merge do PR em `production` ou `homolog`
       |
2. Workflow cd.yml e disparado
       |
3. Semantic Release analisa os commits
   - Se existem commits que geram release: cria tag, atualiza VERSION, CHANGELOG
   - Se nao: prossegue sem nova versao
       |
4. Imagens Docker construidas em paralelo:
   - Imagem da API: docker/Dockerfile (target padrao = production)
   - Imagem runner do Manager: manager-whatsapp-api-golang/Dockerfile --target runner
   - Imagem de migracao do Manager: manager-whatsapp-api-golang/Dockerfile --target migration
       |
5. Versao injetada via build-args (VERSION, COMMIT, BUILD_TIME)
       |
6. Imagens enviadas para ECR com as tags apropriadas:
   - production: :latest (+ :v{version} se nova release)
   - homolog: :homolog
   - migracao: :{tag}-migrate
       |
7. Servicos ECS atualizados com --force-new-deployment
       |
8. Aguarda estabilizacao dos servicos (aws ecs wait services-stable)
       |
9. Health check verifica se a API responde em /health (5 tentativas, intervalo de 10s)
       |
10. Resumo do deploy gerado no GitHub Actions
```

---

## 8. Resolucao de Problemas

### Containers ECS do Manager em Crash (Saida Imediata)

**Causa:** A imagem Docker do Manager foi construida sem `--target runner`. Como o estagio `migration` e o ultimo no Dockerfile, o Docker o constroi por padrao. O container de migracao executa `bun prisma migrate deploy` e encerra imediatamente, fazendo com que o ECS entre em loop de crash.

**Solucao:** Garanta que todos os comandos de build Docker para o runner do Manager incluam `--target runner`:

```bash
docker build --target runner -t manager-whatsapp-api:latest ./manager-whatsapp-api-golang
```

### API Mostrando Versao Incorreta

**Causas possiveis:**

1. O arquivo `VERSION` no repositorio nao foi atualizado pelo semantic release.
2. O build-arg `VERSION` nao foi passado durante o build Docker.
3. As ldflags nao foram aplicadas corretamente durante a compilacao.

**Diagnostico:**

```bash
# Verificar o arquivo VERSION no repositorio
cat VERSION

# Verificar a versao do container em execucao
curl http://<alb-dns>/health | jq '.version'
```

### CI Falhando no Lint

**Solucao:** Execute as ferramentas de formatacao localmente antes do push:

```bash
go mod tidy
goimports -local go.mau.fi/whatsmeow -w .
gofmt -w .
```

### Timeout no Build Docker em arm64

**Causa:** Este projeto so faz build para `linux/amd64`. Se arm64 aparecer no build, verifique o parametro `platforms` no workflow ou no comando de build Docker.

**Solucao:** Garanta que `platforms: linux/amd64` esteja definido em todos os steps de build. Remova qualquer configuracao multi-plataforma.

### Servico ECS Nao Estabilizando

**Causas possiveis:**

1. Health check do container falhando (verifique `/health` para API, `/api/health` para Manager).
2. Variaveis de ambiente ausentes na task definition.
3. Problemas de conectividade com o banco de dados.

**Diagnostico:**

```bash
# Verificar eventos do servico
aws ecs describe-services --cluster <cluster> --services <service> \
  --query 'services[0].events[:10]'

# Verificar motivo de parada da task
aws ecs describe-tasks --cluster <cluster> --tasks <task-arn> \
  --query 'tasks[0].{status:lastStatus,reason:stoppedReason,container:containers[0].reason}'
```

---

## 9. Comandos de Referencia Rapida

```bash
# ------------------------------------------------
# Status dos Servicos ECS
# ------------------------------------------------

# Verificar servicos em um cluster
aws ecs describe-services \
  --cluster production-whatsmeow-cluster \
  --services production-whatsmeow-service production-manager-service \
  --query 'services[*].{name:serviceName,status:status,desired:desiredCount,running:runningCount,pending:pendingCount}'

# Listar tasks em execucao
aws ecs list-tasks --cluster production-whatsmeow-cluster

# Descrever uma task especifica
aws ecs describe-tasks \
  --cluster production-whatsmeow-cluster \
  --tasks <task-arn>

# ------------------------------------------------
# Forcar Redeploy
# ------------------------------------------------

# Forcar redeploy da API
aws ecs update-service \
  --cluster production-whatsmeow-cluster \
  --service production-whatsmeow-service \
  --force-new-deployment

# Forcar redeploy do Manager
aws ecs update-service \
  --cluster production-whatsmeow-cluster \
  --service production-manager-service \
  --force-new-deployment

# ------------------------------------------------
# Imagens ECR
# ------------------------------------------------

# Listar imagens recentes da API
aws ecr describe-images \
  --repository-name whatsapp-api \
  --query 'sort_by(imageDetails,&imagePushedAt)[-5:].[imageTags,imagePushedAt]'

# Listar imagens recentes do Manager
aws ecr describe-images \
  --repository-name manager-whatsapp-api \
  --query 'sort_by(imageDetails,&imagePushedAt)[-5:].[imageTags,imagePushedAt]'

# ------------------------------------------------
# Logs
# ------------------------------------------------

# Visualizar logs do ECS (substitua <stream> pelo nome real do log stream)
aws logs get-log-events \
  --log-group-name /ecs/production-whatsmeow \
  --log-stream-name <stream>

# Listar log streams recentes
aws logs describe-log-streams \
  --log-group-name /ecs/production-whatsmeow \
  --order-by LastEventTime \
  --descending \
  --limit 5

# ------------------------------------------------
# Health Check
# ------------------------------------------------

# Verificar saude da API
curl -s http://<alb-dns>/health | jq .

# Verificar saude do Manager
curl -s http://<manager-alb-dns>/api/health
```

---
---

# Documentacion de Infraestructura & CI/CD (Espanol)

> Ultima actualizacion: 2026-02-13

## Indice

1. [Entornos](#1-entornos)
2. [Infraestructura AWS](#2-infraestructura-aws-1)
3. [Arquitectura de Build Docker](#3-arquitectura-de-build-docker)
4. [Workflows CI/CD](#4-workflows-cicd-1)
5. [Versionado Semantico](#5-versionado-semantico)
6. [Visualizacion de Version](#6-visualizacion-de-version)
7. [Flujo de Deploy](#7-flujo-de-deploy)
8. [Resolucion de Problemas](#8-resolucion-de-problemas-1)
9. [Comandos de Referencia Rapida](#9-comandos-de-referencia-rapida-1)

---

## 1. Entornos

Este proyecto opera con dos entornos de despliegue. La rama `develop` se utiliza exclusivamente para desarrollo y **no** dispara ningun despliegue.

| Entorno     | Rama         | Proposito                   | Disparador de Deploy          |
|-------------|--------------|------------------------------|-------------------------------|
| homolog     | `homolog`    | Pruebas de pre-produccion    | Push a la rama `homolog`      |
| production  | `production` | Produccion / release estable | Push a la rama `production`   |

---

## 2. Infraestructura AWS

### Clusters ECS

| Nombre del Cluster               | Entorno     |
|----------------------------------|-------------|
| `production-whatsmeow-cluster`   | production  |
| `homolog-whatsmeow-cluster`      | homolog     |

### Servicios ECS (4 en total)

| Nombre del Servicio              | Componente | Entorno     |
|----------------------------------|------------|-------------|
| `production-whatsmeow-service`   | API        | production  |
| `production-manager-service`     | Manager    | production  |
| `homolog-whatsmeow-service`      | API        | homolog     |
| `homolog-manager-service`        | Manager    | homolog     |

### Repositorios ECR

| Nombre del Repositorio  | Componente                     |
|--------------------------|--------------------------------|
| `whatsapp-api`           | API (backend Go)               |
| `manager-whatsapp-api`   | Manager (frontend Next.js)     |

### Application Load Balancers (ALBs)

Hay 4 ALBs en total -- uno por servicio por entorno. Cada ALB enruta el trafico hacia su servicio ECS correspondiente a traves de target groups.

### Estrategia de Tags de las Imagenes

| Entorno     | Tag de la Imagen Runner                       | Tag de la Imagen de Migracion |
|-------------|-----------------------------------------------|-------------------------------|
| production  | `:latest` (por defecto) o `:v{version}` (cuando semantic release publica) | `:latest-migrate` o `:v{version}-migrate` |
| homolog     | `:homolog`                                    | `:homolog-migrate`            |

---

## 3. Arquitectura de Build Docker

### API (`docker/Dockerfile`)

El Dockerfile de la API contiene 5 etapas. El **target por defecto** es `production` (Etapa 3).

| Etapa | Nombre        | Imagen Base           | Proposito                                                        |
|-------|---------------|-----------------------|------------------------------------------------------------------|
| 1     | `deps`        | `golang:1.25-alpine`  | Descarga y verificacion de dependencias Go                       |
| 2     | `builder`     | `golang:1.25-alpine`  | Compilacion del binario Go (linux/amd64, CGO habilitado para go-fitz/MuPDF) |
| 3     | `production`  | `alpine:3.21`         | Imagen de runtime con ffmpeg, curl, ca-certificates, usuario no-root |
| 4     | `development` | `golang:1.25-alpine`  | Imagen de desarrollo con herramientas de hot-reload              |
| 5     | `testing`     | `golang:1.25` (Debian) | Imagen de pruebas con race detector (requiere glibc)            |

**Argumentos de build:**

| Argumento    | Descripcion                         |
|--------------|-------------------------------------|
| `VERSION`    | Cadena de version semantica         |
| `COMMIT`     | SHA del commit Git                  |
| `BUILD_TIME` | Marca temporal UTC del build        |

**Inyeccion de version:** La version se compila en el binario mediante ldflags:

```
-X go.mau.fi/whatsmeow/api/internal/version.version=${VERSION}
-X go.mau.fi/whatsmeow/api/internal/version.gitCommit=${COMMIT}
-X go.mau.fi/whatsmeow/api/internal/version.buildTime=${BUILD_TIME}
```

El archivo `VERSION` de la raiz del repositorio tambien se copia a `/app/VERSION` dentro del contenedor como fallback en tiempo de ejecucion.

### Manager (`manager-whatsapp-api-golang/Dockerfile`)

El Dockerfile del Manager contiene 4 etapas. **CRITICO: La ultima etapa es `migration`, no `runner`.**

| Etapa | Nombre      | Imagen Base          | Proposito                                                  |
|-------|-------------|----------------------|------------------------------------------------------------|
| 1     | `deps`      | `oven/bun:1-alpine`  | Instalacion de dependencias npm con Bun, generacion del Prisma client |
| 2     | `builder`   | `oven/bun:1-alpine`  | Build de la aplicacion Next.js                             |
| 3     | `runner`    | `oven/bun:1-alpine`  | Runner de produccion (Next.js standalone con runtime Bun)  |
| 4     | `migration` | `oven/bun:1-alpine`  | Runner de migracion one-shot con Prisma                    |

**Argumentos de build:**

| Argumento                       | Descripcion                                   |
|---------------------------------|-----------------------------------------------|
| `VERSION`                       | Cadena de version semantica                   |
| `COMMIT`                        | SHA del commit Git                            |
| `BUILD_TIME`                    | Marca temporal UTC del build                  |
| `NEXT_PUBLIC_APP_URL`           | URL publica de la aplicacion Manager          |
| `NEXT_PUBLIC_WHATSAPP_API_URL`  | URL publica de la API WhatsApp                |
| `NEXT_PUBLIC_API_BASE_URL`      | URL base para llamadas a la API               |

> **ADVERTENCIA:** La ultima etapa del Dockerfile del Manager es `migration`, NO `runner`. Sin especificar `--target runner`, Docker construira la etapa `migration` por defecto. El contenedor de migracion solo ejecuta `bun prisma migrate deploy` y termina inmediatamente. Esto hara que los contenedores ECS entren en un ciclo de crash. Siempre construya la imagen runner con `--target runner`.

---

## 4. Workflows CI/CD

### ci.yml -- Integracion Continua

**Disparador:** Push en todas las ramas EXCEPTO `production`, `homolog` y `main`. Todos los pull requests.

**Concurrencia:** Cancela ejecuciones en curso para la misma rama.

| Job            | Descripcion                                                     |
|----------------|-----------------------------------------------------------------|
| `lint`         | Formateo de codigo (gofmt, goimports), go vet, pre-commit hooks |
| `build`        | Compilacion de la biblioteca Go y binario de la API (matriz Go 1.24 + 1.25) |
| `test`         | Ejecucion de pruebas con race detector, carga de cobertura (servicios PostgreSQL + Redis) |
| `security`     | Escaner de seguridad Gosec                                      |
| `docker-build` | Prueba de build Docker (sin push)                               |

Este workflow **no** realiza ningun despliegue.

### cd.yml -- Release & Deploy

**Disparador:** Push en `production` u `homolog`.

**Concurrencia:** No cancela ejecuciones en curso (despliegues secuenciales).

| Job              | Depende De                         | Descripcion                                                     |
|------------------|------------------------------------|------------------------------------------------------------------|
| `release`        | --                                 | Semantic Release: analiza commits, crea tags de version, actualiza archivo VERSION y CHANGELOG |
| `docker-api`     | `release`                          | Build y push de la imagen de la API a ECR                        |
| `docker-manager` | `release`                          | Build y push de la imagen runner y de migracion del Manager a ECR |
| `deploy`         | `release`, `docker-api`, `docker-manager` | Actualizacion de servicios ECS con force-new-deployment    |
| `health-check`   | `deploy`                           | Verificacion de salud post-deploy (hasta 5 intentos, 10s de intervalo) |
| `notify`         | todos los jobs                     | Generacion de resumen del deploy en GitHub Actions               |

### manager-deploy.yml -- Deploy Especifico del Manager

**Disparador:** Push en `production` u `homolog` cuando cambian archivos en `manager-whatsapp-api-golang/`. Tambien soporta `workflow_dispatch` con seleccion de entorno.

| Paso                    | Descripcion                                                 |
|-------------------------|-------------------------------------------------------------|
| Build y Pruebas         | Instalacion de deps, Prisma generate, type check, lint, build |
| Build de imagen runner  | Docker build con `--target runner`, push a ECR              |
| Build de imagen migrate | Docker build con `--target migration`, push a ECR           |
| Ejecucion de migraciones| Ejecucion de tarea one-shot Fargate en ECS con imagen de migracion |
| Deploy del servicio ECS | Force new deployment, espera de estabilizacion              |

---

## 5. Versionado Semantico

La configuracion esta definida en `.releaserc.json`.

### Configuracion de Ramas

| Rama         | Canal      | Tipo de Release       |
|--------------|------------|------------------------|
| `production` | por defecto | Releases estables     |
| `homolog`    | `homolog`  | Pre-releases (`x.y.z-homolog.n`) |

### Tipos de Commit e Impacto en la Release

| Tipo de Commit | Genera Release | Incremento de Version |
|----------------|----------------|-----------------------|
| `feat`         | Si             | minor                 |
| `fix`          | Si             | patch                 |
| `perf`         | Si             | patch                 |
| `refactor`     | Si             | patch                 |
| `revert`       | Si             | patch                 |
| Breaking change (cualquier tipo con `BREAKING CHANGE`) | Si | major |
| `docs`         | No             | --                    |
| `style`        | No             | --                    |
| `test`         | No             | --                    |
| `chore`        | No             | --                    |
| `build`        | No             | --                    |
| `ci`           | No             | --                    |

### Plugins (orden de ejecucion)

1. `@semantic-release/commit-analyzer` -- Determina el tipo de release a partir de los commits convencionales
2. `@semantic-release/release-notes-generator` -- Genera notas de release
3. `@semantic-release/changelog` -- Actualiza el `CHANGELOG.md`
4. `@semantic-release/exec` -- Escribe la version en el archivo `VERSION`, actualiza `manager-whatsapp-api-golang/package.json`
5. `@semantic-release/git` -- Hace commit de `CHANGELOG.md`, `VERSION` y `package.json` con `[skip ci]`
6. `@semantic-release/github` -- Crea release en GitHub, agrega comentarios en issues/PRs relacionados

### Formato de Tag

```
v${version}
```

Ejemplos: `v2.0.0`, `v2.1.0-homolog.3`

---

## 6. Visualizacion de Version

### Endpoint `/health` de la API

El endpoint `/health` retorna una respuesta JSON que incluye el campo `version`. La version se resuelve desde `api/internal/version/version.go` con la siguiente prioridad:

1. **ldflags** (tiempo de compilacion) -- Inyectada durante el build Docker mediante flags `-X`
2. **Archivo VERSION** (fallback en runtime) -- Leido desde `/app/VERSION` o rutas relativas
3. **"unknown"** -- Valor por defecto cuando ninguna fuente esta disponible

### Documentacion OpenAPI de la API (`/docs`)

La interfaz Swagger UI en `/docs` inyecta la version dinamicamente en la especificacion OpenAPI. El archivo `api/docs/http.go` llama a `version.String()` para poblar el campo `info.version` en la spec antes de servirla.

### Barra Lateral del Manager

El frontend del Manager muestra la version de la API en su barra lateral a traves del componente `components/layout/api-version.tsx`. Este componente llama al endpoint `/health` de la API cada 60 segundos y renderiza la cadena de version. Cuando la barra lateral esta colapsada, muestra solo la version principal (`vX.Y.Z`); cuando esta expandida, muestra la version completa con el prefijo `API v`.

---

## 7. Flujo de Deploy

Proceso de despliegue paso a paso:

```
1. El desarrollador hace merge del PR en `production` u `homolog`
       |
2. El workflow cd.yml se dispara
       |
3. Semantic Release analiza los commits
   - Si existen commits que generan release: crea tag, actualiza VERSION, CHANGELOG
   - Si no: continua sin nueva version
       |
4. Imagenes Docker construidas en paralelo:
   - Imagen de la API: docker/Dockerfile (target por defecto = production)
   - Imagen runner del Manager: manager-whatsapp-api-golang/Dockerfile --target runner
   - Imagen de migracion del Manager: manager-whatsapp-api-golang/Dockerfile --target migration
       |
5. Version inyectada via build-args (VERSION, COMMIT, BUILD_TIME)
       |
6. Imagenes enviadas a ECR con las tags apropiadas:
   - production: :latest (+ :v{version} si hay nueva release)
   - homolog: :homolog
   - migracion: :{tag}-migrate
       |
7. Servicios ECS actualizados con --force-new-deployment
       |
8. Espera de estabilizacion de servicios (aws ecs wait services-stable)
       |
9. Health check verifica que la API responde en /health (5 intentos, intervalo de 10s)
       |
10. Resumen del deploy generado en GitHub Actions
```

---

## 8. Resolucion de Problemas

### Contenedores ECS del Manager en Crash (Salida Inmediata)

**Causa:** La imagen Docker del Manager fue construida sin `--target runner`. Como la etapa `migration` es la ultima en el Dockerfile, Docker la construye por defecto. El contenedor de migracion ejecuta `bun prisma migrate deploy` y termina inmediatamente, haciendo que ECS entre en un ciclo de crash.

**Solucion:** Asegurese de que todos los comandos de build Docker para el runner del Manager incluyan `--target runner`:

```bash
docker build --target runner -t manager-whatsapp-api:latest ./manager-whatsapp-api-golang
```

### API Mostrando Version Incorrecta

**Posibles causas:**

1. El archivo `VERSION` en el repositorio no fue actualizado por semantic release.
2. El build-arg `VERSION` no fue pasado durante el build Docker.
3. Las ldflags no fueron aplicadas correctamente durante la compilacion.

**Diagnostico:**

```bash
# Verificar el archivo VERSION en el repositorio
cat VERSION

# Verificar la version del contenedor en ejecucion
curl http://<alb-dns>/health | jq '.version'
```

### CI Fallando en Lint

**Solucion:** Ejecute las herramientas de formateo localmente antes del push:

```bash
go mod tidy
goimports -local go.mau.fi/whatsmeow -w .
gofmt -w .
```

### Timeout en Build Docker en arm64

**Causa:** Este proyecto solo construye para `linux/amd64`. Si arm64 aparece en el build, verifique el parametro `platforms` en el workflow o en el comando de build Docker.

**Solucion:** Asegurese de que `platforms: linux/amd64` este configurado en todos los pasos de build. Elimine cualquier configuracion multi-plataforma.

### Servicio ECS No Estabilizandose

**Posibles causas:**

1. Health check del contenedor fallando (verifique `/health` para API, `/api/health` para Manager).
2. Variables de entorno ausentes en la task definition.
3. Problemas de conectividad con la base de datos.

**Diagnostico:**

```bash
# Verificar eventos del servicio
aws ecs describe-services --cluster <cluster> --services <service> \
  --query 'services[0].events[:10]'

# Verificar razon de detencion de la tarea
aws ecs describe-tasks --cluster <cluster> --tasks <task-arn> \
  --query 'tasks[0].{status:lastStatus,reason:stoppedReason,container:containers[0].reason}'
```

---

## 9. Comandos de Referencia Rapida

```bash
# ------------------------------------------------
# Estado de los Servicios ECS
# ------------------------------------------------

# Verificar servicios en un cluster
aws ecs describe-services \
  --cluster production-whatsmeow-cluster \
  --services production-whatsmeow-service production-manager-service \
  --query 'services[*].{name:serviceName,status:status,desired:desiredCount,running:runningCount,pending:pendingCount}'

# Listar tareas en ejecucion
aws ecs list-tasks --cluster production-whatsmeow-cluster

# Describir una tarea especifica
aws ecs describe-tasks \
  --cluster production-whatsmeow-cluster \
  --tasks <task-arn>

# ------------------------------------------------
# Forzar Redespliegue
# ------------------------------------------------

# Forzar redespliegue de la API
aws ecs update-service \
  --cluster production-whatsmeow-cluster \
  --service production-whatsmeow-service \
  --force-new-deployment

# Forzar redespliegue del Manager
aws ecs update-service \
  --cluster production-whatsmeow-cluster \
  --service production-manager-service \
  --force-new-deployment

# ------------------------------------------------
# Imagenes ECR
# ------------------------------------------------

# Listar imagenes recientes de la API
aws ecr describe-images \
  --repository-name whatsapp-api \
  --query 'sort_by(imageDetails,&imagePushedAt)[-5:].[imageTags,imagePushedAt]'

# Listar imagenes recientes del Manager
aws ecr describe-images \
  --repository-name manager-whatsapp-api \
  --query 'sort_by(imageDetails,&imagePushedAt)[-5:].[imageTags,imagePushedAt]'

# ------------------------------------------------
# Logs
# ------------------------------------------------

# Ver logs de ECS (reemplace <stream> con el nombre real del log stream)
aws logs get-log-events \
  --log-group-name /ecs/production-whatsmeow \
  --log-stream-name <stream>

# Listar log streams recientes
aws logs describe-log-streams \
  --log-group-name /ecs/production-whatsmeow \
  --order-by LastEventTime \
  --descending \
  --limit 5

# ------------------------------------------------
# Health Check
# ------------------------------------------------

# Verificar salud de la API
curl -s http://<alb-dns>/health | jq .

# Verificar salud del Manager
curl -s http://<manager-alb-dns>/api/health
```
