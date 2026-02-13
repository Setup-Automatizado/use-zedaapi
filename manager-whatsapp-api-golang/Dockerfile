# ==================================================
# Manager WhatsApp API - Dockerfile
# ==================================================
# Multi-stage build para Next.js 16 com Bun runtime
# Otimizado para AWS ECS Fargate
# ==================================================

# Build Arguments
ARG BUILD_TIME=""
ARG VERSION=""
ARG COMMIT=""
ARG NEXT_PUBLIC_APP_URL="http://localhost:3000"
ARG NEXT_PUBLIC_WHATSAPP_API_URL=""
ARG NEXT_PUBLIC_API_BASE_URL=""

# --------------------------------------------------
# Stage 1: Dependencies
# --------------------------------------------------
FROM oven/bun:1-alpine AS deps

WORKDIR /app

# Copiar arquivos de dependencias (bun.lock para Bun 1.3+)
COPY package.json bun.lock ./
COPY prisma ./prisma/

# Instalar dependencias (incluindo devDependencies para build)
RUN bun install --frozen-lockfile

# Gerar Prisma Client
RUN bun prisma generate

# --------------------------------------------------
# Stage 2: Builder
# --------------------------------------------------
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# Copiar dependencias do stage anterior
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/lib/generated ./lib/generated

# Copiar codigo fonte
COPY . .

# Build arguments para metadata e URLs publicas
ARG BUILD_TIME
ARG VERSION
ARG COMMIT
ARG NEXT_PUBLIC_APP_URL
ARG NEXT_PUBLIC_WHATSAPP_API_URL
ARG NEXT_PUBLIC_API_BASE_URL

# Variaveis de ambiente para build
# NEXT_PUBLIC_* precisa estar definido no build time
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV NEXT_PUBLIC_APP_URL=${NEXT_PUBLIC_APP_URL}
ENV NEXT_PUBLIC_WHATSAPP_API_URL=${NEXT_PUBLIC_WHATSAPP_API_URL}
ENV NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL}

# Executar build do Next.js
RUN bun run build

# --------------------------------------------------
# Stage 3: Production Runner
# --------------------------------------------------
FROM oven/bun:1-alpine AS runner

WORKDIR /app

# Instalar apenas dependencias de runtime necessarias
RUN apk add --no-cache \
    curl \
    tzdata \
    && rm -rf /var/cache/apk/*

# Criar usuario nao-root
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copiar arquivos estaticos e build standalone
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Copiar Prisma para runtime (schema + client gerado em lib/generated/prisma)
COPY --from=builder --chown=nextjs:nodejs /app/prisma ./prisma
COPY --from=builder --chown=nextjs:nodejs /app/lib/generated ./lib/generated
COPY --from=builder --chown=nextjs:nodejs /app/node_modules/@prisma ./node_modules/@prisma

# Metadata
ARG BUILD_TIME
ARG VERSION
ARG COMMIT

LABEL org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.title="Manager WhatsApp API" \
      org.opencontainers.image.description="Frontend Manager para WhatsApp API Golang" \
      org.opencontainers.image.vendor="Setup Automatizado"

# Configurar usuario
USER nextjs

# Expor porta
EXPOSE 3000

# Variaveis de ambiente de runtime
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
ENV NEXT_TELEMETRY_DISABLED=1

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:3000/api/health || exit 1

# Iniciar aplicacao
CMD ["bun", "run", "--bun", "server.js"]

# --------------------------------------------------
# Stage 4: Migration Runner (for ECS one-shot tasks)
# --------------------------------------------------
FROM oven/bun:1-alpine AS migration

WORKDIR /app

# Copiar todas as dependencias (prisma CLI precisa de deps transitivas)
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/prisma ./prisma
COPY --from=deps /app/lib/generated ./lib/generated
COPY --from=deps /app/package.json ./package.json
COPY --from=builder /app/prisma.config.ts ./prisma.config.ts

CMD ["bun", "prisma", "migrate", "deploy"]
