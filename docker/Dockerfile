# syntax=docker/dockerfile:1

# ==========================================
# Stage 1: Build Dependencies
# ==========================================
FROM golang:1.25-alpine AS deps
WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download && go mod verify

# ==========================================
# Stage 2: Build Application
# ==========================================
FROM golang:1.25-alpine AS builder
WORKDIR /build

# Copy dependencies from deps stage
COPY --from=deps /go/pkg /go/pkg

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME

# Build with optimizations
# CGO_ENABLED=0 for static binary
# -ldflags for version info and size reduction
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/api \
    ./api/cmd/server

# Build migrations tool (if exists)
RUN if [ -f "./api/cmd/migrate/main.go" ]; then \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/migrate \
    ./api/cmd/migrate; \
    fi

# ==========================================
# Stage 3: Runtime (Production)
# ==========================================
FROM alpine:3.21 AS production

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && addgroup -g 1000 app \
    && adduser -D -u 1000 -G app app

# Set timezone
ENV TZ=UTC

# Copy certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy compiled binaries
COPY --from=builder /app/api /usr/local/bin/api
COPY --from=builder /app/migrate /usr/local/bin/migrate 2>/dev/null || true

# Copy static files and configs
COPY --from=builder /build/api/.env.example /app/.env.example
COPY --from=builder /build/api/migrations /app/migrations

# Set working directory
WORKDIR /app

# Change ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

# Run application
ENTRYPOINT ["/usr/local/bin/api"]
CMD []

# ==========================================
# Stage 4: Development
# ==========================================
FROM golang:1.25-alpine AS development

# Install development tools
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    make \
    gcc \
    musl-dev

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Install other dev tools
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

EXPOSE 8080

# Run with air (hot reload)
CMD ["air", "-c", "api/.air.toml"]

# ==========================================
# Stage 5: Testing
# ==========================================
FROM golang:1.25-alpine AS testing
WORKDIR /build

# Install build dependencies (gcc needed for -race flag)
RUN apk add --no-cache gcc musl-dev git ca-certificates

# Copy dependencies
COPY --from=deps /go/pkg /go/pkg

# Copy source code
COPY . .

# Run tests with race detector (requires CGO_ENABLED=1)
RUN CGO_ENABLED=1 go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Generate coverage report
RUN go tool cover -html=coverage.out -o coverage.html

# ==========================================
# Metadata
# ==========================================
LABEL maintainer="WhatsApp API"
LABEL description="WhatsApp API"
LABEL version="${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/Funnelchat20/whatsapp-api-golang"
