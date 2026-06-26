# ─── Build Stage ──────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first for dependency caching
COPY shared/go.mod shared/go.sum shared/
RUN cd shared && go mod download

# Copy backend module
COPY backend/go.mod backend/go.sum backend/
RUN cd backend && go mod download

# Copy source
COPY shared/ shared/
COPY backend/ backend/

# Build static binary
RUN cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/server

# ─── Runtime Stage ────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app
COPY --from=builder /app/server .

# Defaults
ENV DATABASE_DRIVER=sqlite
ENV DATABASE_DSN=/data/now-and-again.db
ENV JWT_SECRET=change-me-in-production
ENV DATA_DIR=/data

RUN mkdir -p /data/uploads /data/logs
VOLUME ["/data"]

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/system/status || exit 1

ENTRYPOINT ["./server"]
