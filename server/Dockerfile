# Multi-stage Dockerfile for Go Gin + gRPC backend.
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./main.go

FROM alpine:3.20

RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser

COPY --from=builder /bin/server /usr/local/bin/server

EXPOSE 8080 50051
ENTRYPOINT ["/usr/local/bin/server"]
