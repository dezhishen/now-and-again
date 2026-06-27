# ─── Frontend Build ────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend
RUN corepack enable && corepack prepare pnpm@9 --activate

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ .
RUN pnpm build

# ─── Backend Build ────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app
RUN go env -w GOPROXY='https://goproxy.cn,direct' && go env -w GO111MODULE='auto'
COPY backend/go.mod backend/go.sum backend/
RUN cd backend && go mod download

COPY backend/ backend/

# Embed frontend dist
COPY --from=frontend-builder /frontend/dist /app/backend/internal/webui/dist

RUN cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/server

# ─── Runtime ──────────────────────────────────────────────────────
FROM alpine:3.24
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata
# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/server .

ENV DATA_DIR=/data
ENV GIN_MODE=release

RUN mkdir -p /data/uploads /data/logs \
    && chown -R appuser:appgroup /app /data
VOLUME ["/data"]

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/api/system/status || exit 1

ENTRYPOINT ["./server"]
