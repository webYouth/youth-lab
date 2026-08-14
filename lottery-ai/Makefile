.PHONY: tidy build run migrate export docker-up

tidy:
	cd backend && go mod tidy

build:
	cd backend && go build -o bin/lottery-ai ./cmd/server

run:
	cd backend && go run ./cmd/server -sql ./manifest/sql/001_init.sql

migrate:
	cd backend && go run ./cmd/server -migrate -sql ./manifest/sql/001_init.sql

export:
	cd backend && go run ./cmd/export-finetune --lottery=DLT --output=./finetune/dlt.jsonl --min-hit-rate=0.2

docker-up:
	docker compose -f docker-compose.yml --env-file .env.example up --build -d
