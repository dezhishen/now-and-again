# Docker 部署指南

所有镜像托管在 [GHCR](https://github.com/dezhishen/now-and-again/pkgs/container/now-and-again)。

---

## 快速启动（docker compose）

```bash
# 创建数据目录（运行时需要）
mkdir -p ./na-data

# 拉取并启动
docker compose up -d

# 查看日志
docker compose logs -f
```

服务启动后访问 `http://localhost:8080`。

默认管理员：`admin` / `12345678`（由 `NA_ADMIN_DEFAULT_PASSWORD` 设置）。

---

## 使用 Host 目录持久化（推荐生产环境）

容器内部以非 root 用户 `appuser` 运行，UID/GID 为容 器内自动分配。使用宿主机目录时需提前创建并授权：

```bash
# 创建数据目录并开放权限（容器内用户 UID 不固定，使用 777 最简单可靠）
mkdir -p ./na-data
chmod 777 ./na-data

# 或使用命名卷（容器自动管理权限，更推荐）
docker volume create na-data
```

`docker-compose.yml` 中使用宿主目录：

```yaml
services:
  server:
    image: ghcr.io/dezhishen/now-and-again:latest
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - NA_ADMIN_DEFAULT_PASSWORD=12345678
    volumes:
      - ./na-data:/data    # 宿主目录
    restart: unless-stopped
```

---

## 手动运行

```bash
docker run -d \
  --name now-and-again \
  -p 8080:8080 \
  -v na-data:/data \
  -e NA_ADMIN_DEFAULT_PASSWORD=12345678 \
  -e GIN_MODE=release \
  ghcr.io/dezhishen/now-and-again:latest
```

---

## 指定版本

```bash
# 使用特定版本
ghcr.io/dezhishen/now-and-again:v1.0.0

# 使用大版本（自动获取该大版本最新）
ghcr.io/dezhishen/now-and-again:v1
```

所有可用版本：[GHCR Packages](https://github.com/dezhishen/now-and-again/pkgs/container/now-and-again)

---

## 数据持久化

```
/data/
├── now-and-again.db    # SQLite 数据库
├── .jwt_secret         # JWT 签名密钥
├── uploads/            # 上传文件
├── templates/          # 自定义任务模板
└── logs/               # 运行日志
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `NA_DATA_DIR` | `/data` | 数据目录 |
| `GIN_MODE` | `release` | 运行模式 |
| `NA_PORT` | `8080` | HTTP 端口 |
| `NA_ADMIN_DEFAULT_PASSWORD` | — | 首次启动管理员密码 |
| `NA_JWT_SECRET` | 自动生成 | JWT 签名密钥 |
| `DEFAULT_TIMEZONE` | `Asia/Shanghai` | 后端默认时区（IANA 格式） |

## 健康检查

```bash
curl http://localhost:8080/api/system/status
# → {"status":"ok"}
```

## 升级

```bash
docker compose pull        # 拉取最新镜像
docker compose up -d       # 重新创建容器
```

## CLI 镜像

```bash
# 拉取 CLI 镜像
docker pull ghcr.io/dezhishen/now-and-again-cli:latest

# 运行
docker run --rm ghcr.io/dezhishen/now-and-again-cli:latest --help
```
