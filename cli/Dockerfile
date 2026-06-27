FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY cli/go.mod cli/go.sum cli/
RUN cd cli && go mod download

COPY cli/ cli/

RUN cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/na .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/na /usr/local/bin/na

USER appuser

ENTRYPOINT ["/usr/local/bin/na"]
