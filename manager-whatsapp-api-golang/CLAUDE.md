# Manager WhatsApp API - Claude Code Context

## Project Overview

This is the frontend Manager for the WhatsApp API Golang. A Next.js 16 application with App Router for **exclusive WhatsApp instance management**.

**Project Scope:**
- Create, activate, deactivate, and delete instances
- Generate QR Code and Phone Pairing Code for connection
- Configure webhooks and instance parameters
- Monitor API health (health, ready, metrics)

**OUT of Scope:**
- Message sending
- Webhook event visualization
- Contact or group management

## Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Next.js | 16.x | React Framework with App Router |
| Bun | 1.3.x | Runtime and Package Manager |
| React | 19.x | UI Library |
| TypeScript | 5.x | Type Safety |
| Prisma | 7.x | ORM with Driver Adapters |
| Better Auth | 1.4.x | Authentication with 2FA |
| TailwindCSS | 4.x | Styling |
| shadcn/ui | latest | Component Library |
| Nodemailer | latest | Email sending |

## Theme and Design System

The project uses a custom theme based on Radix UI:

```
base=radix
style=maia
baseColor=neutral
theme=lime
iconLibrary=lucide
font=inter
menuAccent=subtle
menuColor=inverted
radius=large
```

**Main Colors:**
- Primary: `#84cc16` (lime-500)
- Background: `#ffffff` / `#171717` (dark)
- Foreground: `#171717` / `#fafafa` (dark)

## Quick Commands

```bash
# Development
bun dev              # Start with Turbopack
bun build            # Production build
bun lint             # Check linting

# Prisma
bun prisma generate  # Generate client
bun prisma db push   # Sync schema

# Better Auth
bun x @better-auth/cli generate  # Generate/update schema
```

## Folder Structure

```
app/                    # App Router (Next.js 16)
├── (auth)/             # Authentication routes
│   ├── login/          # Login page
│   ├── forgot-password/# Password recovery
│   └── layout.tsx      # Layout with branding
├── (dashboard)/        # Protected routes
│   ├── page.tsx        # Redirect to /dashboard
│   ├── dashboard/      # Main dashboard
│   ├── instances/      # Instance management
│   │   ├── page.tsx    # Instance list
│   │   ├── new/        # Create new instance
│   │   └── [id]/       # Instance details
│   │       ├── page.tsx      # Overview
│   │       ├── webhooks/     # Configure webhooks
│   │       └── settings/     # Settings
│   ├── health/         # API health
│   └── settings/       # Manager settings
├── api/                # API Routes
│   └── auth/[...all]/  # Better Auth handlers
└── layout.tsx

components/             # React Components
├── ui/                 # shadcn/ui base
├── layout/             # Sidebar, Header, Mobile Nav
├── instances/          # Instance components
├── health/             # Health components
└── shared/             # Reusable components

lib/                    # Utilities
├── api/                # API client for WhatsApp API
│   ├── client.ts       # Base HTTP client
│   ├── instances.ts    # Instance endpoints
│   ├── health.ts       # Health endpoints
│   └── webhooks.ts     # Webhook endpoints
├── email/              # Email system
│   ├── config.ts       # SMTP configuration
│   ├── sender.ts       # Email service
│   └── templates/      # HTML templates
├── auth.ts             # Better Auth config (server)
├── auth-client.ts      # Better Auth client
├── prisma.ts           # Prisma client
└── utils.ts

actions/                # Server Actions
├── instances.ts        # Instance CRUD
└── webhooks.ts         # Webhook configuration

schemas/                # Zod Schemas
├── instance.ts         # Instance validation
├── webhook.ts          # Webhook validation
└── auth.ts             # Authentication validation

types/                  # TypeScript types
├── instance.ts
├── health.ts
├── webhook.ts
└── api.ts

docs/                   # Documentation
├── PRD.md              # Product Requirements
├── ARCHITECTURE.md     # Architecture
├── DEVELOPMENT.md      # Development guide
├── COMPONENTS.md       # Component catalog
├── API.md              # API client docs
└── EMAIL.md            # Email system
```

## Authentication System

### Better Auth with 2FA

The project uses Better Auth with support for:
- Email/password login
- OAuth (GitHub, Google)
- Two-Factor Authentication (2FA) via email

### Configuration Files

```typescript
// lib/auth.ts - Server-side configuration
import { betterAuth } from 'better-auth';
import { twoFactor } from 'better-auth/plugins';

// lib/auth-client.ts - Client-side hooks
import { createAuthClient } from 'better-auth/react';
import { twoFactorClient } from 'better-auth/client/plugins';
```

### Authentication Proxy

Next.js 16 uses `proxy.ts` instead of `middleware.ts`:

```typescript
// proxy.ts
export default async function proxy(request: NextRequest) {
  // Public routes: /login, /forgot-password, /api/auth
  // Protected routes: everything in (dashboard)
}
```

## Email System

### Architecture

```
lib/email/
├── config.ts          # SMTP config, brand colors
├── sender.ts          # emailService with all methods
├── index.ts           # Centralized exports
└── templates/
    ├── base.ts        # Base template + helpers
    ├── login-alert.ts # New access alert
    ├── password-reset.ts # Reset + Changed
    ├── two-factor.ts  # 2FA code, enabled, disabled
    ├── user-invite.ts # Invite, accepted, expired
    └── index.ts
```

### Available Templates

| Template | Function | Usage |
|----------|----------|-------|
| `loginAlertTemplate` | `sendLoginAlert` | New access detected |
| `passwordResetTemplate` | `sendPasswordReset` | Recovery link |
| `passwordChangedTemplate` | `sendPasswordChanged` | Change confirmation |
| `twoFactorCodeTemplate` | `sendTwoFactorCode` | 2FA code |
| `twoFactorEnabledTemplate` | `sendTwoFactorEnabled` | 2FA enabled |
| `twoFactorDisabledTemplate` | `sendTwoFactorDisabled` | 2FA disabled |
| `userInviteTemplate` | `sendUserInvite` | New user invite |
| `inviteAcceptedTemplate` | `sendInviteAccepted` | Notify inviter |
| `inviteExpiredTemplate` | `sendInviteExpired` | Expired invite |

### Usage

```typescript
import { sendLoginAlert, sendTwoFactorCode } from '@/lib/email';

// Send login alert
await sendLoginAlert('user@email.com', {
  userName: 'John',
  userEmail: 'user@email.com',
  device: 'Chrome on Windows',
  ipAddress: '192.168.1.1',
  loginTime: new Date(),
});

// Send 2FA code
await sendTwoFactorCode('user@email.com', {
  userName: 'John',
  userEmail: 'user@email.com',
  verificationCode: '123456',
  expiresIn: 10,
});
```

## Development Patterns

### Server Components (Default)
- All WhatsApp API communication via server-side
- Never expose tokens to client
- Use `apiClient` from `@/lib/api/client`

### Client Components (When Necessary)
- Add `'use client'` at the top
- Only for interactivity (forms, modals, local state)
- Keep minimal state

### Server Actions
- Use for data mutations
- Always validate input with Zod
- Return `{ success: true, data }` or `{ success: false, error }`
- Call `revalidatePath` after mutations

### Typing
- Prefer `interface` for objects
- Use `type` for unions and functions
- Never use `any`, prefer `unknown`

## WhatsApp API Communication

The WhatsApp API Backend runs separately. Always communicate via `apiClient`:

```typescript
// Instance endpoints require instanceId and instanceToken in path
const status = await apiClient('/status', {
  instanceId: 'uuid',
  instanceToken: 'token',
});

// Partner endpoints use Partner-Token header
const instances = await apiClient('/instances', {
  usePartnerToken: true,
});
```

### Authentication Headers
- `Client-Token`: Global token for instance operations
- `Partner-Token`: Token for creating/deleting instances

## API Endpoints Used

### Health & Monitoring
| Endpoint | Method | Auth |
|----------|--------|------|
| `/health` | GET | None |
| `/ready` | GET | None |
| `/metrics` | GET | None |

### Partner/Integrator (CRUD)
| Endpoint | Method | Auth |
|----------|--------|------|
| `/instances` | GET | Partner-Token |
| `/instances/integrator/on-demand` | POST | Partner-Token |
| `/instances/{id}/token/{token}/integrator/on-demand/subscription` | POST | Partner-Token |
| `/instances/{id}/token/{token}/integrator/on-demand/cancel` | POST | Partner-Token |
| `/instances/{id}` | DELETE | Partner-Token |

### Instance Management
| Endpoint | Method | Auth |
|----------|--------|------|
| `/instances/{id}/token/{token}/status` | GET | Client-Token |
| `/instances/{id}/token/{token}/qr-code` | GET | Client-Token |
| `/instances/{id}/token/{token}/qr-code/image` | GET | Client-Token |
| `/instances/{id}/token/{token}/phone-code/{phone}` | GET | Client-Token |
| `/instances/{id}/token/{token}/device` | GET | Client-Token |
| `/instances/{id}/token/{token}/restart` | POST | Client-Token |
| `/instances/{id}/token/{token}/disconnect` | POST | Client-Token |

### Webhook Configuration
| Endpoint | Method | Auth |
|----------|--------|------|
| `/instances/{id}/token/{token}/update-webhook-delivery` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-webhook-received` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-webhook-received-delivery` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-webhook-message-status` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-webhook-connected` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-webhook-disconnected` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-webhook-chat-presence` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-every-webhooks` | PUT | Client-Token |
| `/instances/{id}/token/{token}/update-notify-sent-by-me` | PUT | Client-Token |

## Environment Variables

```env
# Database (required)
DATABASE_URL="postgresql://..."

# Better Auth (required)
BETTER_AUTH_SECRET="..."
BETTER_AUTH_URL="http://localhost:3000"

# WhatsApp API (required)
WHATSAPP_API_URL="http://localhost:8080"
WHATSAPP_CLIENT_TOKEN="..."     # Min 16 chars
WHATSAPP_PARTNER_TOKEN="..."    # Bearer token

# SMTP Configuration (required for emails)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER=your@email.com
SMTP_PASSWORD=your-password

# Email From Configuration
EMAIL_FROM_NAME=WhatsApp Manager
EMAIL_FROM_ADDRESS=noreply@whatsapp-manager.com
SUPPORT_EMAIL=support@whatsapp-manager.com

# OAuth (optional)
GITHUB_CLIENT_ID=""
GITHUB_CLIENT_SECRET=""
GOOGLE_CLIENT_ID=""
GOOGLE_CLIENT_SECRET=""

# Public URL
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Complete Documentation

- **@docs/PRD.md**: Product Requirements Document (instance scope)
- **@docs/ARCHITECTURE.md**: System architecture
- **@docs/DEVELOPMENT.md**: Development guide
- **@docs/COMPONENTS.md**: Component catalog
- **@docs/API.md**: API client documentation
- **@docs/EMAIL.md**: Transactional email system

## Development Checklist

- [ ] Use Server Components whenever possible
- [ ] Validate input in Server Actions with Zod
- [ ] Never expose tokens to client
- [ ] Call revalidatePath after mutations
- [ ] Type all functions and components
- [ ] Use existing shadcn/ui components
- [ ] Follow kebab-case naming for files
- [ ] Test on mobile (responsiveness)
- [ ] Use email system for notifications

## Important Notes

1. **Prisma Client Location**: Client is generated in `lib/generated/prisma/`, import from `@/lib/generated/prisma/client`

2. **Better Auth**: Use `@/lib/prisma` as singleton, don't create new PrismaClient

3. **WhatsApp API**: All endpoints require `Client-Token` header except `/health`, `/ready`, and `/metrics`

4. **QR Code**: Expires after 60 seconds, implement polling to check connection

5. **Phone Pairing**: Alternative to QR Code, generate code via `/phone-code/{phone}`

6. **Webhooks**: 7 configurable types per instance (delivery, received, status, connected, disconnected, presence, received-delivery)

7. **Limited Scope**: This manager does NOT send messages or view events. Only manages instances.

8. **Next.js 16**: Uses `proxy.ts` instead of `middleware.ts` for authentication

9. **Email System**: Use `lib/email` for all transactional emails. See `docs/EMAIL.md` for details

10. **2FA**: Implemented via Better Auth plugin `twoFactor`. Codes sent by email
