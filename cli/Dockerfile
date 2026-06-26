FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY shared/go.mod shared/go.sum shared/
RUN cd shared && go mod download

COPY cli/go.mod cli/go.sum cli/
RUN cd cli && go mod download

COPY shared/ shared/
COPY cli/ cli/

RUN cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/na .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/na /usr/local/bin/na

ENTRYPOINT ["/usr/local/bin/na"]
