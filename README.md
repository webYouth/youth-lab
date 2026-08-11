# Youth Blog Monorepo

This repository is a full project skeleton for a personal blog system:

- Go backend (`Gin + gRPC`)
- Next.js frontend (`App Router + BFF API Route`)
- Nginx reverse proxy
- Internal Happy8 (`快乐8`) lottery checker (email only, not public)
- `docker-compose` one-command startup

The blog is served under `/blog`, and backend HTTP endpoints are served under `/api`.

## Project Structure

```text
.
├── docker-compose.yml
├── docker-compose.dev.yml
├── docker-compose.prod.yml
├── .env.prod.example
├── .github/workflows/deploy.yml
├── nginx/
│   └── default.conf
├── kl8/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── tickets.example.yaml
│   └── ...
├── server/
│   ├── Dockerfile
│   ├── .air.toml
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── proto/
│   │   └── blog.proto
│   ├── internal/
│   │   ├── handler/
│   │   │   └── health.go
│   │   ├── service/
│   │   │   └── blog_service.go
│   │   └── client/
│   │       └── .gitkeep
│   └── gen/
│       └── blogpb/
│           ├── blog.pb.go
│           └── blog_grpc.pb.go
└── web/
    ├── Dockerfile
    ├── package.json
    ├── pnpm-lock.yaml
    ├── next.config.js
    ├── tsconfig.json
    ├── next-env.d.ts
    ├── app/
    │   ├── layout.tsx
    │   ├── page.tsx
    │   ├── globals.css
    │   └── api/
    │       └── posts/
    │           └── route.ts
    ├── lib/
    │   └── grpc-client.ts
    ├── proto/
    │   └── blog.proto
    └── public/
        └── .gitkeep
```

## Local Development

### 1) Start backend

```bash
cd server
go mod tidy
go run ./main.go
```

Backend endpoints:

- HTTP health: `http://localhost:8080/health`
- gRPC: `localhost:50051`

### 2) Start frontend

```bash
cd web
npx pnpm@9.15.4 install
GRPC_SERVER_ADDR=localhost:50051 npx pnpm@9.15.4 dev
```

Frontend URL:

- `http://localhost:3000/blog`

### 3) Run Happy8 checker once (dry-run)

```bash
cd kl8
cp tickets.example.yaml tickets.yaml
go run . -once -dry-run -tickets ./tickets.yaml
```

## Docker Compose One-command Deployment

From repository root:

```bash
docker-compose up --build
```

Then access:

- Blog homepage: `http://localhost/blog`
- Backend health via nginx: `http://localhost/api/health`

## Full-stack Hot Reload with Docker

Use the development override file to run both backend and frontend in hot reload mode:

```bash
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

During development, edit files under `web/` and Next.js will hot reload automatically.
Edit files under `server/` and `air` will rebuild/restart the Go service automatically.
The backend reuses the `air` binary from a Docker volume (`/go/bin`) and only installs it on first run.
You can also access the frontend directly at `http://localhost:3000/blog`.

## Production Deploy (GitHub -> ACR -> ECS)

Push to `main` triggers `.github/workflows/deploy.yml`:

1. Build/push `youth-lab-server` / `youth-lab-web` / `youth-lab-kl8` to ACR
2. Sync `docker-compose.prod.yml` + nginx config to `/opt/youth-lab`
3. SSH into ECS and `docker-compose -f docker-compose.prod.yml up -d`

### ECS bootstrap for Happy8 mail

```bash
mkdir -p /opt/youth-lab/kl8
cp /opt/youth-lab/kl8/tickets.example.yaml /opt/youth-lab/kl8/tickets.yaml
# edit /opt/youth-lab/kl8/tickets.yaml every day with your numbers

cp .env.prod.example /opt/youth-lab/.env
# fill SMTP_USER / SMTP_PASS (QQ mailbox authorization code)
```

Required ACR repositories:

- `webyouth/youth-lab-server`
- `webyouth/youth-lab-web`
- `webyouth/youth-lab-kl8`

## Happy8 Checker (internal)

- Data source: China Welfare Lottery public API (`findDrawNotice?name=kl8`)
- Schedule: Asia/Shanghai daily from 22:00
- If today's draw is not ready, retry every 5 minutes until success
- Sends result email to `webyouth@qq.com`
- No public port / no nginx route

Prize notes:

- Prefer current issue `prizegrades` amounts from official API
- Fallback fixed table includes 选九/选十 common awards
- 选十中十 is floating (cap 5,000,000)
- 复式 is expanded into single bets before checking

## Proto Regeneration

The files in `server/gen/blogpb/` are compile-safe placeholders for bootstrap only.
Regenerate with `protoc` when your environment is ready.

### Generate Go code

```bash
cd server
protoc \
  --proto_path=./proto \
  --go_out=./gen/blogpb \
  --go_opt=paths=source_relative \
  --go-grpc_out=./gen/blogpb \
  --go-grpc_opt=paths=source_relative \
  blog.proto
```

### Generate Node code (optional)

If you prefer pre-generated JS stubs, use:

```bash
cd web
protoc \
  --proto_path=./proto \
  --js_out=import_style=commonjs,binary:./gen/node \
  --grpc_out=grpc_js:./gen/node \
  blog.proto
```

Current implementation uses runtime loading (`@grpc/proto-loader`), so this step is optional.

## Notes about Go 1.26.2

- This project is configured for Go `1.26.2` in `server/go.mod` and Dockerfile base image.
- If `golang:1.26.2-alpine` is unavailable in your environment, replace it with a valid nearby tag (for example `golang:1.26-alpine`) in `server/Dockerfile`.
