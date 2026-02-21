#!/bin/sh
# =============================================================================
# Zé da API Manager — Entrypoint
# DB push + Seed rodam ANTES do server. Se falharem, server inicia mesmo assim.
# =============================================================================

echo "========================================="
echo "  Zé da API Manager — Startup"
echo "========================================="

# --- 1. Database migration (Prisma db push) ---
echo ""
echo "[1/3] Running Prisma DB push..."
(
  cd /seed && \
  bunx --bun prisma db push --skip-generate 2>&1 && \
  echo "  ✓ Database schema synced"
) || echo "  ⚠ DB push skipped (may already be in sync or DB unavailable)"

# --- 2. Seed (plans, admin, webhook, settings) ---
echo ""
echo "[2/3] Running seed..."
(
  cd /seed && \
  bun run prisma/seed.ts 2>&1 && \
  echo "  ✓ Seed complete"
) || echo "  ⚠ Seed skipped (may already be seeded or DB unavailable)"

# --- 3. Start Next.js server (ALWAYS runs) ---
echo ""
echo "[3/3] Starting Next.js server..."
echo "========================================="
cd /app
exec bun --bun run server.js
