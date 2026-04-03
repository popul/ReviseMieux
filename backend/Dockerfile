# ============================================================================
# Stage 1: Build
# ============================================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /server ./cmd/server

# ============================================================================
# Stage 2: Runtime
# ============================================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /server /server
COPY migrations/ /migrations/

EXPOSE 8080

# ── Environment variables ────────────────────────────────────────────────────
# DATABASE_URL        — PostgreSQL connection string (required)
# PORT                — HTTP listen port (default: 8080)
# GIN_MODE            — Gin framework mode: debug | release | test
# MIGRATIONS_DIR      — Path to SQL migration files inside the container
# REDIS_URL           — Redis connection string (optional)
# S3_ENDPOINT         — S3-compatible endpoint for object storage (optional)
# S3_BUCKET           — S3 bucket name (optional)
# S3_ACCESS_KEY       — S3 access key (optional)
# S3_SECRET_KEY       — S3 secret key (optional)
# ANTHROPIC_API_KEY   — Anthropic API key for LLM features (optional)
# OCR_PROVIDER        — OCR provider: gemini | google-vision (optional)
ENV GIN_MODE=release
ENV MIGRATIONS_DIR=/migrations

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/server"]
