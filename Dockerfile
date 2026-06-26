# ─── Frontend Build Stage ─────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend
RUN corepack enable && corepack prepare pnpm@9 --activate

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ .
RUN pnpm build

# ─── Backend Build Stage ─────────────────────────────────────────
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

COPY shared/go.mod shared/go.sum shared/
RUN cd shared && go mod download

COPY backend/go.mod backend/go.sum backend/
RUN cd backend && go mod download

COPY shared/ shared/
COPY backend/ backend/

# Copy frontend dist into backend for embedding
COPY --from=frontend-builder /frontend/dist /app/backend/internal/webui/dist

# Build static binary
RUN cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/server

# ─── CLI Build Stage ──────────────────────────────────────────────
FROM golang:1.22-alpine AS cli-builder

WORKDIR /app
COPY shared/go.mod shared/go.sum shared/
RUN cd shared && go mod download
COPY cli/go.mod cli/go.sum cli/
RUN cd cli && go mod download
COPY shared/ shared/
COPY cli/ cli/
RUN cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/na .

# ─── Server Runtime Stage ─────────────────────────────────────────
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=backend-builder /app/server .
COPY --from=cli-builder /app/na /usr/local/bin/na

ENV DATABASE_DRIVER=sqlite
ENV DATABASE_DSN=/data/now-and-again.db
ENV JWT_SECRET=change-me-in-production
ENV DATA_DIR=/data
ENV GIN_MODE=release

RUN mkdir -p /data/uploads /data/logs
VOLUME ["/data"]

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/system/status || exit 1

ENTRYPOINT ["./server"]

# ─── CLI Runtime Stage ────────────────────────────────────────────
FROM alpine:3.20 AS cli-runtime

RUN apk add --no-cache ca-certificates tzdata

COPY --from=cli-builder /app/na /usr/local/bin/na

ENTRYPOINT ["/usr/local/bin/na"]
