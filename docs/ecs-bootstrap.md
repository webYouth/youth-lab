# ECS 首次初始化（在服务器上执行）

你当前在 `~`（家目录），所以 `cd opt` 会失败。
要用绝对路径 `/opt`。

## 1. 创建部署目录

```bash
sudo mkdir -p /opt/youth-lab/{nginx,ssl,certbot/www,kl8}
cd /opt/youth-lab
pwd
# 应输出：/opt/youth-lab
```

## 2. 准备环境变量（QQ 邮箱发信）

```bash
cat > /opt/youth-lab/.env <<'EOF'
SMTP_HOST=smtp.qq.com
SMTP_PORT=465
SMTP_USER=webyouth@qq.com
SMTP_PASS=替换成你的QQ邮箱授权码
MAIL_FROM=webyouth@qq.com
MAIL_TO=webyouth@qq.com
EOF

chmod 600 /opt/youth-lab/.env
```

## 3. 准备今日号码文件

等 GitHub Actions 首次部署同步 `tickets.example.yaml` 后执行：

```bash
cp /opt/youth-lab/kl8/tickets.example.yaml /opt/youth-lab/kl8/tickets.yaml
vi /opt/youth-lab/kl8/tickets.yaml
```

如果还没部署过，可先放一个空占位：

```bash
touch /opt/youth-lab/kl8/tickets.yaml
```

## 4. 确认目录结构

```bash
ls -la /opt/youth-lab
```

期望至少有：

```text
/opt/youth-lab/
  .env
  docker-compose.prod.yml   # 首次部署后由 CI 同步
  nginx/
  ssl/
  certbot/www/
  kl8/
```

## 5. 之后每次部署

把代码推到 `main`，GitHub Actions 会自动：

1. 构建并推送镜像到 ACR
2. 把 compose/nginx 同步到 `/opt/youth-lab`
3. 在 ECS 执行 `docker-compose -f docker-compose.prod.yml up -d`
