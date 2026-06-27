FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN go env -w GOPROXY='https://goproxy.cn,direct' && go env -w GO111MODULE='auto'
COPY backend/ backend/
COPY cli/go.mod cli/go.sum cli/
RUN cd cli && go mod download
COPY cli/ cli/
RUN cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/na .

FROM alpine:3.24

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/na /usr/local/bin/na

USER appuser

ENTRYPOINT ["/usr/local/bin/na"]
