#!/bin/sh
# =============================================================================
# Zé da API Manager — Entrypoint
# DB push + Seed rodam ANTES do server. Se falharem, o server NAO inicia.
# =============================================================================

echo "========================================="
echo "  Zé da API Manager — Startup"
echo "========================================="

# --- 0. Preconditions ---
if [ -z "$DATABASE_URL" ]; then
  echo "  ✗ DATABASE_URL is not set. Aborting."
  exit 1
fi

# --- 1. Database migration (Prisma db push) ---
echo ""
echo "[1/3] Running Prisma DB push..."
cd /seed || exit 1
DB_PUSH_OK=0
for i in 1 2 3 4 5; do
  if bunx --bun prisma db push 2>&1; then
    DB_PUSH_OK=1
    break
  fi
  echo "  ⚠ DB push failed (attempt $i/5). Retrying in 3s..."
  sleep 3
done
if [ "$DB_PUSH_OK" -ne 1 ]; then
  echo "  ✗ DB push failed after retries. Aborting."
  exit 1
fi
echo "  ✓ Database schema synced"

# --- 2. Seed (plans, admin, webhook, settings) ---
echo ""
echo "[2/3] Running seed..."
SEED_OK=0
for i in 1 2 3; do
  if bun run prisma/seed.ts 2>&1; then
    SEED_OK=1
    break
  fi
  echo "  ⚠ Seed failed (attempt $i/3). Retrying in 3s..."
  sleep 3
done
if [ "$SEED_OK" -ne 1 ]; then
  echo "  ✗ Seed failed after retries. Aborting."
  exit 1
fi
echo "  ✓ Seed complete"

# --- 3. Start Next.js server (ALWAYS runs) ---
echo ""
echo "[3/3] Starting Next.js server..."
echo "========================================="
cd /app
exec bun --bun run server.js
