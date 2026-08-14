# 多阶段构建：彩票 AI 预测后端
FROM golang:1.26.2-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/lottery-ai ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/lottery-ai /usr/local/bin/lottery-ai
COPY manifest/sql /app/manifest/sql
ENV TZ=Asia/Shanghai
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/lottery-ai", "-sql", "/app/manifest/sql/001_init.sql"]
