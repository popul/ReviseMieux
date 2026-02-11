# Backend Dockerfile
# Multi-stage build for Go application

# ============================================================================
# Build stage
# ============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies (CGO required for Tesseract bindings)
RUN apk add --no-cache git ca-certificates gcc g++ musl-dev tesseract-ocr-dev

# Copy go mod files first for caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Resolve any new dependencies (e.g. gosseract)
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# ============================================================================
# Production stage
# ============================================================================
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies (Tesseract OCR + French language data)
RUN apk add --no-cache ca-certificates tzdata tesseract-ocr tesseract-ocr-data-fra

# Create non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -D app

# Create storage directory with correct permissions
RUN mkdir -p /app/storage && chown -R app:app /app/storage

USER app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations folder
COPY --from=builder /app/migrations ./migrations

# Copy config file
COPY --from=builder /app/config.yaml ./config.yaml

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]
