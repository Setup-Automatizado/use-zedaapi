# =============================================================================
# Zé da API Manager — Next.js 16 + Bun + Prisma 7 (standalone)
# =============================================================================

# ---------------------------------------------------------------------------
# Stage 1: Install dependencies
# ---------------------------------------------------------------------------
FROM oven/bun:1.3-alpine AS deps
WORKDIR /app

# Copy dependency manifests
COPY package.json bun.lock ./

# Copy Prisma schema + config (required by postinstall: bunx --bun prisma generate)
COPY prisma/schema.prisma ./prisma/schema.prisma
COPY prisma.config.ts ./

# Dummy DATABASE_URL: prisma.config.ts calls env("DATABASE_URL") at parse time,
# but prisma generate only reads the schema — it does NOT connect to the DB.
ENV DATABASE_URL="postgresql://build:build@localhost:5432/build"

# Install ALL deps (devDependencies needed for prisma CLI + next build)
# postinstall hook runs: bunx --bun prisma generate
RUN bun install --frozen-lockfile

# ---------------------------------------------------------------------------
# Stage 2: Build Next.js
# ---------------------------------------------------------------------------
FROM oven/bun:1.3-alpine AS builder
WORKDIR /app

# Copy deps + generated Prisma client from stage 1
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/generated ./generated

# Copy full source
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Build Next.js (standalone output)
# --bun: forces Next.js worker subprocesses to use Bun runtime (faster compilation)
RUN bun --bun run build

# ---------------------------------------------------------------------------
# Stage 3: Production runner (Bun)
# ---------------------------------------------------------------------------
FROM oven/bun:1.3-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Security: non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy public assets
COPY --from=builder /app/public ./public

# Copy standalone server + static assets
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# ---------------------------------------------------------------------------
# Seed environment: /seed dir with Prisma CLI + schema + seed script
# Runs on startup BEFORE Next.js — safe to fail (server still starts)
# ---------------------------------------------------------------------------
COPY --from=deps --chown=nextjs:nodejs /app/node_modules /seed/node_modules
COPY --from=deps --chown=nextjs:nodejs /app/generated /seed/generated
COPY --from=builder --chown=nextjs:nodejs /app/prisma /seed/prisma
COPY --from=builder --chown=nextjs:nodejs /app/prisma.config.ts /seed/prisma.config.ts
COPY --from=builder --chown=nextjs:nodejs /app/package.json /seed/package.json
COPY --from=builder --chown=nextjs:nodejs /app/tsconfig.json /seed/tsconfig.json

# Entrypoint: db push + seed + server
COPY --chown=nextjs:nodejs docker-entrypoint.sh ./
RUN chmod +x docker-entrypoint.sh

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD wget -qO- http://localhost:3000/api/health || exit 1

ENTRYPOINT ["./docker-entrypoint.sh"]
