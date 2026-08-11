# Multi-stage Dockerfile for internal 快乐8 checker daemon.
FROM golang:1.26.2-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/kl8-checker .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -g '' appuser \
  && mkdir -p /data \
  && chown appuser:appuser /data
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/kl8-checker /usr/local/bin/kl8-checker
# No ports exposed on purpose: this service is internal-only.
ENTRYPOINT ["/usr/local/bin/kl8-checker"]
