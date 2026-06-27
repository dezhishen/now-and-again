FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY cli/go.mod cli/go.sum cli/
RUN cd cli && go mod download

COPY cli/ cli/

RUN cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/na .

FROM alpine:3.21

# Copy CA certificates and timezone data from builder (avoids apk network issues)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/na /usr/local/bin/na

USER appuser

ENTRYPOINT ["/usr/local/bin/na"]
