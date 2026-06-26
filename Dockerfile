# ─── Frontend Build ────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend
RUN corepack enable && corepack prepare pnpm@9 --activate

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ .
RUN pnpm build

# ─── Backend Build ────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY shared/go.mod shared/go.sum shared/
RUN cd shared && go mod download

COPY backend/go.mod backend/go.sum backend/
RUN cd backend && go mod download

COPY shared/ shared/
COPY backend/ backend/

# Embed frontend dist
COPY --from=frontend-builder /frontend/dist /app/backend/internal/webui/dist

RUN cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/server

# ─── Runtime ──────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app
COPY --from=builder /app/server .

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
