# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

SaaS management dashboard for the ZedaAPI WhatsApp API platform. Users sign up (with waitlist gating), subscribe to plans (Stripe + Sicredi PIX/Boleto), provision WhatsApp instances via ZedaAPI's HTTP API, and manage them through a web UI. Includes an admin panel, affiliate system, NFS-e (Brazilian electronic invoice) issuance, and background job processing.

## Stack

- **Framework**: Next.js 16 App Router with Turbopack
- **Runtime**: Bun (build, dev, workers, scripts)
- **Language**: TypeScript (strict)
- **Database**: PostgreSQL via Prisma 7 (with `@prisma/adapter-pg` driver adapter)
- **Auth**: Better Auth with plugins: admin, organization, two-factor, waitlist
- **Payments**: Stripe (international) + Sicredi PIX/Boleto (Brazil)
- **Queue**: BullMQ with Redis (ioredis)
- **UI**: shadcn/ui (radix-maia style) + Tailwind CSS 4 + Framer Motion
- **State**: Zustand (client), React Query (server state)
- **Email**: React Email + Nodemailer

## Build & Development Commands

```bash
# Dev server (Turbopack)
bun dev

# Build
bun run build

# Lint
bun run lint

# Database
bun run db:generate        # Generate Prisma client (output: generated/prisma/)
bun run db:push            # Push schema to DB (no migration files)
bun run db:migrate         # Create migration
bun run db:migrate:deploy  # Apply migrations in production
bun run db:studio          # Open Prisma Studio
bun run db:seed            # Seed database
bun run db:setup           # Push schema + seed (initial setup)
bun run db:reset           # Reset database (destructive)

# Workers (separate process from Next.js)
bun run worker:dev         # BullMQ workers with --watch
bun run worker:all         # BullMQ workers (production)

# Seed only
bun run prisma/seed.ts
```

## Architecture

### Two-Process Model

The app runs as two separate processes:
1. **Next.js server** (`bun dev` / `next start`) — handles UI, API routes, Server Actions
2. **BullMQ workers** (`bun run workers/index.ts`) — processes background jobs via Redis queues

### Route Groups (App Router)

- `app/(auth)/` — Sign-in, sign-up, waitlist, forgot-password, two-factor, verify-email. Unauthenticated.
- `app/(dashboard)/` — Main user area. Protected via `requireAuth()` in layout. Instances, billing, subscriptions, profile, settings, affiliates, API keys.
- `app/(admin)/admin/` — Admin panel. Protected via `requireAdmin()`. Users, instances, subscriptions, plans, feature flags, invoices, NFS-e, waitlist, activity logs, affiliates.
- `app/api/` — Route handlers for auth (`[...all]`), webhooks (Stripe, Sicredi), and Brazil API proxies (CEP, CNPJ).

### Server-Side Pattern

```
Server Actions (server/actions/*.ts)
  └── Services (server/services/*.ts)
       └── DB (lib/db.ts) + External clients (lib/zedaapi-client.ts, lib/stripe.ts)
```

- **Server Actions** (`server/actions/`): Thin layer called from client components. Handles auth, validation (Zod), revalidation. Returns `ActionResult<T>`.
- **Services** (`server/services/`): Business logic. Instance provisioning/deprovisioning, subscription lifecycle, billing, affiliates, notifications.
- **Auth helpers** (`lib/auth-server.ts`): `requireAuth()` redirects to sign-in; `requireAdmin()` redirects to dashboard.

### ZedaAPI Integration

`lib/zedaapi-client.ts` is the HTTP client for the Go-based WhatsApp API backend (the sibling `api/` project). Features:
- Two auth modes: Partner (Bearer token) for admin ops, Instance (Client-Token) for per-instance ops
- Retry with exponential backoff (3 retries)
- Circuit breaker (5 failures → 30s open)
- Singleton with global cache for dev HMR

### Background Workers

Six BullMQ workers defined in `workers/processors/`:

| Worker | Queue | Purpose |
|--------|-------|---------|
| `stripe-webhook` | `stripe-webhooks` | Process Stripe webhook events |
| `nfse.processor` | `nfse-issuance` | Issue NFS-e (Brazilian invoices) |
| `email-sending` | `email-sending` | Send transactional emails |
| `sicredi-billing` | `sicredi-billing` | Sicredi PIX/Boleto payment processing |
| `instance-sync` | `instance-sync` | Sync instance status from ZedaAPI |
| `affiliate-payout` | `affiliate-payouts` | Process affiliate commission payouts |

Queue config in `lib/queue/config.ts`. Job producers in `lib/queue/producers.ts`.

### Prisma Setup

- **Schema**: `prisma/schema.prisma`
- **Generated client output**: `generated/prisma/` (not default `node_modules`)
- **Config**: `prisma.config.ts` at project root (Prisma 7 style)
- **Driver adapter**: Uses `@prisma/adapter-pg` with raw `pg` driver (not default Prisma connection)
- **Import**: Always `import { PrismaClient } from "@/generated/prisma/client"`

### Key Models

Better Auth core: `User`, `Session`, `Account`, `Verification`, `TwoFactor`, `Organization`, `Member`, `Invitation`

Business domain: `Plan`, `Subscription`, `Instance`, `Invoice`, `PaymentMethod`, `ApiKey`, `SicrediCharge`, `Waitlist`

Affiliate system: `Affiliate`, `Commission`, `Referral`, `Payout`

Support: `Notification`, `NotificationPreference`, `ActivityLog`, `FeatureFlag`, `SystemSetting`, `EmailTemplate`, `NfseConfig`, `NfseSequence`, `UsageRecord`

### Payment Dual-Gateway

- **Stripe**: International cards. Webhook at `app/api/webhooks/stripe/route.ts`.
- **Sicredi**: Brazilian PIX and Boleto Hibrido. Webhook at `app/api/webhooks/sicredi/route.ts`. Requires mTLS certificates in `certs/`.

### Docker

- `Dockerfile` — Next.js production build. Uses Bun for build, Node 22 for runtime (standalone output).
- `Dockerfile.workers` — BullMQ workers. Runs entirely on Bun.

## Conventions

- **Path alias**: `@/` maps to project root
- **Client components**: Named `*-client.tsx`, placed next to their page
- **UI primitives**: `components/ui/` (shadcn/ui, do not edit manually — use `bunx shadcn add`)
- **Domain components**: `components/{domain}/` (auth, billing, instances, dashboard, affiliates, etc.)
- **Shared components**: `components/shared/` (data-table, confirm-dialog, empty-state, loading-skeleton, theme-toggle)
- **Layout components**: `components/layout/` (sidebar, topbar, shells)
- **Hooks**: `hooks/` (use-auth, use-instances, use-debounce, use-mobile)
- **Stores**: `stores/` (Zustand stores)
- **Types**: `types/index.ts` (shared), `types/zedaapi.ts` (ZedaAPI response/request types)
- **Env validation**: `lib/env.ts` with Zod schema. Use `getEnv()` for runtime validation.
- **Error types**: `lib/errors.ts` defines `AppError`, `NotFoundError`, `ZedaAPIError`
- **Commits**: Conventional Commits format (`feat:`, `fix:`, `chore:`, etc.)

## Environment Variables

Copy `.env.example` to `.env`. Required for dev:
- `DATABASE_URL` — PostgreSQL connection string
- `BETTER_AUTH_SECRET` — Min 32 chars
- `BETTER_AUTH_URL` — Base URL for auth
- `ZEDAAPI_BASE_URL`, `ZEDAAPI_PARTNER_TOKEN`, `ZEDAAPI_CLIENT_TOKEN` — ZedaAPI backend
- `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY`
- `REDIS_URL` — Required for BullMQ workers
- `ENCRYPTION_KEY` — 64-char hex string for field encryption

## External Packages in Next.js Config

`next.config.ts` marks these as `serverExternalPackages` (not bundled by Turbopack): `node-forge`, `xml-crypto`, `@xmldom/xmldom`, `nodemailer`, `bullmq`, `ioredis`.
