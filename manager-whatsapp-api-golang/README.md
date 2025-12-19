# WhatsApp Manager

WhatsApp instance management panel built with Next.js 16, Better Auth, and shadcn/ui.

## Features

- **Complete authentication** with email/password, OAuth (GitHub, Google), and 2FA
- **WhatsApp instance management** (create, activate, deactivate, delete)
- **QR Code connection** or Phone Pairing Code
- **Webhook configuration** (7 event types)
- **API health monitoring** (health, ready, metrics)
- **Transactional email system** with modern templates

## Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| [Next.js](https://nextjs.org) | 16.x | React Framework |
| [Bun](https://bun.sh) | 1.3.x | Runtime and Package Manager |
| [React](https://react.dev) | 19.x | UI Library |
| [TypeScript](https://www.typescriptlang.org) | 5.x | Type Safety |
| [Prisma](https://prisma.io) | 7.x | ORM |
| [Better Auth](https://better-auth.com) | 1.4.x | Authentication |
| [TailwindCSS](https://tailwindcss.com) | 4.x | Styling |
| [shadcn/ui](https://ui.shadcn.com) | latest | Components |
| [Nodemailer](https://nodemailer.com) | latest | Emails |

## Quick Start

### Prerequisites

- [Bun](https://bun.sh) 1.3.4 or higher
- PostgreSQL database
- SMTP server (optional, for emails)

### Installation

```bash
# Clone repository
git clone <repo-url>
cd manager-whatsapp-api-golang

# Install dependencies
bun install

# Configure environment variables
cp .env.example .env
# Edit .env with your settings

# Generate Prisma client
bun prisma generate

# Sync schema with database
bun prisma db push

# Start development server
bun dev
```

### Environment Variables

```env
# Database
DATABASE_URL="postgresql://user:pass@localhost:5432/whatsapp_manager"

# Better Auth
BETTER_AUTH_SECRET="your-secret-key-32-chars"
BETTER_AUTH_URL="http://localhost:3000"

# WhatsApp API Backend
WHATSAPP_API_URL="http://localhost:8080"
WHATSAPP_CLIENT_TOKEN="your-client-token-min-16-chars"
WHATSAPP_PARTNER_TOKEN="your-partner-token"

# SMTP (for email sending)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER=your@email.com
SMTP_PASSWORD=your-password-or-app-password

# Email Configuration
EMAIL_FROM_NAME=WhatsApp Manager
EMAIL_FROM_ADDRESS=noreply@yourdomain.com
SUPPORT_EMAIL=support@yourdomain.com

# OAuth (optional)
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# Public
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Available Scripts

```bash
bun dev        # Development with Turbopack
bun build      # Production build
bun start      # Start production
bun lint       # Check linting
```

## Project Structure

```
app/
├── (auth)/          # Authentication pages
│   ├── login/       # Login (email, OAuth)
│   └── forgot-password/
├── (dashboard)/     # Protected pages
│   ├── dashboard/   # Main dashboard
│   ├── instances/   # Instance management
│   ├── health/      # API health
│   └── settings/    # Settings
└── api/auth/        # Better Auth handlers

components/
├── ui/              # shadcn/ui base
├── layout/          # Sidebar, Header
├── instances/       # Instance components
├── health/          # Health components
└── shared/          # Reusable components

lib/
├── api/             # WhatsApp API client
├── email/           # Email system
├── auth.ts          # Better Auth config
└── prisma.ts        # Prisma client
```

## Email System

The project includes a complete transactional email system:

- **Login alert** - New access detected
- **Password recovery** - Reset link
- **Password changed** - Confirmation
- **2FA code** - Two-factor verification
- **2FA enabled/disabled** - Confirmations
- **User invite** - For new members

See [docs/EMAIL.md](docs/EMAIL.md) for complete documentation.

## Authentication

### Supported Methods

- Email and password
- OAuth (GitHub, Google)
- Two-Factor Authentication (2FA) via email

### Login Flow

1. User accesses `/login`
2. Chooses authentication method
3. If 2FA enabled, code is sent by email
4. After authentication, redirected to `/dashboard`

## Theme and Design

The project uses a custom theme based on Radix UI:

- **Base**: Radix
- **Style**: Maia
- **Base Color**: Neutral
- **Theme**: Lime
- **Radius**: Large

## Project Scope

**Included:**
- WhatsApp instance management
- QR Code / Phone Pairing connection
- Webhook configuration
- Health monitoring

**Not included:**
- Message sending
- Event visualization
- Contact/group management

## License

Proprietary - All rights reserved.
