FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/ backend/
COPY cli/go.mod cli/go.sum cli/
RUN cd cli && go mod download
COPY cli/ cli/
RUN cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/na .

FROM registry.alpinelinux.org/img/alpine:3.24
RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
ENV TZ=Asia/Shanghai

COPY --from=builder /app/na /usr/local/bin/na

USER appuser

ENTRYPOINT ["/usr/local/bin/na"]
