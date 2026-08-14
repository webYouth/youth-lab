# 彩票 AI 预测系统（个人学习研究用途）

> **免责声明：预测结果仅供个人学习研究，不构成投注建议；系统不能保证中奖。**

本目录是 `youth-lab` 下的独立子系统，包含：

- `backend/`：Golang + Gin + PostgreSQL 16（目录按 GoFrame 风格组织，便于扩展）
- `frontend/`：Expo + React Native + TypeScript
- `docker-compose.yml`：一键启动 PostgreSQL + 后端

## 功能概览

1. 抓取大乐透 / 排列三 / 快乐8 历史开奖（可配置 URL，失败重试，保留原始日志）
2. 计算热号/冷号/遗漏/和值等统计特征
3. 并发调用 5 个大模型（DeepSeek / Qwen / Kimi / GLM / MiniMax），失败优雅降级
4. 按历史命中率加权投票生成最终预测
5. 定时同步 / 预测 / 评估
6. 导出 JSONL 微调样本

## 环境变量

复制 `.env.example` 为 `.env`，填入：

| 变量 | 说明 |
|---|---|
| `DEEP_SEEK_API_KEY` | DeepSeek |
| `QWEN_API_KEY` | 通义千问 |
| `KIMI_API_KEY` | Moonshot Kimi |
| `GLM_API_KEY` | 智谱 GLM |
| `MINIMAX_API_KEY` | MiniMax |
| `ADMIN_TOKEN` | 管理接口 `X-Admin-Token` |
| `API_TOKEN` | 可选 Bearer Token |
| `DATABASE_URL` | PostgreSQL 连接串 |

GitHub Actions Secrets 已预留同名 Key，部署时会注入容器。

## 本地启动后端

```bash
cd lottery-ai
cp .env.example .env
# 编辑 .env 填入 Key

docker compose up -d postgres
cd backend
go mod tidy
go run ./cmd/server -sql ./manifest/sql/001_init.sql
```

健康检查：

```bash
curl http://127.0.0.1:8090/api/v1/health
```

手动同步 / 预测 / 评估：

```bash
curl -X POST -H "X-Admin-Token: change-me-admin" http://127.0.0.1:8090/api/v1/admin/sync
curl -X POST -H "X-Admin-Token: change-me-admin" http://127.0.0.1:8090/api/v1/admin/generate
curl -X POST -H "X-Admin-Token: change-me-admin" http://127.0.0.1:8090/api/v1/admin/evaluate
```

## API 一览

统一响应：`{code, message, data, disclaimer}`

- `GET /api/v1/health`
- `GET /api/v1/lottery-types`
- `GET /api/v1/predictions/today?lottery_code=DLT`
- `GET /api/v1/predictions?lottery_code=DLT&page=1`
- `GET /api/v1/draw-results?lottery_code=DLT&page=1`
- `GET /api/v1/accuracy?lottery_code=DLT&days=30`
- `POST /api/v1/admin/sync`（需 `X-Admin-Token`）
- `POST /api/v1/admin/generate`
- `POST /api/v1/admin/evaluate`

## 定时任务（Asia/Shanghai）

| 任务 | Cron | 说明 |
|---|---|---|
| sync_history | 06:00 / 21:45 | 同步开奖 |
| generate_predictions | 06:10 | 生成当日预测（大乐透非开奖日跳过） |
| evaluate_predictions | 22:05 | 评估命中率 |

## 微调导出

```bash
cd backend
go run ./cmd/export-finetune --lottery=DLT --output=./finetune/dlt.jsonl --min-hit-rate=0.2
```

## 前端（React Native / Expo）

```bash
cd frontend
pnpm install   # 或 npm install
pnpm start
```

设置页可配置服务器地址与 Token。所有页面底部/顶部均有免责声明。

## Docker 一键

```bash
cd lottery-ai
cp .env.example .env
docker compose --env-file .env up --build -d
```

## 项目结构

```text
lottery-ai/
├── backend/
│   ├── cmd/server/main.go
│   ├── cmd/export-finetune/main.go
│   ├── internal/...
│   ├── manifest/sql/001_init.sql
│   └── Dockerfile
├── frontend/
├── docker-compose.yml
├── .env.example
├── Makefile
└── README.md
```

## 重要说明

1. 个人使用，不对外发布
2. API Key 只读环境变量，不入库明文
3. 模型调用失败不影响主流程
4. 预测不能保证中奖
